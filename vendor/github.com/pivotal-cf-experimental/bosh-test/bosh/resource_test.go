package bosh_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {
	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			Expect(req.URL.Path).To(Equal("/resources/some-resource-guid"))
			Expect(req.Method).To(Equal("GET"))
			username, password, ok := req.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("I am a banana!"))
		}))
	})

	It("returns the specified resource", func() {
		client := bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		resource, err := client.Resource("some-resource-guid")
		Expect(err).NotTo(HaveOccurred())
		defer resource.Close()

		contents, err := ioutil.ReadAll(resource)
		Expect(err).NotTo(HaveOccurred())

		Expect(contents).To(Equal([]byte("I am a banana!")))
	})

	Context("failure cases", func() {
		Context("when the request cannot be created", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "%%%%%",
				})

				_, err := client.Resource("some-resource-guid")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		Context("when the request fails to be made", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "",
				})

				_, err := client.Resource("some-resource-guid")
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
			})
		})

		Context("when the request returns an unexpected status code", func() {
			It("returns an error with the body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
					w.Write([]byte("More Info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.Resource("some-resource-guid")
				Expect(err).To(MatchError("unexpected response 418 I'm a teapot:\nMore Info"))
			})
		})

		Context("when the response body cannot be read", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
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

				_, err := client.Resource("some-resource-guid")

				Expect(err).To(MatchError("a bad read happened"))
			})
		})
	})
})
