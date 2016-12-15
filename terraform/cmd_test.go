package terraform_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Cmd", func() {
	var (
		stdout *bytes.Buffer
		stderr *bytes.Buffer

		cmd terraform.Cmd

		fakeTerraformBackendServer *httptest.Server
		pathToFakeTerraform        string
		pathToTerraform            string
		fastFailTerraform          bool
		fastFailTerraformMutex     sync.Mutex
	)

	var setFastFailTerraform = func(on bool) {
		fastFailTerraformMutex.Lock()
		defer fastFailTerraformMutex.Unlock()
		fastFailTerraform = on
	}

	var getFastFailTerraform = func() bool {
		fastFailTerraformMutex.Lock()
		defer fastFailTerraformMutex.Unlock()
		return fastFailTerraform
	}

	BeforeEach(func() {
		stdout = bytes.NewBuffer([]byte{})
		stderr = bytes.NewBuffer([]byte{})

		cmd = terraform.NewCmd(stderr)

		fakeTerraformBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			if getFastFailTerraform() {
				responseWriter.WriteHeader(http.StatusInternalServerError)
			}
		}))

		var err error
		pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
		err = os.Rename(pathToFakeTerraform, pathToTerraform)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), os.Getenv("PATH")}, ":"))
	})

	It("runs terraform with args", func() {
		err := cmd.Run(stdout, "/tmp", []string{"apply", "some-arg"})
		Expect(err).NotTo(HaveOccurred())

		Expect(stdout).To(MatchRegexp("working directory: (.*)/tmp"))
		Expect(stdout).To(ContainSubstring("apply some-arg"))
	})

	Context("failure case", func() {
		BeforeEach(func() {
			setFastFailTerraform(true)
		})

		AfterEach(func() {
			setFastFailTerraform(false)
		})

		It("returns an error when terraform fails", func() {
			err := cmd.Run(stdout, "", []string{"fast-fail"})
			Expect(err).To(MatchError("exit status 1"))

			Expect(stderr).To(ContainSubstring("failed to terraform"))
		})
	})
})
