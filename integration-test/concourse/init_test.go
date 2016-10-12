package integration_test

import (
	"strings"
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
	pathToBBL, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl",
		"-ldflags", strings.Join([]string{
			"-X main.BOSHURL=https://bosh.io/d/github.com/cloudfoundry/bosh?v=257.15",
			"-X main.BOSHSHA1=f4cf3579bfac994cd3bde4a9d8cbee3ad189c8b2",
			"-X main.BOSHAWSCPIURL=https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=60",
			"-X main.BOSHAWSCPISHA1=8e40a9ff892204007889037f094a1b0d23777058",
			"-X main.StemcellURL=https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent",
			"-X main.StemcellSHA1=a4a3b387ee81cd0e1b73debffc39f0907b80f9c6",
		}, " "))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
