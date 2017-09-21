package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
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
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "up-env")
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		session := bbl.Down()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("bbl's up a new bosh director and jumpbox", func() {
		acceptance.SkipUnless("bbl-up")
		session := bbl.Up("--name", bbl.PredefinedEnvID())
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		By("creating an ssh tunnel to the director in print-env", func() {
			evalBBLPrintEnv(bbl)
		})

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

func evalBBLPrintEnv(bbl actors.BBL) {
	stdout := fmt.Sprintf("#!/bin/bash\n%s", bbl.PrintEnv())
	Expect(stdout).To(ContainSubstring("ssh -f -N"))

	stdout = strings.Replace(stdout, "-f -N", "", 1)

	dir, err := ioutil.TempDir("", "bosh-print-env-command")
	Expect(err).NotTo(HaveOccurred())

	printEnvCommandPath := filepath.Join(dir, "eval-print-env")

	err = ioutil.WriteFile(printEnvCommandPath, []byte(stdout), 0700)
	Expect(err).NotTo(HaveOccurred())

	cmd := exec.Command(printEnvCommandPath)
	cmdIn, err := cmd.StdinPipe()

	go func() {
		defer GinkgoRecover()
		cmdOut, err := cmd.Output()
		if err != nil {
			switch err.(type) {
			case *exec.ExitError:
				exitErr := err.(*exec.ExitError)
				fmt.Println(string(exitErr.Stderr))
			}
		}
		Expect(err).NotTo(HaveOccurred())

		output := string(cmdOut)
		Expect(output).To(ContainSubstring("Welcome to Ubuntu"))
	}()

	cmdIn.Close()
}
