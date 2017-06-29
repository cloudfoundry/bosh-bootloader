package bosh_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("locks", func() {

	It("returns active locks from the bosh director", func() {
		var (
			client          bosh.Client
			serverCallCount int
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverCallCount++
			Expect(r.URL.Path).To(Equal("/locks"))
			Expect(r.Method).To(Equal("GET"))

			username, password, ok := r.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))

			w.Write([]byte(`[{"type":"deployment","resource":["some-deployment"],"timeout":"1475796348.793560"}]`))
		}))

		client = bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		locks, err := client.Locks()
		Expect(err).NotTo(HaveOccurred())
		Expect(serverCallCount).To(Equal(1))

		Expect(locks).To(Equal([]bosh.Lock{
			{Type: "deployment", Resource: []string{"some-deployment"}, Timeout: "1475796348.793560"},
		}))
	})

	Context("failure cases", func() {
		It("returns an error when url is invalid", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "%%%%%%",
			})

			_, err := client.Locks()
			Expect(err).To(MatchError("parse %%%%%%/locks: invalid URL escape \"%%%\""))
		})

		It("returns an error when Get request fails", func() {
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.CloseClientConnections()
			}))

			client := bosh.NewClient(bosh.Config{
				URL: server.URL,
			})

			_, err := client.Locks()
			Expect(err.Error()).To(ContainSubstring("EOF"))
		})

		It("returns an error when the status code is not 200", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))

			client := bosh.NewClient(bosh.Config{
				URL: server.URL,
			})

			_, err := client.Locks()
			Expect(err).To(MatchError("unexpected response 404 Not Found"))
		})

		It("returns an error when client cannot parse the response", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Some invalid JSON"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL: server.URL,
			})
			_, err := client.Locks()
			Expect(err).To(MatchError(`invalid character 'S' looking for beginning of value`))
		})
	})
})
