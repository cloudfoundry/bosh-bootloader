package acceptance_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl", func() {
	Context("when passed an unsupported --iaas flag", func() {
		It("prints out an error", func() {
			cmd := exec.Command(pathToBBL, "up", "--iaas", "banana")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))

			Expect(string(session.Err.Contents())).To(ContainSubstring("--iaas [gcp, aws, azure, vsphere, openstack, cloudstack] must be provided or BBL_IAAS must be set"))
			Expect(string(session.Err.Contents())).NotTo(ContainSubstring("panic"))
		})
	})

	Context("when passed invalid GCP credentials", func() {
		It("prints out an error", func() {
			cmd := exec.Command(
				pathToBBL, "up",
				"--iaas", "gcp",
				"--gcp-service-account-key", `{"project_id": "idk"}`,
				"--gcp-region", "us-central1",
			)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).NotTo(BeEmpty())
			Expect(string(session.Err.Contents())).To(ContainSubstring("google:"))
			Expect(string(session.Err.Contents())).NotTo(ContainSubstring("panic"))
		})
	})
})
