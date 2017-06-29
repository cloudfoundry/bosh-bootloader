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

var _ = Describe("stemcell", func() {
	Describe("StemcellByName", func() {
		It("fetches the stemcell from the director", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/stemcells"))
				Expect(r.Method).To(Equal("GET"))

				username, password, ok := r.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				w.Write([]byte(`[
					{"name": "some-stemcell-name","version": "1"},
					{"name": "some-stemcell-name","version": "2"},
					{"name": "some-other-stemcell-name","version": "100"}
				]`))

			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			stemcell, err := client.StemcellByName("some-stemcell-name")

			Expect(err).NotTo(HaveOccurred())
			Expect(stemcell.Name).To(Equal("some-stemcell-name"))
			Expect(stemcell.Versions).To(Equal([]string{"1", "2"}))
		})

		Context("failure cases", func() {
			It("should error on a non 200 status code with a body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/stemcells"))
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("More Info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.StemcellByName("some-stemcell-name")

				Expect(err).To(MatchError("unexpected response 400 Bad Request:\nMore Info"))
			})

			It("should error with a helpful message on 404 status code", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/stemcells"))
					w.WriteHeader(http.StatusNotFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.StemcellByName("some-stemcell-name")

				Expect(err).To(MatchError("stemcell some-stemcell-name could not be found"))
			})

			It("should error on an unsupported protocol", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "banana://example.com",
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.StemcellByName("some-stemcell-name")
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("should error on a malformed url", func() {
				client := bosh.NewClient(bosh.Config{
					URL:                 "&&&&&%%%&%&%&%&%&",
					TaskPollingInterval: time.Nanosecond,
				})

				_, err := client.StemcellByName("some-stemcell-name")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
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

				_, err := client.StemcellByName("some-stemcell-name")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})

			It("returns an error on a bogus response body", func() {
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/stemcells"))

					w.WriteHeader(http.StatusTeapot)
					w.Write([]byte("More info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL: testServer.URL,
				})

				bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
					return nil, errors.New("a bad read happened")
				})

				_, err := client.StemcellByName("some-stemcell-name")
				Expect(err).To(MatchError("a bad read happened"))
			})
		})
	})

	Describe("StemcellByOS", func() {
		It("fetches the stemcell from the director", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/stemcells"))
				Expect(r.Method).To(Equal("GET"))

				username, password, ok := r.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				w.Write([]byte(`[
					{"operating_system": "some-stemcell-os","version": "1"},
					{"operating_system": "some-stemcell-os","version": "2"},
					{"operating_system": "some-other-stemcell-os","version": "100"}
				]`))

			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			stemcell, err := client.StemcellByOS("some-stemcell-os")

			Expect(err).NotTo(HaveOccurred())
			Expect(stemcell.OS).To(Equal("some-stemcell-os"))
			Expect(stemcell.Versions).To(Equal([]string{"1", "2"}))
		})

		Context("failure cases", func() {
			It("should error on a non 200 status code with a body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/stemcells"))
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("More Info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.StemcellByOS("some-stemcell-os")

				Expect(err).To(MatchError("unexpected response 400 Bad Request:\nMore Info"))
			})

			It("should error with a helpful message on 404 status code", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/stemcells"))
					w.WriteHeader(http.StatusNotFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.StemcellByOS("some-stemcell-os")

				Expect(err).To(MatchError("stemcell some-stemcell-os could not be found"))
			})

			It("should error on an unsupported protocol", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "banana://example.com",
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.StemcellByOS("some-stemcell-os")
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("should error on a malformed url", func() {
				client := bosh.NewClient(bosh.Config{
					URL:                 "&&&&&%%%&%&%&%&%&",
					TaskPollingInterval: time.Nanosecond,
				})

				_, err := client.StemcellByOS("some-stemcell-os")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
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

				_, err := client.StemcellByOS("some-stemcell-os")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})

			It("returns an error on a bogus response body", func() {
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/stemcells"))

					w.WriteHeader(http.StatusTeapot)
					w.Write([]byte("More info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL: testServer.URL,
				})

				bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
					return nil, errors.New("a bad read happened")
				})

				_, err := client.StemcellByOS("some-stemcell-os")
				Expect(err).To(MatchError("a bad read happened"))
			})
		})
	})

	Describe("Latest", func() {
		It("should return the latest stemcell available", func() {
			stemcell := bosh.NewStemcell()
			stemcell.Versions = []string{
				"2127",
				"3147",
				"3126.11",
				"389",
				"3147.2",
				"3126",
				"3263.8",
				"3263.10",
			}

			version, err := stemcell.Latest()
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("3263.10"))
		})

		It("should handle no installed stemcells", func() {
			stemcell := bosh.NewStemcell()
			stemcell.Versions = []string{}

			_, err := stemcell.Latest()
			Expect(err).To(MatchError("no stemcell versions found, cannot get latest"))
		})

		It("returns an error when the stemcell version cannot be parsed", func() {
			stemcell := bosh.NewStemcell()
			stemcell.Versions = []string{
				"baseball",
			}

			_, err := stemcell.Latest()
			Expect(err).To(MatchError(`Invalid character(s) found in major number "baseball"`))
		})
	})
})
