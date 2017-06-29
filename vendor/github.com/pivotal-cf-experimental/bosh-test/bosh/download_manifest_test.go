package bosh_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DownloadManifest", func() {
	It("downloads manifest", func() {
		testServerCallCount := 0

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			testServerCallCount++

			Expect(r.Method).To(Equal("GET"))
			Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name"))

			username, password, ok := r.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"manifest": "some-manifest-contents"}`))
		}))

		client := bosh.NewClient(bosh.Config{
			URL:      testServer.URL,
			Username: "some-username",
			Password: "some-password",
		})

		rawManifest, err := client.DownloadManifest("some-deployment-name")

		Expect(err).NotTo(HaveOccurred())
		Expect(testServerCallCount).To(Equal(1))
		Expect(rawManifest).To(Equal([]byte("some-manifest-contents")))
	})

	Context("failure cases", func() {
		It("returns an error when request is malformed", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "%%%%%",
			})

			_, err := client.DownloadManifest("some-deployment-name")
			Expect(err).To(MatchError(`parse %%%%%/deployments/some-deployment-name: invalid URL escape "%%%"`))
		})

		It("returns an error when the request fails", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "",
			})

			_, err := client.DownloadManifest("")

			Expect(err).To(MatchError(`unsupported protocol scheme ""`))
		})

		It("errors on an unexpected status code with a body", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name"))

				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL: testServer.URL,
			})

			_, err := client.DownloadManifest("some-deployment-name")
			Expect(err).To(MatchError("unexpected response 418 I'm a teapot:\nMore Info"))
		})

		It("returns an error on a bogus response body", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name"))

				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL: testServer.URL,
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("a bad read happened")
			})

			_, err := client.DownloadManifest("some-deployment-name")
			Expect(err).To(MatchError("a bad read happened"))
		})

		It("returns an error when server returns malformed json", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name"))

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("%%%%%%"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL: testServer.URL,
			})

			_, err := client.DownloadManifest("some-deployment-name")
			Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
		})
	})
})
