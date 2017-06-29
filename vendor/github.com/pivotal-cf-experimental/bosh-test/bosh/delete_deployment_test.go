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

var _ = Context("DeleteDeployment", func() {
	It("deletes the given deployment", func() {
		callCount := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/deployments/some-deployment-name":
				Expect(r.Method).To(Equal("DELETE"))

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
				Fail("could not match any URL endpoints")
			}
		}))

		client := bosh.NewClient(bosh.Config{
			URL:                 server.URL,
			Username:            "some-username",
			Password:            "some-password",
			TaskPollingInterval: time.Nanosecond,
		})

		err := client.DeleteDeployment("some-deployment-name")

		Expect(err).NotTo(HaveOccurred())
		Expect(callCount).To(Equal(4))
	})

	Context("failure cases", func() {
		It("should error if the name is empty", func() {
			client := bosh.NewClient(bosh.Config{
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("")
			Expect(err).To(MatchError("a valid deployment name is required"))
		})

		It("should error on a non 302 redirect response with a body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name":
					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("More Info"))
				case "/tasks/1":
					Fail("should not have redirected to this task")
				default:
					Fail("could not match any URL endpoints")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("some-deployment-name")

			Expect(err).To(MatchError("unexpected response 400 Bad Request:\nMore Info"))
		})

		It("returns an error on a bogus response body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name":
					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("More Info"))
				case "/tasks/1":
					Fail("should not have redirected to this task")
				default:
					Fail("could not match any URL endpoints")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("a bad read happened")
			})

			err := client.DeleteDeployment("some-deployment-name")

			Expect(err).To(MatchError("a bad read happened"))
		})

		It("should error on an errored task status", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name":
					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusFound)
				case "/tasks/1":
					w.Write([]byte(`{"Id": 1, "state": "errored", "result": "some-error-message"}`))
				case "/tasks/1/output":
					if r.URL.RawQuery == "type=event" {
						w.Write([]byte(`
								{"state": "some-state"}
								{"error": {"code": 100, "message": "some-better-error-message"}}
							`))
					}
				default:
					Fail("could not match any URL endpoints")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("some-deployment-name")
			Expect(err).To(MatchError(`task error: 100 has occurred: some-better-error-message`))
		})

		It("should error on an error task status", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name":
					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusFound)
				case "/tasks/1":
					w.Write([]byte(`{"Id": 1, "state": "error", "result": "some-error-message"}`))
				case "/tasks/1/output":
					if r.URL.RawQuery == "type=event" {
						w.Write([]byte(`
								{"state": "some-state"}
								{"error": {"code": 100, "message": "some-better-error-message"}}
							`))
					}
				default:
					Fail("could not match any URL endpoints")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("some-deployment-name")
			Expect(err).To(MatchError(`task error: 100 has occurred: some-better-error-message`))
		})

		It("should error on a cancelled task status", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name":
					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusFound)
				case "/tasks/1":
					w.Write([]byte(`{"state": "cancelled"}`))
				default:
					Fail("could not match any URL endpoints")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("some-deployment-name")
			Expect(err).To(MatchError(errors.New("bosh task was cancelled")))
		})

		It("should error on a malformed redirect location", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", fmt.Sprintf("http://%s/%%%%%%%%%%%%%%", r.Host))
				w.WriteHeader(http.StatusFound)
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("some-deployment-name")
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("should error on a malformed url", func() {
			client := bosh.NewClient(bosh.Config{
				URL:                 "&&&&&%%%&%&%&%&%&",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("some-deployment-name")
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("should error on an unsupported protocol", func() {
			client := bosh.NewClient(bosh.Config{
				URL:                 "banana://some-url",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("some-deployment-name")
			Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
		})

		It("should error on malformed json", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
				w.WriteHeader(http.StatusFound)
				w.Write([]byte(`&&%%%%%&%&%&%&%&%&%&%&`))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.DeleteDeployment("some-deployment-name")

			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})
	})
})
