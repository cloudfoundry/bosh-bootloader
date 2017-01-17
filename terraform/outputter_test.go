package terraform_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Outputter", func() {
	var (
		cmd       *fakes.TerraformCmd
		outputter terraform.Outputter
		tempDir   string
	)

	BeforeEach(func() {
		cmd = &fakes.TerraformCmd{}

		outputter = terraform.NewOutputter(cmd)

		terraform.SetTempDir(func(dir, prefix string) (string, error) {
			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			return tempDir, nil
		})

		terraform.SetReadFile(func(string) ([]byte, error) {
			return []byte(""), nil
		})
	})

	AfterEach(func() {
		terraform.ResetTempDir()
		terraform.ResetReadFile()
		terraform.ResetWriteFile()
	})

	It("returns an output from the terraform state", func() {
		cmd.RunCall.Stub = func(stdout io.Writer) {
			fmt.Fprintf(stdout, "some-external-ip\n")
		}
		output, err := outputter.Get("some-tf-state", "external_ip")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal("some-external-ip"))

		Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
		Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"output", "external_ip"}))
		Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
	})

	Context("failure cases", func() {
		It("returns an error when it fails to create a temp dir", func() {
			terraform.SetTempDir(func(dir, prefix string) (string, error) {
				return "", errors.New("failed to make temp dir")
			})
			_, err := outputter.Get("some-tf-state", "external_ip")
			Expect(err).To(MatchError("failed to make temp dir"))
		})

		It("returns an error when it fails to write the tfstate file", func() {
			terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
				if strings.Contains(file, "terraform.tfstate") {
					return errors.New("failed to write tf state file")
				}

				return nil
			})

			_, err := outputter.Get("some-tf-state", "external_ip")
			Expect(err).To(MatchError("failed to write tf state file"))
		})

		It("returns an error when it fails to call terraform command run", func() {
			cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

			_, err := outputter.Get("some-tf-state", "external_ip")
			Expect(err).To(MatchError("failed to run terraform command"))
		})
	})
})
