package bosh_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ScanAndFix", func() {
	Describe("ScanAndFix", func() {
		It("scans and fixes the specified instance", func() {
			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name/scan_and_fix":
					Expect(r.Method).To(Equal("PUT"))
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					username, password, ok := r.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(username).To(Equal("some-username"))
					Expect(password).To(Equal("some-password"))

					body, err := ioutil.ReadAll(r.Body)
					Expect(err).NotTo(HaveOccurred())
					defer r.Body.Close()

					Expect(string(body)).To(MatchJSON(`{
						"jobs":{
							"some-job": [3,7]
						}
					}`))
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

			err := client.ScanAndFix("some-deployment-name", "some-job", []int{3, 7})
			Expect(err).NotTo(HaveOccurred())
			Expect(callCount).To(Equal(4))
		})
	})

	Describe("ScanAndFixAll", func() {
		It("scans and fixes all instances in a deployment", func() {
			var callCount int

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name/scan_and_fix":
					Expect(r.Method).To(Equal("PUT"))
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					username, password, ok := r.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(username).To(Equal("some-username"))
					Expect(password).To(Equal("some-password"))

					body, err := ioutil.ReadAll(r.Body)
					Expect(err).NotTo(HaveOccurred())
					defer r.Body.Close()

					Expect(string(body)).To(MatchJSON(`{
						"jobs":{
							"consul_z1": [0,1],
							"consul_z3": [0]
						}
					}`))
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

			yaml := `---
name: some-deployment-name
jobs:
  - name: consul_z1
    instances: 2
  - name: consul_z2
    instances: 0
  - name: consul_z3
    instances: 1
`

			err := client.ScanAndFixAll([]byte(yaml))
			Expect(err).NotTo(HaveOccurred())
			Expect(callCount).To(Equal(4))
		})

		Context("failure cases", func() {
			It("errors on malformed yaml", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "http://example.com",
					Username: "some-username",
					Password: "some-password",
				})

				err := client.ScanAndFixAll([]byte("%%%%%%%%%%%%%%%"))
				Expect(err).To(MatchError(ContainSubstring("yaml: ")))
			})
		})
	})

	Context("failure cases", func() {
		It("errors when the bosh URL is malformed", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      "banana://example.com",
				Username: "some-username",
				Password: "some-password",
			})

			err := client.ScanAndFixAll([]byte("---\njobs: []"))
			Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
		})

		It("errors when the deployment name contains invalid URL characters", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      "http://example.com%%%%%%%%%",
				Username: "some-username",
				Password: "some-password",
			})

			err := client.ScanAndFixAll([]byte("---\njobs: []"))
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
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

			err := client.ScanAndFixAll([]byte("---\njobs: []"))
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("errors when the response is not a redirect with a body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			err := client.ScanAndFixAll([]byte("---\njobs: []"))
			Expect(err).To(MatchError("unexpected response 400 Bad Request:\nMore Info"))
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

			err := client.ScanAndFixAll([]byte("---\njobs: []"))
			Expect(err).To(MatchError("a bad read happened"))
		})
	})
})
