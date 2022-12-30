package e2e_tests_test

import (
	"fmt"
	"immi/pkg/immi"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

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
	NumImmis = 10000
	J        = "application/json"
)

var (
	userTokens []string
)

var _ = Describe("Immi backend testing", func() {
	It("Initialise", func() {
		userTokens = make([]string, NumUsers)
	})

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

	It("Signin users", func() {
		// This could be moved to the accounts package unit-test
		req := `{"Username": "user1", "Password": "user2"}`
		resp, err := http.Post(ImmiURL+"/accounts/login", J,
			strings.NewReader(req))
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

		req = `{"Username": "abcdef", "Password": "user2"}`
		resp, err = http.Post(ImmiURL+"/accounts/login", J,
			strings.NewReader(req))
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

		req = `{"Username": "user1", "Password": "user1"}`
		resp, err = http.Post(ImmiURL+"/accounts/login", J,
			strings.NewReader(req))
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		err = resp.Body.Close()
		Expect(err).To(BeNil())
	})

	It("Save login tokens to array", func() {
		for i := 0; i < NumUsers; i++ {
			req := fmt.Sprintf(
				`{"Username": "user%d", "Password": "user%d"}`,
				i, i)
			resp, err := http.Post(ImmiURL+"/accounts/login", J,
				strings.NewReader(req))
			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			body, err := io.ReadAll(resp.Body)
			Expect(err).To(BeNil())
			loginToken := string(body)
			Expect(loginToken).ToNot(BeEmpty())
			userTokens[i] = loginToken
			err = resp.Body.Close()
			Expect(err).To(BeNil())
		}
	})

	It("Create Listys for each user", func() {
		for i := 0; i < NumUsers; i++ {
			for j := 0; j < 10; j++ {
				log.Printf("Creating list 'list%d' for user 'user%d'", j, i)
				body := fmt.Sprintf(
					`{
						"DisplayName": "list%d",
						"RouteName":   "list%d"
					}`,
					j, j)

				req, err := http.NewRequest(
					http.MethodPost,
					ImmiURL+"/listys/create-listy",
					strings.NewReader(body),
				)
				Expect(err).To(BeNil())

				// TODO: This would work only when the tests were run
				// on a clean vanilla database. Also, this MUST fail in prod,
				// as the UserHeader MUST be over-written at the port of entry.
				req.Header.Add(immi.UserHeader, fmt.Sprintf("%d", i+1))
				resp, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				err = resp.Body.Close()
				Expect(err).To(BeNil())
			}
		}
	})

	It("Add Users to Listys", func() {
		for i := 0; i < NumUsers; i++ {
			for j := 0; j < 10; j++ {
				for k := j; k < NumUsers; k += 10 {
					// log.Printf("Adding user%d to list%d of user%d", k, j, i)

					body := fmt.Sprintf(
						`{
							"ListRouteName":   "list%d",
							"Username": "user%d"
						}`,
						j, k)

					req, err := http.NewRequest(
						http.MethodPost,
						ImmiURL+"/listys/add-to-listy",
						strings.NewReader(body),
					)
					Expect(err).To(BeNil())

					// TODO: This would work only when the tests were run
					// on a clean vanilla database. Also, this MUST fail in prod,
					// as the UserHeader MUST be over-written at the port of entry.
					req.Header.Add(immi.UserHeader, fmt.Sprintf("%d", i+1))
					resp, err := http.DefaultClient.Do(req)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					err = resp.Body.Close()
					Expect(err).To(BeNil())
				}
			}
		}
	})

	It("Post Immies", func() {
		var wg sync.WaitGroup
		for i := 0; i < NumUsers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer GinkgoRecover()
				defer wg.Done()
				for j := 0; j < NumImmis; j++ {
					body := fmt.Sprintf(`{"Msg":   "User%d TestMsg: %d"}`,
						workerID, j)

					req, err := http.NewRequest(
						http.MethodPost,
						ImmiURL+"/funnel",
						strings.NewReader(body),
					)
					Expect(err).To(BeNil())

					// TODO: This would work only when the tests were run
					// on a clean vanilla database. Also, this MUST fail in prod,
					// as the UserHeader MUST be over-written at the port of entry.
					req.Header.Add(
						immi.UserHeader,
						fmt.Sprintf("%d", workerID+1),
					)
					resp, err := http.DefaultClient.Do(req)
					Expect(err).To(BeNil())

					if resp.StatusCode != http.StatusOK {
						rbody, err := io.ReadAll(resp.Body)
						log.Println(string(rbody), err)
					}
					Expect(resp.StatusCode).To(Equal(http.StatusOK))

					err = resp.Body.Close()
					Expect(err).To(BeNil())

					// Sleep a random time (until 3 seconds) in between requests
					time.Sleep(
						time.Duration(rand.Intn(3000)) * time.Millisecond,
					)
				}

			}(i)
		}
		wg.Wait()
	})
})
