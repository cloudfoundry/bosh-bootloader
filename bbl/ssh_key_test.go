package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("ssh-key", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when bosh/jumpbox variables does not exist", func() {
		Context("when keypair.privateKey exists", func() {
			It("returns the ssh key from keypair.privateKey", func() {
				state := []byte(`{
					"version": 3,
					"keyPair": {
						"privateKey": "some-ssh-private-key"
					}
				}`)
				err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				args := []string{
					"--state-dir", tempDirectory,
					"ssh-key",
				}

				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("some-ssh-private-key"))
			})
		})
	})

	It("returns the ssh key from the given state file", func() {
		state := []byte(`{
			"version": 3,
			"bosh": {
				"variables": "jumpbox_ssh:\n  private_key: some-ssh-private-key"
			}
		}`)
		err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"--state-dir", tempDirectory,
			"ssh-key",
		}

		session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("some-ssh-private-key"))
	})

	Context("when a jumpbox was deployed", func() {
		It("returns the ssh key from the given state file", func() {
			state := []byte(`{
				"version": 3,
				"jumpbox": {
					"enabled": true,
					"variables": "jumpbox_ssh:\n  private_key: some-ssh-private-key"
				}
			}`)
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"ssh-key",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("some-ssh-private-key"))
		})
	})

	Context("failure cases", func() {
		It("returns a non zero exit code when the bbl-state.json does not exist", func() {
			tempDirectory, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"ssh-key",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))

			expectedErrorMessage := fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)
			Expect(session.Err.Contents()).To(ContainSubstring(expectedErrorMessage))
		})

		It("returns a non zero exit code when the ssh key does not exist", func() {
			state := []byte(`{"version":3}`)
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"ssh-key",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err.Contents()).To(ContainSubstring("Could not retrieve the ssh key"))
		})
	})
})
