package terraform_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Run", func() {
	var (
		fakeStdout   *bytes.Buffer
		fakeStderr   *bytes.Buffer
		outputBuffer *bytes.Buffer

		defaultArgs []string

		cmd terraform.Cmd

		fakeTerraformBackendServer *ghttp.Server
		pathToTerraform            string
	)

	BeforeEach(func() {
		fakeStdout = bytes.NewBuffer([]byte{})
		fakeStderr = bytes.NewBuffer([]byte{})
		outputBuffer = bytes.NewBuffer([]byte{})

		defaultArgs = []string{"apply", "-state=/tmp/terraform.tfstate", "/tmp"}
		cmd = terraform.NewCmd(fakeStderr, io.MultiWriter(outputBuffer, GinkgoWriter), "some-terraform-dir")

		fakeTerraformBackendServer = ghttp.NewServer()

		var err error
		pathToTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/fakes/terraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL()))
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), originalPath}, ":"))
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)
	})

	Context("when the terraform server succeeds", func() {
		BeforeEach(func() {
			fakeTerraformBackendServer.AppendHandlers(
				ghttp.RespondWith(200, "we good"),
				ghttp.VerifyJSONRepresenting(defaultArgs),
			)
		})

		It("runs terraform with args", func() {
			err := cmd.Run(fakeStdout, defaultArgs, false)
			Expect(err).NotTo(HaveOccurred())

			By("does not write terraform output to stdout", func() {
				fakeStdoutContents := string(fakeStdout.Bytes())
				Expect(fakeStdoutContents).NotTo(ContainSubstring("-state=/tmp/terraform.tfstate"))
			})
		})

		It("sets TF_DATA_DIR to the provided .terraform directory", func() {
			err := cmd.Run(nil, defaultArgs, false)
			Expect(err).NotTo(HaveOccurred())

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

		Context("when called with extra envs", func() {
			It("doesn't fail.", func() {
				err := cmd.RunWithEnv(fakeStdout, defaultArgs, []string{"WHATEVER=1"}, true)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStdout).To(ContainSubstring("WHATEVER=1"))
			})
		})
	})

	Context("when terraform fails", func() {
		BeforeEach(func() {
			fakeTerraformBackendServer.AppendHandlers(ghttp.RespondWith(http.StatusInternalServerError, "intentional 500"))
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
