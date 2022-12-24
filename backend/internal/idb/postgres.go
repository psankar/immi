package idb

import (
	"context"
	"fmt"
	"immi/internal/common"
	"immi/pkg/dao"
	"immi/pkg/immi"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgerrcode"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type pg struct {
	conn *pgxpool.Pool
}

func NewPGDB() (*pg, error) {
	connStr, err := DBConnStr()
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	return &pg{conn: conn}, nil
}

func PGErrMsg(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Error()
	}

	return err.Error()
}

func (pg *pg) AppendImmis(ctx context.Context, immis []dao.Immi) *common.Error {
	_, err := pg.conn.CopyFrom(
		ctx,
		pgx.Identifier{"immis"},
		[]string{"id", "user_id", "msg", "ctime"},
		pgx.CopyFromSlice(len(immis), func(i int) ([]any, error) {
			return []any{
				immis[i].ID,
				immis[i].UserID,
				immis[i].Msg,
				immis[i].CTime,
			}, nil
		}),
	)
	return common.Err(err, http.StatusInternalServerError)
}

func (pg *pg) CreateUser(ctx context.Context, user dao.User) *common.Error {
	query := `
INSERT INTO users (username, email_address, password_hash, user_state)
	VALUES ($1, $2, $3, $4)`

	_, err := pg.conn.Exec(ctx, query, user.Username, user.EmailAddress,
		user.PasswordHash, user.UserState)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				switch pgErr.ConstraintName {
				case "users_unique_username":
					return immi.ErrDuplicateUsername
				default:
					// If we add a new unique constraint later,
					// we can handle it here.
					return immi.ErrImmiInternal
				}
			}
			return immi.ErrImmiInternal
		}
		return immi.ErrImmiInternal
	}

	return nil
}

// Private Errors for backend; For Public errors for clients, see immi package
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

func EnsureTestDB() error {
	for _, i := range []struct {
		env   string
		value string
	}{
		{PostgresDB, "immi"},
		{PostgresUser, "immi"},
		{PostgresPassword, "password"},
		{PostgresHost, "localhost"},
		{PostgresPort, "5432"},
	} {
		err := os.Setenv(i.env, i.value)
		if err != nil {
			return err
		}
	}

	return nil
}
