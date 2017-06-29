package bosh_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cleanup", func() {
	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/cleanup":
				Expect(req.Method).To(Equal("POST"))
				username, password, ok := req.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				var contents map[string]interface{}
				err := json.NewDecoder(req.Body).Decode(&contents)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(Equal(map[string]interface{}{
					"config": map[string]interface{}{
						"remove_all": true,
					},
				}))

				w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/5", req.Host))
				w.WriteHeader(http.StatusFound)

			case "/tasks/5":
				Expect(req.Method).To(Equal("GET"))
				username, password, ok := req.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				w.Write([]byte(`{"id": 5, "state": "done"}`))

			default:
				Fail(fmt.Sprintf("unhandled request to %s", req.URL.Path))
			}
		}))
	})

	It("cleans up the bosh director", func() {
		client := bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		taskID, err := client.Cleanup()
		Expect(err).NotTo(HaveOccurred())
		Expect(taskID).To(Equal(5))
	})

	Context("failure cases", func() {
		Context("when the request cannot be created", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "%%%%%",
				})

				_, err := client.Cleanup()
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		Context("when the request cannot be made", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "",
				})

				_, err := client.Cleanup()
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
			})
		})

		Context("when the request returns an unexpected response status", func() {
			It("returns an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				}))

				client := bosh.NewClient(bosh.Config{
					URL: server.URL,
				})

				_, err := client.Cleanup()
				Expect(err).To(MatchError("unexpected response 418 I'm a teapot"))
			})
		})
	})
})
