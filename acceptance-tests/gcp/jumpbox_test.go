package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

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

		bbl.Up("gcp", []string{"--name", bbl.PredefinedEnvID(), "--jumpbox"})
	})

	AfterEach(func() {
		bbl.Destroy()
	})

	It("bbl's up a new jumpbox and a new bosh director", func() {
		By("parsing the output of print-env", func() {
			stdout := fmt.Sprintf("#!/bin/bash\n%s", bbl.PrintEnv())

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
		})
	})
})
