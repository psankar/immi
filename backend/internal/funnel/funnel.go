package funnel

import (
	"encoding/json"
	"immi/pkg/dao"
	"immi/pkg/immi"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

type FunnelServer struct {
}

func NewServer() (*FunnelServer, error) {
	return &FunnelServer{}, nil
}

func (s *FunnelServer) Handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/immis", s.immiHandler)
	return r
}

func (s *FunnelServer) immiHandler(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Header.Get(immi.UserHeader)
	userID, err := strconv.ParseInt(userIDRaw, 0, 64)
	if err != nil {
		log.Error().Msg("userID not set in incoming request")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var newImmi immi.NewImmi
	err = json.NewDecoder(r.Body).Decode(&newImmi)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	immiID := xid.New().String()

	immiDAO := dao.Immi{
		ID:     immiID,
		UserID: userID,
		Msg:    newImmi.Msg,
		CTime:  time.Now().UTC(),
	}

	_ = immiDAO // TODO: Write to DB

	w.Write([]byte(immiID))
}
