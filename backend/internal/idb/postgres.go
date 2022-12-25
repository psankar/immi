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
	"github.com/rs/zerolog"
)

type pg struct {
	conn *pgxpool.Pool
	log  *zerolog.Logger
}

func NewPGDB(log *zerolog.Logger) (*pg, error) {
	connStr, err := DBConnStr()
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	return &pg{conn: conn, log: log}, nil
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
	pg.log.Err(err).Msg("COPY failed")
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
					pg.log.Err(err).Msg("CreateUser failed 1")
					return immi.ErrImmiInternal
				}
			}
			pg.log.Err(err).Msg("CreateUser failed 2")
			return immi.ErrImmiInternal
		}
		pg.log.Err(err).Msg("CreateUser failed 3")
		return immi.ErrImmiInternal
	}

	return nil
}

func (pg *pg) CreateListy(ctx context.Context,
	newListy dao.Listy) *common.Error {
	query := `
INSERT INTO listys (id, display_name, route_name, user_id, ctime)
VALUES ($1, $2, $3, $4, $5)`

	_, err := pg.conn.Exec(ctx, query, newListy.ID, newListy.DisplayName,
		newListy.RouteName, newListy.UserID, newListy.CTime)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				switch pgErr.ConstraintName {
				case "listys_unique_user_id__display_name":
					return immi.ErrDuplicateDisplayName
				case "listys_unique_user_id__route_name":
					return immi.ErrDuplicateRouteName
				default:
					// If we add a new unique constraint later,
					// we can handle it here.
					pg.log.Err(err).Msg("CreateListy failed 1")
					return immi.ErrImmiInternal
				}
			} else if pgErr.Code == pgerrcode.ForeignKeyViolation {
				switch pgErr.ConstraintName {
				case "listys_fk_accounts":
					pg.log.Err(err).
						Int("UserID", int(newListy.UserID)).
						Msg("INVALID UserID detected")
					return immi.ErrImmiInternal
				default:
					// If we add a new foreign key constraint later,
					// we can handle it here.
					pg.log.Err(err).Msg("CreateListy failed 2")
					return immi.ErrImmiInternal
				}
			}
			pg.log.Err(err).Msg("CreateListy failed 3")
			return immi.ErrImmiInternal
		}
		pg.log.Err(err).Msg("CreateListy failed 4")
		return immi.ErrImmiInternal
	}

	return nil
}

func (pg *pg) AddGraf(ctx context.Context, graf immi.Graf) *common.Error {
	query := `
WITH l AS (
	SELECT id, user_id FROM listys WHERE route_name = $1
), u AS (
	SELECT id FROM users WHERE username = $2
)
INSERT INTO graf (listy_id, user_id, ctime) VALUES (
	(SELECT l.id FROM l, u WHERE l.user_id = u.id),
	(SELECT l.user_id FROM l, u WHERE l.user_id = u.id),
	TIMEZONE('utc', NOW())
) ON CONFLICT DO NOTHING;
	`
	_, err := pg.conn.Exec(ctx, query, graf.ListRouteName, graf.Username)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.NotNullViolation {
				return immi.ErrListAddFailed
			}
			pg.log.Err(pgErr).Msg("AddGraf failed with pgErr")
			return immi.ErrImmiInternal
		}
		pg.log.Err(err).Msg("AddGraf failed")
		return immi.ErrImmiInternal
	}

	return nil
}

func (pg *pg) GetUser(ctx context.Context, username string) (
	dao.User, *common.Error) {
	query := `
SELECT username, email_address, password_hash, user_state
FROM users
WHERE username = $1`

	var user dao.User
	err := pg.conn.QueryRow(ctx, query, username).
		Scan(&user.Username, &user.EmailAddress,
			&user.PasswordHash, &user.UserState)
	if err != nil {
		// Returning a zero value instead of pointer, may
		// put less pressure on the GC and avoids crashes
		if err == pgx.ErrNoRows {
			return user, immi.ErrAuthenticationFailed
		}

		pg.log.Err(err).Msg("pg.GetUser failed")
		return user, immi.ErrImmiInternal
	}

	return user, nil
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
