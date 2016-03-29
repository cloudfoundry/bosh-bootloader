package bosh_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
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
