package listys

import (
	"context"
	"encoding/json"
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
