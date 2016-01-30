package main_test

import (
	"os/exec"

	"github.com/gomega/gexec"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl", func() {
	Describe("bbl -h", func() {
		It("prints out the usage", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, "-h"), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("usage"))
		})
	})

	Describe("bbl -v", func() {
		It("prints out the current version", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, "-v"), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("bbl 0.0.1"))
		})
	})

	It("prints an error when an unknown flag is provided", func() {
		session, err := gexec.Start(exec.Command(pathToBBL, "--some-unknown-flag"), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err.Contents()).To(ContainSubstring("unknown flag `some-unknown-flag'"))
	})
})
