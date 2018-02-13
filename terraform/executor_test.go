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

		tfStatePath string

		tfVarsPath string
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

		tfVarsPath = filepath.Join(varsDir, "bbl.tfvars")

		input = map[string]interface{}{"project_id": "some-project-id"}
	})

	Describe("Init", func() {
		It("runs terraform init", func() {
			err := executor.Init()
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.CallCount).To(Equal(1))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"init", terraformDir}))
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
			fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
				fakes.FileInfo{
					FileName: "bbl.tfvars",
				},
			}
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
				Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
					"apply",
					"--auto-approve",
					"-var", "some-cert=some-cert-value",
					"-state", tfStatePath,
					"-var-file", tfVarsPath,
					terraformDir,
				}))
				Expect(cmd.RunCall.Receives.Debug).To(BeTrue())
			})
		})

		Context("when other vars files are in the directory", func() {
			var (
				userProvidedVarsPathA string
				userProvidedVarsPathC string
			)
			BeforeEach(func() {
				fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
					fakes.FileInfo{
						FileName: "bbl.tfvars",
					},
					fakes.FileInfo{
						FileName: "awesome-user-vars.tfvars",
					},
					fakes.FileInfo{
						FileName: "custom-user-vars.tfvars",
					},
					fakes.FileInfo{
						FileName: "definitely-not-a-tf-vars-file",
					},
				}

				userProvidedVarsPathA = filepath.Join(varsDir, "awesome-user-vars.tfvars")
				userProvidedVarsPathC = filepath.Join(varsDir, "custom-user-vars.tfvars")
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
					"-state", tfStatePath,
					"-var-file", userProvidedVarsPathA,
					"-var-file", tfVarsPath,
					"-var-file", userProvidedVarsPathC,
					terraformDir,
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

			fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
				fakes.FileInfo{
					FileName: "bbl.tfvars",
				},
			}

			err := executor.Init()
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes the template and tf state to a temp dir", func() {
			err := executor.Destroy(credentials)
			Expect(err).NotTo(HaveOccurred())

			By("passing the correct args and dir to run command", func() {
				Expect(cmd.RunCall.Receives.Args).To(ConsistOf([]string{
					"destroy",
					"-force",
					"-var", "some-cert=some-cert-value",
					"-state", tfStatePath,
					"-var-file", tfVarsPath,
					terraformDir,
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

			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"output", "external_ip", "-state", tfStatePath, terraformDir}))
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

			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{
				"output", "--json", "-state", tfStatePath,
			}))
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

	Describe("IsPaved", func() {
		Context("when the state store fails to return the vars directory", func() {
			It("returns an error", func() {
				stateStore.GetVarsDirCall.Returns.Error = errors.New("guava")

				_, err := executor.IsPaved()

				Expect(err).To(MatchError("Get vars dir: guava"))
			})
		})
		Context("when the terraform.tfstate file does not exist", func() {
			It("returns false", func() {
				fileIO.StatCall.Returns.Error = errors.New("pear")

				isPaved, err := executor.IsPaved()

				Expect(err).NotTo(HaveOccurred())
				Expect(isPaved).To(Equal(false))
			})
		})
		Context("when the terraform.tfstate file does exist", func() {
			It("returns true", func() {
				isPaved, err := executor.IsPaved()

				Expect(err).NotTo(HaveOccurred())
				Expect(isPaved).To(Equal(true))
			})
		})
	})
})
