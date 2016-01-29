package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestBbl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bbl Suite")
}

var (
	pathToBBL string
)

var _ = BeforeSuite(func() {
	var err error

	pathToBBL, err = gexec.Build("github.com/pivotal-cf-experimental/bosh-bootloader/bbl")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
