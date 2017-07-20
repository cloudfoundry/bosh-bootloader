package proxy_test

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/proxy"

	"golang.org/x/crypto/ssh"
	goproxy "golang.org/x/net/proxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Socks5Proxy", func() {
	Describe("Start", func() {
		var (
			socks5Proxy   *proxy.Socks5Proxy
			hostKeyGetter *fakes.HostKeyGetter
			logger        *fakes.Logger

			sshServerURL       string
			httpServerHostPort string
			httpServer         *httptest.Server
		)

		BeforeEach(func() {
			httpServer = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			httpServerHostPort = strings.Split(httpServer.URL, "http://")[1]

			sshServerURL = startSSHServer(httpServerHostPort)

			signer, err := ssh.ParsePrivateKey([]byte(sshPrivateKey))
			Expect(err).NotTo(HaveOccurred())

			hostKeyGetter = &fakes.HostKeyGetter{}
			hostKeyGetter.GetCall.Returns.HostKey = signer.PublicKey()

			logger = &fakes.Logger{}
			socks5Proxy = proxy.NewSocks5Proxy(logger, hostKeyGetter, 0)
		})

		AfterEach(func() {
			proxy.ResetNetListen()
		})

		It("starts a proxy to the jumpbox", func() {
			err := socks5Proxy.Start(sshPrivateKey, sshServerURL)
			Expect(err).NotTo(HaveOccurred())

			// Wait for socks5 proxy to start
			time.Sleep(1 * time.Second)

			socks5Addr := socks5Proxy.Addr()
			socks5Client, err := goproxy.SOCKS5("tcp", socks5Addr, nil, goproxy.Direct)
			Expect(err).NotTo(HaveOccurred())

			Expect(hostKeyGetter.GetCall.CallCount).To(Equal(1))
			Expect(hostKeyGetter.GetCall.Receives.PrivateKey).To(Equal(sshPrivateKey))
			Expect(hostKeyGetter.GetCall.Receives.ServerURL).To(Equal(sshServerURL))

			conn, err := socks5Client.Dial("tcp", httpServerHostPort)
			Expect(err).NotTo(HaveOccurred())

			_, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			status, err := bufio.NewReader(conn).ReadString('\n')
			Expect(status).To(Equal("HTTP/1.0 200 OK\r\n"))
		})

		Context("when starting the proxy a second time", func() {
			It("no-ops on the second run", func() {
				err := socks5Proxy.Start(sshPrivateKey, sshServerURL)
				Expect(err).NotTo(HaveOccurred())

				// Wait for socks5 proxy to start
				time.Sleep(1 * time.Second)

				err = socks5Proxy.Start(sshPrivateKey, sshServerURL)
				Expect(err).NotTo(HaveOccurred())

				socks5Addr := socks5Proxy.Addr()
				socks5Client, err := goproxy.SOCKS5("tcp", socks5Addr, nil, goproxy.Direct)
				Expect(err).NotTo(HaveOccurred())

				conn, err := socks5Client.Dial("tcp", httpServerHostPort)
				Expect(err).NotTo(HaveOccurred())

				_, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
				Expect(err).NotTo(HaveOccurred())
				defer conn.Close()

				status, err := bufio.NewReader(conn).ReadString('\n')
				Expect(status).To(Equal("HTTP/1.0 200 OK\r\n"))
			})
		})

		Context("failure cases", func() {
			It("returns an error when it cannot parse the private key", func() {
				err := socks5Proxy.Start("some-bad-private-key", sshServerURL)
				Expect(err).To(MatchError("ssh: no key found"))
			})

			It("returns an error when it cannot get the host key", func() {
				hostKeyGetter.GetCall.Returns.Error = errors.New("failed to get host key")
				err := socks5Proxy.Start(sshPrivateKey, sshServerURL)
				Expect(err).To(MatchError("failed to get host key"))
			})

			It("returns an error when it cannot dial the jumpbox url", func() {
				err := socks5Proxy.Start(sshPrivateKey, "some-bad-url")
				Expect(err).To(MatchError("dial tcp: address some-bad-url: missing port in address"))
			})

			Context("when it cannot start a socks5 proxy server", func() {
				var (
					fakeServer net.Listener
				)

				BeforeEach(func() {
					var err error
					fakeServer, err = net.Listen("tcp", "127.0.0.1:9999")
					Expect(err).NotTo(HaveOccurred())

					socks5Proxy = proxy.NewSocks5Proxy(logger, hostKeyGetter, 9999)
				})

				AfterEach(func() {
					fakeServer.Close()
				})

				It("logs a helpful error message", func() {
					err := socks5Proxy.Start(sshPrivateKey, sshServerURL)
					Expect(err).NotTo(HaveOccurred())
					Eventually(func() []string {
						return logger.PrintlnMessages()
					}, "10s").Should(ContainElement("err: failed to start socks5 proxy: listen tcp 127.0.0.1:9999: bind: address already in use"))
				})
			})

			It("returns an error when netListen fails", func() {
				proxy.SetNetListen(func(string, string) (net.Listener, error) {
					return nil, errors.New("failed to listen")
				})

				err := socks5Proxy.Start(sshPrivateKey, sshServerURL)
				Expect(err).To(MatchError("failed to listen"))
			})
		})
	})

	Describe("Addr", func() {
		var (
			socks5Proxy   *proxy.Socks5Proxy
			logger        *fakes.Logger
			hostKeyGetter *fakes.HostKeyGetter
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			socks5Proxy = proxy.NewSocks5Proxy(logger, hostKeyGetter, 9999)
		})

		It("returns a valid address of the socks5 proxy", func() {
			Expect(socks5Proxy.Addr()).To(Equal("127.0.0.1:9999"))
		})
	})
})
