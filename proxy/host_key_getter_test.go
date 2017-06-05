package proxy_test

import (
	"github.com/cloudfoundry/bosh-bootloader/proxy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("HostKeyGetter", func() {
	Describe("Get", func() {
		var (
			hostKeyGetter proxy.HostKeyGetter
			key           ssh.PublicKey
			sshServerAddr string
		)

		BeforeEach(func() {
			signer, err := ssh.ParsePrivateKey([]byte(sshPrivateKey))
			Expect(err).NotTo(HaveOccurred())
			key = signer.PublicKey()

			sshServerAddr = startSSHServer("")

			hostKeyGetter = proxy.NewHostKeyGetter()
		})

		It("returns the host key", func() {
			hostKey, err := hostKeyGetter.Get(sshPrivateKey, sshServerAddr)
			Expect(err).NotTo(HaveOccurred())
			Expect(hostKey).To(Equal(key))
		})

		Context("failure cases", func() {
			It("returns an error when parse private key fails", func() {
				_, err := hostKeyGetter.Get("%%%", sshServerAddr)
				Expect(err).To(MatchError("ssh: no key found"))
			})

			It("returns an error when dial fails", func() {
				_, err := hostKeyGetter.Get(sshPrivateKey, "some-bad-url")
				Expect(err).To(MatchError("dial tcp: address some-bad-url: missing port in address"))
			})
		})
	})
})
