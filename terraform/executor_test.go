package terraform_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	var (
		cmd      *fakes.TerraformCmd
		executor terraform.Executor

		tempDir string
		input   map[string]string
	)

	BeforeEach(func() {
		cmd = &fakes.TerraformCmd{}

		executor = terraform.NewExecutor(cmd, true)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		terraform.SetTempDir(func(dir, prefix string) (string, error) {
			return tempDir, nil
		})

		terraform.SetReadFile(func(filename string) ([]byte, error) {
			return []byte{}, nil
		})

		input = map[string]string{
			"env_id":                      "some-env-id",
			"project_id":                  "some-project-id",
			"region":                      "some-region",
			"zone":                        "some-zone",
			"credentials":                 "some/credentials/path",
			"system_domain":               "some-domain",
			"ssl_certificate":             "some/certificate/path",
			"ssl_certificate_private_key": "some/key/path",
		}
	})

	AfterEach(func() {
		terraform.ResetTempDir()
		terraform.ResetReadFile()
		terraform.ResetWriteFile()
	})

	Describe("Apply", func() {
		It("writes the terraform template to a file", func() {
			_, err := executor.Apply(input, "some-template", "")
			Expect(err).NotTo(HaveOccurred())

			fileContents, err := ioutil.ReadFile(filepath.Join(tempDir, "template.tf"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(fileContents)).To(Equal("some-template"))
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Apply(input, "some-template", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
			Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
				"apply",
				"-var", "project_id=some-project-id",
				"-var", "env_id=some-env-id",
				"-var", "region=some-region",
				"-var", "zone=some-zone",
				"-var", "ssl_certificate=some/certificate/path",
				"-var", "ssl_certificate_private_key=some/key/path",
				"-var", "credentials=some/credentials/path",
				"-var", "system_domain=some-domain",
			}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
		})

		It("reads and returns the terraform state written by the command", func() {
			var actualFilename string

			terraform.SetReadFile(func(filename string) ([]byte, error) {
				actualFilename = filename
				return []byte("some-terraform-state"), nil
			})

			terraformState, err := executor.Apply(input, "some-template", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(actualFilename).To(ContainSubstring("terraform.tfstate"))
			Expect(terraformState).To(Equal("some-terraform-state"))
		})

		Context("when previous tf state is blank", func() {
			var (
				writeTFStateFileCallCount int
			)

			BeforeEach(func() {
				cmd.RunCall.Stub = func(stdout io.Writer) {
					err := ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("some-tfstate"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
				}

				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if file == "terraform.tfstate" {
						writeTFStateFileCallCount++
					}
					return nil
				})
			})

			AfterEach(func() {
				terraform.ResetWriteFile()
			})

			It("does not write the previous tf state file", func() {
				_, err := executor.Apply(input, "some-template", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(writeTFStateFileCallCount).To(Equal(0))
			})
		})

		Context("when previous tf state is not blank", func() {
			It("writes the tf state to a file", func() {
				_, err := executor.Apply(input, "some-template", "some-tf-state")
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
				_, err := executor.Apply(input, "some-template", "")
				Expect(err).To(MatchError("failed to make temp dir"))
			})

			It("returns an error when it fails to write the template file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if file == filepath.Join(tempDir, "template.tf") {
						return errors.New("failed to write template file")
					}

					return nil
				})

				_, err := executor.Apply(input, "some-template", "")
				Expect(err).To(MatchError("failed to write template file"))
			})

			It("returns an error when it fails to write the previous tfstate file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if file == filepath.Join(tempDir, "terraform.tfstate") {
						return errors.New("failed to write tf state file")
					}

					return nil
				})

				_, err := executor.Apply(input, "some-template", "some-tf-state")
				Expect(err).To(MatchError("failed to write tf state file"))
			})

			It("returns an error and the current tf state when it fails to call terraform command run", func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("some-tf-state"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

				_, err = executor.Apply(input, "some-template", "")
				taErr := err.(terraform.ExecutorError)
				Expect(taErr).To(MatchError("failed to run terraform command"))

				tfState, err := taErr.TFState()
				Expect(err).NotTo(HaveOccurred())
				Expect(tfState).To(Equal("some-tf-state"))
			})

			It("returns an error when it fails to read the tf state file", func() {
				terraform.SetReadFile(func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read tf state file")
				})

				_, err := executor.Apply(input, "some-template", "")
				Expect(err).To(MatchError("failed to read tf state file"))
			})

			Context("when --debug is false", func() {
				BeforeEach(func() {
					executor = terraform.NewExecutor(cmd, false)
				})

				It("returns an error and the current tf state when it fails to call terraform command run", func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("some-tf-state"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

					_, err = executor.Apply(input, "some-template", "")
					taErr := err.(terraform.ExecutorError)

					tfState, err := taErr.TFState()
					Expect(err).NotTo(HaveOccurred())
					Expect(tfState).To(Equal("some-tf-state"))
				})
			})
		})
	})

	Describe("Destroy", func() {
		It("writes the template and tf state to a temp dir", func() {
			_, err := executor.Destroy(input, "some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			templateContents, err := ioutil.ReadFile(filepath.Join(tempDir, "template.tf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(templateContents)).To(Equal("some-template"))

			tfStateContents, err := ioutil.ReadFile(filepath.Join(tempDir, "terraform.tfstate"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(tfStateContents)).To(Equal("some-tf-state"))
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Destroy(input, "some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
			Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
				"destroy",
				"-force",
				"-var", "project_id=some-project-id",
				"-var", "env_id=some-env-id",
				"-var", "region=some-region",
				"-var", "zone=some-zone",
				"-var", "ssl_certificate=some/certificate/path",
				"-var", "ssl_certificate_private_key=some/key/path",
				"-var", "credentials=some/credentials/path",
				"-var", "system_domain=some-domain",
			}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
		})

		It("reads and returns the tf state", func() {
			terraform.SetReadFile(func(filename string) ([]byte, error) {
				return []byte{}, nil
			})

			tfState, err := executor.Destroy(input, "some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			Expect(tfState).To(Equal(""))
		})

		Context("failure cases", func() {
			It("returns an error when it fails to create a temp dir", func() {
				terraform.SetTempDir(func(dir, prefix string) (string, error) {
					return "", errors.New("failed to make temp dir")
				})

				_, err := executor.Destroy(input, "some-template", "")
				Expect(err).To(MatchError("failed to make temp dir"))
			})

			It("returns an error when it fails to write the template file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "template.tf") {
						return errors.New("failed to write template file")
					}

					return nil
				})

				_, err := executor.Destroy(input, "some-template", "")
				Expect(err).To(MatchError("failed to write template file"))
			})

			It("returns an error when it fails to write the tfstate file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "terraform.tfstate") {
						return errors.New("failed to write tf state file")
					}

					return nil
				})

				_, err := executor.Destroy(input, "some-template", "some-tf-state")
				Expect(err).To(MatchError("failed to write tf state file"))
			})

			It("returns an error and the current tf state when it fails to call terraform command run", func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("some-tf-state"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

				_, err = executor.Destroy(input, "some-template", "")
				tdErr := err.(terraform.ExecutorError)
				Expect(tdErr).To(MatchError("failed to run terraform command"))

				tfState, err := tdErr.TFState()
				Expect(err).NotTo(HaveOccurred())
				Expect(tfState).To(Equal("some-tf-state"))
			})

			It("returns an error when it fails to read out the resulting tf state", func() {
				terraform.SetReadFile(func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read tf state file")
				})

				_, err := executor.Destroy(input, "some-template", "")
				Expect(err).To(MatchError("failed to read tf state file"))
			})

			Context("when --debug is false", func() {
				BeforeEach(func() {
					executor = terraform.NewExecutor(cmd, false)
				})

				It("returns an error and the current tf state when it fails to call terraform command run", func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("some-tf-state"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

					_, err = executor.Destroy(input, "some-template", "")
					tdErr := err.(terraform.ExecutorError)

					tfState, err := tdErr.TFState()
					Expect(tfState).To(Equal("some-tf-state"))
				})
			})
		})
	})

	Describe("Import", func() {
		var (
			receivedTFState string
		)

		BeforeEach(func() {
			cmd.RunCall.Stub = func(stdout io.Writer) {
				fileContents, err := ioutil.ReadFile(filepath.Join(tempDir, "terraform.tfstate"))
				Expect(err).NotTo(HaveOccurred())
				receivedTFState = string(fileContents)

				err = ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte("some-other-tfstate"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			}

			terraform.ResetReadFile()
		})

		It("writes the tfState to a file", func() {
			_, err := executor.Import("some-addr", "some-id", "some-tf-state", storage.AWS{})
			Expect(err).NotTo(HaveOccurred())

			Expect(receivedTFState).To(Equal("some-tf-state"))
		})

		It("shells out to terraform import and returns the tfState", func() {
			tfState, err := executor.Import("some-addr", "some-id", "some-tf-state", storage.AWS{})
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"import", "some-addr", "some-id"}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
			Expect(tfState).To(Equal("some-other-tfstate"))
		})

		Context("failure cases", func() {
			Context("when it fails to create a temp dir", func() {
				It("returns an error", func() {
					terraform.SetTempDir(func(dir, prefix string) (string, error) {
						return "", errors.New("failed to make temp dir")
					})
					_, err := executor.Import("addr", "id", "some-tf-state", storage.AWS{})
					Expect(err).To(MatchError("failed to make temp dir"))
				})
			})

			Context("when the file cannot be written", func() {
				It("returns an error", func() {
					terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
						if strings.Contains(file, "terraform.tfstate") {
							return errors.New("failed to write tfstate")
						}

						return nil
					})

					_, err := executor.Import("addr", "id", "some-tf-state", storage.AWS{})
					Expect(err).To(MatchError("failed to write tfstate"))
				})
			})

			Context("when the command fails to run", func() {
				It("returns an error", func() {
					cmd.RunCall.Returns.Error = errors.New("bad import")

					_, err := executor.Import("addr", "id", "some-tf-state", storage.AWS{})
					Expect(err).To(MatchError("failed to import: bad import"))
				})
			})

			Context("when the state cannot be read", func() {
				It("returns an error", func() {
					terraform.SetReadFile(func(filePath string) ([]byte, error) {
						if strings.Contains(filePath, "terraform.tfstate") {
							return []byte{}, errors.New("failed to read tf state file")
						}

						return []byte{}, nil
					})

					_, err := executor.Import("addr", "id", "some-tf-state", storage.AWS{})
					Expect(err).To(MatchError("failed to read tf state file"))
				})
			})
		})
	})

	Describe("Version", func() {
		BeforeEach(func() {
			cmd.RunCall.Stub = func(stdout io.Writer) {
				stdout.Write([]byte("some-text v0.8.9 some-other-text"))
			}
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"version"}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
		})

		It("returns the correctly trimmed version", func() {
			version, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("0.8.9"))
		})

		Context("failure cases", func() {
			It("returns an error when the run command fails", func() {
				cmd.RunCall.Returns.Error = errors.New("run cmd failed")
				_, err := executor.Version()
				Expect(err).To(MatchError("run cmd failed"))
			})

			It("returns an error when the version cannot be parsed", func() {
				cmd.RunCall.Stub = func(stdout io.Writer) {
					stdout.Write([]byte(""))
				}
				_, err := executor.Version()
				Expect(err).To(MatchError("Terraform version could not be parsed"))
			})
		})
	})

	Describe("Output", func() {
		It("returns an output from the terraform state", func() {
			cmd.RunCall.Stub = func(stdout io.Writer) {
				fmt.Fprintf(stdout, "some-external-ip\n")
			}
			output, err := executor.Output("some-tf-state", "external_ip")
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
				_, err := executor.Output("some-tf-state", "external_ip")
				Expect(err).To(MatchError("failed to make temp dir"))
			})

			It("returns an error when it fails to write the tfstate file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "terraform.tfstate") {
						return errors.New("failed to write tf state file")
					}

					return nil
				})

				_, err := executor.Output("some-tf-state", "external_ip")
				Expect(err).To(MatchError("failed to write tf state file"))
			})

			It("returns an error when it fails to call terraform command run", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

				_, err := executor.Output("some-tf-state", "external_ip")
				Expect(err).To(MatchError("failed to run terraform command"))
			})
		})
	})

	Describe("Outputs", func() {
		It("returns all outputs from the terraform state", func() {
			cmd.RunCall.Stub = func(stdout io.Writer) {
				fmt.Fprintf(stdout, `{
					"director_address": {
						"sensitive": false,
						"type": "string",
						"value": "some-director-address"
					},
					"external_ip": {
						"sensitive": false,
						"type": "string",
						"value": "some-external-ip"
					}
				}`)
			}
			outputs, err := executor.Outputs("some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			Expect(outputs).To(Equal(map[string]interface{}{
				"director_address": "some-director-address",
				"external_ip":      "some-external-ip",
			}))

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"output", "--json"}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
		})

		Context("failure cases", func() {
			It("returns an error when it fails to create a temp dir", func() {
				terraform.SetTempDir(func(dir, prefix string) (string, error) {
					return "", errors.New("failed to make temp dir")
				})
				_, err := executor.Outputs("some-tf-state")
				Expect(err).To(MatchError("failed to make temp dir"))
			})

			It("returns an error when it fails to write the tfstate file", func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					if strings.Contains(file, "terraform.tfstate") {
						return errors.New("failed to write tf state file")
					}

					return nil
				})

				_, err := executor.Outputs("some-tf-state")
				Expect(err).To(MatchError("failed to write tf state file"))
			})

			It("returns an error when it fails to call terraform command run", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")

				_, err := executor.Outputs("some-tf-state")
				Expect(err).To(MatchError("failed to run terraform command"))
			})

			It("returns an error when it fails to unmarshal the terraform outputs", func() {
				cmd.RunCall.Stub = func(stdout io.Writer) {
					fmt.Fprintf(stdout, "%%%")
				}
				_, err := executor.Outputs("some-tf-state")
				Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
			})
		})
	})
})
