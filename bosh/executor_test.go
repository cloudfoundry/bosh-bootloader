package bosh_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	Describe("JumpboxInterpolate", func() {
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

		It("interpolates the jumpbox and bosh manifests", func() {
			interpolateInput.DeploymentVars = "internal_cidr: 10.0.0.0/24"
			interpolateInput.OpsFile = ""

			jumpboxInterpolateOutput, err := executor.JumpboxInterpolate(interpolateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd.RunCallCount()).To(Equal(1))

			sharedArgs := append([]string{
				"--vars-store", fmt.Sprintf("%s/jumpbox-variables.yml", varsDir),
				"--vars-file", fmt.Sprintf("%s/jumpbox-deployment-vars.yml", varsDir),
				"-o", fmt.Sprintf("%s/cpi.yml", deploymentDir),
			})

			By("running with the expected args in the vars directory", func() {
				_, workingDir, args := cmd.RunArgsForCall(0)
				Expect(args).To(Equal(jumpboxInterpolateArgs(sharedArgs, deploymentDir)))
				Expect(workingDir).To(Equal(varsDir))
			})

			Expect(jumpboxInterpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "jumpbox")))
			Expect(jumpboxInterpolateOutput.Manifest).To(Equal("some-manifest"))
			Expect(jumpboxInterpolateOutput.Variables).To(gomegamatchers.MatchYAML("key: value"))
		})

		Describe("failure cases", func() {
			Context("when trying to run a command fails", func() {
				BeforeEach(func() {
					cmd.RunReturnsOnCall(0, errors.New("kiwi"))
				})

				It("returns an error", func() {
					executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
					_, err := executor.JumpboxInterpolate(bosh.InterpolateInput{
						DeploymentDir: deploymentDir,
						VarsDir:       varsDir,
						IAAS:          "aws",
					})
					Expect(err).To(MatchError("Jumpbox interpolate: kiwi: "))
				})
			})

			Context("when the variables file fails to be read", func() {
				It("returns an error", func() {
					readFileFunc := func(path string) ([]byte, error) {
						return []byte{}, errors.New("kiwi")
					}

					executor = bosh.NewExecutor(cmd, readFileFunc, json.Unmarshal, json.Marshal, ioutil.WriteFile)
					_, err := executor.JumpboxInterpolate(bosh.InterpolateInput{
						DeploymentDir: deploymentDir,
						VarsDir:       varsDir,
						IAAS:          "aws",
					})
					Expect(err).To(MatchError("Jumpbox read file: kiwi"))
				})
			})
		})
	})

	Describe("DirectorInterpolate", func() {
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

				interpolateOutput, err := executor.DirectorInterpolate(azureInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(1))

				sharedArgs := []string{
					"--vars-store", fmt.Sprintf("%s/director-variables.yml", varsDir),
					"--vars-file", fmt.Sprintf("%s/director-deployment-vars.yml", varsDir),
					"-o", fmt.Sprintf("%s/cpi.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/jumpbox-user.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/uaa.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/credhub.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/user-ops-file.yml", varsDir),
				}

				_, _, args := cmd.RunArgsForCall(0)
				Expect(args).To(Equal(interpolateArgs(sharedArgs, deploymentDir)))

				Expect(interpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "bosh")))
				Expect(interpolateOutput.Manifest).To(Equal("some-manifest"))
				Expect(interpolateOutput.Variables).To(Equal("key: value"))
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

			It("interpolates the jumpbox and bosh manifests", func() {
				awsInterpolateInput.DeploymentVars = "internal_cidr: 10.0.0.0/24"

				interpolateOutput, err := executor.DirectorInterpolate(awsInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(1))

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

				_, _, args := cmd.RunArgsForCall(0)
				Expect(args).To(Equal(interpolateArgs(sharedArgs, deploymentDir)))

				Expect(interpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "bosh")))

				Expect(interpolateOutput.Manifest).To(Equal("some-manifest"))
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

			It("interpolates the jumpbox and bosh manifests", func() {
				gcpInterpolateInput.DeploymentVars = "internal_cidr: 10.0.0.0/24"
				gcpInterpolateInput.OpsFile = ""

				interpolateOutput, err := executor.DirectorInterpolate(gcpInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(1))

				sharedArgs := append([]string{
					"--vars-store", fmt.Sprintf("%s/director-variables.yml", varsDir),
					"--vars-file", fmt.Sprintf("%s/director-deployment-vars.yml", varsDir),
					"-o", fmt.Sprintf("%s/cpi.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/jumpbox-user.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/uaa.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/credhub.yml", deploymentDir),
					"-o", fmt.Sprintf("%s/gcp-bosh-director-ephemeral-ip-ops.yml", deploymentDir),
				})

				_, _, args := cmd.RunArgsForCall(0)
				Expect(args).To(Equal(interpolateArgs(sharedArgs, deploymentDir)))

				Expect(interpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "bosh")))

				Expect(interpolateOutput.Manifest).To(Equal("some-manifest"))
				Expect(interpolateOutput.Variables).To(Equal("key: value"))
			})

			Context("when a user opsfile is provided", func() {
				It("interpolates the bosh manifest once", func() {
					interpolateOutput, err := executor.DirectorInterpolate(gcpInterpolateInput)
					Expect(err).NotTo(HaveOccurred())

					Expect(cmd.RunCallCount()).To(Equal(1))

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

					_, _, args := cmd.RunArgsForCall(0)
					Expect(args).To(Equal(interpolateArgs(sharedArgs, deploymentDir)))

					Expect(interpolateOutput.Args).To(Equal(createEnvArgs(sharedArgs, deploymentDir, varsDir, "bosh")))
					Expect(interpolateOutput.Manifest).To(Equal("some-manifest"))
					Expect(interpolateOutput.Variables).To(Equal("key: value"))
				})
			})
		})

		Describe("failure cases", func() {
			Context("when trying to run a command fails", func() {
				BeforeEach(func() {
					cmd.RunReturnsOnCall(0, errors.New("failed to run command"))
				})

				It("returns an error", func() {
					executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
					_, err := executor.DirectorInterpolate(bosh.InterpolateInput{
						DeploymentDir: deploymentDir,
						VarsDir:       varsDir,
						IAAS:          "aws",
						OpsFile:       "some-ops-file",
					})
					Expect(err).To(MatchError("failed to run command"))
				})
			})
			Context("when the variables file fails to be read", func() {
				It("returns an error", func() {
					readFileFunc := func(path string) ([]byte, error) {
						return []byte{}, errors.New("failed to read variables file")
					}

					executor = bosh.NewExecutor(cmd, readFileFunc, json.Unmarshal, json.Marshal, ioutil.WriteFile)
					_, err := executor.DirectorInterpolate(bosh.InterpolateInput{
						DeploymentDir: deploymentDir,
						VarsDir:       varsDir,
						IAAS:          "aws",
					})
					Expect(err).To(MatchError("failed to read variables file"))
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
		})

		It("creates a bosh environment", func() {
			err := executor.CreateEnv(createEnvInput)
			Expect(err).NotTo(HaveOccurred())

			writer, dir, args := cmd.RunArgsForCall(0)
			Expect(writer).To(Equal(os.Stdout))
			Expect(dir).To(Equal(varsDir))
			Expect(args).To(Equal([]string{"some", "command", "args"}))
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
				err := executor.CreateEnv(createEnvInput)
				Expect(err).To(MatchError("Create env: apricot"))
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

func jumpboxInterpolateArgs(sharedArgs []string, deploymentDir string) []string {
	return append(
		[]string{
			"interpolate", fmt.Sprintf("%s/jumpbox.yml", deploymentDir),
			"--var-errs",
		},
		sharedArgs...,
	)
}

func interpolateArgs(sharedArgs []string, deploymentDir string) []string {
	return append(
		[]string{
			"interpolate", fmt.Sprintf("%s/bosh.yml", deploymentDir),
			"--var-errs",
			"--var-errs-unused",
		},
		sharedArgs...,
	)
}

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
