package listys

import (
	"context"
	"encoding/json"
	"fmt"
	"immi/internal/idb"
	"immi/pkg/dao"
	"immi/pkg/immi"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

type ListysConfig struct {
	Logger *zerolog.Logger
	DB     idb.IDB
}

type ListyServer struct {
	logger *zerolog.Logger
	db     idb.IDB
}

func NewServer(config ListysConfig) (ListyServer, error) {
	return ListyServer{
		logger: config.Logger,
		db:     config.DB,
	}, nil
}

func (s *ListyServer) Handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/create-listy", s.createListyHandler)
	r.HandleFunc("/add-to-listy", s.addToListyHandler)
	r.HandleFunc("/rm-from-listy", s.rmFromListyHandler)
	r.HandleFunc("/get-timeline", s.getTimelineHandler)
	r.HandleFunc("/subscribe", s.subscribeHandler)
	return r
}

func (s *ListyServer) createListyHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userIDRaw := r.Header.Get(immi.UserHeader)
	userID, err := strconv.ParseInt(userIDRaw, 0, 64)
	if err != nil {
		s.logger.Error().Msgf("invalid userID %q", userIDRaw)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var newList immi.NewListy
	err = json.NewDecoder(r.Body).Decode(&newList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Fix context usage
	dbErr := s.db.CreateListy(context.Background(), dao.Listy{
		UserID:      userID,
		DisplayName: newList.DisplayName,
		RouteName:   newList.RouteName,
		CTime:       time.Now().UTC(),
	})
	if dbErr != nil {
		http.Error(w, dbErr.Err, dbErr.HTTPCode)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *ListyServer) addToListyHandler(w http.ResponseWriter,
	r *http.Request) {
	userIDRaw := r.Header.Get(immi.UserHeader)
	userID, err := strconv.ParseInt(userIDRaw, 0, 64)
	if err != nil {
		s.logger.Error().Msgf("invalid userID %q", userIDRaw)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var graf immi.Graf
	err = json.NewDecoder(r.Body).Decode(&graf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Fix context usage
	dbErr := s.db.AddGraf(context.Background(), graf, userID)
	if dbErr != nil {
		http.Error(w, dbErr.Err, dbErr.HTTPCode)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *ListyServer) rmFromListyHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

}

func (s *ListyServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Header.Get(immi.UserHeader)
	userID, err := strconv.ParseInt(userIDRaw, 0, 64)
	if err != nil {
		s.logger.Error().Msgf("invalid userID %q", userIDRaw)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var req immi.SubscribeListyTL
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Fix context usage
	listy, dbErr := s.db.GetListy(context.Background(),
		userID, req.ListyRouteName)
	if dbErr != nil {
		http.Error(w, dbErr.Err, dbErr.HTTPCode)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.logger.Error().Msg("Could not init http.Flusher")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	defer func() {
		s.logger.Debug().Msg("Client closed connection. All cleanups complete")
		// TODO: Remove listyID from the active listys eventually
	}()

	for {
		select {
		// TODO: Listen for listy.ID updates and send to the client
		case message := <-time.After(time.Second * 5):
			s.logger.Printf("sending %q after Nsec to %#v", message, listy)
			fmt.Fprintf(w, "data: %s\n\n", message)
			flusher.Flush()
		case <-r.Context().Done():
			s.logger.Print("Client closed connection")
			return
		}
	}
	// Then add the ListyID+SSEStreamHandle to a queue of
	// active Listys to refresh.

	// A new component Tywin should loop on the queue of active Listys
	// and refresh the Listy.

	const refreshListSQL = `
INSERT INTO tl(listy_id, immi_id)
  SELECT 111, id FROM immis WHERE user_id IN (
	  SELECT user_id FROM graf WHERE listy_id = 111
	)
	AND ctime > (SELECT GREATEST(
	  (SELECT TIMEZONE('utc', NOW() - INTERVAL '7 DAYS')),
      (SELECT last_refresh_time FROM listys WHERE id = 111)
	))
	AND ctime < (SELECT TIMEZONE('utc', NOW()))
  ON CONFLICT DO NOTHING
  ;
  `
	_ = refreshListSQL
}

func (s *ListyServer) getTimelineHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.subscribeHandler(w, r)
}
