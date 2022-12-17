package funnel

import (
	"context"
	"encoding/json"
	"errors"
	"immi/internal/idb"
	"immi/pkg/dao"
	"immi/pkg/immi"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type FunnelConfig struct {
	BatchSize     int
	BatchDuration time.Duration
	DB            idb.IDB
	Logger        *zerolog.Logger
}

type FunnelServer struct {
	batchSize     int
	batchDuration time.Duration
	batchChan     chan dao.Immi
	batch         []dao.Immi
	ctx           context.Context
	db            idb.IDB
	log           *zerolog.Logger
}

func NewServer(config FunnelConfig) (*FunnelServer, error) {
	// TODO: validate config
	server := &FunnelServer{
		batchSize:     config.BatchSize,
		batchDuration: config.BatchDuration,
		batchChan:     make(chan dao.Immi),
		batch:         make([]dao.Immi, 0, config.BatchSize),
		ctx:           context.TODO(),
		db:            config.DB,
		log:           config.Logger,
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
			timer.Stop() // This is not needed perhaps, just to be safe
			s.batch = append(s.batch, immi)
		case <-s.ctx.Done():
			// TODO: Graceful shutdown
		case <-timer.C:
			if len(s.batch) == 0 {
				// No Immis to write as of now
				continue
			}
			immis := s.batch
			s.batch = make([]dao.Immi, 0, s.batchSize)
			err := s.db.AppendImmis(immis)
			if err != nil {
				if errors.Is(err, idb.ErrInternal) {
					// Likely case as validations would have
					// handled user errors already
					s.log.Error().Err(err)
				} else {
					// TODO: Handle user errors
					_ = err
				}
			}
		}
	}
}
