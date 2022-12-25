package accounts

import (
	"context"
	"encoding/json"
	"immi/internal/idb"
	"immi/pkg/dao"
	"immi/pkg/immi"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
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
	r.HandleFunc("/signup", s.signupHandler)
	r.HandleFunc("/login", s.loginHandler)
	r.HandleFunc("/logout", s.logoutHandler)
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
		s.logger.Err(err).Msg("Password hash generation failed")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// TODO: validate fields

	// TODO: Fix context usage
	dbErr := s.db.CreateUser(context.Background(), dao.User{
		Username:     signupReq.Username,
		EmailAddress: signupReq.EmailAddress,
		PasswordHash: string(passwordBytes),
		UserState:    dao.ActiveUser,
	})
	if dbErr != nil {
		http.Error(w, dbErr.Err, dbErr.HTTPCode)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *AccountsServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	var loginReq immi.Login
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, dbErr := s.db.GetUser(context.Background(), loginReq.Username)
	if dbErr != nil {
		http.Error(w, dbErr.Err, dbErr.HTTPCode)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash),
		[]byte(loginReq.Password))
	if err != nil {
		http.Error(w, immi.ErrAuthenticationFailed.Err, http.StatusUnauthorized)
		return
	}

	// TODO: We need a better system here with refresh tokens
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		s.logger.Err(err).Msg("JWT generation failed")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(tokenString))
}

func (s *AccountsServer) logoutHandler(w http.ResponseWriter, r *http.Request) {

}
