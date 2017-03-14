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

	"github.com/cloudfoundry/bosh-bootloader/bbl/gcpbackend"
	"github.com/cloudfoundry/bosh-bootloader/storage"
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
	gcpBackend         gcpbackend.GCPBackend
	fakeGCPServer      *httptest.Server
	serviceAccountKey  string
	originalPath       string
)

var _ = BeforeSuite(func() {
	var err error
	gcpBackend = gcpbackend.GCPBackend{}
	fakeGCPServer, serviceAccountKey = gcpBackend.StartFakeGCPBackend()

	pathToBBL, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl", "--ldflags", fmt.Sprintf("-X main.gcpBasePath=%s", fakeGCPServer.URL))
	Expect(err).NotTo(HaveOccurred())

	pathToFakeBOSHInit, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakeboshinit")
	Expect(err).NotTo(HaveOccurred())

	pathToBOSHInit = filepath.Join(filepath.Dir(pathToFakeBOSHInit), "bosh-init")
	err = os.Rename(pathToFakeBOSHInit, pathToBOSHInit)
	Expect(err).NotTo(HaveOccurred())

	os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSHInit), os.Getenv("PATH")}, ":"))

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
