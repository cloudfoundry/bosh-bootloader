package bosh_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe("Info", func() {
		It("returns the director info", func() {
			fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Write([]byte(`{
					"name": "some-bosh-director",
					"uuid": "some-uuid",
					"version": "some-version"
				}`))
			}))

			client := bosh.NewClient(fakeBOSH.URL, "some-username", "some-password")
			info, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(Equal(bosh.Info{
				Name:    "some-bosh-director",
				UUID:    "some-uuid",
				Version: "some-version",
			}))
		})

		Context("failure cases", func() {
			It("returns an error when the response is not StatusOK", func() {
				fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusNotFound)
				}))

				client := bosh.NewClient(fakeBOSH.URL, "some-username", "some-password")
				_, err := client.Info()
				Expect(err).To(MatchError("unexpected http response 404 Not Found"))
			})

			It("returns an error when the url cannot be parsed", func() {
				client := bosh.NewClient("%%%", "some-username", "some-password")
				_, err := client.Info()
				Expect(err.(*url.Error).Op).To(Equal("parse"))
			})

			It("returns an error when the request fails", func() {
				client := bosh.NewClient("fake://some-url", "some-username", "some-password")
				_, err := client.Info()
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
			})

			It("returns an error when it cannot parse info json", func() {
				fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.Write([]byte(`%%%`))
				}))

				client := bosh.NewClient(fakeBOSH.URL, "some-username", "some-password")
				_, err := client.Info()
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})

		})

	})

	Describe("UpdateCloudConfig", func() {
		It("uploads the given cloud config", func() {
			var (
				cloudConfig []byte
				contentType string
				username    string
				password    string
			)

			fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				var (
					err error
				)

				username, password, _ = request.BasicAuth()
				contentType = request.Header.Get("Content-Type")

				cloudConfig, err = ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())

				responseWriter.WriteHeader(http.StatusCreated)
			}))

			client := bosh.NewClient(fakeBOSH.URL, "some-username", "some-password")

			err := client.UpdateCloudConfig([]byte("cloud: config"))
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudConfig).To(Equal([]byte("cloud: config")))
			Expect(contentType).To(Equal("text/yaml"))
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))
		})

		Context("failure cases", func() {
			It("returns an error when the status code is not StatusCreated", func() {
				fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusInternalServerError)
				}))

				client := bosh.NewClient(fakeBOSH.URL, "", "")

				err := client.UpdateCloudConfig([]byte("cloud: config"))
				Expect(err).To(MatchError("unexpected http response 500 Internal Server Error"))
			})

			It("returns an error when the director address is malformed", func() {
				client := bosh.NewClient("%%%%%%%%%%%%%%%", "", "")

				err := client.UpdateCloudConfig([]byte("cloud: config"))
				Expect(err.(*url.Error).Op).To(Equal("parse"))
			})

			It("returns an error when the director address is malformed", func() {
				fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.WriteHeader(http.StatusInternalServerError)
				}))

				client := bosh.NewClient(fakeBOSH.URL, "", "")

				fakeBOSH.Close()

				err := client.UpdateCloudConfig([]byte("cloud: config"))
				Expect(err).To(MatchError(ContainSubstring("connection refused")))
			})
		})
	})
})
