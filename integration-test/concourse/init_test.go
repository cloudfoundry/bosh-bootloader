package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "concourse")
}

var pathToBBL string

var _ = BeforeSuite(func() {
	var err error
	pathToBBL, err = gexec.Build("github.com/pivotal-cf-experimental/bosh-bootloader/bbl")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
