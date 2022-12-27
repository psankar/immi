package funnel_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"immi/internal/funnel"
	"immi/internal/idb"
	"immi/pkg/immi"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	pgx "github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

func TestFunnel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Funnel Suite")
}

var user1ID string

var db idb.IDB
var dbConn *pgx.Conn

var msg string

var testServer *httptest.Server
var immiURL string
var logger zerolog.Logger

var _ = BeforeSuite(func() {
	logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

	err := idb.EnsureTestDB()
	Expect(err).To(BeNil())

	db, err = idb.NewPGDB(&logger)
	Expect(err).To(BeNil())

	config := funnel.FunnelConfig{
		BatchSize:     3,
		BatchDuration: time.Second * 3,
		DB:            db,
		Logger:        &logger,
	}
	server, err := funnel.NewServer(config)
	Expect(err).To(BeNil())

	testServer = httptest.NewServer(server.Handler())
	immiURL = fmt.Sprintf("%s/immis", testServer.URL)

	// Seed Data
	dbStr, err := idb.DBConnStr()
	Expect(err).To(BeNil())
	Expect(dbStr).ToNot(BeEmpty())
	dbConn, err = pgx.Connect(context.Background(), dbStr)
	Expect(err).To(BeNil())
	Expect(dbConn).ToNot(BeNil())

	var user1IDRaw int64
	err = dbConn.QueryRow(context.Background(), `
INSERT INTO users(username, email_address, password_hash, user_state)
	VALUES ('funnelU1', 'blah', 'blah', 'blah')	RETURNING id
`).Scan(&user1IDRaw)
	user1ID = fmt.Sprintf("%d", user1IDRaw)
	Expect(err).To(BeNil())
})

var _ = Describe("immis", func() {
	var _ = It("test without user ID", func() {
		resp, err := testServer.Client().Get(immiURL)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
	})

	var _ = It("test with invalid userID", func() {
		body, err := json.Marshal(immi.NewImmi{
			Msg: "This is a test message",
		})
		Expect(err).To(BeNil())

		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(body),
		)
		req.Header.Add(immi.UserHeader, "123asd")
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
	})

	var _ = It("test with empty userID", func() {
		body, err := json.Marshal(immi.NewImmi{
			Msg: "This is a test message",
		})
		Expect(err).To(BeNil())

		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(body),
		)
		req.Header.Add(immi.UserHeader, "")
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
	})

	var _ = It("test with invalid body", func() {
		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			strings.NewReader(`{myjson: "need not be valid"}`),
		)
		req.Header.Add(immi.UserHeader, "123")
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())

		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	var _ = It("test with 1 immi", func() {
		msg = fmt.Sprintf("This is a test message: %q", xid.New().String())
		body, err := json.Marshal(immi.NewImmi{
			Msg: msg,
		})
		Expect(err).To(BeNil())

		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(body),
		)
		// TODO: Seed some test users
		req.Header.Add(immi.UserHeader, user1ID)
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		b, err := io.ReadAll(resp.Body)
		Expect(err).To(BeNil())
		log.Println("Response is: ", string(b))

		// Wait for some time, so the batch write would complete
		<-time.After(time.Second * 10)

		var dbUserID int
		var dbMsg string
		err = dbConn.QueryRow(context.Background(),
			"SELECT user_id, msg FROM immis WHERE id = $1",
			string(b)).Scan(&dbUserID, &dbMsg)
		Expect(err).To(BeNil())

		user1IDInt64, err := strconv.ParseInt(user1ID, 10, 64)
		Expect(err).To(BeNil())

		Expect(dbUserID).To(BeEquivalentTo(user1IDInt64))
		Expect(dbMsg).To(BeEquivalentTo(msg))
	})
})

var _ = AfterSuite(func() {
	// cleanup seed data
	_, err := dbConn.Exec(context.Background(),
		`DELETE FROM users WHERE username = 'funnelU1'`)
	Expect(err).To(BeNil())
})
