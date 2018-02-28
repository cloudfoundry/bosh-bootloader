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
	"time"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		tlsConfig   *tls.Config
		fakeBOSH    *httptest.Server
		ca          []byte
		cloudConfig []byte
		token       string
		httpClient  *http.Client
		failStatus  int
	)

	BeforeEach(func() {
		bosh.MAX_RETRIES = 1
		bosh.RETRY_DELAY = 1 * time.Millisecond

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

				token = req.Header.Get("Authorization")

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

		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial:            (&net.Dialer{}).Dial,
				TLSClientConfig: tlsConfig,
			},
		}
	})

	AfterEach(func() {
		failStatus = 0
	})

	Describe("Info", func() {
		It("returns the director info", func() {
			fakeBOSH.StartTLS()

			client := bosh.NewClient(httpClient, fakeBOSH.URL, "some-username", "some-password", string(ca))
			info, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(Equal(bosh.Info{
				Name:    "some-bosh-director",
				UUID:    "some-uuid",
				Version: "some-version",
			}))
		})

		Context("failure cases", func() {
			Context("when the response is not StatusOK", func() {
				BeforeEach(func() {
					failStatus = http.StatusNotFound
				})

				It("returns an error", func() {
					fakeBOSH.StartTLS()
					client := bosh.NewClient(httpClient, fakeBOSH.URL, "some-username", "some-password", string(ca))
					_, err := client.Info()
					Expect(err).To(MatchError("unexpected http response 404 Not Found"))
				})
			})

			Context("when the url cannot be parsed", func() {
				It("returns an error", func() {
					fakeBOSH.StartTLS()

					client := bosh.NewClient(httpClient, "%%%", "some-username", "some-password", "some-false")
					_, err := client.Info()
					Expect(err.(*url.Error).Op).To(Equal("parse"))
				})
			})

			Context("when the request fails", func() {
				It("returns an error", func() {
					fakeBOSH.StartTLS()

					client := bosh.NewClient(httpClient, "fake://some-url", "some-username", "some-password", string(ca))
					_, err := client.Info()
					Expect(err).To(MatchError("made 1 attempts, last error: Get fake://some-url/info: unsupported protocol scheme \"fake\""))
				})
			})

			Context("when it cannot parse info json", func() {
				It("returns an error", func() {
					failStatus = http.StatusOK

					fakeBOSH.StartTLS()
					client := bosh.NewClient(httpClient, fakeBOSH.URL, "some-username", "some-password", string(ca))
					_, err := client.Info()
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})

	Describe("UpdateCloudConfig", func() {
		Context("when a jumpbox is enabled", func() {
			It("uses UAA to get a token in order to upload the cloud-config", func() {
				dialer := &fakes.Socks5Client{}
				dialer.DialCall.Stub = func(network, addr string) (net.Conn, error) {
					u, _ := url.Parse(fakeBOSH.URL)
					return net.Dial(network, u.Host)
				}

				httpClient = &http.Client{
					Transport: &http.Transport{
						Dial:            dialer.Dial,
						TLSClientConfig: tlsConfig,
					},
				}

				fakeBOSH.StartTLS()

				client := bosh.NewClient(httpClient, fakeBOSH.URL, "some-username", "some-password", string(ca))

				err := client.UpdateCloudConfig([]byte("cloud: config"))
				Expect(err).NotTo(HaveOccurred())

				Expect(token).To(Equal("Bearer some-uaa-token"))
				Expect(cloudConfig).To(Equal([]byte("cloud: config")))
			})

			Context("when an error occurs", func() {
				Context("when a non-201 occurs", func() {
					It("returns an error", func() {
						fakeBOSH.StartTLS()

						client := bosh.NewClient(httpClient, fakeBOSH.URL, "", "", string(ca))

						err := client.UpdateCloudConfig([]byte("cloud: config"))
						Expect(err).To(MatchError(ContainSubstring("made 1 attempts, last error: Post")))
						Expect(err).To(MatchError(ContainSubstring("connection refused")))
					})
				})
			})
		})
	})
})
