package bosh_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info", func() {
	It("fetches the director info", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/info"))
			Expect(r.Method).To(Equal("GET"))

			w.Write([]byte(`{"uuid":"some-director-uuid", "cpi":"some-cpi"}`))
		}))

		client := bosh.NewClient(bosh.Config{
			URL:                 server.URL,
			TaskPollingInterval: time.Nanosecond,
		})

		info, err := client.Info()

		Expect(err).NotTo(HaveOccurred())
		Expect(info).To(Equal(bosh.DirectorInfo{
			UUID: "some-director-uuid",
			CPI:  "some-cpi",
		}))
	})

	Context("failure cases", func() {
		It("should error on malformed json", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`&&%%%%%&%&%&%&%&%&%&%&`))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				TaskPollingInterval: time.Nanosecond,
			})

			_, err := client.Info()

			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})

		It("should error on an unsupported protocol", func() {
			client := bosh.NewClient(bosh.Config{
				URL:                 "banana://example.com",
				TaskPollingInterval: time.Nanosecond,
			})

			_, err := client.Info()
			Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
		})

		It("returns an error on an unexpected status code with body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Info()
			Expect(err).To(MatchError("unexpected response 502 Bad Gateway:\nMore Info"))
		})

		It("returns an error on a bogus response body", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL: testServer.URL,
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("a bad read happened")
			})

			_, err := client.Info()
			Expect(err).To(MatchError("a bad read happened"))
		})
	})
})
