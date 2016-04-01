package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestBbl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bbl")
}

var (
	pathToBBL          string
	pathToBOSHInit     string
	pathToFakeBOSHInit string
)

var _ = BeforeSuite(func() {
	var err error

	pathToBBL, err = gexec.Build("github.com/pivotal-cf-experimental/bosh-bootloader/bbl")
	Expect(err).NotTo(HaveOccurred())

	pathToFakeBOSHInit, err = gexec.Build("github.com/pivotal-cf-experimental/bosh-bootloader/bbl/fakeboshinit")
	Expect(err).NotTo(HaveOccurred())

	pathToBOSHInit = filepath.Join(filepath.Dir(pathToFakeBOSHInit), "bosh-init")
	err = os.Rename(pathToFakeBOSHInit, pathToBOSHInit)
	Expect(err).NotTo(HaveOccurred())

	os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSHInit), os.Getenv("PATH")}, ":"))
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func executeCommand(args []string, exitCode int) *gexec.Session {
	cmd := exec.Command(pathToBBL, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(exitCode))

	return session
}
