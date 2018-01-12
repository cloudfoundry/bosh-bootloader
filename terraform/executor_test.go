package terraform_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

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
		fileIO     *fakes.FileIO
		executor   terraform.Executor

		tempDir      string
		terraformDir string
		varsDir      string
		input        map[string]interface{}

		tfStatePath       string
		relativeStatePath string

		tfVarsPath       string
		relativeVarsPath string
	)

	BeforeEach(func() {
		cmd = &fakes.TerraformCmd{}
		stateStore = &fakes.StateStore{}
		fileIO = &fakes.FileIO{}

		executor = terraform.NewExecutor(cmd, stateStore, fileIO, true)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		terraformDir, err = ioutil.TempDir("", "terraform")
		Expect(err).NotTo(HaveOccurred())
		stateStore.GetTerraformDirCall.Returns.Directory = terraformDir

		varsDir, err = ioutil.TempDir("", "vars")
		Expect(err).NotTo(HaveOccurred())
		stateStore.GetVarsDirCall.Returns.Directory = varsDir

		tfStatePath = filepath.Join(varsDir, "terraform.tfstate")
		relativeStatePath, err = filepath.Rel(terraformDir, tfStatePath)
		Expect(err).NotTo(HaveOccurred())

		tfVarsPath = filepath.Join(varsDir, "bbl.tfvars")
		relativeVarsPath, err = filepath.Rel(terraformDir, tfVarsPath)
		Expect(err).NotTo(HaveOccurred())

		input = map[string]interface{}{"project_id": "some-project-id"}
	})

	Describe("Init", func() {
		It("runs terraform init", func() {
			err := executor.Init()
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.CallCount).To(Equal(1))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"init"}))
		})

		Context("when getting terraform dir fails", func() {
			BeforeEach(func() {
				stateStore.GetTerraformDirCall.Returns.Error = errors.New("canteloupe")
			})

			It("returns an error", func() {
				err := executor.Init()
				Expect(err).To(MatchError("Get terraform dir: canteloupe"))
			})
		})

		Context("when terraform init fails", func() {
			BeforeEach(func() {
				cmd.RunCall.Returns.Errors = []error{errors.New("guava")}
			})

			It("returns an error", func() {
				err := executor.Init()
				Expect(err).To(MatchError("Run terraform init: guava"))
			})
		})
	})

	Describe("Setup", func() {
		It("writes existing terraform state", func() {
			err := executor.Setup("some-template", input)
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.GetTerraformDirCall.CallCount).To(Equal(1))

			Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(terraformDir, "bbl-template.tf")))
			Expect(string(fileIO.WriteFileCall.Receives[0].Contents)).To(Equal("some-template"))

			Expect(fileIO.WriteFileCall.Receives[1].Filename).To(Equal(filepath.Join(terraformDir, ".terraform", ".gitignore")))
			Expect(string(fileIO.WriteFileCall.Receives[1].Contents)).To(Equal("*\n"))

			Expect(fileIO.WriteFileCall.Receives[2].Filename).To(Equal(tfVarsPath))
			Expect(string(fileIO.WriteFileCall.Receives[2].Contents)).To(ContainSubstring(`project_id="some-project-id"`))

			Expect(cmd.RunCall.CallCount).To(Equal(0))
		})

		Context("when an error occurs", func() {
			Context("when getting terraform dir fails", func() {
				BeforeEach(func() {
					stateStore.GetTerraformDirCall.Returns.Error = errors.New("canteloupe")
				})

				It("returns an error", func() {
					err := executor.Setup("some-template", input)
					Expect(err).To(MatchError("Get terraform dir: canteloupe"))
				})
			})

			Context("when writing the template file fails", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("pear")}}
				})

				It("returns an error", func() {
					err := executor.Setup("some-template", input)
					Expect(err).To(MatchError("Write terraform template: pear"))
				})
			})

			Context("when getting vars dir fails", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("coconut")
				})

				It("returns an error", func() {
					err := executor.Setup("", input)
					Expect(err).To(MatchError("Get vars dir: coconut"))
				})
			})

			Context("when writing the vars file fails", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{}, {}, {Error: errors.New("apple")}}
				})

				It("returns an error", func() {
					err := executor.Setup("some-template", input)
					Expect(err).To(MatchError("Write terraform vars: apple"))
				})
			})

			Context("when creating the .terraform directory fails", func() {
				BeforeEach(func() {
					_, err := os.Create(filepath.Join(terraformDir, ".terraform"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					err := executor.Setup("some-template", input)
					Expect(err.Error()).To(ContainSubstring("Create .terraform directory: "))
				})
			})

			Context("when writing the .gitignore for terraform binaries fails", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{}, {Error: errors.New("nectarine")}}
				})

				It("returns an error", func() {
					err := executor.Setup("some-template", input)
					Expect(err).To(MatchError("Write .gitignore for terraform binaries: nectarine"))
				})
			})
		})
	})

	Describe("Apply", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(tfStatePath, []byte("some-updated-terraform-state"), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(tfVarsPath, []byte("some-tfvars"), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())

			err = executor.Init() // We need to run the terraform init command.
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(varsDir)
		})

		It("runs terraform apply", func() {
			err := executor.Apply(map[string]string{
				"some-cert": "some-cert-value",
			})
			Expect(err).NotTo(HaveOccurred())

			By("passing the correct args and dir to run command", func() {
				Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(terraformDir))
				Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
					"apply",
					"--auto-approve",
					"-var", "some-cert=some-cert-value",
					"-state", relativeStatePath,
					"-var-file", relativeVarsPath,
				}))
				Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
			})
		})

		Context("when other vars files are in the directory", func() {
			var (
				relativeUserProvidedVarsPathA string
				relativeUserProvidedVarsPathC string
			)
			BeforeEach(func() {
				userProvidedVarsPathA := filepath.Join(varsDir, "awesome-user-vars.tfvars")
				err := ioutil.WriteFile(userProvidedVarsPathA, []byte("user-provided tfvars"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
				userProvidedVarsPathC := filepath.Join(varsDir, "custom-user-vars.tfvars")
				err = ioutil.WriteFile(userProvidedVarsPathC, []byte("user-provided tfvars"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(filepath.Join(varsDir, "not-a-tfvars-file.yml"), []byte("definitely not a tfvars file"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())

				relativeUserProvidedVarsPathA, err = filepath.Rel(terraformDir, userProvidedVarsPathA)
				Expect(err).NotTo(HaveOccurred())
				relativeUserProvidedVarsPathC, err = filepath.Rel(terraformDir, userProvidedVarsPathC)
				Expect(err).NotTo(HaveOccurred())
			})

			It("passes all user provided tfvars files to the run command in alphabetic order", func() {
				err := executor.Apply(map[string]string{
					"some-cert": "some-cert-value",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
					"apply",
					"--auto-approve",
					"-var", "some-cert=some-cert-value",
					"-state", relativeStatePath,
					"-var-file", relativeUserProvidedVarsPathA,
					"-var-file", relativeVarsPath,
					"-var-file", relativeUserProvidedVarsPathC,
				}))
			})
		})

		Context("when terraform command run fails", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())

				cmd.RunCall.Returns.Errors = []error{nil, errors.New("the-executor-error")}
			})

			It("returns the error", func() {
				err := executor.Apply(map[string]string{})
				Expect(err).To(MatchError("the-executor-error"))
			})

			Context("and --debug is false", func() {
				BeforeEach(func() {
					executor = terraform.NewExecutor(cmd, stateStore, fileIO, false)
				})

				It("returns a redacted error message", func() {
					err := executor.Apply(map[string]string{})
					Expect(err).To(MatchError("Some output has been redacted, use `bbl latest-error` to see it or run again with --debug for additional debug output"))
				})
			})
		})
	})

	Describe("Destroy", func() {
		var credentials map[string]string

		BeforeEach(func() {
			credentials = map[string]string{
				"some-cert": "some-cert-value",
			}

			err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(tfVarsPath, []byte("some-tf-vars"), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())

			err = executor.Init()
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes the template and tf state to a temp dir", func() {
			err := executor.Destroy(credentials)
			Expect(err).NotTo(HaveOccurred())

			By("passing the correct args and dir to run command", func() {
				Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(terraformDir))
				Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
					"destroy",
					"-force",
					"-var", "some-cert=some-cert-value",
					"-state", relativeStatePath,
					"-var-file", relativeVarsPath,
				}))
				Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
			})
		})

		Context("when an error occurs", func() {
			Context("when getting terraform dir fails", func() {
				BeforeEach(func() {
					stateStore.GetTerraformDirCall.Returns.Error = errors.New("kiwi")
				})

				It("returns an error", func() {
					err := executor.Destroy(credentials)
					Expect(err).To(MatchError("Get terraform dir: kiwi"))
				})
			})

			Context("when getting vars dir fails", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("banana")
				})

				It("returns an error", func() {
					err := executor.Destroy(credentials)
					Expect(err).To(MatchError("Get vars dir: banana"))
				})
			})

			Context("when command run fails", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), storage.StateMode)
					Expect(err).NotTo(HaveOccurred())
					cmd.RunCall.Returns.Errors = []error{nil, errors.New("the-executor-error")}
				})

				It("returns an error", func() {
					err := executor.Destroy(credentials)
					Expect(err).To(MatchError("the-executor-error"))
				})

				Context("when --debug is false", func() {
					BeforeEach(func() {
						executor = terraform.NewExecutor(cmd, stateStore, fileIO, false)
						err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), storage.StateMode)
						Expect(err).NotTo(HaveOccurred())

						cmd.RunCall.Returns.Errors = []error{nil, errors.New("failed to run terraform command")}
					})

					It("returns a redacted error", func() {
						err := executor.Destroy(credentials)
						Expect(err).To(MatchError("Some output has been redacted, use `bbl latest-error` to see it or run again with --debug for additional debug output"))
					})
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
		BeforeEach(func() {
			err := ioutil.WriteFile(tfStatePath, []byte("some-tf-state"), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an output from the terraform state", func() {
			cmd.RunCall.Stub = func(stdout io.Writer) {
				fmt.Fprintf(stdout, "some-external-ip\n")
			}
			output, err := executor.Output("external_ip")
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("some-external-ip"))

			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(terraformDir))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"output", "external_ip", "-state", tfStatePath}))
			Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
		})

		Context("when an error occurs", func() {
			Context("when it fails to get terraform dir", func() {
				BeforeEach(func() {
					stateStore.GetTerraformDirCall.Returns.Error = errors.New("failed")
				})

				It("returns an error", func() {
					_, err := executor.Output("external_ip")
					Expect(err).To(MatchError("Get terraform dir: failed"))
				})
			})

			Context("when it fails to get vars dir", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("failed")
				})

				It("returns an error", func() {
					_, err := executor.Output("external_ip")
					Expect(err).To(MatchError("Get vars dir: failed"))
				})
			})

			Context("when terraform init fails", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{errors.New("failed")}
				})

				It("returns an error", func() {
					_, err := executor.Output("external_ip")
					Expect(err).To(MatchError("Run terraform init in terraform dir: failed"))
				})
			})

			Context("when it fails to call terraform command run", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{nil, errors.New("failed")}
				})

				It("returns an error", func() {
					_, err := executor.Output("external_ip")
					Expect(err).To(MatchError("Run terraform output -state: failed"))
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
			outputs, err := executor.Outputs()
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
					stateStore.GetVarsDirCall.Returns.Error = errors.New("failed")
				})

				It("returns an error", func() {
					_, err := executor.Outputs()
					Expect(err).To(MatchError("Get vars dir: failed"))
				})
			})

			Context("when terraform init fails", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{errors.New("failed")}
				})

				It("returns an error", func() {
					_, err := executor.Outputs()
					Expect(err).To(MatchError("Run terraform init in vars dir: failed"))
				})
			})

			Context("when it fails to call terraform command run", func() {
				BeforeEach(func() {
					cmd.RunCall.Returns.Errors = []error{nil, errors.New("failed")}
				})

				It("returns an error", func() {
					_, err := executor.Outputs()
					Expect(err).To(MatchError("Run terraform output --json in vars dir: failed"))
				})
			})

			Context("when it fails to unmarshal the terraform outputs", func() {
				BeforeEach(func() {
					cmd.RunCall.Stub = func(stdout io.Writer) {
						fmt.Fprintf(stdout, "%%%")
					}
				})

				It("returns an error", func() {
					_, err := executor.Outputs()
					Expect(err).To(MatchError("Unmarshal terraform output: invalid character '%' looking for beginning of value"))
				})
			})
		})
	})
})
