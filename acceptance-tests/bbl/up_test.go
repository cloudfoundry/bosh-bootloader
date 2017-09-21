package acceptance_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("up", func() {
	var (
		bbl             actors.BBL
		boshcli         actors.BOSHCLI
		directorAddress string
		caCertPath      string
		sshSession      *gexec.Session
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "up-env")
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		sshSession.Interrupt()
		Eventually(sshSession, "5s").Should(gexec.Exit())
		session := bbl.Down()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("bbl's up a new bosh director and jumpbox", func() {
		acceptance.SkipUnless("bbl-up")
		session := bbl.Up("--name", bbl.PredefinedEnvID())
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		By("creating an ssh tunnel to the director in print-env", func() {
			sshSession = startSSHTunnel(bbl)
		})

		By("checking if the bosh director exists", func() {
			directorAddress = bbl.DirectorAddress()
			caCertPath = bbl.SaveDirectorCA()

			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			if err != nil {
				fmt.Println(string(err.(*exec.ExitError).Stderr))
			}
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

		By("checking if bbl print-env prints the bosh environment variables", func() {
			stdout := bbl.PrintEnv()

			Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT="))
			Expect(stdout).To(ContainSubstring("export BOSH_CLIENT="))
			Expect(stdout).To(ContainSubstring("export BOSH_CLIENT_SECRET="))
			Expect(stdout).To(ContainSubstring("export BOSH_CA_CERT="))
		})

		By("rotating the jumpbox's ssh key", func() {
			sshKey := bbl.SSHKey()
			Expect(sshKey).NotTo(BeEmpty())

			session = bbl.Rotate()
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

			rotatedKey := bbl.SSHKey()
			Expect(rotatedKey).NotTo(BeEmpty())
			Expect(rotatedKey).NotTo(Equal(sshKey))
		})

		By("checking bbl up is idempotent", func() {
			session := bbl.Up()
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("destroying the director and the jumpbox", func() {
			session := bbl.Down()
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		})
	})
})

func startSSHTunnel(bbl actors.BBL) *gexec.Session {
	printEnvLines := strings.Split(bbl.PrintEnv(), "\n")
	os.Setenv("BOSH_ALL_PROXY", getExport("BOSH_ALL_PROXY", printEnvLines))

	var sshArgs []string
	for i := 0; i < len(printEnvLines); i++ {
		if strings.HasPrefix(printEnvLines[i], "ssh ") {
			sshCmd := strings.TrimPrefix(printEnvLines[i], "ssh ")
			sshCmd = strings.Replace(sshCmd, "$BOSH_GW_PRIVATE_KEY", getExport("BOSH_GW_PRIVATE_KEY", printEnvLines), -1)
			sshCmd = strings.Replace(sshCmd, "-f ", "", -1)
			sshArgs = strings.Split(sshCmd, " ")
		}
	}

	cmd := exec.Command("ssh", sshArgs...)
	sshSession, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return sshSession
}

func getExport(keyName string, lines []string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, "export ") {
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				Fail(fmt.Sprintf("Unexpected print-env output: %s\n", line))
			}
			keyValue := parts[1]
			keyValueParts := strings.Split(keyValue, "=")
			key := keyValueParts[0]
			value := keyValueParts[1]

			if key == keyName {
				return value
			}
		}
	}
	return ""
}
