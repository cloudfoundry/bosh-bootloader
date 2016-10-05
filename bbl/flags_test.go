package main_test

import (
	"os/exec"
	"strings"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("flags test", func() {
	Context("GlobalFlags", func() {
		DescribeTable("exits with non-zero status code when state-dir is specified twice", func(arguments string) {
			args := strings.Split(arguments, " ")

			cmd := exec.Command(pathToBBL, args...)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Invalid usage: cannot specify global 'state-dir' flag more than once."))
		},
			Entry("two --state-dir flags with space/equal sign", "--state-dir /some/fake/dir --state-dir=/some/other/fake/dir up"),
			Entry("two -state-dir flags with space/equal sign", "-state-dir /some/fake/dir -state-dir=/some/other/fake/dir up"),
		)
	})

	Context("Up", func() {
		Context("failure cases", func() {
			It("exits with non-zero status when invalid flags are passed", func() {
				args := []string{
					"up",
					"--aws-access-key-id", "some-aws-access-key-id",
					"--aws-secret-access-key", "aws-secret-access-key",
					"--aws-region", "aws-region",
					"--some-invalid-flag", "some-value",
				}
				cmd := exec.Command(pathToBBL, args...)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("flag provided but not defined: -some-invalid-flag"))
			})

			It("fails when unknown global flags are passed", func() {
				args := []string{
					"-some-global-flag",
					"up",
					"--aws-access-key-id", "some-aws-access-key-id",
					"--aws-secret-access-key", "aws-secret-access-key",
					"--aws-region", "aws-region",
				}
				cmd := exec.Command(pathToBBL, args...)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("flag provided but not defined: -some-global-flag"))
			})

			It("fails when unknown commands are passed", func() {
				args := []string{
					"-h",
					"badcmd",
					"--aws-access-key-id", "some-aws-access-key-id",
					"--aws-secret-access-key", "aws-secret-access-key",
					"--aws-region", "aws-region",
				}
				cmd := exec.Command(pathToBBL, args...)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("Unrecognized command 'badcmd'"))
			})
		})
	})

	Context("Delete-lbs", func() {
		It("exits with non-zero status when aws creds are passed to it", func() {
			args := []string{
				"delete-lbs",
				"--aws-access-key-id", "some-aws-access-key-id",
				"--aws-secret-access-key", "aws-secret-access-key",
				"--aws-region", "aws-region",
			}
			cmd := exec.Command(pathToBBL, args...)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("flag provided but not defined: -aws-access-key-id"))
		})
	})
})
