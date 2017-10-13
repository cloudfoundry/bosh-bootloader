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
		cmd        *fakes.TerraformCmd
		stateStore *fakes.StateStore
		executor   terraform.Executor

		tempDir      string
		terraformDir string
		varsDir      string
		input        map[string]string

		tfStatePath       string
		relativeStatePath string
	)

	BeforeEach(func() {
		cmd = &fakes.TerraformCmd{}
		stateStore = &fakes.StateStore{}

		executor = terraform.NewExecutor(cmd, stateStore, true)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		terraform.SetReadFile(func(filename string) ([]byte, error) {
			return []byte{}, nil
		})

		terraformDir, err = ioutil.TempDir("", "terraform")
		Expect(err).NotTo(HaveOccurred())
		stateStore.GetTerraformDirCall.Returns.Directory = terraformDir

		varsDir, err = ioutil.TempDir("", "vars")
		Expect(err).NotTo(HaveOccurred())
		stateStore.GetVarsDirCall.Returns.Directory = varsDir

		tfStatePath = filepath.Join(varsDir, "terraform.tfstate")
		relativeStatePath, err = filepath.Rel(terraformDir, tfStatePath)
		Expect(err).NotTo(HaveOccurred())

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
		terraform.ResetReadFile()
		terraform.ResetWriteFile()
	})

	Describe("Apply", func() {
		It("writes the terraform template to a file", func() {
			_, err := executor.Apply(input, "some-template", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.GetTerraformDirCall.CallCount).To(Equal(1))

			fileContents, err := ioutil.ReadFile(filepath.Join(terraformDir, "template.tf"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(fileContents)).To(Equal("some-template"))
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Apply(input, "some-template", "")
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(terraformDir))
			Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
				"apply",
				"-state", relativeStatePath,
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
			var writeTFStateFileCallCount int

			BeforeEach(func() {
				cmd.RunCall.Stub = func(stdout io.Writer) {
					err := ioutil.WriteFile(tfStatePath, []byte("some-tfstate"), os.ModePerm)
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

				fileContents, err := ioutil.ReadFile(tfStatePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(fileContents)).To(Equal("some-tf-state"))
			})
		})

		Context("when an error occurs", func() {
			Context("when getting terraform dir fails", func() {
				BeforeEach(func() {
					stateStore.GetTerraformDirCall.Returns.Error = errors.New("canteloupe")
				})

				It("returns an error", func() {
					_, err := executor.Apply(input, "some-template", "")
					Expect(err).To(MatchError("Get terraform dir: canteloupe"))
				})
			})

			Context("when writing the template file fails", func() {
				BeforeEach(func() {
					terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
						if file == filepath.Join(terraformDir, "template.tf") {
							return errors.New("pear")
						}
						return nil
					})
				})

				It("returns an error", func() {
					_, err := executor.Apply(input, "some-template", "")
					Expect(err).To(MatchError("Write terraform template: pear"))
				})
			})

			Context("when getting vars dir fails", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("coconut")
				})

				It("returns an error", func() {
					_, err := executor.Apply(input, "", "")
					Expect(err).To(MatchError("Get vars dir: coconut"))
				})
			})

			Context("when writing the previous tfstate file fails", func() {
				BeforeEach(func() {
					terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
						if file == tfStatePath {
							return errors.New("peach")
						}
						return nil
					})
				})

				It("returns an error", func() {
					_, err := executor.Apply(input, "some-template", "some-tf-state")
					Expect(err).To(MatchError("Write previous terraform state: peach"))
				})
			})

			Context("when terraform init fails", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{errors.New("guava")}
				})

				It("returns an error", func() {
					_, err := executor.Apply(input, "some-template", "")
					Expect(err).To(MatchError("Run terraform init: guava"))
				})
			})

			Context("when terraform command run fails", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					cmd.RunCall.Returns.Errors = []error{nil, errors.New("the-executor-error")}
				})

				It("returns an error and the current tf state", func() {
					_, err := executor.Apply(input, "some-template", "")
					taErr := err.(terraform.ExecutorError)
					Expect(taErr).To(MatchError("the-executor-error"))

					tfState, err := taErr.TFState()
					Expect(err).NotTo(HaveOccurred())
					Expect(tfState).To(Equal("some-tf-state"))
				})
			})

			Context("when reading the tfstate file fails", func() {
				BeforeEach(func() {
					terraform.SetReadFile(func(filename string) ([]byte, error) {
						return []byte{}, errors.New("lychee")
					})
				})

				It("returns an error", func() {
					_, err := executor.Apply(input, "some-template", "")
					Expect(err).To(MatchError("Read terraform state: lychee"))
				})
			})

			Context("when --debug is false", func() {
				BeforeEach(func() {
					executor = terraform.NewExecutor(cmd, stateStore, false)
				})

				Context("when terraform command run fails", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), os.ModePerm)
						Expect(err).NotTo(HaveOccurred())

						cmd.RunCall.Returns.Errors = []error{nil, errors.New("failed to run terraform command")}
					})

					It("returns an error and the current tf state", func() {
						_, err := executor.Apply(input, "some-template", "")
						taErr := err.(terraform.ExecutorError)

						tfState, err := taErr.TFState()
						Expect(err).NotTo(HaveOccurred())
						Expect(tfState).To(Equal("some-tf-state"))
					})
				})
			})
		})
	})

	Describe("Destroy", func() {
		It("writes the template and tf state to a temp dir", func() {
			_, err := executor.Destroy(input, "some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			templateContents, err := ioutil.ReadFile(filepath.Join(terraformDir, "template.tf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(templateContents)).To(Equal("some-template"))

			tfStateContents, err := ioutil.ReadFile(tfStatePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(tfStateContents)).To(Equal("some-tf-state"))
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Destroy(input, "some-template", "some-tf-state")
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(terraformDir))
			Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
				"destroy",
				"-force",
				"-state", relativeStatePath,
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

		Context("when an error occurs", func() {
			Context("when getting terraform dir fails", func() {
				BeforeEach(func() {
					stateStore.GetTerraformDirCall.Returns.Error = errors.New("kiwi")
				})

				It("returns an error", func() {
					_, err := executor.Destroy(input, "some-template", "")
					Expect(err).To(MatchError("Get terraform dir: kiwi"))
				})
			})

			Context("when getting vars dir fails", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("banana")
				})

				It("returns an error", func() {
					_, err := executor.Destroy(input, "some-template", "")
					Expect(err).To(MatchError("Get vars dir: banana"))
				})
			})

			Context("when writing the template file fails", func() {
				BeforeEach(func() {
					terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
						if strings.Contains(file, "template.tf") {
							return errors.New("nectarine")
						}
						return nil
					})
				})

				It("returns an error", func() {
					_, err := executor.Destroy(input, "some-template", "")
					Expect(err).To(MatchError("Write terraform template: nectarine"))
				})
			})

			Context("when writing the tfstate file fails", func() {
				BeforeEach(func() {
					terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
						if strings.Contains(file, "terraform.tfstate") {
							return errors.New("grape")
						}
						return nil
					})
				})

				It("returns an error", func() {
					_, err := executor.Destroy(input, "some-template", "some-tf-state")
					Expect(err).To(MatchError("Write previous terraform state: grape"))
				})
			})

			Context("when terraform init fails", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{errors.New("coconut")}
				})

				It("returns an error", func() {
					_, err := executor.Destroy(input, "some-template", "")
					Expect(err).To(MatchError("Run terraform init: coconut"))
				})
			})

			Context("when it fails to call terraform command run", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
					cmd.RunCall.Returns.Errors = []error{nil, errors.New("the-executor-error")}
				})

				It("returns an error and the current tf state", func() {
					_, err := executor.Destroy(input, "some-template", "")
					tdErr := err.(terraform.ExecutorError)
					Expect(tdErr).To(MatchError("the-executor-error"))

					tfState, err := tdErr.TFState()
					Expect(err).NotTo(HaveOccurred())
					Expect(tfState).To(Equal("some-tf-state"))
				})
			})

			Context("when it fails to read out the resulting tf state", func() {
				BeforeEach(func() {
					terraform.SetReadFile(func(filename string) ([]byte, error) {
						return []byte{}, errors.New("blueberry")
					})
				})

				It("returns an error", func() {
					_, err := executor.Destroy(input, "some-template", "")
					Expect(err).To(MatchError("Read terraform state: blueberry"))
				})
			})

			Context("when --debug is false", func() {
				BeforeEach(func() {
					executor = terraform.NewExecutor(cmd, stateStore, false)
				})

				Context("when it fails to call terraform command run", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), os.ModePerm)
						Expect(err).NotTo(HaveOccurred())

						cmd.RunCall.Returns.Errors = []error{nil, errors.New("failed to run terraform command")}
					})

					It("returns an error and the current tf state", func() {
						_, err := executor.Destroy(input, "some-template", "")
						tdErr := err.(terraform.ExecutorError)

						tfState, err := tdErr.TFState()
						Expect(tfState).To(Equal("some-tf-state"))
					})
				})
			})
		})
	})

	Describe("Import", func() {
		var (
			receivedTFState    string
			receivedTFTemplate string
		)

		BeforeEach(func() {
			cmd.RunCall.Stub = func(stdout io.Writer) {
				fileContents, err := ioutil.ReadFile(tfStatePath)
				Expect(err).NotTo(HaveOccurred())
				receivedTFState = string(fileContents)

				fileContents, err = ioutil.ReadFile(filepath.Join(terraformDir, "template.tf"))
				Expect(err).NotTo(HaveOccurred())
				receivedTFTemplate = string(fileContents)

				err = ioutil.WriteFile(tfStatePath, []byte("some-other-tfstate"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			}

			terraform.ResetReadFile()
		})

		It("writes the tfState to a file", func() {
			_, err := executor.Import(terraform.ImportInput{
				TerraformAddr: "some-resource-type.some-addr",
				AWSResourceID: "some-id",
				TFState:       "some-tf-state",
				Creds:         storage.AWS{},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(receivedTFState).To(Equal("some-tf-state"))
		})

		It("writes a terraform template to a file", func() {
			_, err := executor.Import(terraform.ImportInput{
				TerraformAddr: "some-resource-type.some-addr[i]",
				AWSResourceID: "some-id",
				TFState:       "some-tf-state",
				Creds: storage.AWS{
					Region:          "some-region",
					AccessKeyID:     "some-access-key",
					SecretAccessKey: "some-secret",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(receivedTFTemplate).To(Equal(`
provider "aws" {
	region     = "some-region"
	access_key = "some-access-key"
	secret_key = "some-secret"
}

resource "some-resource-type" "some-addr" {
}`))
		})

		It("shells out to terraform import and returns the tfState", func() {
			tfState, err := executor.Import(terraform.ImportInput{
				TerraformAddr: "some-resource-type.some-addr",
				AWSResourceID: "some-id",
				TFState:       "some-tf-state",
				Creds:         storage.AWS{},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{
				"import", "some-resource-type.some-addr", "some-id",
				"-state", relativeStatePath,
			}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
			Expect(tfState).To(Equal("some-other-tfstate"))
		})

		Context("when an error occurs", func() {
			Context("when it fails to get terraform dir", func() {
				BeforeEach(func() {
					stateStore.GetTerraformDirCall.Returns.Error = errors.New("failed to get terraform dir")
				})

				It("returns an error", func() {
					_, err := executor.Import(terraform.ImportInput{
						TerraformAddr: "some-resource-type.some-addr",
						AWSResourceID: "some-id",
						TFState:       "some-tf-state",
						Creds:         storage.AWS{},
					})
					Expect(err).To(MatchError("failed to get terraform dir"))
				})
			})

			Context("when it fails to get vars dir", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("failed to get vars dir")
				})

				It("returns an error", func() {
					_, err := executor.Import(terraform.ImportInput{
						TerraformAddr: "some-resource-type.some-addr",
						AWSResourceID: "some-id",
						TFState:       "some-tf-state",
						Creds:         storage.AWS{},
					})
					Expect(err).To(MatchError("failed to get vars dir"))
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

					_, err := executor.Import(terraform.ImportInput{
						TerraformAddr: "some-resource-type.some-addr",
						AWSResourceID: "some-id",
						TFState:       "some-tf-state",
						Creds:         storage.AWS{},
					})
					Expect(err).To(MatchError("failed to write tfstate"))
				})
			})

			Context("when terraform init fails", func() {
				It("returns an error", func() {
					cmd.RunCall.Returns.Errors = []error{errors.New("failed to initialize terraform")}

					_, err := executor.Import(terraform.ImportInput{
						TerraformAddr: "some-resource-type.some-addr",
						AWSResourceID: "some-id",
						TFState:       "some-tf-state",
						Creds:         storage.AWS{},
					})
					Expect(err).To(MatchError("failed to initialize terraform"))
				})
			})

			Context("when the command fails to run", func() {
				It("returns an error", func() {
					cmd.RunCall.Returns.Errors = []error{nil, errors.New("bad import")}

					_, err := executor.Import(terraform.ImportInput{
						TerraformAddr: "some-resource-type.some-addr",
						AWSResourceID: "some-id",
						TFState:       "some-tf-state",
						Creds:         storage.AWS{},
					})
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

					_, err := executor.Import(terraform.ImportInput{
						TerraformAddr: "some-resource-type.some-addr",
						AWSResourceID: "some-id",
						TFState:       "some-tf-state",
						Creds:         storage.AWS{},
					})
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

		Context("when an error occurs", func() {
			Context("when the run command fails", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{errors.New("run cmd failed")}
				})

				It("returns an error", func() {
					_, err := executor.Version()
					Expect(err).To(MatchError("run cmd failed"))
				})
			})

			Context("when the version cannot be parsed", func() {
				BeforeEach(func() {
					cmd.RunCall.Stub = func(stdout io.Writer) {
						stdout.Write([]byte(""))
					}
				})

				It("returns an error", func() {
					_, err := executor.Version()
					Expect(err).To(MatchError("Terraform version could not be parsed"))
				})
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

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(terraformDir))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"output", "external_ip", "-state", tfStatePath}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
		})

		Context("when an error occurs", func() {
			Context("when it fails to get terraform dir", func() {
				BeforeEach(func() {
					stateStore.GetTerraformDirCall.Returns.Error = errors.New("failed to get terraform dir")
				})

				It("returns an error", func() {
					_, err := executor.Output("some-tf-state", "external_ip")
					Expect(err).To(MatchError("failed to get terraform dir"))
				})
			})

			Context("when it fails to get vars dir", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("failed to get vars dir")
				})

				It("returns an error", func() {
					_, err := executor.Output("some-tf-state", "external_ip")
					Expect(err).To(MatchError("failed to get vars dir"))
				})
			})

			Context("when it fails to write the tfstate file", func() {
				BeforeEach(func() {
					terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
						if strings.Contains(file, "terraform.tfstate") {
							return errors.New("failed to write tf state file")
						}

						return nil
					})
				})

				It("returns an error", func() {
					_, err := executor.Output("some-tf-state", "external_ip")
					Expect(err).To(MatchError("failed to write tf state file"))
				})
			})

			Context("when terraform init fails", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{errors.New("failed to initialize terraform")}
				})

				It("returns an error", func() {
					_, err := executor.Output("some-template", "external_ip")
					Expect(err).To(MatchError("failed to initialize terraform"))
				})
			})

			Context("when it fails to call terraform command run", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{nil, errors.New("failed to run terraform command")}
				})

				It("returns an error", func() {
					_, err := executor.Output("some-tf-state", "external_ip")
					Expect(err).To(MatchError("failed to run terraform command"))
				})
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

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(varsDir))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"output", "--json"}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
		})

		Context("when an error occurs", func() {
			Context("when it fails to get vars dir", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("failed to get vars dir")
				})

				It("returns an error", func() {
					_, err := executor.Outputs("some-tf-state")
					Expect(err).To(MatchError("failed to get vars dir"))
				})
			})

			Context("when it fails to write the tfstate file", func() {
				BeforeEach(func() {
					terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
						if strings.Contains(file, "terraform.tfstate") {
							return errors.New("failed to write tf state file")
						}

						return nil
					})
				})

				It("returns an error", func() {
					_, err := executor.Outputs("some-tf-state")
					Expect(err).To(MatchError("failed to write tf state file"))
				})
			})

			Context("when terraform init fails", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{errors.New("failed to initialize terraform")}
				})

				It("returns an error", func() {
					_, err := executor.Outputs("some-tf-state")
					Expect(err).To(MatchError("failed to initialize terraform"))
				})
			})

			Context("when it fails to call terraform command run", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{nil, errors.New("failed to run terraform command")}
				})

				It("returns an error", func() {
					_, err := executor.Outputs("some-tf-state")
					Expect(err).To(MatchError("failed to run terraform command"))
				})
			})

			Context("when it fails to unmarshal the terraform outputs", func() {
				BeforeEach(func() {
					cmd.RunCall.Stub = func(stdout io.Writer) {
						fmt.Fprintf(stdout, "%%%")
					}
				})

				It("returns an error", func() {
					_, err := executor.Outputs("some-tf-state")
					Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
				})
			})
		})
	})
})
