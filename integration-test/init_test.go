package integration_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "integration")
}

var (
	config    integration.Config
	pathToBBL string
)

var _ = BeforeSuite(func() {
	configPath, err := integration.ConfigPath()
	Expect(err).NotTo(HaveOccurred())

	config, err = integration.LoadConfig(configPath)
	Expect(err).NotTo(HaveOccurred())

	pathToBBL, err = gexec.Build("github.com/pivotal-cf-experimental/bosh-bootloader/bbl")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func bbl(args []string) *gexec.Session {
	cmd := exec.Command(pathToBBL, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
