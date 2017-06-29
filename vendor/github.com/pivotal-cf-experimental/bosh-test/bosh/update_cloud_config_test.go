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

var _ = Describe("UpdateCloudConfig", func() {
	It("updates cloud config", func() {
		testServerCallCount := 0
		cloudConfig := "some-cloud-config-yaml"

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			testServerCallCount++
			Expect(r.Method).To(Equal("POST"))
			Expect(r.URL.Path).To(Equal("/cloud_configs"))
			Expect(r.Header.Get("Content-Type")).To(Equal("text/yaml"))

			username, password, ok := r.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))

			rawBody, err := ioutil.ReadAll(r.Body)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(rawBody)).To(Equal(cloudConfig))

			w.WriteHeader(http.StatusCreated)
		}))

		client := bosh.NewClient(bosh.Config{
			URL:      testServer.URL,
			Username: "some-username",
			Password: "some-password",
		})

		err := client.UpdateCloudConfig([]byte(cloudConfig))

		Expect(err).NotTo(HaveOccurred())
		Expect(testServerCallCount).To(Equal(1))
	})

	Context("failure cases", func() {
		It("returns an error when request creation fails", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "%%%%%",
			})

			err := client.UpdateCloudConfig([]byte(""))

			Expect(err).To(MatchError(`parse %%%%%/cloud_configs: invalid URL escape "%%%"`))
		})

		It("returns an error when request fails", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "",
			})

			err := client.UpdateCloudConfig([]byte(""))

			Expect(err).To(MatchError(`unsupported protocol scheme ""`))
		})

		It("returns an error on a bogus response body", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/cloud_configs"))

				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("More info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL: testServer.URL,
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("a bad read happened")
			})

			err := client.UpdateCloudConfig([]byte(""))
			Expect(err).To(MatchError("a bad read happened"))
		})

		It("errors on an unexpected status code with body", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/cloud_configs"))

				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("More info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL: testServer.URL,
			})

			err := client.UpdateCloudConfig([]byte(""))
			Expect(err).To(MatchError("unexpected response 418 I'm a teapot:\nMore info"))
		})
	})
})
