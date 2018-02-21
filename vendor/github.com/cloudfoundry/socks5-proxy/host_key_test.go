package proxy_test

import (
	proxy "github.com/cloudfoundry/socks5-proxy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("HostKey", func() {
	Describe("Get", func() {
		var (
			hostKey       proxy.HostKey
			key           ssh.PublicKey
			sshServerAddr string
		)

		BeforeEach(func() {
			signer, err := ssh.ParsePrivateKey([]byte(privateKey))
			Expect(err).NotTo(HaveOccurred())
			key = signer.PublicKey()

			sshServerAddr = proxy.StartTestSSHServer("", privateKey, "")

			hostKey = proxy.NewHostKey()
		})

		It("returns the host key", func() {
			hostKey, err := hostKey.Get("", privateKey, sshServerAddr)
			Expect(err).NotTo(HaveOccurred())
			Expect(hostKey).To(Equal(key))
		})

		Context("when a username has been set", func() {
			BeforeEach(func() {
				sshServerAddr = proxy.StartTestSSHServer("", privateKey, "different-username")
				hostKey = proxy.NewHostKey()
			})

			It("returns the host key", func() {
				hostKey, err := hostKey.Get("different-username", privateKey, sshServerAddr)
				Expect(err).NotTo(HaveOccurred())
				Expect(hostKey).To(Equal(key))
			})
		})

		Context("failure cases", func() {
			Context("when parse private key fails", func() {
				It("returns an error", func() {
					_, err := hostKey.Get("", "%%%", sshServerAddr)
					Expect(err).To(MatchError("ssh: no key found"))
				})
			})

			Context("when dial fails", func() {
				It("returns an error", func() {
					_, err := hostKey.Get("", privateKey, "some-bad-url")
					Expect(err).To(MatchError("dial tcp: address some-bad-url: missing port in address"))
				})
			})
		})
	})
})
