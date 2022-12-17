package funnel_test

import (
	"bytes"
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
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

func TestFunnel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Funnel Suite")
}

var db idb.IDB
var testServer *httptest.Server
var immiURL string
var logger zerolog.Logger

var _ = BeforeSuite(func() {
	logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

	db, err := idb.NewPGDB()
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
		body, err := json.Marshal(immi.NewImmi{
			Msg: "This is a test message",
		})
		Expect(err).To(BeNil())

		req, err := http.NewRequest(
			http.MethodPost,
			immiURL,
			bytes.NewReader(body),
		)
		req.Header.Add(immi.UserHeader, "123")
		Expect(err).To(BeNil())

		resp, err := testServer.Client().Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		b, err := io.ReadAll(resp.Body)
		Expect(err).To(BeNil())
		log.Println("Response is: ", string(b))

		// Wait for some time, so the batch write would complete
		<-time.After(time.Second * 10)
	})
})
