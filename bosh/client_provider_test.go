package bosh_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/proxy"
)

var _ = Describe("Client Provider", func() {
	var (
		clientProvider bosh.ClientProvider
		jumpbox        storage.Jumpbox
		noJumpbox      storage.Jumpbox
		socks5Proxy    *fakes.Socks5Proxy
	)
	BeforeEach(func() {
		noJumpbox = storage.Jumpbox{}
		socks5Proxy = &fakes.Socks5Proxy{}
		clientProvider = bosh.NewClientProvider(socks5Proxy)
	})

	Describe("Dialer", func() {
		var (
			socks5Network    string
			socks5Addr       string
			socks5Auth       *proxy.Auth
			socks5Forward    proxy.Dialer
			fakeSocks5Client *fakes.Socks5Client
		)
		BeforeEach(func() {
			socks5Proxy.AddrCall.Returns.Addr = "some-socks-proxy-addr"
			bosh.SetProxySOCKS5(func(network, addr string, auth *proxy.Auth, forward proxy.Dialer) (proxy.Dialer, error) {
				socks5Network = network
				socks5Addr = addr
				socks5Auth = auth
				socks5Forward = forward

				return fakeSocks5Client, nil
			})
		})
		AfterEach(func() {
			bosh.ResetProxySOCKS5()
		})
		Context("when not using a jumpbox", func() {
			var (
				err      error
				listener net.Listener
				addr     string
			)
			BeforeEach(func() {
				listener, err = net.Listen("tcp", ":0")
				Expect(err).NotTo(HaveOccurred())

				_, port, err := net.SplitHostPort(listener.Addr().String())
				Expect(err).NotTo(HaveOccurred())
				addr = fmt.Sprintf(":%s", port)
			})
			AfterEach(func() {
				listener.Close()
			})

			It("does not start a socks5 proxy", func() {
				_, err := clientProvider.Dialer(noJumpbox)
				Expect(err).NotTo(HaveOccurred())
				Expect(socks5Proxy.StartCall.CallCount).To(Equal(0))
				Expect(socks5Proxy.AddrCall.CallCount).To(Equal(0))
			})

			It("returns the default http dialer", func() {
				socks5Client, err := clientProvider.Dialer(noJumpbox)
				Expect(err).NotTo(HaveOccurred())

				_, err = socks5Client.Dial("tcp", addr)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when using a jumpbox", func() {
			BeforeEach(func() {
				jumpbox = storage.Jumpbox{Enabled: true, URL: "https://some-jumpbox", Variables: "jumpbox_ssh: { private_key: some-private-key }"}
				clientProvider = bosh.NewClientProvider(socks5Proxy)
			})

			It("starts the socks 5 proxy to the jumpbox and returns a socks 5 client", func() {
				socks5Client, err := clientProvider.Dialer(jumpbox)
				Expect(err).NotTo(HaveOccurred())
				Expect(socks5Client).To(Equal(fakeSocks5Client))

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(1))
				Expect(socks5Proxy.StartCall.Receives.JumpboxPrivateKey).To(Equal("some-private-key"))
				Expect(socks5Proxy.StartCall.Receives.JumpboxExternalURL).To(Equal("https://some-jumpbox"))

				Expect(socks5Proxy.AddrCall.CallCount).To(Equal(1))

				Expect(socks5Network).To(Equal("tcp"))
				Expect(socks5Addr).To(Equal("some-socks-proxy-addr"))
				Expect(socks5Auth).To(BeNil())
				Expect(socks5Forward).To(Equal(proxy.Direct))
			})

			Context("when the private key does not exist", func() {
				BeforeEach(func() {
					jumpbox = storage.Jumpbox{Enabled: true, URL: "https://some-jumpbox"}
					clientProvider = bosh.NewClientProvider(socks5Proxy)
				})
				It("returns an error", func() {
					_, err := clientProvider.Dialer(jumpbox)
					Expect(err).To(MatchError(ContainSubstring("get jumpbox ssh key: private key not found")))
				})
			})

			Context("when the private key cannot be unmarshaled", func() {
				BeforeEach(func() {
					jumpbox = storage.Jumpbox{Enabled: true, URL: "https://some-jumpbox", Variables: "%%%%"}
					clientProvider = bosh.NewClientProvider(socks5Proxy)
				})
				It("returns an error", func() {
					_, err := clientProvider.Dialer(jumpbox)
					Expect(err).To(MatchError(ContainSubstring("get jumpbox ssh key: yaml")))
				})
			})

			Context("when starting the proxy returns an error", func() {
				BeforeEach(func() {
					socks5Proxy.StartCall.Returns.Error = errors.New("coconut")
				})
				It("returns an error", func() {
					_, err := clientProvider.Dialer(jumpbox)
					Expect(err).To(MatchError(ContainSubstring("start proxy: coconut")))
				})
			})

			Context("when the socks5 client cannot be created", func() {
				BeforeEach(func() {
					bosh.SetProxySOCKS5(func(network, addr string, auth *proxy.Auth, forward proxy.Dialer) (proxy.Dialer, error) {
						return nil, errors.New("banana")
					})
				})
				It("returns an error", func() {
					_, err := clientProvider.Dialer(jumpbox)
					Expect(err).To(MatchError(ContainSubstring("create socks5 client: banana")))
				})
			})
		})
	})

	Describe("HttpClient", func() {
		var (
			err    error
			ca     []byte
			dialer *fakes.Socks5Client
		)

		BeforeEach(func() {
			ca, err = ioutil.ReadFile("fixtures/some-fake-ca.crt")
			Expect(err).NotTo(HaveOccurred())

			clientProvider = bosh.NewClientProvider(socks5Proxy)
			dialer = &fakes.Socks5Client{}
		})

		It("returns an http client that uses the dialer", func() {
			httpClient := clientProvider.HTTPClient(dialer, ca)

			_, err := httpClient.Transport.(*http.Transport).Dial("some-network", "some-addr")
			Expect(dialer.DialCall.CallCount).To(Equal(1))
			Expect(dialer.DialCall.Receives.Network).To(Equal("some-network"))
			Expect(dialer.DialCall.Receives.Addr).To(Equal("some-addr"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("adds the CA to the http client's certs pool", func() {
			httpClient := clientProvider.HTTPClient(dialer, ca)

			certsPool := httpClient.Transport.(*http.Transport).TLSClientConfig.RootCAs
			Expect(certsPool).NotTo(BeNil())
			Expect(certsPool.Subjects()).NotTo(BeEmpty())
			Expect(string(certsPool.Subjects()[0])).To(ContainSubstring("some-fake-ca"))
		})
	})
})
