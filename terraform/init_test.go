package terraform_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestTerraform(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "terraform")
}

var (
	originalPath string
)

var _ = BeforeSuite(func() {
	originalPath = os.Getenv("PATH")
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
