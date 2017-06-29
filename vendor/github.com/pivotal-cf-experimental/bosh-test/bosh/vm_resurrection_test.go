package bosh_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetVMResurrection", func() {
	DescribeTable("sets the resurrection state of an instance in a job",
		func(enable bool) {
			var body []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				username, password, _ := req.BasicAuth()

				if username == "some-username" && password == "some-password" {
					if req.URL.Path == "/deployments/some-deployment-name/jobs/some-job-name/1/resurrection" {
						if req.Method == "PUT" {
							var err error
							body, err = ioutil.ReadAll(req.Body)
							if err != nil {
								w.WriteHeader(http.StatusInternalServerError)
								return
							}
							w.WriteHeader(http.StatusOK)
							return
						}
					}
				}

				w.WriteHeader(http.StatusTeapot)
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			err := client.SetVMResurrection("some-deployment-name", "some-job-name", 1, enable)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(MatchJSON(fmt.Sprintf(`{"resurrection_paused":%t}`, !enable)))
		},
		Entry("enable resurrection", true),
		Entry("disableresurrection", false),
	)

	Context("failure cases", func() {
		It("returns an error when the request cannot be created", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "%%%%%",
			})

			err := client.SetVMResurrection("", "", 0, true)
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("returns an error when the request fails to be made", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "",
			})

			err := client.SetVMResurrection("", "", 0, true)
			Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
		})

		It("returns an error with the body when the request returns an unexpected status code", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("something bad happened"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			err := client.SetVMResurrection("", "", 0, true)
			Expect(err).To(MatchError("unexpected response 418 I'm a teapot: something bad happened"))
		})

		It("returns an error when the body fails to read", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte(""))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("failed to read")
			})

			err := client.SetVMResurrection("", "", 0, true)
			Expect(err).To(MatchError("failed to read"))
		})
	})
})
