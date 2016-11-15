package gcp_test

import (
	"net/http"

	"github.com/cloudfoundry/bosh-bootloader/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2/jwt"
)

var _ = Describe("ClientProvider", func() {
	var (
		clientProvider *gcp.ClientProvider
	)

	BeforeEach(func() {
		clientProvider = gcp.NewClientProvider("http://example.com")
	})

	Describe("SetConfig", func() {
		AfterEach(func() {
			gcp.ResetGCPHTTPClient()
		})

		It("returns an error when the service account key is not valid json", func() {
			err := clientProvider.SetConfig("1231:123")
			Expect(err).To(MatchError("invalid character ':' after top-level value"))
		})

		It("returns an error when a service could not be created", func() {
			gcp.SetGCPHTTPClient(func(*jwt.Config) *http.Client {
				return nil
			})
			err := clientProvider.SetConfig(`{"type": "service_account"}`)
			Expect(err).To(MatchError("client is nil"))
		})
	})
})
