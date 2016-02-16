package main_test

import (
	"io/ioutil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl", func() {
	Describe("bbl unsupported-print-concourse-aws-template", func() {
		It("prints a CloudFomation template", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, "unsupported-print-concourse-aws-template"), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			buf, err := ioutil.ReadFile("../aws/cloudformation/fixtures/cloudformation.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(session.Out.Contents()).To(MatchJSON(string(buf)))
		})
	})
})
