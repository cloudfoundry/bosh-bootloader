package bosh_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restart", func() {
	Describe("Restart", func() {
		It("restarts a job instance", func() {
			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name/jobs/some-job/0":
					Expect(r.Method).To(Equal("PUT"))
					Expect(r.Header.Get("Content-Type")).To(Equal("text/yaml"))
					Expect(r.URL.RawQuery).To(Equal("state=restart"))

					username, password, ok := r.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(username).To(Equal("some-username"))
					Expect(password).To(Equal("some-password"))

					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusFound)
				case "/tasks/1":
					Expect(r.Method).To(Equal("GET"))

					username, password, ok := r.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(username).To(Equal("some-username"))
					Expect(password).To(Equal("some-password"))

					if callCount == 3 {
						w.Write([]byte(`{"state": "done"}`))
					} else {
						w.Write([]byte(`{"state": "processing"}`))
					}
					callCount++
				default:
					Fail("unexpected route")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.Restart("some-deployment-name", "some-job", 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(callCount).To(Equal(4))
		})

		Context("failure cases", func() {
			It("errors when the deployment name contains invalid URL characters", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("something bad happened"))
				}))
				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				err := client.Restart("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError(ContainSubstring("unexpected response 500 Internal Server Error:\nsomething bad happened")))
			})

			It("errors when the deployment name contains invalid URL characters", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "http://example.com%%%%%%%%%",
					Username: "some-username",
					Password: "some-password",
				})

				err := client.Restart("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("errors when the bosh URL is malformed", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "banana://example.com",
					Username: "some-username",
					Password: "some-password",
				})

				err := client.Restart("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("errors when the redirect location is bad", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Location", "%%%%%%%%%%%")
					w.WriteHeader(http.StatusFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				err := client.Restart("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("returns an error on a bogus response body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("More Info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
					return nil, errors.New("a bad read happened")
				})

				err := client.Restart("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError("a bad read happened"))
			})
		})
	})
})
