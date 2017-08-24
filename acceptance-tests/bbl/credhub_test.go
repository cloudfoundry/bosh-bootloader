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

var _ = Describe("credhub test", func() {
	var (
		bbl   actors.BBL
		state acceptance.State
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "credhub-env")
		state = acceptance.NewState(configuration.StateFileDir)
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
	})

	It("creates a director with a jumpbox, credhub, and UAA", func() {
		session := bbl.Up("--credhub", "--name", bbl.PredefinedEnvID())
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		By("creating an ssh tunnel to the director in print-env", func() {
			evalBBLPrintEnv(bbl)
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

		By("creating a new ssh tunnel to the director in print-env", func() {
			evalBBLPrintEnv(bbl)
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
