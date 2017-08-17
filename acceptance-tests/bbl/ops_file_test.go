package acceptance_test

import (
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("ops file test", func() {
	var (
		bbl     actors.BBL
		boshcli actors.BOSHCLI
		state   acceptance.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "ops-file-env")
		boshcli = actors.NewBOSHCLI()
		state = acceptance.NewState(configuration.StateFileDir)

		session := bbl.Up(configuration.IAAS, []string{
			"--name", bbl.PredefinedEnvID(),
			"--ops-file", filepath.Join("fixtures", "jumpbox_user_other.yml"),
		})
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("bbl's up a new bosh director", func() {
		By("checking if the bosh director exists", func() {
			directorAddress := bbl.DirectorAddress()
			caCertPath := bbl.SaveDirectorCA()

			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("checking if ssh'ing works", func() {
			privateKey, err := ssh.ParsePrivateKey([]byte(bbl.SSHKey()))
			Expect(err).NotTo(HaveOccurred())

			directorAddressURL, err := url.Parse(bbl.DirectorAddress())
			Expect(err).NotTo(HaveOccurred())

			address := fmt.Sprintf("%s:22", directorAddressURL.Hostname())
			_, err = ssh.Dial("tcp", address, &ssh.ClientConfig{
				User: "jumpbox_other",
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(privateKey),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
