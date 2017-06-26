package main_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

type fakeBOSHDirector struct {
	mutex           sync.Mutex
	cloudConfig     []byte
	cloudConfigFail bool
}

func (b *fakeBOSHDirector) SetCloudConfig(cloudConfig []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.cloudConfig = cloudConfig
}

func (b *fakeBOSHDirector) GetCloudConfig() []byte {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.cloudConfig
}

func (b *fakeBOSHDirector) SetCloudConfigEndpointFail(fail bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.cloudConfigFail = fail
}

func (b *fakeBOSHDirector) GetCloudConfigEndpointFail() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.cloudConfigFail
}

func (b *fakeBOSHDirector) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/info":
		responseWriter.Write([]byte(`{
     "name": "some-bosh-director",
     "uuid": "some-uuid",
     "version": "some-version"
   }`))

		return
	case "/cloud_configs":
		if b.GetCloudConfigEndpointFail() {
			responseWriter.WriteHeader(0)
			return
		}
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		b.SetCloudConfig(buf)
		responseWriter.WriteHeader(http.StatusCreated)

		return
	default:
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
}

var (
	pathToBBL             string
	pathToFakeBOSH        string
	pathToBOSH            string
	pathToFakeTerraform   string
	pathToTerraform       string
	pathToPreMigrationBBL string

	gcpBackend    gcpbackend.GCPBackend
	fakeGCPServer *httptest.Server

	fakeTerraformBackendServer *terraformbackend.Backend
	fakeBOSHCLIBackendServer   *boshbackend.Backend

	serviceAccountKey string
	originalPath      string
	noFakesPath       string
)

func TestBbl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bbl")
}

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

	var bblBinaryLocation string
	if runtime.GOOS == "darwin" {
		bblBinaryLocation = "https://www.github.com/cloudfoundry/bosh-bootloader/releases/download/v3.2.4/bbl-v3.2.4_osx"
	} else {
		bblBinaryLocation = "https://www.github.com/cloudfoundry/bosh-bootloader/releases/download/v3.2.4/bbl-v3.2.4_linux_x86-64"
	}

	resp, err := http.Get(bblBinaryLocation)
	Expect(err).NotTo(HaveOccurred())

	f, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())

	_, err = io.Copy(f, resp.Body)
	Expect(err).NotTo(HaveOccurred())

	err = os.Chmod(f.Name(), 0700)
	Expect(err).NotTo(HaveOccurred())

	pathToPreMigrationBBL = f.Name()
})

var _ = AfterSuite(func() {
	err := os.Remove(pathToPreMigrationBBL)
	Expect(err).NotTo(HaveOccurred())

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

func upAWS(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	return upAWSWithAdditionalFlags(serverURL, tempDirectory, []string{}, exitCode)
}

func upAWSWithAdditionalFlags(serverURL string, tempDirectory string, additionalArgs []string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--state-dir", tempDirectory,
		"--debug",
		"up",
		"--iaas", "aws",
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
	}
	args = append(args, additionalArgs...)

	return executeCommand(args, exitCode)
}

func upAWSCloudFormation(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--state-dir", tempDirectory,
		"--debug",
		"up",
		"--iaas", "aws",
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
	}

	cmd := exec.Command(pathToPreMigrationBBL, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(exitCode))

	return session
}

func createLB(serverURL string, tempDirectory string, lbType string, certPath string, keyPath string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--state-dir", tempDirectory,
		"create-lbs",
		"--type", lbType,
		"--cert", certPath,
		"--key", keyPath,
	}

	return executeCommand(args, exitCode)
}
