package bosh_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Context("DeleteRelease", func() {
	It("deletes the given stemcell", func() {
		callCount := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/releases/otherthing":
				Expect(r.Method).To(Equal("DELETE"))

				Expect(r.URL.Query().Get("version")).To(Equal("42"))

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
				req, err := httputil.DumpRequest(r, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(string(req))
			}
		}))

		client := bosh.NewClient(bosh.Config{
			URL:                 server.URL,
			Username:            "some-username",
			Password:            "some-password",
			TaskPollingInterval: time.Nanosecond,
		})

		err := client.DeleteRelease("otherthing", "42")

		Expect(err).NotTo(HaveOccurred())
		Expect(callCount).To(Equal(4))
	})

	Context("when an error occurs", func() {
		Context("when the status is not 302", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/releases/otherthing":
						w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("More Info"))
					default:
						req, err := httputil.DumpRequest(r, true)
						Expect(err).NotTo(HaveOccurred())
						Fail(string(req))
					}
				}))

				client := bosh.NewClient(bosh.Config{
					URL:                 server.URL,
					Username:            "some-username",
					Password:            "some-password",
					TaskPollingInterval: time.Nanosecond,
				})

				err := client.DeleteRelease("otherthing", "42")

				Expect(err).To(MatchError("unexpected response 400 Bad Request:\nMore Info"))
			})
		})

		Context("when any other error code comes back", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/releases/otherthing":
						w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
						w.WriteHeader(http.StatusFound)
					case "/tasks/1":
						w.Write([]byte(`{"Id": 1, "state": "error", "result": "Stemcell something/42 is still in use"}`))
					case "/tasks/1/output":
						if r.URL.RawQuery == "type=event" {
							w.Write([]byte(`
								{"state": "some-state"}
								{"error": {"code": 4999, "message": "some other random error"}}
							`))
						}
					default:
						req, err := httputil.DumpRequest(r, true)
						Expect(err).NotTo(HaveOccurred())
						Fail(string(req))
					}
				}))

				client := bosh.NewClient(bosh.Config{
					URL:                 server.URL,
					Username:            "some-username",
					Password:            "some-password",
					TaskPollingInterval: time.Nanosecond,
				})

				err := client.DeleteRelease("otherthing", "42")
				Expect(err).To(MatchError("task error: 4999 has occurred: some other random error"))
			})
		})
	})
})
