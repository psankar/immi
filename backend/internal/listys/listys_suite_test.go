package listys_test

import (
	"bytes"
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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

func TestListys(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Listys Suite")
}

var db idb.IDB
var testServer *httptest.Server
var immiURL string
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
	immiURL = fmt.Sprintf("%s/create-listy", testServer.URL)
})

var _ = Describe("listys", func() {
	var _ = It("test without user ID", func() {
		resp, err := testServer.Client().Get(immiURL)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
	})

	var _ = It("test with invalid userID", func() {
		body, err := json.Marshal(immi.NewListy{
			DisplayName: "list1",
			RouteName:   "list1",
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
		body, err := json.Marshal(immi.NewListy{
			DisplayName: "list1",
			RouteName:   "list1",
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

	var _ = It("test with 1 list", func() {
		body, err := json.Marshal(immi.NewListy{
			DisplayName: "list1",
			RouteName:   "list1",
		})
		Expect(err).To(BeNil())

		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(body),
		)
		// TODO: Seed some test users
		req.Header.Add(immi.UserHeader, "1")
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		b, err := io.ReadAll(resp.Body)
		Expect(err).To(BeNil())
		log.Println("Response is: ", string(b))
	})
})
