package integration_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bbl")
}

var pathToBBL string

var _ = BeforeSuite(func() {
	var err error

	pathToBBL, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl",
		"-ldflags", strings.Join([]string{
			fmt.Sprintf("-X main.BOSHURL=%s", os.Getenv("BOSH_URL")),
			fmt.Sprintf("-X main.BOSHSHA1=%s", os.Getenv("BOSH_SHA1")),
			fmt.Sprintf("-X main.BOSHAWSCPIURL=%s", os.Getenv("BOSH_AWS_CPI_URL")),
			fmt.Sprintf("-X main.BOSHAWSCPISHA1=%s", os.Getenv("BOSH_AWS_CPI_SHA1")),
			fmt.Sprintf("-X main.StemcellURL=%s", os.Getenv("STEMCELL_URL")),
			fmt.Sprintf("-X main.StemcellSHA1=%s", os.Getenv("STEMCELL_SHA1")),
		}, " "))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
