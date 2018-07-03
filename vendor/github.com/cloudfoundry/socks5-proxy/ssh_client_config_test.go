package proxy_test

import (
	"github.com/cloudfoundry/socks5-proxy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

var _ = Describe("NewSSHClientConfig", func() {
	It("creates an ssh client config with a timeout", func() {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		Expect(err).NotTo(HaveOccurred())
		hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}

		config := proxy.NewSSHClientConfig("some-username", hostKeyCallback, ssh.PublicKeys(signer))

		Expect(config.Timeout).To(Equal(30 * time.Second))
	})
})
