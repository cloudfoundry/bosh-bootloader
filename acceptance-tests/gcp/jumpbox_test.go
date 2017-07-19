package acceptance_test

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/proxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("jumpbox test", func() {
	var (
		bbl     actors.BBL
		bosh    actors.BOSH
		boshcli actors.BOSHCLI
		state   acceptance.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "jumpbox-env")
		bosh = actors.NewBOSH()
		boshcli = actors.NewBOSHCLI()
		state = acceptance.NewState(configuration.StateFileDir)

		bbl.Up(actors.GCPIAAS, []string{"--name", bbl.PredefinedEnvID(), "--jumpbox"})
	})

	AfterEach(func() {
		bbl.Destroy()
	})

	It("bbl's up a new jumpbox and a new bosh director", func() {
		var jumpboxAddress string

		By("checking if the bosh jumpbox exists", func() {
			jumpboxAddress = bbl.JumpboxAddress()

			hostKeyGetter := proxy.NewHostKeyGetter()
			socks5Proxy := proxy.NewSocks5Proxy(nil, hostKeyGetter, 0)
			err := socks5Proxy.Start(bbl.SSHKey(), jumpboxAddress)
			Expect(err).NotTo(HaveOccurred())

			os.Setenv("BOSH_ALL_PROXY", fmt.Sprintf("socks5://%s", socks5Proxy.Addr()))
		})

		By("checking that the director is running", func() {
			directorAddress := bbl.DirectorAddress()
			caCertPath := bbl.SaveDirectorCA()

			env, err := boshcli.Env(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(env).To(ContainSubstring(bbl.PredefinedEnvID()))
		})

		By("checking if ssh'ing works", func() {
			privateKey, err := ssh.ParsePrivateKey([]byte(bbl.SSHKey()))
			Expect(err).NotTo(HaveOccurred())

			_, err = ssh.Dial("tcp", jumpboxAddress, &ssh.ClientConfig{
				User: "jumpbox",
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(privateKey),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("checking if bbl print-env prints the bosh environment variables", func() {
			stdout := bbl.PrintEnv()

			Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT="))
			Expect(stdout).To(ContainSubstring("export BOSH_CLIENT="))
			Expect(stdout).To(ContainSubstring("export BOSH_CLIENT_SECRET="))
			Expect(stdout).To(ContainSubstring("export BOSH_CA_CERT="))
		})
	})
})
