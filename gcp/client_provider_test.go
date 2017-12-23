package gcp_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2/jwt"
)

var _ = Describe("NewClient", func() {
	var (
		basePath              string
		serviceAccountKey     string
		serviceAccountKeyPath string
		fakeFileIO            *fakes.FileIO
	)

	BeforeEach(func() {
		serviceAccountKeyPath = "/some/path"
		fakeFileIO = &fakes.FileIO{}
		privateKeyContents, err := ioutil.ReadFile("fixtures/service-account-key")
		Expect(err).NotTo(HaveOccurred())
		serviceAccountKey = fmt.Sprintf(`{
				"type": "service_account",
				"private_key": %q
			}`, string(privateKeyContents))
		fakeFileIO.ReadFileCall.Returns.Contents = []byte(serviceAccountKey)

		gcp.SetGCPHTTPClient(func(*jwt.Config) *http.Client {
			return &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
		})

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/proj-id/regions/some-region":
				w.Write([]byte(`{}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))

		basePath = server.URL

	})

	AfterEach(func() {
		gcp.ResetGCPHTTPClient()
	})

	It("works", func() {
		_, err := gcp.NewClient(storage.GCP{
			ServiceAccountKey: serviceAccountKeyPath,
			ProjectID:         "proj-id",
			Region:            "some-region",
			Zone:              "some-zone",
		}, basePath, fakeFileIO)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the service account key is not valid json", func() {
		BeforeEach(func() {
			fakeFileIO.ReadFileCall.Returns.Contents = []byte("%%")
		})
		It("returns an error", func() {
			_, err := gcp.NewClient(storage.GCP{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "proj-id",
				Region:            "some-region",
				Zone:              "some-zone",
			}, basePath, fakeFileIO)
			Expect(err).To(MatchError("parse service account key: invalid character '%' looking for beginning of value"))
		})
	})

	Context("when a service could not be created", func() {
		BeforeEach(func() {
			gcp.SetGCPHTTPClient(func(*jwt.Config) *http.Client {
				return nil
			})
		})

		It("returns an error", func() {
			_, err := gcp.NewClient(storage.GCP{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "proj-id",
				Region:            "some-region",
				Zone:              "some-zone",
			}, basePath, fakeFileIO)
			Expect(err).To(MatchError("create gcp client: client is nil"))
		})
	})

	Context("when the region is invalid", func() {
		It("returns an error", func() {
			_, err := gcp.NewClient(storage.GCP{
				ServiceAccountKey: serviceAccountKeyPath,
				ProjectID:         "proj-id",
				Region:            "bad-region",
				Zone:              "some-zone",
			}, basePath, fakeFileIO)
			Expect(err).To(MatchError(ContainSubstring("get region: ")))
			Expect(err).To(MatchError(ContainSubstring("googleapi")))
			Expect(err).To(MatchError(ContainSubstring("404")))
		})
	})
})
