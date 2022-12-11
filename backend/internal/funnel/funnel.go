package funnel

import (
	"context"
	"encoding/json"
	"immi/pkg/dao"
	"immi/pkg/immi"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

type FunnelConfig struct {
	BatchSize     int
	BatchDuration time.Duration
}

type FunnelServer struct {
	batchSize     int
	batchDuration time.Duration
	batchChan     chan dao.Immi
	batch         []dao.Immi
	ctx           context.Context
}

func NewServer(config FunnelConfig) (*FunnelServer, error) {
	// TODO: validate config
	server := &FunnelServer{
		batchSize:     config.BatchSize,
		batchDuration: config.BatchDuration,
		batchChan:     make(chan dao.Immi),
		batch:         make([]dao.Immi, 0, config.BatchSize),
		ctx:           context.TODO(),
	}

	go server.batcher()
	return server, nil
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

	// TODO: Validate newImmi

	immiID := xid.New().String()

	immiDAO := dao.Immi{
		ID:     immiID,
		UserID: userID,
		Msg:    newImmi.Msg,
		CTime:  time.Now().UTC(),
	}

	s.batchChan <- immiDAO

	w.Write([]byte(immiID))
}

func (s *FunnelServer) batcher() {
	// We use a for-select here instead of a Mutex for s.batch,
	// because the SELECT does not lead to starvation of the
	// DB Writes operation, due to pseudo-randomness. Locks
	// do not guarantee against the DB Writer starvation.
	for {
		timer := time.NewTimer(s.batchDuration)
		select {
		case immi := <-s.batchChan:
			s.batch = append(s.batch, immi)
		case <-s.ctx.Done():
			// TODO: Graceful shutdown
		case <-timer.C:
			if len(s.batch) == 0 {
				// No Immis to write as of now
				continue
			}
			x := s.batch
			s.batch = make([]dao.Immi, 0, s.batchSize)

			// TODO: Write to DB
			log.Printf("Write to DB: %v", len(x))
		}
	}
}
