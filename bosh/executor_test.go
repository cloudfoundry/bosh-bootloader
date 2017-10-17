package bosh_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	Describe("JumpboxCreateEnvArgs", func() {
		var (
			cmd *fakes.BOSHCommand

			deploymentDir string
			varsDir       string

			executor         bosh.Executor
			interpolateInput bosh.InterpolateInput
		)

		BeforeEach(func() {
			cmd = &fakes.BOSHCommand{}
			cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				stdout.Write([]byte("some-manifest"))
				return nil
			}

			var err error
			deploymentDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			interpolateInput = bosh.InterpolateInput{
				IAAS:          "aws",
				DeploymentDir: deploymentDir,
				VarsDir:       varsDir,
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: "key: value",
				OpsFile:   "some-ops-file",
			}

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		It("generates create-env args for jumpbox", func() {
			interpolateInput.DeploymentVars = "internal_cidr: 10.0.0.0/24"
			interpolateInput.OpsFile = ""

			jumpboxInterpolateOutput, err := executor.JumpboxCreateEnvArgs(interpolateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd.RunCallCount()).To(Equal(0))

			sharedArgs := append([]string{
				"--vars-store", fmt.Sprintf("%s/jumpbox-variables.yml", varsDir),
				"--vars-file", fmt.Sprintf("%s/jumpbox-deployment-vars.yml", varsDir),
				"-o", fmt.Sprintf("%s/cpi.yml", deploymentDir),
			})

			Expect(jumpboxInterpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "jumpbox")))
		})
	})

	Describe("DirectorCreateEnvArgs", func() {
		var (
			cmd *fakes.BOSHCommand

			deploymentDir string
			varsDir       string

			executor         bosh.Executor
			interpolateInput bosh.InterpolateInput
		)

		BeforeEach(func() {
			cmd = &fakes.BOSHCommand{}

			var err error
			deploymentDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			interpolateInput = bosh.InterpolateInput{
				DeploymentDir:  deploymentDir,
				VarsDir:        varsDir,
				DeploymentVars: "internal_cidr: 10.0.0.0/24",
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: "key: value",
				OpsFile:   "some-ops-file",
			}

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		Context("azure", func() {
			var azureInterpolateInput bosh.InterpolateInput

			BeforeEach(func() {
				azureInterpolateInput = interpolateInput
				azureInterpolateInput.IAAS = "azure"
			})

			It("generates a bosh manifest", func() {
				cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
					stdout.Write([]byte("some-manifest"))
					return nil
				}

				interpolateOutput, err := executor.DirectorCreateEnvArgs(azureInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(0))

				sharedArgs := []string{
					"--vars-store", fmt.Sprintf("%s/director-variables.yml", varsDir),
					"--vars-file", fmt.Sprintf("%s/director-deployment-vars.yml", varsDir),
					"-o", fmt.Sprintf("%s/cpi.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/jumpbox-user.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/uaa.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/credhub.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/user-ops-file.yml", varsDir),
				}

				Expect(interpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "bosh")))
			})
		})

		Context("aws", func() {
			var awsInterpolateInput bosh.InterpolateInput

			BeforeEach(func() {
				awsInterpolateInput = interpolateInput
				awsInterpolateInput.IAAS = "aws"

				cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
					stdout.Write([]byte("some-manifest"))
					return nil
				}
			})

			It("generates create-env args for director", func() {
				awsInterpolateInput.DeploymentVars = "internal_cidr: 10.0.0.0/24"

				interpolateOutput, err := executor.DirectorCreateEnvArgs(awsInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(0))

				sharedArgs := append([]string{
					"--vars-store", fmt.Sprintf("%s/director-variables.yml", varsDir),
					"--vars-file", fmt.Sprintf("%s/director-deployment-vars.yml", varsDir),
					"-o", fmt.Sprintf("%s/cpi.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/jumpbox-user.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/uaa.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/credhub.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/aws-bosh-director-ephemeral-ip-ops.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/iam-instance-profile.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/aws-bosh-director-encrypt-disk-ops.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/user-ops-file.yml", varsDir),
				})

				Expect(interpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "bosh")))
			})
		})

		Context("gcp", func() {
			var gcpInterpolateInput bosh.InterpolateInput

			BeforeEach(func() {
				gcpInterpolateInput = interpolateInput
				gcpInterpolateInput.IAAS = "gcp"

				cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
					stdout.Write([]byte("some-manifest"))
					return nil
				}
			})

			It("generates create-env args for director", func() {
				gcpInterpolateInput.DeploymentVars = "internal_cidr: 10.0.0.0/24"
				gcpInterpolateInput.OpsFile = ""

				interpolateOutput, err := executor.DirectorCreateEnvArgs(gcpInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(0))

				sharedArgs := append([]string{
					"--vars-store", fmt.Sprintf("%s/director-variables.yml", varsDir),
					"--vars-file", fmt.Sprintf("%s/director-deployment-vars.yml", varsDir),
					"-o", fmt.Sprintf("%s/cpi.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/jumpbox-user.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/uaa.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/credhub.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/gcp-bosh-director-ephemeral-ip-ops.yml", deploymentDir),
				})

				Expect(interpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "bosh")))
			})

			Context("when a user opsfile is provided", func() {
				It("puts the user-provided opsfile in create-env args", func() {
					interpolateOutput, err := executor.DirectorCreateEnvArgs(gcpInterpolateInput)
					Expect(err).NotTo(HaveOccurred())

					Expect(cmd.RunCallCount()).To(Equal(0))

					sharedArgs := append([]string{
						"--vars-store", fmt.Sprintf("%s/director-variables.yml", varsDir),
						"--vars-file", fmt.Sprintf("%s/director-deployment-vars.yml", varsDir),
						"-o", fmt.Sprintf("%s/cpi.yml", deploymentDir),
						"-o", fmt.Sprintf("%s/jumpbox-user.yml", deploymentDir),
						"-o", fmt.Sprintf("%s/uaa.yml", deploymentDir),
						"-o", fmt.Sprintf("%s/credhub.yml", deploymentDir),
						"-o", fmt.Sprintf("%s/gcp-bosh-director-ephemeral-ip-ops.yml", deploymentDir),
						"-o", fmt.Sprintf("%s/user-ops-file.yml", varsDir),
					})

					Expect(interpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "bosh")))
				})
			})
		})
	})

	Describe("CreateEnv", func() {
		var (
			cmd      *fakes.BOSHCommand
			executor bosh.Executor

			varsDir string

			createEnvInput bosh.CreateEnvInput
		)

		BeforeEach(func() {
			var err error

			cmd = &fakes.BOSHCommand{}
			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			createEnvInput = bosh.CreateEnvInput{
				Args:       []string{"some", "command", "args"},
				Deployment: "some-deployment",
				Directory:  varsDir,
			}

			cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				varsStore := filepath.Join(varsDir, "some-deployment-variables.yml")
				ioutil.WriteFile(varsStore, []byte("some-vars-store-contents"), os.ModePerm)
				return nil
			}
		})

		AfterEach(func() {
			os.Remove(filepath.Join(varsDir, "some-deployment-variables.yml"))
		})

		It("creates a bosh environment", func() {
			vars, err := executor.CreateEnv(createEnvInput)
			Expect(err).NotTo(HaveOccurred())

			writer, dir, args := cmd.RunArgsForCall(0)
			Expect(writer).To(Equal(os.Stdout))
			Expect(dir).To(Equal(varsDir))
			Expect(args).To(Equal([]string{"some", "command", "args"}))

			By("returning the contents of the vars store", func() {
				Expect(vars).To(Equal("some-vars-store-contents"))
			})
		})

		Context("when the run command returns an error", func() {
			BeforeEach(func() {
				cmd.RunReturns(errors.New("apricot"))
			})

			It("returns an error", func() {
				createEnvInput := bosh.CreateEnvInput{
					Args:       []string{"some", "command", "args"},
					Deployment: "some-deployment",
					Directory:  varsDir,
				}
				vars, err := executor.CreateEnv(createEnvInput)
				Expect(err).To(MatchError("Create env: apricot"))
				Expect(vars).To(Equal(""))
			})
		})
	})

	Describe("DeleteEnv", func() {
		var (
			cmd      *fakes.BOSHCommand
			executor bosh.Executor

			varsDir string

			deleteEnvInput bosh.DeleteEnvInput
		)

		BeforeEach(func() {
			var err error

			cmd = &fakes.BOSHCommand{}
			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			deleteEnvInput = bosh.DeleteEnvInput{
				Args:       []string{"create-env", "command", "args"},
				Deployment: "some-deployment",
				Directory:  varsDir,
			}
		})

		It("deletes a bosh environment", func() {
			err := executor.DeleteEnv(deleteEnvInput)
			Expect(err).NotTo(HaveOccurred())

			writer, dir, args := cmd.RunArgsForCall(0)
			Expect(writer).To(Equal(os.Stdout))
			Expect(dir).To(Equal(varsDir))
			Expect(args).To(Equal([]string{"delete-env", "command", "args"}))
		})

		Context("when the run command returns an error", func() {
			BeforeEach(func() {
				cmd.RunReturns(errors.New("tangerine"))
			})

			It("returns an error", func() {
				deleteEnvInput := bosh.DeleteEnvInput{
					Args:       []string{"some", "command", "args"},
					Deployment: "some-deployment",
					Directory:  varsDir,
				}
				err := executor.DeleteEnv(deleteEnvInput)
				Expect(err).To(MatchError("Delete env: tangerine"))
			})
		})
	})

	Describe("Version", func() {
		var (
			cmd      *fakes.BOSHCommand
			executor bosh.Executor
		)
		BeforeEach(func() {
			cmd = &fakes.BOSHCommand{}
			cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				stdout.Write([]byte("some-text version 2.0.24 some-other-text"))
				return nil
			}

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())

			_, _, args := cmd.RunArgsForCall(0)
			Expect(args).To(Equal([]string{"-v"}))
		})

		It("returns the correctly trimmed version", func() {
			version, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("2.0.24"))
		})

		Context("failure cases", func() {
			Context("when the run cmd fails", func() {
				BeforeEach(func() {
					cmd.RunReturns(errors.New("failed to run cmd"))
				})

				It("returns an error", func() {
					_, err := executor.Version()
					Expect(err).To(MatchError("failed to run cmd"))
				})
			})

			Context("when the version cannot be parsed", func() {
				var expectedError error

				BeforeEach(func() {
					expectedError = bosh.NewBOSHVersionError(errors.New("BOSH version could not be parsed"))
					cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
						stdout.Write([]byte(""))
						return nil
					}
				})

				It("returns a bosh version error", func() {
					_, err := executor.Version()
					Expect(err).To(Equal(expectedError))
				})
			})
		})
	})
})

func createEnvArgs(sharedArgs []string, deploymentDir, varsDir, deployment string) []string {
	return append(
		[]string{
			"create-env",
			fmt.Sprintf("%s/%s.yml", deploymentDir, deployment),
			"--state", fmt.Sprintf("%s/%s-state.json", varsDir, deployment),
		},
		sharedArgs...,
	)
}
