package funnel_test

import (
	"fmt"
	"immi/internal/funnel"
	"io"
	"log"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFunnel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Funnel Suite")
}

var testServer *httptest.Server

var _ = BeforeSuite(func() {
	server, err := funnel.NewServer()
	Expect(err).To(BeNil())

	testServer = httptest.NewServer(server.Handler())
})

var _ = Describe("immis", func() {
	var _ = It("test with 1 immi", func() {
		immiURL := fmt.Sprintf("%s/immis", testServer.URL)
		resp, err := testServer.Client().Get(immiURL)
		Expect(err).To(BeNil())

		b, err := io.ReadAll(resp.Body)
		Expect(err).To(BeNil())

		log.Println("Response is: ", string(b))
	})
})
