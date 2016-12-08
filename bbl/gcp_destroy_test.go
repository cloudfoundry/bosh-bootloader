package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl destroy gcp", func() {
	var (
		tempDirectory       string
		statePath           string
		pathToFakeTerraform string
		pathToTerraform     string
		fakeBOSHServer      *httptest.Server
		fakeBOSH            *fakeBOSHDirector
	)

	BeforeEach(func() {
		var err error

		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeTerraformBackendServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
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
				responseWriter.Write([]byte("some-tag"))
			case "/output/bosh_open_tag_name":
				responseWriter.Write([]byte("some-bosh-open-tag"))
			}
		}))

		pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
		err = os.Rename(pathToFakeTerraform, pathToTerraform)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), os.Getenv("PATH")}, ":"))

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		state := storage.State{
			IAAS:    "gcp",
			TFState: `{"key": "value"}`,
			GCP: storage.GCP{
				ProjectID:         "some-project-id",
				ServiceAccountKey: serviceAccountKey,
				Region:            "some-region",
				Zone:              "some-zone",
			},
			KeyPair: storage.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: testhelpers.BBL_KEY,
			},
			BOSH: storage.BOSH{
				DirectorName: "some-bosh-director-name",
			},
		}

		stateContents, err := json.Marshal(state)
		Expect(err).NotTo(HaveOccurred())

		statePath = filepath.Join(tempDirectory, "bbl-state.json")
		err = ioutil.WriteFile(statePath, stateContents, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	It("deletes the bbl-state", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"destroy",
		}
		cmd := exec.Command(pathToBBL, args...)

		stdin, err := cmd.StdinPipe()
		Expect(err).NotTo(HaveOccurred())

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		_, err = stdin.Write([]byte("yes\n"))
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))

		_, err = os.Stat(statePath)
		Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
	})

	It("calls out to terraform", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"destroy",
		}
		cmd := exec.Command(pathToBBL, args...)

		stdin, err := cmd.StdinPipe()
		Expect(err).NotTo(HaveOccurred())

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		_, err = stdin.Write([]byte("yes\n"))
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))

		Expect(session.Out.Contents()).To(ContainSubstring("terraform destroy"))
	})
})
