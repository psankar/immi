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

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

type ListyConfig struct {
	Logger *zerolog.Logger
	DB     idb.IDB
}

type ListyServer struct {
	logger *zerolog.Logger
	db     idb.IDB
}

func NewServer(config ListyConfig) (ListyServer, error) {
	return ListyServer{
		logger: config.Logger,
		db:     config.DB,
	}, nil
}

func (s *ListyServer) Handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/create-list", s.createListHandler)
	r.HandleFunc("/add-to-list", s.addToListHandler)
	r.HandleFunc("/rm-from-list", s.rmFromListHandler)
	return r
}

func (s *ListyServer) createListHandler(
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

	var newList immi.NewList
	err = json.NewDecoder(r.Body).Decode(&newList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	listID := xid.New().String()

	// TODO: Fix context usage
	dbErr := s.db.CreateListy(context.Background(), dao.Listy{
		ID:        listID,
		UserID:    userID,
		ListyName: newList.Name,
		CTime:     time.Now().UTC(),
	})
	if dbErr != nil {
		http.Error(w, dbErr.Err, dbErr.HTTPCode)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *ListyServer) addToListHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *ListyServer) rmFromListHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

}
