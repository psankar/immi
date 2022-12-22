package accounts

import (
	"immi/internal/idb"
	"net/http"

	"github.com/rs/zerolog"
)

type AccountsConfig struct {
	Logger *zerolog.Logger
	DB     idb.IDB
}

type AccountsServer struct {
	logger *zerolog.Logger
	db     idb.IDB
}

func NewServer(config AccountsConfig) (AccountsServer, error) {
	return AccountsServer{
		logger: config.Logger,
		db:     config.DB,
	}, nil
}

func (s *AccountsServer) Handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/accounts/signup", s.signupHandler)
	r.HandleFunc("/accounts/login", s.loginHandler)
	r.HandleFunc("/accounts/logout", s.logoutHandler)
	return r
}

func (s *AccountsServer) signupHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *AccountsServer) loginHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *AccountsServer) logoutHandler(w http.ResponseWriter, r *http.Request) {

}
