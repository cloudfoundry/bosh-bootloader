package bosh_test

import (
	"errors"
	"io/ioutil"
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
		allProxyGetter *fakes.AllProxyGetter
		socks5Proxy    *fakes.Socks5Proxy
		sshKeyGetter   *fakes.SSHKeyGetter
	)

	BeforeEach(func() {
		allProxyGetter = &fakes.AllProxyGetter{}
		socks5Proxy = &fakes.Socks5Proxy{}
		sshKeyGetter = &fakes.SSHKeyGetter{}
		sshKeyGetter.GetCall.Returns.PrivateKey = "some-private-key"

		clientProvider = bosh.NewClientProvider(allProxyGetter, socks5Proxy, sshKeyGetter, "some-path-to-bosh")
	})

	Describe("Dialer", func() {
		var (
			socks5Network string
			socks5Addr    string
			socks5Auth    *proxy.Auth
			socks5Forward proxy.Dialer
			fakeDialer    *fakes.Dialer
		)

		BeforeEach(func() {
			socks5Proxy.AddrCall.Returns.Addr = "some-socks-proxy-addr"
			bosh.SetProxySOCKS5(func(network, addr string, auth *proxy.Auth, forward proxy.Dialer) (proxy.Dialer, error) {
				socks5Network = network
				socks5Addr = addr
				socks5Auth = auth
				socks5Forward = forward

				return fakeDialer, nil
			})
		})

		AfterEach(func() {
			bosh.ResetProxySOCKS5()
		})

		Context("when using a jumpbox", func() {
			It("starts the socks 5 proxy to the jumpbox and returns a socks 5 client", func() {
				proxyDialer, err := clientProvider.Dialer(storage.Jumpbox{URL: "https://some-jumpbox"})
				Expect(err).NotTo(HaveOccurred())
				Expect(proxyDialer).To(Equal(fakeDialer))

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(1))
				Expect(socks5Proxy.StartCall.Receives.Username).To(Equal(""))
				Expect(socks5Proxy.StartCall.Receives.PrivateKey).To(Equal("some-private-key"))
				Expect(socks5Proxy.StartCall.Receives.ExternalURL).To(Equal("https://some-jumpbox"))

				Expect(socks5Proxy.AddrCall.CallCount).To(Equal(1))

				Expect(socks5Network).To(Equal("tcp"))
				Expect(socks5Addr).To(Equal("some-socks-proxy-addr"))
				Expect(socks5Auth).To(BeNil())
				Expect(socks5Forward).To(Equal(proxy.Direct))
			})

			Context("when retrieving the private key fails", func() {
				BeforeEach(func() {
					sshKeyGetter.GetCall.Returns.Error = errors.New("tamarind")
				})
				It("returns an error", func() {
					_, err := clientProvider.Dialer(jumpbox)
					Expect(err).To(MatchError("get jumpbox ssh key: tamarind"))
					Expect(sshKeyGetter.GetCall.Receives.Deployment).To(Equal("jumpbox"))
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

			Context("when getting the proxy address returns an error", func() {
				BeforeEach(func() {
					socks5Proxy.AddrCall.Returns.Error = errors.New("mango")
				})
				It("returns an error", func() {
					_, err := clientProvider.Dialer(jumpbox)
					Expect(err).To(MatchError(ContainSubstring("get proxy address: mango")))
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
			ca     []byte
			dialer *fakes.Dialer
		)

		BeforeEach(func() {
			var err error
			ca, err = ioutil.ReadFile("fixtures/some-fake-ca.crt")
			Expect(err).NotTo(HaveOccurred())
			sshKeyGetter := &fakes.SSHKeyGetter{}

			clientProvider = bosh.NewClientProvider(allProxyGetter, socks5Proxy, sshKeyGetter, "some-path-to-bosh")
			dialer = &fakes.Dialer{}
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

	Describe("Client", func() {
		It("returns a well-formed bosh client", func() {
			client, err := clientProvider.Client(storage.Jumpbox{URL: "https://some-jumpbox"}, "https://director:9999", "user", "pass", "some-fake-ca")
			Expect(err).NotTo(HaveOccurred())

			structClient := client.(bosh.Client)
			Expect(structClient.DirectorAddress).To(Equal("https://director:9999"))
			Expect(structClient.UAAAddress).To(Equal("https://director:8443"))
		})

		Context("given a director address without a port", func() {
			It("Errors", func() {
				_, err := clientProvider.Client(storage.Jumpbox{URL: "https://some-jumpbox"}, "https://director", "user", "pass", "some-fake-ca")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("BoshCLI", func() {
		It("returns an authenticated bosh cli", func() {
			allProxyGetter.BoshAllProxyCall.Returns.URL = "some-all-proxy-url"
			cli, err := clientProvider.BoshCLI(storage.Jumpbox{URL: "https://some-jumpbox"}, nil, "some-address", "some-username", "some-password", "some-fake-ca")
			Expect(err).NotTo(HaveOccurred())

			boshCLI := cli.(bosh.BOSHCLI)
			Expect(boshCLI.GlobalArgs).To(Equal([]string{
				"--environment", "some-address",
				"--client", "some-username",
				"--client-secret", "some-password",
				"--ca-cert", "some-fake-ca",
				"--non-interactive",
			}))
			Expect(boshCLI.BOSHAllProxy).To(Equal("some-all-proxy-url"))
		})

		Context("when it can not get the correct key", func() {
			It("Errors", func() {
				allProxyGetter.GeneratePrivateKeyCall.Returns.Error = errors.New("fruit")
				_, err := clientProvider.BoshCLI(storage.Jumpbox{URL: "https://some-jumpbox"}, nil, "some-address", "some-username", "some-password", "some-fake-ca")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("fruit"))
			})
		})
	})
})
