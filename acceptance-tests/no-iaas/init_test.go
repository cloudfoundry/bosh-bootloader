package acceptance_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "no-iaas")
}

var pathToBBL string

var _ = BeforeSuite(func() {
	var err error

	pathToBBL, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
