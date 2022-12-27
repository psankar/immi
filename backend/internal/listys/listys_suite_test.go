package listys_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"immi/internal/idb"
	"immi/internal/listys"
	"immi/pkg/immi"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

func TestListys(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Listys Suite")
}

const J = "application/json"

const (
	user2 = "listysU2"
)

var user1ID string

var db idb.IDB
var dbConn *pgx.Conn

var testServer *httptest.Server
var logger zerolog.Logger

var _ = BeforeSuite(func() {
	logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

	err := idb.EnsureTestDB()
	Expect(err).To(BeNil())

	db, err = idb.NewPGDB(&logger)
	Expect(err).To(BeNil())

	config := listys.ListysConfig{
		DB:     db,
		Logger: &logger,
	}
	server, err := listys.NewServer(config)
	Expect(err).To(BeNil())

	testServer = httptest.NewServer(server.Handler())

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
	VALUES ('listysU1', 'blah', 'blah', 'blah')	RETURNING id
`).Scan(&user1IDRaw)
	user1ID = fmt.Sprintf("%d", user1IDRaw)
	Expect(err).To(BeNil())

	_, err = dbConn.Exec(context.Background(), `
INSERT INTO users(username, email_address, password_hash, user_state)
	VALUES ('listysU2', 'blah', 'blah', 'blah')
`)
	Expect(err).To(BeNil())
})

var _ = Describe("create listy", func() {
	var immiURL string
	var body []byte

	var _ = It("initialize body", func() {
		immiURL = fmt.Sprintf("%s/create-listy", testServer.URL)
		var err error
		body, err = json.Marshal(immi.NewListy{
			DisplayName: "list1",
			RouteName:   "list1",
		})
		Expect(err).To(BeNil())
	})

	var _ = It("test without user ID", func() {
		resp, err := testServer.Client().Get(immiURL)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
	})

	var _ = It("test with invalid userID", func() {
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
		req.Header.Add(immi.UserHeader, user1ID)
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())

		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	var _ = It("test with 1 list", func() {
		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(body),
		)

		req.Header.Add(immi.UserHeader, user1ID)
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		b, err := io.ReadAll(resp.Body)
		Expect(err).To(BeNil())
		log.Println("Response is: ", string(b))
	})
})

var _ = Describe("add to listy", func() {
	var immiURL string
	var body []byte

	var _ = It("Initialize body", func() {
		var err error
		immiURL = fmt.Sprintf("%s/add-to-listy", testServer.URL)
		body, err = json.Marshal(immi.Graf{
			ListRouteName: "list1",
			Username:      user2,
		})
		Expect(err).To(BeNil())
	})

	var _ = It("test without user ID", func() {
		resp, err := testServer.Client().Post(immiURL, J, bytes.NewReader(body))
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
	})

	var _ = It("test with invalid userID", func() {
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
		req.Header.Add(immi.UserHeader, user1ID)
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())

		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	var _ = It("test with non-existent list", func() {
		invBody, err := json.Marshal(immi.Graf{
			ListRouteName: "list2",
			Username:      user2,
		})
		Expect(err).To(BeNil())
		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(invBody),
		)

		req.Header.Add(immi.UserHeader, user1ID)
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	var _ = It("test with non-existent list across users", func() {
		invBody, err := json.Marshal(immi.Graf{
			ListRouteName: "list2234324234242342423",
			Username:      user2,
		})
		Expect(err).To(BeNil())
		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(invBody),
		)

		req.Header.Add(immi.UserHeader, user1ID)
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	var _ = It("test with non-existent username", func() {
		invBody, err := json.Marshal(immi.Graf{
			ListRouteName: "list1",
			Username:      "user22323232323232",
		})
		Expect(err).To(BeNil())
		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(invBody),
		)

		req.Header.Add(immi.UserHeader, user1ID)
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	var _ = It("test with non-existent username and list name", func() {
		invBody, err := json.Marshal(immi.Graf{
			ListRouteName: "list1342424243",
			Username:      "user22323232323232",
		})
		Expect(err).To(BeNil())
		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(invBody),
		)

		req.Header.Add(immi.UserHeader, user1ID)
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})
})

var _ = AfterSuite(func() {
	// cleanup seed data
	_, err := dbConn.Exec(context.Background(),
		`DELETE FROM users WHERE username IN ('listysU1', 'listysU2')`)
	Expect(err).To(BeNil())
})
