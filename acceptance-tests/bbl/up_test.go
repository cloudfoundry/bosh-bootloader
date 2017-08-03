package acceptance_test

import (
	"fmt"
	"net/url"

	"golang.org/x/crypto/ssh"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("up test", func() {
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

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "up-env")
		bosh = actors.NewBOSH()
		boshcli = actors.NewBOSHCLI()
		state = acceptance.NewState(configuration.StateFileDir)

		bbl.Up(configuration.IAAS, []string{"--name", bbl.PredefinedEnvID()})
	})

	AfterEach(func() {
		bbl.Destroy()
	})

	It("bbl's up a new bosh director", func() {
		var (
			directorAddress string
			caCertPath      string
		)

		By("checking if the bosh director exists", func() {
			directorAddress = bbl.DirectorAddress()
			caCertPath = bbl.SaveDirectorCA()

			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("checking that the cloud config exists", func() {
			directorUsername := bbl.DirectorUsername()
			directorPassword := bbl.DirectorPassword()

			cloudConfig, err := boshcli.CloudConfig(directorAddress, caCertPath, directorUsername, directorPassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudConfig).NotTo(BeEmpty())
		})

		By("checking if ssh'ing works", func() {
			privateKey, err := ssh.ParsePrivateKey([]byte(bbl.SSHKey()))
			Expect(err).NotTo(HaveOccurred())

			directorAddressURL, err := url.Parse(bbl.DirectorAddress())
			Expect(err).NotTo(HaveOccurred())

			address := fmt.Sprintf("%s:22", directorAddressURL.Hostname())
			_, err = ssh.Dial("tcp", address, &ssh.ClientConfig{
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
