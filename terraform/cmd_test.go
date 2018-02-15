package terraform_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
		fakeStdout   *bytes.Buffer
		fakeStderr   *bytes.Buffer
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
		fakeStdout = bytes.NewBuffer([]byte{})
		fakeStderr = bytes.NewBuffer([]byte{})
		outputBuffer = bytes.NewBuffer([]byte{})

		defaultArgs = []string{"apply", "-state=/tmp/terraform.tfstate", "/tmp"}

		cmd = terraform.NewCmd(fakeStderr, io.MultiWriter(outputBuffer, GinkgoWriter), "some-terraform-dir")

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
		err := cmd.Run(fakeStdout, defaultArgs, false)
		Expect(err).NotTo(HaveOccurred())

		Expect(terraformArgs).To(Equal([]string{"apply", "-state=/tmp/terraform.tfstate", "/tmp"}))

		By("does not write terraform output to stdout", func() {
			fakeStdoutContents := string(fakeStdout.Bytes())
			Expect(fakeStdoutContents).NotTo(ContainSubstring("-state=/tmp/terraform.tfstate"))
		})
	})

	It("sets TF_DATA_DIR to the provided .terraform directory", func() {
		err := cmd.Run(nil, defaultArgs, false)
		Expect(err).NotTo(HaveOccurred())

		Expect(terraformArgs).To(Equal([]string{"apply", "-state=/tmp/terraform.tfstate", "/tmp"}))

		outputBufferContents := string(outputBuffer.Bytes())
		Expect(outputBufferContents).To(ContainSubstring("data directory: some-terraform-dir"))
	})

	Context("when debug is true", func() {
		It("redirects command stdout to provided stdout", func() {
			err := cmd.Run(fakeStdout, defaultArgs, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStdout).To(ContainSubstring("apply -state=/tmp/terraform.tfstate /tmp"))
		})
	})

	Context("when terraform fails", func() {
		BeforeEach(func() {
			fastFailMtx.Lock()
			fastFailTerraform = true
			fastFailMtx.Unlock()
		})

		It("returns an error and redirects command stderr to the provided buffer", func() {
			err := cmd.Run(fakeStdout, []string{"-state=/tmp/terraform.tfstate", "fast-fail"}, false)
			Expect(err).To(MatchError("exit status 1"))

			outputBufferContents := string(outputBuffer.Bytes())
			Expect(outputBufferContents).To(ContainSubstring("failed to terraform"))
		})

		Context("when debug is true", func() {
			It("redirects command stderr to provided stderr and buffer", func() {
				_ = cmd.Run(fakeStdout, []string{"-state=/tmp/terraform.tfstate", "fast-fail"}, true)
				Expect(fakeStderr).To(ContainSubstring("failed to terraform"))

				outputBufferContents := string(outputBuffer.Bytes())
				Expect(outputBufferContents).To(ContainSubstring("failed to terraform"))
			})
		})
	})
})
