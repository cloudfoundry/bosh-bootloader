package terraform_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestTerraform(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "terraform")
}

var (
	pathToFakeTerraform string
	pathToTerraform     string
)

var _ = BeforeSuite(func() {
	var err error
	pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform")
	Expect(err).NotTo(HaveOccurred())

	pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
	err = os.Rename(pathToFakeTerraform, pathToTerraform)
	Expect(err).NotTo(HaveOccurred())

	os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), os.Getenv("PATH")}, ":"))
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
