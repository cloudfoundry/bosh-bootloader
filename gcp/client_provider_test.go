package gcp_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry/bosh-bootloader/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2/jwt"
)

var _ = Describe("ClientProvider", func() {
	var (
		clientProvider *gcp.ClientProvider
		privateKey     string
	)

	BeforeEach(func() {
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
			case "/proj-id/zones/zone":
				w.Write([]byte(`{}`))
			case "/proj-id/regions/region":
				w.Write([]byte(`{}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))

		clientProvider = gcp.NewClientProvider(server.URL)

		privateKeyContents, err := ioutil.ReadFile("fixtures/service-account-key")
		Expect(err).NotTo(HaveOccurred())
		privateKey = string(privateKeyContents)
	})

	Describe("SetConfig", func() {
		AfterEach(func() {
			gcp.ResetGCPHTTPClient()
		})

		It("returns an error when the service account key is not valid json", func() {
			err := clientProvider.SetConfig("1231:123", "proj-id", "region", "zone")
			Expect(err).To(MatchError("invalid character ':' after top-level value"))
		})

		It("returns an error when a service could not be created", func() {
			gcp.SetGCPHTTPClient(func(*jwt.Config) *http.Client {
				return nil
			})
			err := clientProvider.SetConfig(`{"type": "service_account"}`, "proj-id", "region", "zone")
			Expect(err).To(MatchError("client is nil"))
		})

		It("returns an error when the zone is invalid", func() {
			serviceAccountKey := fmt.Sprintf(`{
				"type": "service_account",
				"private_key": %q
			}`, privateKey)

			err := clientProvider.SetConfig(serviceAccountKey, "proj-id", "region", "bad-zone")
			Expect(err).To(MatchError(ContainSubstring("googleapi")))
			Expect(err).To(MatchError(ContainSubstring("404")))
		})

		It("returns an error when the region is invalid", func() {
			serviceAccountKey := fmt.Sprintf(`{
				"type": "service_account",
				"private_key": %q
			}`, privateKey)
			err := clientProvider.SetConfig(serviceAccountKey, "proj-id", "bad-region", "zone")
			Expect(err).To(MatchError(ContainSubstring("googleapi")))
			Expect(err).To(MatchError(ContainSubstring("404")))
		})
	})
})
