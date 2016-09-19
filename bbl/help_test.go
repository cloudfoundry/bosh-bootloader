package main_test

import (
	"fmt"
	"os/exec"

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

			It("prints out the help and ignores any sub-commands passed to it", func() {
				args := []string{
					"--help",
					"some-invalid-command",
				}

				cmd := exec.Command(pathToBBL, args...)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Eventually(session.Out.Contents()).Should(ContainSubstring("bbl [GLOBAL OPTIONS] COMMAND [OPTIONS]"))
				Eventually(session.Out.Contents()).ShouldNot(ContainSubstring("some-invalid-command"))
			})
		})
	})

	DescribeTable("help command", func(command, expectedDescription string) {
		args := []string{
			"help",
			command,
		}

		cmd := exec.Command(pathToBBL, args...)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Eventually(session.Out.Contents()).Should(ContainSubstring(fmt.Sprintf("bbl [GLOBAL OPTIONS] %s [OPTIONS]", command)))
		Eventually(session.Out.Contents()).Should(ContainSubstring(expectedDescription))
	},
		Entry("Up", "up", "--aws-access-key-id"),
		Entry("Destroy", "destroy", "--no-confirm"),
		Entry("Create LBs", "create-lbs", "Attaches a load balancer"),
		Entry("Update LBs", "update-lbs", "Updates a load balancer"),
		Entry("Delete LBs", "delete-lbs", "Deletes the load balancers"),
		Entry("Version", "version", "Prints version"),
		Entry("Director Address", "director-address", "Prints the BOSH director address"),
		Entry("Director Username", "director-username", "Prints the BOSH director username"),
		Entry("Director Username", "director-username", "Prints the BOSH director username"),
		Entry("Director Password", "director-password", "Prints the BOSH director password"),
		Entry("BOSH CA Cert", "bosh-ca-cert", "Prints the BOSH director CA certificate"),
		Entry("ENV ID", "env-id", "environment ID"),
		Entry("Help", "help", "Prints helpful message for the given command"),
		Entry("LBs", "lbs", "Lists attached load balancers"),
		Entry("SSH Key", "ssh-key", "Prints the SSH private key"),
	)

	DescribeTable("command --help", func(command, expectedDescription string) {
		args := []string{
			command,
			"--help",
		}

		cmd := exec.Command(pathToBBL, args...)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Eventually(session.Out.Contents()).Should(ContainSubstring(fmt.Sprintf("bbl [GLOBAL OPTIONS] %s [OPTIONS]", command)))
		Eventually(session.Out.Contents()).Should(ContainSubstring(expectedDescription))
	},
		Entry("Up", "up", "--aws-access-key-id"),
		Entry("Destroy", "destroy", "--no-confirm"),
		Entry("Create LBs", "create-lbs", "Attaches a load balancer"),
		Entry("Update LBs", "update-lbs", "Updates a load balancer"),
		Entry("Delete LBs", "delete-lbs", "Deletes the load balancers"),
		Entry("Version", "version", "Prints version"),
		Entry("Director Address", "director-address", "Prints the BOSH director address"),
		Entry("Director Username", "director-username", "Prints the BOSH director username"),
		Entry("Director Username", "director-username", "Prints the BOSH director username"),
		Entry("Director Password", "director-password", "Prints the BOSH director password"),
		Entry("BOSH CA Cert", "bosh-ca-cert", "Prints the BOSH director CA certificate"),
		Entry("ENV ID", "env-id", "environment ID"),
		Entry("Help", "help", "Prints helpful message for the given command"),
		Entry("LBs", "lbs", "Lists attached load balancers"),
		Entry("SSH Key", "ssh-key", "Prints the SSH private key"),
	)
})
