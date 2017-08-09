package bosh_test

import (
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
	Describe("ConfigureHttpClient", func() {
		var (
			socks5Client *fakes.Socks5Client
			fakeBOSH     *httptest.Server
		)

		BeforeEach(func() {
			socks5Client = &fakes.Socks5Client{}
			socks5Client.DialCall.Stub = func(network, addr string) (net.Conn, error) {
				return net.Dial(network, addr)
			}

			fakeBOSH = httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Write([]byte(`{
					"name": "some-bosh-director",
					"uuid": "some-uuid",
					"version": "some-version"
				}`))
			}))
		})

		It("configures the http client to use the socks5 proxy", func() {
			client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", "some-fake-ca")
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
			fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Write([]byte(`{
					"name": "some-bosh-director",
					"uuid": "some-uuid",
					"version": "some-version"
				}`))
			}))

			client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", "some-fake-ca")
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

				client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", "some-fake-ca")
				_, err := client.Info()
				Expect(err).To(MatchError("unexpected http response 404 Not Found"))
			})

			It("returns an error when the url cannot be parsed", func() {
				client := bosh.NewClient(false, "%%%", "some-username", "some-password", "some-false")
				_, err := client.Info()
				Expect(err.(*url.Error).Op).To(Equal("parse"))
			})

			It("returns an error when the request fails", func() {
				client := bosh.NewClient(false, "fake://some-url", "some-username", "some-password", "some-fake-ca")
				_, err := client.Info()
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
			})

			It("returns an error when it cannot parse info json", func() {
				fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					responseWriter.Write([]byte(`%%%`))
				}))

				client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", "some-fake-ca")
				_, err := client.Info()
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})

	Describe("UpdateCloudConfig", func() {
		Context("when a jumpbox is enabled", func() {
			It("uploads the cloud-config", func() {
				var (
					cloudConfig []byte
					token       string
					username    string
					password    string
				)

				fakeUAABOSH := httptest.NewTLSServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					switch req.URL.Path {
					case "/oauth/token":
						var ok bool
						username, password, ok = req.BasicAuth()
						Expect(ok).To(BeTrue())

						resp.Header().Set("Content-Type", "application/json")

						resp.Write([]byte(`{
				          "access_token": "some-uaa-token",
				          "token_type": "bearer",
				          "expires_in": 3600
		                }`))
					case "/cloud_configs":
						token = req.Header.Get("Authorization")

						resp.WriteHeader(http.StatusCreated)

						var err error
						cloudConfig, err = ioutil.ReadAll(req.Body)
						Expect(err).NotTo(HaveOccurred())
					default:
						dump, err := httputil.DumpRequest(req, true)
						Expect(err).NotTo(HaveOccurred())
						Fail(fmt.Sprintf("received unknown request: %s\n", string(dump)))
					}
				}))

				socks5Client := &fakes.Socks5Client{}
				socks5Client.DialCall.Stub = func(network, addr string) (net.Conn, error) {
					u, _ := url.Parse(fakeUAABOSH.URL)
					return net.Dial(network, u.Host)
				}

				client := bosh.NewClient(true, fakeUAABOSH.URL, "some-username", "some-password", "some-fake-ca")
				client.ConfigureHTTPClient(socks5Client)

				err := client.UpdateCloudConfig([]byte("cloud: config"))
				Expect(err).NotTo(HaveOccurred())

				Expect(token).To(Equal("Bearer some-uaa-token"))
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))
				Expect(cloudConfig).To(Equal([]byte("cloud: config")))
			})

			Context("when an error occurs", func() {
				Context("when a non-201 occurs", func() {
					It("returns an error ", func() {
						fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
							responseWriter.WriteHeader(http.StatusInternalServerError)
						}))

						client := bosh.NewClient(true, fakeBOSH.URL, "", "", "")

						err := client.UpdateCloudConfig([]byte("cloud: config"))
						Expect(err).To(MatchError(ContainSubstring("connection refused")))
					})
				})
			})
		})

		Context("when a jumpbox is not enabled", func() {
			It("uploads the cloud-config", func() {
				var (
					cloudConfig []byte
					contentType string
					username    string
					password    string
				)

				fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					var err error

					username, password, _ = request.BasicAuth()
					contentType = request.Header.Get("Content-Type")

					cloudConfig, err = ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())

					responseWriter.WriteHeader(http.StatusCreated)
				}))

				client := bosh.NewClient(false, fakeBOSH.URL, "some-username", "some-password", "some-fake-ca")

				err := client.UpdateCloudConfig([]byte("cloud: config"))
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfig).To(Equal([]byte("cloud: config")))
				Expect(contentType).To(Equal("text/yaml"))
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))
			})

			Context("when an error occurs", func() {
				Context("when a non-201 occurs", func() {
					It("returns an error ", func() {
						fakeBOSH := httptest.NewTLSServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
							responseWriter.WriteHeader(http.StatusInternalServerError)
						}))

						client := bosh.NewClient(false, fakeBOSH.URL, "", "", "")

						err := client.UpdateCloudConfig([]byte("cloud: config"))
						Expect(err).To(MatchError("unexpected http response 500 Internal Server Error"))
					})
				})

				Context("when the director address is malformed", func() {
					It("returns an error", func() {
						client := bosh.NewClient(false, "%%%%%%%%%%%%%%%", "", "", "")

						err := client.UpdateCloudConfig([]byte("cloud: config"))
						Expect(err.(*url.Error).Op).To(Equal("parse"))
					})
				})
			})
		})
	})
})
