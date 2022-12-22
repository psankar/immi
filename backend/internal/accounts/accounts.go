package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"immi/internal/idb"
	"immi/pkg/dao"
	"immi/pkg/immi"
	"net/http"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
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
	var signupReq immi.SignUp
	err := json.NewDecoder(r.Body).Decode(&signupReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	passwordBytes, err := bcrypt.GenerateFromPassword(
		[]byte(signupReq.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// TODO: validate fields

	// TODO: Fix context usage
	err = s.db.CreateUser(context.Background(), dao.User{
		Username:     signupReq.Username,
		EmailAddress: signupReq.EmailAddress,
		PasswordHash: string(passwordBytes),
		UserState:    dao.ActiveUser,
	})
	if err != nil {
		var userError immi.UserError
		if errors.As(err, &userError) {
			http.Error(w, userError.Error(), http.StatusBadRequest)
			return
		}

		s.logger.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *AccountsServer) loginHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *AccountsServer) logoutHandler(w http.ResponseWriter, r *http.Request) {

}
