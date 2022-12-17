package idb

import (
	"fmt"
	"immi/pkg/dao"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type pg struct {
}

func NewPGDB() (*pg, error) {

	return &pg{}, nil
}

func (pg *pg) AppendImmis(immis []dao.Immi) error {
	return nil
}

var (
	ErrPGDB       = errors.New("invalid POSTGRES_DB")
	ErrPGUser     = errors.New("invalid POSTGRES_USER")
	ErrPGPassword = errors.New("invalid POSTGRES_PASSWORD")
	ErrPGHost     = errors.New("invalid POSTGRES_HOST")
	ErrPGPort     = errors.New("invalid POSTGRES_PORT")
)

const (
	PostgresDB       = "POSTGRES_DB"
	PostgresUser     = "POSTGRES_USER"
	PostgresPassword = "POSTGRES_PASSWORD"
	PostgresHost     = "POSTGRES_HOST"
	PostgresPort     = "POSTGRES_PORT"
)

func DBConnStr() (string, error) {
	db, ok := os.LookupEnv(PostgresDB)
	if !ok || len(strings.TrimSpace(db)) == 0 {
		return "", ErrPGDB
	}

	user, ok := os.LookupEnv(PostgresUser)
	if !ok || len(strings.TrimSpace(user)) == 0 {
		return "", ErrPGUser
	}

	password, ok := os.LookupEnv(PostgresPassword)
	if !ok || len(strings.TrimSpace(password)) == 0 {
		return "", ErrPGPassword
	}

	host, ok := os.LookupEnv(PostgresHost)
	if !ok || len(strings.TrimSpace(host)) == 0 {
		return "", ErrPGHost
	}

	port, ok := os.LookupEnv(PostgresPort)
	if !ok || len(strings.TrimSpace(port)) == 0 {
		return "", ErrPGPort
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, db)
	return connStr, nil
}
