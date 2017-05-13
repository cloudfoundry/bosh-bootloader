package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"testing"

	boshbackend "github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh/backend"
	terraformbackend "github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform/backend"
	"github.com/cloudfoundry/bosh-bootloader/bbl/gcpbackend"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBbl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bbl")
}

var (
	pathToBBL           string
	pathToFakeBOSH      string
	pathToBOSH          string
	pathToFakeTerraform string
	pathToTerraform     string

	gcpBackend    gcpbackend.GCPBackend
	fakeGCPServer *httptest.Server

	fakeTerraformBackendServer *terraformbackend.Backend
	fakeBOSHCLIBackendServer   *boshbackend.Backend

	serviceAccountKey string
	originalPath      string
	noFakesPath       string
)

var _ = BeforeSuite(func() {
	var err error
	gcpBackend = gcpbackend.GCPBackend{}
	fakeGCPServer, serviceAccountKey = gcpBackend.StartFakeGCPBackend()
	fakeBOSHCLIBackendServer = boshbackend.NewBackend()
	fakeTerraformBackendServer = terraformbackend.NewBackend()

	pathToBBL, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl", "--ldflags", fmt.Sprintf("-X main.gcpBasePath=%s", fakeGCPServer.URL))
	Expect(err).NotTo(HaveOccurred())

	pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
		"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHCLIBackendServer.ServerURL()))
	Expect(err).NotTo(HaveOccurred())

	pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
	err = os.Rename(pathToFakeBOSH, pathToBOSH)
	Expect(err).NotTo(HaveOccurred())

	pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
		"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.ServerURL()))
	Expect(err).NotTo(HaveOccurred())

	pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
	err = os.Rename(pathToFakeTerraform, pathToTerraform)
	Expect(err).NotTo(HaveOccurred())

	noFakesPath = os.Getenv("PATH")
	fakeBOSHCLIBackendServer.SetPath(noFakesPath)

	os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSH), filepath.Dir(pathToTerraform), os.Getenv("PATH")}, ":"))

	originalPath = os.Getenv("PATH")
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

func writeStateJson(state storage.State, tempDirectory string) {
	buf, err := json.Marshal(state)
	Expect(err).NotTo(HaveOccurred())

	ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), buf, os.ModePerm)
}

func readStateJson(tempDirectory string) storage.State {
	buf, err := ioutil.ReadFile(filepath.Join(tempDirectory, storage.StateFileName))
	Expect(err).NotTo(HaveOccurred())

	var state storage.State
	err = json.Unmarshal(buf, &state)
	Expect(err).NotTo(HaveOccurred())

	return state
}
