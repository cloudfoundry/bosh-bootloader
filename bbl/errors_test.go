package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl", func() {
	It("prints an error when an unknown flag is provided", func() {
		session, err := gexec.Start(exec.Command(pathToBBL, "--some-unknown-flag"), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err.Contents()).To(ContainSubstring("flag provided but not defined: -some-unknown-flag"))
		Expect(session.Out.Contents()).To(ContainSubstring("Usage"))
	})

	It("prints an error when an unknown command is provided", func() {
		session, err := gexec.Start(exec.Command(pathToBBL, "some-unknown-command"), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err.Contents()).To(ContainSubstring("unknown command: some-unknown-command"))
		Expect(session.Out.Contents()).To(ContainSubstring("Usage"))
	})
})
