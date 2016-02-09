package main_test

import (
	"io/ioutil"
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

	Describe("bbl -v", func() {
		It("prints out the current version", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, "-v"), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("bbl 0.0.1"))
		})
	})

	Describe("bbl unsupported-print-concourse-aws-template", func() {
		It("prints a CloudFomation template", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, "unsupported-print-concourse-aws-template"), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			buf, err := ioutil.ReadFile("../commands/fixtures/cloudformation.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(session.Out.Contents()).To(MatchJSON(string(buf)))
		})
	})

	It("prints an error when an unknown flag is provided", func() {
		session, err := gexec.Start(exec.Command(pathToBBL, "--some-unknown-flag"), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err.Contents()).To(ContainSubstring("unknown flag `some-unknown-flag'"))
	})

	It("prints an error when an unknown command is provided", func() {
		session, err := gexec.Start(exec.Command(pathToBBL, "some-unknown-flag"), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err.Contents()).To(ContainSubstring("Unknown command `some-unknown-flag'"))
	})
})
