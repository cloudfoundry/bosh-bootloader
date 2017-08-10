package bosh_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		tlsConfig              *tls.Config
		fakeBOSH               *httptest.Server
		ca                     []byte
		cloudConfig            []byte
		token                  string
		username               string
		password               string
		cloudConfigContentType string
		failStatus             int
	)

	BeforeEach(func() {
		var err error
		ca, err = ioutil.ReadFile("fixtures/some-fake-ca.crt")
		Expect(err).NotTo(HaveOccurred())

		pool := x509.NewCertPool()
		ok := pool.AppendCertsFromPEM(ca)
		Expect(ok).To(BeTrue())

		clientCert, err := ioutil.ReadFile("fixtures/some-cert.crt")
		Expect(err).NotTo(HaveOccurred())

		clientKey, err := ioutil.ReadFile("fixtures/some-cert.key")
		Expect(err).NotTo(HaveOccurred())

		cert, err := tls.X509KeyPair(clientCert, clientKey)
		Expect(err).NotTo(HaveOccurred())

		fakeBOSH = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/oauth/token":
				username, password, _ = req.BasicAuth()

				w.Header().Set("Content-Type", "application/json")

				w.Write([]byte(`{
				          "access_token": "some-uaa-token",
				          "token_type": "bearer",
				          "expires_in": 3600
		                }`))
			case "/info":
				if failStatus != 0 {
					w.WriteHeader(failStatus)
					w.Write([]byte("%%%%%%%%%%%%%%%%"))
					return
				}

				w.Write([]byte(`{
				          "name": "some-bosh-director",
				          "uuid": "some-uuid",
				          "version": "some-version"
		                }`))
			case "/cloud_configs":
				if failStatus != 0 {
					w.WriteHeader(failStatus)
					return
				}

				username, password, _ = req.BasicAuth()

				token = req.Header.Get("Authorization")
				cloudConfigContentType = req.Header.Get("Content-Type")

				w.WriteHeader(http.StatusCreated)

				var err error
				cloudConfig, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
			default:
				dump, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("received unknown request: %s\n", string(dump)))
			}
		}))

		tlsConfig = &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cert},
		}

		fakeBOSH.TLS = tlsConfig
	})

	AfterEach(func() {
		failStatus = 0
	})

	Describe("ConfigureHttpClient", func() {
		It("configures the http client to use the socks5 proxy", func() {
			socks5Client := &fakes.Socks5Client{}
			socks5Client.DialCall.Stub = func(network, addr string) (net.Conn, error) {
				u, _ := url.Parse(fakeBOSH.URL)
				return net.Dial(network, u.Host)
			}

			fakeBOSH.StartTLS()

			client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", string(ca))
			client.ConfigureHTTPClient(socks5Client)
			info, err := client.Info()

			Expect(socks5Client.DialCall.CallCount).To(Equal(1))
			Expect(socks5Client.DialCall.Receives.Network).To(Equal("tcp"))
			Expect(socks5Client.DialCall.Receives.Addr).To(Equal(fakeBOSH.Listener.Addr().String()))
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(Equal(bosh.Info{
				Name:    "some-bosh-director",
				UUID:    "some-uuid",
				Version: "some-version",
			}))
		})
	})

	Describe("Info", func() {
		It("returns the director info", func() {
			fakeBOSH.StartTLS()

			client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", string(ca))
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
				failStatus = http.StatusNotFound

				fakeBOSH.StartTLS()

				client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", string(ca))
				_, err := client.Info()
				Expect(err).To(MatchError("unexpected http response 404 Not Found"))
			})

			It("returns an error when the url cannot be parsed", func() {
				fakeBOSH.StartTLS()

				client := bosh.NewClient(false, "%%%", "some-username", "some-password", "some-false")
				_, err := client.Info()
				Expect(err.(*url.Error).Op).To(Equal("parse"))
			})

			It("returns an error when the request fails", func() {
				fakeBOSH.StartTLS()

				client := bosh.NewClient(false, "fake://some-url", "some-username", "some-password", string(ca))
				_, err := client.Info()
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
			})

			It("returns an error when it cannot parse info json", func() {
				failStatus = http.StatusOK

				fakeBOSH.StartTLS()
				client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", string(ca))
				_, err := client.Info()
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})

	Describe("UpdateCloudConfig", func() {
		Context("when a jumpbox is enabled", func() {
			It("uploads the cloud-config", func() {
				socks5Client := &fakes.Socks5Client{}
				socks5Client.DialCall.Stub = func(network, addr string) (net.Conn, error) {
					u, _ := url.Parse(fakeBOSH.URL)
					return net.Dial(network, u.Host)
				}

				fakeBOSH.StartTLS()

				client := bosh.NewClient(true, fakeBOSH.URL, "some-username", "some-password", string(ca))
				client.ConfigureHTTPClient(socks5Client)

				err := client.UpdateCloudConfig([]byte("cloud: config"))
				Expect(err).NotTo(HaveOccurred())

				Expect(token).To(Equal("Bearer some-uaa-token"))
				Expect(cloudConfig).To(Equal([]byte("cloud: config")))
			})

			Context("when an error occurs", func() {
				Context("when a non-201 occurs", func() {
					It("returns an error ", func() {
						fakeBOSH.StartTLS()

						client := bosh.NewClient(true, fakeBOSH.URL, "", "", string(ca))

						err := client.UpdateCloudConfig([]byte("cloud: config"))
						Expect(err).To(MatchError(ContainSubstring("connection refused")))
					})
				})
			})
		})

		Context("when a jumpbox is not enabled", func() {
			It("uploads the cloud-config", func() {
				fakeBOSH.StartTLS()

				client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", string(ca))

				err := client.UpdateCloudConfig([]byte("cloud: config"))
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfig).To(Equal([]byte("cloud: config")))
				Expect(cloudConfigContentType).To(Equal("text/yaml"))
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))
			})

			Context("when an error occurs", func() {
				Context("when a non-201 occurs", func() {
					It("returns an error ", func() {
						failStatus = http.StatusInternalServerError
						fakeBOSH.StartTLS()

						client := bosh.NewClient(false, fakeBOSH.URL, "", "", string(ca))

						err := client.UpdateCloudConfig([]byte("cloud: config"))
						Expect(err).To(MatchError("unexpected http response 500 Internal Server Error"))
					})
				})

				Context("when the director address is malformed", func() {
					It("returns an error", func() {
						fakeBOSH.StartTLS()

						client := bosh.NewClient(false, "%%%%%%%%%%%%%%%", "", "", "")

						err := client.UpdateCloudConfig([]byte("cloud: config"))
						Expect(err.(*url.Error).Op).To(Equal("parse"))
					})
				})
			})
		})
	})
})
