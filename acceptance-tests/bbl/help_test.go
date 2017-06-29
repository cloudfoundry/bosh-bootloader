package acceptance_test

import (
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl", func() {
	Describe("help", func() {
		Describe("bbl -h", func() {
			It("prints out the usage", func() {
				session, err := gexec.Start(exec.Command(pathToBBL, "-h"), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("usage"))
			})
		})

		Describe("bbl --help", func() {
			It("prints out the usage", func() {
				session, err := gexec.Start(exec.Command(pathToBBL, "--help"), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("usage"))
			})
		})

		Describe("bbl help", func() {
			It("prints out the usage", func() {
				session, err := gexec.Start(exec.Command(pathToBBL, "help"), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("usage"))
			})

			It("prints out the help and errors on any unkown commands passed to it", func() {
				args := []string{
					"--help",
					"some-invalid-command",
				}

				cmd := exec.Command(pathToBBL, args...)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Out.Contents()).Should(ContainSubstring("bbl [GLOBAL OPTIONS] COMMAND [OPTIONS]"))
				Eventually(session.Err.Contents()).Should(ContainSubstring("some-invalid-command"))
			})

			It("prints out the help and ignores other args passed to it", func() {
				args := []string{
					"--help",
					"up",
					"--aws-gibberish",
				}

				cmd := exec.Command(pathToBBL, args...)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Eventually(session.Out.Contents()).Should(ContainSubstring("bbl [GLOBAL OPTIONS] up [OPTIONS]"))
				Expect(session.Out.Contents()).NotTo(ContainSubstring("aws-gibberish"))
			})
		})

		Describe("bbl", func() {
			It("prints out the usage", func() {
				session, err := gexec.Start(exec.Command(pathToBBL), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("usage"))
			})
		})
	})

	Context("command specific help", func() {
		DescribeTable("when passing --help flag or help command", func(command, expectedDescription string, args []string) {
			cmd := exec.Command(pathToBBL, args...)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out.Contents()).Should(ContainSubstring(fmt.Sprintf("bbl [GLOBAL OPTIONS] %s [OPTIONS]", command)))
			Eventually(session.Out.Contents()).Should(ContainSubstring(expectedDescription))
		},
			Entry("Up", "up", "--aws-access-key-id", []string{"help", "up"}),
			Entry("Up", "up", "--aws-access-key-id", []string{"up", "--help"}),
			Entry("Destroy", "destroy", "--no-confirm", []string{"help", "destroy"}),
			Entry("Destroy", "destroy", "--no-confirm", []string{"destroy", "--help"}),
			Entry("Create LBs", "create-lbs", "Attaches load balancer(s)", []string{"help", "create-lbs"}),
			Entry("Create LBs", "create-lbs", "Attaches load balancer(s)", []string{"create-lbs", "--help"}),
			Entry("Update LBs", "update-lbs", "Updates load balancer(s)", []string{"help", "update-lbs"}),
			Entry("Update LBs", "update-lbs", "Updates load balancer(s)", []string{"update-lbs", "--help"}),
			Entry("Delete LBs", "delete-lbs", "Deletes load balancer(s)", []string{"help", "delete-lbs"}),
			Entry("Delete LBs", "delete-lbs", "Deletes load balancer(s)", []string{"delete-lbs", "--help"}),
			Entry("Version", "version", "Prints version", []string{"help", "version"}),
			Entry("Version", "version", "Prints version", []string{"version", "--help"}),
			Entry("Director Address", "director-address", "Prints BOSH director address", []string{"help", "director-address"}),
			Entry("Director Address", "director-address", "Prints BOSH director address", []string{"director-address", "--help"}),
			Entry("Director Username", "director-username", "Prints BOSH director username", []string{"help", "director-username"}),
			Entry("Director Username", "director-username", "Prints BOSH director username", []string{"director-username", "--help"}),
			Entry("Director Password", "director-password", "Prints BOSH director password", []string{"help", "director-password"}),
			Entry("Director Password", "director-password", "Prints BOSH director password", []string{"director-password", "--help"}),
			Entry("Director CA Cert", "director-ca-cert", "Prints BOSH director CA certificate", []string{"help", "director-ca-cert"}),
			Entry("Director CA Cert", "director-ca-cert", "Prints BOSH director CA certificate", []string{"director-ca-cert", "--help"}),
			Entry("ENV ID", "env-id", "environment ID", []string{"help", "env-id"}),
			Entry("ENV ID", "env-id", "environment ID", []string{"env-id", "--help"}),
			Entry("Help", "help", "Prints helpful message for the given command", []string{"help", "help"}),
			Entry("Help", "help", "Prints helpful message for the given command", []string{"help", "--help"}),
			Entry("Latest Error", "latest-error", "Prints the output from the latest call to terraform", []string{"help", "latest-error"}),
			Entry("Latest Error", "latest-error", "Prints the output from the latest call to terraform", []string{"latest-error", "--help"}),
			Entry("LBs", "lbs", "Prints attached load balancer(s)", []string{"help", "lbs"}),
			Entry("LBs", "lbs", "Prints attached load balancer(s)", []string{"lbs", "--help"}),
			Entry("SSH Key", "ssh-key", "Prints SSH private key", []string{"help", "ssh-key"}),
			Entry("SSH Key", "ssh-key", "Prints SSH private key", []string{"ssh-key", "--help"}),
		)
	})
})

func executeCommand(args []string, exitCode int) *gexec.Session {
	cmd := exec.Command(pathToBBL, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(exitCode))

	return session
}
