package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("bbl cloud-config", func() {
	var (
		pathToFakeTerraform        string
		pathToTerraform            string
		pathToFakeBOSH             string
		pathToBOSH                 string
		fakeBOSHCLIBackendServer   *httptest.Server
		fakeTerraformBackendServer *httptest.Server
		fakeBOSHServer             *httptest.Server
		fakeBOSH                   *fakeBOSHDirector

		tempDirectory         string
		serviceAccountKeyPath string
		createEnvArgs         string
		interpolateArgs       []string

		callRealInterpolate      bool
		callRealInterpolateMutex sync.Mutex
	)

	BeforeEach(func() {
		var err error
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeBOSHCLIBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/path":
				responseWriter.Write([]byte(originalPath))
			case "/create-env/args":
				body, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				createEnvArgs = string(body)
			case "/interpolate/args":
				body, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				interpolateArgs = append(interpolateArgs, string(body))
			case "/call-real-interpolate":
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				if callRealInterpolate {
					responseWriter.Write([]byte("true"))
				} else {
					responseWriter.Write([]byte("false"))
				}
			}
		}))

		fakeTerraformBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/output/external_ip":
				responseWriter.Write([]byte("127.0.0.1"))
			case "/output/director_address":
				responseWriter.Write([]byte(fakeBOSHServer.URL))
			case "/output/network_name":
				responseWriter.Write([]byte("some-network-name"))
			case "/output/subnetwork_name":
				responseWriter.Write([]byte("some-subnetwork-name"))
			case "/output/internal_tag_name":
				responseWriter.Write([]byte("some-internal-tag"))
			case "/output/bosh_open_tag_name":
				responseWriter.Write([]byte("some-bosh-tag"))
			case "/version":
				responseWriter.Write([]byte("0.8.6"))
			case "/output/concourse_target_pool":
				responseWriter.Write([]byte("concourse-target-pool"))
			case "/output/router_backend_service":
				responseWriter.Write([]byte("router-backend-service"))
			case "/output/ssh_proxy_target_pool":
				responseWriter.Write([]byte("ssh-proxy-target-pool"))
			case "/output/tcp_router_target_pool":
				responseWriter.Write([]byte("tcp-router-target-pool"))
			case "/output/ws_target_pool":
				responseWriter.Write([]byte("ws-target-pool"))
			case "/output/router_lb_ip":
				responseWriter.Write([]byte("some-router-lb-ip"))
			case "/output/ssh_proxy_lb_ip":
				responseWriter.Write([]byte("some-ssh-proxy-lb-ip"))
			case "/output/tcp_router_lb_ip":
				responseWriter.Write([]byte("some-tcp-router-lb-ip"))
			case "/output/concourse_lb_ip":
				responseWriter.Write([]byte("some-concourse-lb-ip"))
			case "/output/ws_lb_ip":
				responseWriter.Write([]byte("some-ws-lb-ip"))
			}
		}))

		pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
		err = os.Rename(pathToFakeTerraform, pathToTerraform)
		Expect(err).NotTo(HaveOccurred())

		pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHCLIBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
		err = os.Rename(pathToFakeBOSH, pathToBOSH)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), filepath.Dir(pathToBOSH), originalPath}, ":"))

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		callRealInterpolateMutex.Lock()
		defer callRealInterpolateMutex.Unlock()
		callRealInterpolate = true

		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--no-director",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "some-zone",
			"--gcp-region", "us-east1",
		}

		executeCommand(args, 0)
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)

		callRealInterpolateMutex.Lock()
		defer callRealInterpolateMutex.Unlock()
		callRealInterpolate = false
	})

	It("returns the cloud config of the bbl environment", func() {
		contents, err := ioutil.ReadFile("../cloudconfig/fixtures/gcp-cloud-config-no-lb.yml")
		Expect(err).NotTo(HaveOccurred())
		args := []string{
			"--state-dir", tempDirectory,
			"cloud-config",
		}

		session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(MatchYAML(string(contents)))
	})
})
