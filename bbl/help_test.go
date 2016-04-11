package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
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
	})
})
