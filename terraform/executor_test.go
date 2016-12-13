package terraform_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	var (
		cmd      *fakes.TerraformCmd
		executor terraform.Executor
		tempDir  string
	)

	BeforeEach(func() {
		cmd = &fakes.TerraformCmd{}

		executor = terraform.NewExecutor(cmd)

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

	Describe("Apply", func() {
		It("writes the terraform template to a file", func() {
			_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
			Expect(err).NotTo(HaveOccurred())

			fileContents, err := ioutil.ReadFile(filepath.Join(tempDir, "template.tf"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(fileContents)).To(Equal("some-template"))
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{
				"apply",
				"-var", "project_id=some-project-id",
				"-var", "env_id=some-env-id",
				"-var", "region=some-region",
				"-var", "zone=some-zone",
				"-var", fmt.Sprintf("credentials=%s/credentials.json", tempDir),
			}))
		})

		It("reads and returns the terraform state written by the command", func() {
			var actualFilename string

			terraform.SetReadFile(func(filename string) ([]byte, error) {
				actualFilename = filename
				return []byte("some-terraform-state"), nil
			})

			terraformState, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(actualFilename).To(ContainSubstring("terraform.tfstate"))
			Expect(terraformState).To(Equal("some-terraform-state"))
		})

		Context("when previous tf state is blank", func() {
			It("does not write the previous tf state file", func() {
				_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(filepath.Join(tempDir, "terraform.tfstate"))
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		Context("when previous tf state is not blank", func() {
			It("writes the tf state to a file", func() {
				_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "some-tf-state")
				Expect(err).NotTo(HaveOccurred())

				fileContents, err := ioutil.ReadFile(filepath.Join(tempDir, "terraform.tfstate"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(fileContents)).To(Equal("some-tf-state"))
			})
		})

		Context("failure case", func() {
			It("returns an error when it fails to create a temp dir", func() {
				terraform.SetTempDir(func(dir, prefix string) (string, error) {
					return "", errors.New("failed to make temp dir")
				})
				_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to make temp dir"))
			})

			It("returns an error when it fails to write the template file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "template.tf") {
						return errors.New("failed to write template file")
					}

					return nil
				})

				_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to write template file"))
			})

			It("returns an error when it fails to write the previous tfstate file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "terraform.tfstate") {
						return errors.New("failed to write tf state file")
					}

					return nil
				})

				_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "some-tf-state")
				Expect(err).To(MatchError("failed to write tf state file"))
			})

			It("returns an error and the current tf state when it fails to call terraform command run", func() {
				terraform.SetReadFile(func(string) ([]byte, error) {
					return []byte("some-tf-state"), nil
				})
				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

				tfState, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to run terraform command"))
				Expect(tfState).To(Equal("some-tf-state"))
			})

			It("returns an error when it fails to call terraform command run and read out the resulting tf state", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")
				terraform.SetReadFile(func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read tf state file")
				})

				_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("the following errors occurred:\nfailed to run terraform command,\nfailed to read tf state file"))
			})

			It("returns an error when it fails to read the tf state file", func() {
				terraform.SetReadFile(func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read tf state file")
				})

				_, err := executor.Apply("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to read tf state file"))
			})
		})
	})

	Describe("Destroy", func() {
		It("writes the template and tf state to a temp dir", func() {
			_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region",
				"some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			templateContents, err := ioutil.ReadFile(filepath.Join(tempDir, "template.tf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(templateContents)).To(Equal("some-template"))

			tfStateContents, err := ioutil.ReadFile(filepath.Join(tempDir, "terraform.tfstate"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(tfStateContents)).To(Equal("some-tf-state"))
		})

		It("writes credentials to a file", func() {
			_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region",
				"some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			templateContents, err := ioutil.ReadFile(filepath.Join(tempDir, "credentials.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(templateContents)).To(Equal("some-credentials-json"))
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region",
				"some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{
				"destroy",
				"-force",
				"-var", "project_id=some-project-id",
				"-var", "env_id=some-env-id",
				"-var", "region=some-region",
				"-var", "zone=some-zone",
				"-var", fmt.Sprintf("credentials=%s/credentials.json", tempDir),
			}))
		})

		It("reads and returns the tf state", func() {
			terraform.SetReadFile(func(filename string) ([]byte, error) {
				return []byte{}, nil
			})

			tfState, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region",
				"some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			Expect(tfState).To(Equal(""))
		})

		Context("failure cases", func() {
			It("returns an error when it fails to create a temp dir", func() {
				terraform.SetTempDir(func(dir, prefix string) (string, error) {
					return "", errors.New("failed to make temp dir")
				})

				_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to make temp dir"))
			})

			It("returns an error when it fails to write the credentials file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "credentials.json") {
						return errors.New("failed to write credentials file")
					}

					return nil
				})

				_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to write credentials file"))
			})

			It("returns an error when it fails to write the template file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "template.tf") {
						return errors.New("failed to write template file")
					}

					return nil
				})

				_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to write template file"))
			})

			It("returns an error when it fails to write the tfstate file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "terraform.tfstate") {
						return errors.New("failed to write tf state file")
					}

					return nil
				})

				_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "some-tf-state")
				Expect(err).To(MatchError("failed to write tf state file"))
			})

			It("returns an error and the current tf state when it fails to call terraform command run", func() {
				terraform.SetReadFile(func(filename string) ([]byte, error) {
					return []byte("some-tf-state"), nil
				})
				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

				tfState, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to run terraform command"))
				Expect(tfState).To(Equal("some-tf-state"))
			})

			It("returns an error when it fails to call terraform command run and read out the resulting tf state", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")
				terraform.SetReadFile(func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read tf state file")
				})

				_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("the following errors occurred:\nfailed to run terraform command,\nfailed to read tf state file"))
			})

			It("returns an error when it fails to read out the resulting tf state", func() {
				terraform.SetReadFile(func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read tf state file")
				})

				_, err := executor.Destroy("some-credentials-json", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template", "")
				Expect(err).To(MatchError("failed to read tf state file"))
			})

		})
	})
})
