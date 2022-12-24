package e2e_tests_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2eTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2eTests Suite")
}

const (
	ImmiURL  = "http://localhost"
	NumUsers = 100
	json     = "application/json"
)

var _ = Describe("Accounts testing", func() {
	It("Signup valid users", func() {
		for i := 0; i < NumUsers; i++ {
			req := fmt.Sprintf(
				`{
					"Username": "user%d",
					"EmailAddress": "user%d@example.com",
					"Password": "user%d"
				}`,
				i, i, i)
			resp, err := http.Post(
				ImmiURL+"/accounts/signup",
				"application/json",
				strings.NewReader(req),
			)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}
	})

	// TODO: Tests for invalid signups that would fail validation

	It("Signin users", func() {
		req := `{"Username": "user1", "Password": "user2"}`
		resp, err := http.Post(ImmiURL+"/accounts/login", json,
			strings.NewReader(req))
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

		req = `{"Username": "user1", "Password": "user1"}`
		resp, err = http.Post(ImmiURL+"/accounts/login", json,
			strings.NewReader(req))
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		body, err := io.ReadAll(resp.Body)
		log.Println(string(body), err)
	})
})
