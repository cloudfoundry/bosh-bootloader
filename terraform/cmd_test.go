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

		defaultArgs []string

		cmd terraform.Cmd

		fakeTerraformBackendServer *httptest.Server
		pathToTerraform            string
		fastFailTerraform          bool
		fastFailMtx                sync.Mutex

		terraformArgs []string
	)

	BeforeEach(func() {
		stdout = bytes.NewBuffer([]byte{})
		stderr = bytes.NewBuffer([]byte{})
		outputBuffer = bytes.NewBuffer([]byte{})

		defaultArgs = []string{"apply", "-state=/tmp/terraform.tfstate", "/tmp"}

		cmd = terraform.NewCmd(stderr, outputBuffer)

		fakeTerraformBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fastFailMtx.Lock()
			if fastFailTerraform {
				responseWriter.WriteHeader(http.StatusInternalServerError)
			}
			fastFailMtx.Unlock()

			if request.Method == "POST" {
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
		pathToTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/fakes/terraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), originalPath}, ":"))
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)
	})

	It("runs terraform with args", func() {
		err := cmd.Run(stdout, defaultArgs, false)
		Expect(err).NotTo(HaveOccurred())

		Expect(terraformArgs).To(Equal([]string{"apply", "-state=/tmp/terraform.tfstate", "/tmp"}))
	})

	It("redirects command stdout to the provided buffer", func() {
		err := cmd.Run(nil, defaultArgs, false)
		Expect(err).NotTo(HaveOccurred())

		Expect(terraformArgs).To(Equal([]string{"apply", "-state=/tmp/terraform.tfstate", "/tmp"}))
	})

	Context("when debug is true", func() {
		It("redirects command stdout to provided stdout", func() {
			err := cmd.Run(stdout, defaultArgs, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout).To(ContainSubstring("apply -state=/tmp/terraform.tfstate /tmp"))
		})
	})

	Context("when terraform fails", func() {
		BeforeEach(func() {
			fastFailMtx.Lock()
			fastFailTerraform = true
			fastFailMtx.Unlock()
		})

		It("returns an error and redirects command stderr to the provided buffer", func() {
			err := cmd.Run(stdout, []string{"-state=/tmp/terraform.tfstate", "fast-fail"}, false)
			Expect(err).To(MatchError("exit status 1"))

			outputBufferContents := string(outputBuffer.Bytes())
			Expect(outputBufferContents).To(ContainSubstring("failed to terraform"))
		})

		Context("when debug is true", func() {
			It("redirects command stderr to provided stderr and buffer", func() {
				_ = cmd.Run(stdout, []string{"-state=/tmp/terraform.tfstate", "fast-fail"}, true)
				Expect(stderr).To(ContainSubstring("failed to terraform"))

				outputBufferContents := string(outputBuffer.Bytes())
				Expect(outputBufferContents).To(ContainSubstring("failed to terraform"))
			})
		})
	})
})
