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

var _ = Describe("release", func() {
	Context("Release", func() {
		It("fetches the release from the director", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/releases/some-release-name"))
				Expect(r.Method).To(Equal("GET"))

				username, password, ok := r.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				w.Write([]byte(`{"versions":["some-version","some-version.1","some-version.2"]}`))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			release, err := client.Release("some-release-name")

			Expect(err).NotTo(HaveOccurred())
			Expect(release.Name).To(Equal("some-release-name"))
			Expect(release.Versions).To(Equal([]string{"some-version", "some-version.1", "some-version.2"}))
		})

		Context("failure cases", func() {
			It("should error on a non 200 status code with a body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/releases/some-release-name"))
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("More Info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.Release("some-release-name")

				Expect(err).To(MatchError("unexpected response 400 Bad Request:\nMore Info"))
			})

			It("should error with a helpful message on 404 status code", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/releases/some-release-name"))
					w.WriteHeader(http.StatusNotFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.Release("some-release-name")

				Expect(err).To(MatchError("release some-release-name could not be found"))
			})

			It("should error on an unsupported protocol", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "banana://example.com",
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.Release("some-release-name")
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("should error on malformed json", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(`&&%%%%%&%&%&%&%&%&%&%&`))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.Release("some-release-name")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})

			It("should error on a malformed url", func() {
				client := bosh.NewClient(bosh.Config{
					URL:                 "&&&&&%%%&%&%&%&%&",
					TaskPollingInterval: time.Nanosecond,
				})

				_, err := client.Release("some-release-name")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("returns an error on a bogus response body", func() {
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/releases/some-release-name"))

					w.WriteHeader(http.StatusTeapot)
					w.Write([]byte("More info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL: testServer.URL,
				})

				bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
					return nil, errors.New("a bad read happened")
				})

				_, err := client.Release("some-release-name")
				Expect(err).To(MatchError("a bad read happened"))
			})
		})
	})
	Context("Latest", func() {
		It("should return the latest release available", func() {
			release := bosh.NewRelease()

			release.Versions = []string{
				"21+dev.2",
				"21+dev.3",
				"21+dev.4",
				"21+dev.5",
				"21+dev.6",
				"21+dev.7",
				"21+dev.8",
				"21+dev.9",
				"21+dev.10",
				"21+dev.11",
				"21+dev.12",
				"21+dev.13",
				"21+dev.14",
				"21+dev.15",
				"21+dev.16",
				"21+dev.17",
				"21+dev.18",
				"21+dev.19",
				"21+dev.20",
				"21+dev.21",
				"21+dev.22",
				"21+dev.23",
				"21+dev.24",
				"21+dev.25",
				"21+dev.26",
				"21+dev.27",
				"21+dev.28",
			}

			Expect(release.Latest()).To(Equal("21+dev.28"))
		})
	})
})
