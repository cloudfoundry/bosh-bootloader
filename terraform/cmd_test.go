package terraform_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

var _ = Describe("Run", func() {
	var (
		stdout       *bytes.Buffer
		stderr       *bytes.Buffer
		outputBuffer *bytes.Buffer

		cmd terraform.Cmd

		fakeTerraformBackendServer *httptest.Server
		pathToFakeTerraform        string
		pathToTerraform            string
		fastFailTerraform          bool
		fastFailTerraformMutex     sync.Mutex

		terraformArgs      []string
		terraformArgsMutex sync.Mutex
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
		outputBuffer = bytes.NewBuffer([]byte{})

		cmd = terraform.NewCmd(stderr, outputBuffer)

		fakeTerraformBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			if getFastFailTerraform() {
				responseWriter.WriteHeader(http.StatusInternalServerError)
			}

			if request.Method == "POST" {
				terraformArgsMutex.Lock()
				defer terraformArgsMutex.Unlock()
				body, err := ioutil.ReadAll(request.Body)
				if err != nil {
					panic(err)
				}

				err = json.Unmarshal(body, &terraformArgs)
				if err != nil {
					panic(err)
				}
			}
		}))

		var err error
		pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
		err = os.Rename(pathToFakeTerraform, pathToTerraform)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), originalPath}, ":"))
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)
	})

	It("runs terraform with args", func() {
		err := cmd.Run(stdout, "/tmp", []string{"apply", "some-arg"}, false)
		Expect(err).NotTo(HaveOccurred())

		terraformArgsMutex.Lock()
		defer terraformArgsMutex.Unlock()
		Expect(terraformArgs).To(Equal([]string{"apply", "some-arg"}))

		Expect(stdout).NotTo(MatchRegexp("working directory: (.*)/tmp"))
		Expect(stdout).NotTo(ContainSubstring("apply some-arg"))
	})

	It("redirects command stdout to the provided buffer", func() {
		err := cmd.Run(nil, "/tmp", []string{"apply", "some-arg"}, false)
		Expect(err).NotTo(HaveOccurred())

		terraformArgsMutex.Lock()
		defer terraformArgsMutex.Unlock()
		Expect(terraformArgs).To(Equal([]string{"apply", "some-arg"}))

		outputBufferContents := string(outputBuffer.Bytes())
		Expect(outputBufferContents).To(MatchRegexp("working directory: (.*)/tmp"))
		Expect(outputBufferContents).To(ContainSubstring("apply some-arg"))
	})

	It("redirects command stdout to provided stdout when debug is true", func() {
		err := cmd.Run(stdout, "/tmp", []string{"apply", "some-arg"}, true)
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

		It("returns an error and redirects command stderr to the provided buffer when terraform fails", func() {
			err := cmd.Run(stdout, "", []string{"fast-fail"}, false)
			Expect(err).To(MatchError("exit status 1"))

			outputBufferContents := string(outputBuffer.Bytes())
			Expect(outputBufferContents).To(ContainSubstring("failed to terraform"))
		})

		It("redirects command stderr to provided stderr and buffer when debug is true", func() {
			_ = cmd.Run(stdout, "", []string{"fast-fail"}, true)
			Expect(stderr).To(ContainSubstring("failed to terraform"))

			outputBufferContents := string(outputBuffer.Bytes())
			Expect(outputBufferContents).To(ContainSubstring("failed to terraform"))
		})
	})
})
