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
	Describe("DirectorInterpolate", func() {
		var (
			cmd *fakes.BOSHCommand

			tempDir          string
			tempDirFunc      func(string, string) (string, error)
			tempDirCallCount int

			executor bosh.Executor

			stateJSONContents    string
			variablesYMLContents string
			awsInterpolateInput  bosh.InterpolateInput
			gcpInterpolateInput  bosh.InterpolateInput
		)

		BeforeEach(func() {
			cmd = &fakes.BOSHCommand{}

			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDirFunc = func(prefix, dir string) (string, error) {
				tempDirCallCount++
				return tempDir, nil
			}

			stateJSONContents = `{"key":"value"}`
			variablesYMLContents = "key: value"

			awsInterpolateInput = bosh.InterpolateInput{
				IAAS:           "aws",
				DeploymentVars: "internal_cidr: 10.0.0.0/24",
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: variablesYMLContents,
				OpsFile:   "some-ops-file",
			}

			gcpInterpolateInput = awsInterpolateInput
			gcpInterpolateInput.IAAS = "gcp"

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		AfterEach(func() {
			tempDirCallCount = 0
		})

		Context("aws", func() {
			It("generates a bosh manifest", func() {
				cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
					stdout.Write([]byte("some-manifest"))
					return nil
				}

				interpolateOutput, err := executor.DirectorInterpolate(awsInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(2))
				Expect(tempDirCallCount).To(Equal(1))

				expectedArgs := append([]string{
					"interpolate", fmt.Sprintf("%s/bosh.yml", tempDir),
					"--var-errs",
					"--var-errs-unused",
					"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
					"--vars-file", fmt.Sprintf("%s/deployment-vars.yml", tempDir),
					"-o", fmt.Sprintf("%s/cpi.yml", tempDir),
					"-o", fmt.Sprintf("%s/jumpbox-user.yml", tempDir),
					"-o", fmt.Sprintf("%s/aws-external-ip-not-recommended.yml", tempDir),
					"-o", fmt.Sprintf("%s/iam-instance-profile.yml", tempDir)})

				_, _, args := cmd.RunArgsForCall(0)
				Expect(args).To(Equal(expectedArgs))

				expectedArgs = append([]string{
					"interpolate", fmt.Sprintf("%s/bosh.yml", tempDir),
					"--var-errs",
					"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
					"--vars-file", fmt.Sprintf("%s/deployment-vars.yml", tempDir),
					"-o", fmt.Sprintf("%s/user-ops-file.yml", tempDir)})

				_, _, args = cmd.RunArgsForCall(1)
				Expect(args).To(Equal(expectedArgs))

				Expect(interpolateOutput.Manifest).To(Equal("some-manifest"))
				Expect(interpolateOutput.Variables).To(gomegamatchers.MatchYAML(variablesYMLContents))
			})
		})

		Context("gcp", func() {
			It("generates a bosh manifest", func() {
				cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
					stdout.Write([]byte("some-manifest"))
					return nil
				}

				interpolateOutput, err := executor.DirectorInterpolate(gcpInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(2))
				Expect(tempDirCallCount).To(Equal(1))

				expectedArgs := append([]string{
					"interpolate", fmt.Sprintf("%s/bosh.yml", tempDir),
					"--var-errs",
					"--var-errs-unused",
					"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
					"--vars-file", fmt.Sprintf("%s/deployment-vars.yml", tempDir),
					"-o", fmt.Sprintf("%s/cpi.yml", tempDir),
					"-o", fmt.Sprintf("%s/jumpbox-user.yml", tempDir),
					"-o", fmt.Sprintf("%s/gcp-external-ip-not-recommended.yml", tempDir)})

				_, _, args := cmd.RunArgsForCall(0)
				Expect(args).To(Equal(expectedArgs))

				expectedArgs = append([]string{
					"interpolate", fmt.Sprintf("%s/bosh.yml", tempDir),
					"--var-errs",
					"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
					"--vars-file", fmt.Sprintf("%s/deployment-vars.yml", tempDir),
					"-o", fmt.Sprintf("%s/user-ops-file.yml", tempDir)})

				_, _, args = cmd.RunArgsForCall(1)
				Expect(args).To(Equal(expectedArgs))

				Expect(interpolateOutput.Manifest).To(Equal("some-manifest"))
				Expect(interpolateOutput.Variables).To(gomegamatchers.MatchYAML(variablesYMLContents))
			})

			Context("when there are jumpbox deployment vars", func() {
				It("interpolates the jumpbox and bosh manifests", func() {
					gcpInterpolateInput.JumpboxDeploymentVars = "internal_cidr: 10.0.0.0/24"
					gcpInterpolateInput.OpsFile = ""

					cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
						stdout.Write([]byte("some-manifest"))
						return nil
					}

					jumpboxInterpolateOutput, err := executor.JumpboxInterpolate(gcpInterpolateInput)
					Expect(err).NotTo(HaveOccurred())

					Expect(cmd.RunCallCount()).To(Equal(1))
					Expect(tempDirCallCount).To(Equal(1))

					expectedArgs := append([]string{
						"interpolate", fmt.Sprintf("%s/jumpbox.yml", tempDir),
						"--var-errs",
						"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
						"--vars-file", fmt.Sprintf("%s/jumpbox-deployment-vars.yml", tempDir),
						"-o", fmt.Sprintf("%s/cpi.yml", tempDir),
					})

					_, _, args := cmd.RunArgsForCall(0)
					Expect(args).To(Equal(expectedArgs))

					Expect(jumpboxInterpolateOutput.Manifest).To(Equal("some-manifest"))
					Expect(jumpboxInterpolateOutput.Variables).To(gomegamatchers.MatchYAML("key: value"))

					interpolateOutput, err := executor.DirectorInterpolate(gcpInterpolateInput)
					Expect(err).NotTo(HaveOccurred())

					Expect(cmd.RunCallCount()).To(Equal(2))
					Expect(tempDirCallCount).To(Equal(2))

					expectedArgs = append([]string{
						"interpolate", fmt.Sprintf("%s/bosh.yml", tempDir),
						"--var-errs",
						"--var-errs-unused",
						"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
						"--vars-file", fmt.Sprintf("%s/deployment-vars.yml", tempDir),
						"-o", fmt.Sprintf("%s/cpi.yml", tempDir),
						"-o", fmt.Sprintf("%s/bosh-director-ephemeral-ip-ops.yml", tempDir),
						"-o", fmt.Sprintf("%s/uaa.yml", tempDir),
						"-o", fmt.Sprintf("%s/credhub.yml", tempDir),
					})

					_, _, args = cmd.RunArgsForCall(1)
					Expect(args).To(Equal(expectedArgs))

					Expect(interpolateOutput.Manifest).To(Equal("some-manifest"))
					Expect(jumpboxInterpolateOutput.Variables).To(gomegamatchers.MatchYAML("key: value"))
				})
			})
		})

		Context("when a user opsfile is provided", func() {
			It("re-interpolates the bosh manifest", func() {
				interpolateInput := bosh.InterpolateInput{
					IAAS: "gcp",
					DeploymentVars: `internal_cidr: 10.0.0.0/24
		tags: [some-bosh-tag, some-internal-tag]'`,
					BOSHState: map[string]interface{}{
						"key": "value",
					},
					Variables: variablesYMLContents,
					OpsFile: `
---
- type: replace
path: /networks/name=default/subnets/0/cloud_properties/tags/-
value: sabeti-bosh-isolation
		`,
				}

				manifest := `
---
networks
- name: default
  subnets:
  - az: z1
  cloud_properties:
    tags:
      - some-bosh-tag
      - some-internal-tag
`
				manifestWithUserOpsFile := `
---
networks
- name: default
  subnets:
  - az: z1
  cloud_properties:
    tags:
      - some-bosh-tag
      - some-internal-tag
      - sabeti-bosh-isolation
`

				writtenManifest := []byte{}
				cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
					for _, arg := range args {
						if arg == fmt.Sprintf("%s/user-ops-file.yml", tempDir) {
							var err error
							writtenManifest, err = ioutil.ReadFile(fmt.Sprintf("%s/bosh.yml", tempDir))
							if err != nil {
								return err
							}
							stdout.Write([]byte(manifestWithUserOpsFile))
							return nil
						}
					}
					stdout.Write([]byte(manifest))
					return nil
				}

				interpolateOutput, err := executor.DirectorInterpolate(interpolateInput)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(2))
				Expect(tempDirCallCount).To(Equal(1))

				expectedArgs := append([]string{
					"interpolate", fmt.Sprintf("%s/bosh.yml", tempDir),
					"--var-errs",
					"--var-errs-unused",
					"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
					"--vars-file", fmt.Sprintf("%s/deployment-vars.yml", tempDir),
					"-o", fmt.Sprintf("%s/cpi.yml", tempDir),
					"-o", fmt.Sprintf("%s/jumpbox-user.yml", tempDir),
					"-o", fmt.Sprintf("%s/gcp-external-ip-not-recommended.yml", tempDir)})

				_, _, args := cmd.RunArgsForCall(0)
				Expect(args).To(Equal(expectedArgs))

				expectedArgsWithUserOpsfile := append([]string{
					"interpolate", fmt.Sprintf("%s/bosh.yml", tempDir),
					"--var-errs",
					"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
					"--vars-file", fmt.Sprintf("%s/deployment-vars.yml", tempDir),
					"-o", fmt.Sprintf("%s/user-ops-file.yml", tempDir)})

				_, _, args = cmd.RunArgsForCall(1)
				Expect(args).To(Equal(expectedArgsWithUserOpsfile))

				opsFileContents, err := ioutil.ReadFile(fmt.Sprintf("%s/user-ops-file.yml", tempDir))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(opsFileContents)).To(Equal(interpolateInput.OpsFile))
				Expect(string(writtenManifest)).To(Equal(manifest))

				Expect(interpolateOutput.Manifest).To(Equal(manifestWithUserOpsFile))
				Expect(interpolateOutput.Variables).To(gomegamatchers.MatchYAML("key: value"))
			})
		})

		It("does not pass in false to run command on interpolate", func() {
			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
			_, err := executor.DirectorInterpolate(awsInterpolateInput)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("failure cases", func() {
			It("fails when trying to run command", func() {
				cmd.RunReturnsOnCall(0, errors.New("failed to run command"))

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.DirectorInterpolate(bosh.InterpolateInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to run command"))
			})

			It("fails when trying to run the command to interpolate with the user opsfile", func() {
				cmd.RunReturnsOnCall(1, errors.New("failed to run command"))

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.DirectorInterpolate(bosh.InterpolateInput{
					IAAS:    "aws",
					OpsFile: "some-ops-file",
				})
				Expect(err).To(MatchError("failed to run command"))
			})

			It("fails when the variables file fails to be read", func() {
				readFileFunc := func(path string) ([]byte, error) {
					return []byte{}, errors.New("failed to read variables file")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, readFileFunc, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.DirectorInterpolate(bosh.InterpolateInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to read variables file"))
			})
		})
	})

	var createEnvDeleteEnvFailureCases = func(callback func(executor bosh.Executor) error) {
		var (
			cmd *fakes.BOSHCommand

			tempDir          string
			tempDirFunc      func(string, string) (string, error)
			tempDirCallCount int

			executor bosh.Executor
		)

		BeforeEach(func() {
			var err error

			cmd = &fakes.BOSHCommand{}
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDirFunc = func(prefix, dir string) (string, error) {
				tempDirCallCount++
				return tempDir, nil
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		It("fails when the temporary directory cannot be created", func() {
			tempDirFunc = func(prefix, dir string) (string, error) {
				return "", errors.New("failed to create temp dir")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
			err := callback(executor)
			Expect(err).To(MatchError("failed to create temp dir"))
		})

		It("fails when the state fails to marshal", func() {
			marshalFunc := func(input interface{}) ([]byte, error) {
				return []byte{}, errors.New("failed to marshal state")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, marshalFunc, ioutil.WriteFile)
			err := callback(executor)
			Expect(err).To(MatchError("failed to marshal state"))
		})

		It("fails when the state cannot be written to a file", func() {
			writeFile := func(filename string, contents []byte, mode os.FileMode) error {
				return errors.New("failed to write file")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, writeFile)
			err := callback(executor)
			Expect(err).To(MatchError("failed to write file"))
		})

		It("fails when the run command returns an error", func() {
			cmd.RunReturnsOnCall(0, errors.New("failed to run"))
			err := callback(executor)
			Expect(err).To(MatchError("failed to run"))
		})
	}

	Describe("CreateEnv", func() {
		var (
			cmd *fakes.BOSHCommand

			tempDir          string
			tempDirFunc      func(string, string) (string, error)
			tempDirCallCount int

			executor bosh.Executor

			createEnvInput bosh.CreateEnvInput
			manifestPath   string
			variablesPath  string
			statePath      string
		)

		BeforeEach(func() {
			var err error

			cmd = &fakes.BOSHCommand{}
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDirFunc = func(prefix, dir string) (string, error) {
				tempDirCallCount++
				return tempDir, nil
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			createEnvInput = bosh.CreateEnvInput{
				Manifest:  "some-manifest",
				Variables: "some-variables",
				State:     map[string]interface{}{},
			}

			manifestPath = fmt.Sprintf("%s/manifest.yml", tempDir)
			variablesPath = fmt.Sprintf("%s/variables.yml", tempDir)
			statePath = fmt.Sprintf("%s/state.json", tempDir)

			cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				return ioutil.WriteFile(statePath, []byte(`{"key": "value"}`), os.ModePerm)
			}
		})

		AfterEach(func() {
			tempDirCallCount = 0
		})

		It("creates a bosh environment", func() {
			createEnvOutput, err := executor.CreateEnv(createEnvInput)
			Expect(err).NotTo(HaveOccurred())

			Expect(tempDirCallCount).To(Equal(1))

			manifestContents, err := ioutil.ReadFile(manifestPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(manifestContents)).To(Equal("some-manifest"))

			variablesContents, err := ioutil.ReadFile(variablesPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(variablesContents)).To(Equal("some-variables"))

			writer, dir, args := cmd.RunArgsForCall(0)
			Expect(writer).To(Equal(os.Stdout))
			Expect(dir).To(Equal(tempDir))
			Expect(args).To(Equal([]string{
				"create-env", manifestPath,
				"--vars-store", variablesPath,
				"--state", statePath,
			}))

			Expect(createEnvOutput.State).To(Equal(map[string]interface{}{
				"key": "value",
			}))
		})

		Context("failure cases", func() {
			createEnvDeleteEnvFailureCases(func(executor bosh.Executor) error {
				createEnvInput := bosh.CreateEnvInput{
					Manifest:  "some-manifest",
					Variables: "some-variables",
					State:     map[string]interface{}{},
				}
				_, err := executor.CreateEnv(createEnvInput)
				return err
			})

			Context("when command run fails", func() {
				BeforeEach(func() {
					cmd.RunReturns(errors.New("failed to run"))
					executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

					cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
						ioutil.WriteFile(statePath, []byte(`{"key": "value"}`), os.ModePerm)
						return errors.New("failed to run")
					}
				})

				It("returns a create env error with a valid bosh state", func() {
					expectedError := bosh.NewCreateEnvError(map[string]interface{}{
						"key": "value",
					}, errors.New("failed to run"))
					_, err := executor.CreateEnv(createEnvInput)
					Expect(err).To(MatchError(expectedError))
				})

				Context("when the state cannot be read", func() {
					BeforeEach(func() {
						readFile := func(filename string) ([]byte, error) {
							return []byte{}, errors.New("failed to read file")
						}

						executor = bosh.NewExecutor(cmd, tempDirFunc, readFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
					})

					It("returns an error", func() {
						_, err := executor.CreateEnv(createEnvInput)
						Expect(err).To(MatchError("the following errors occurred:\nfailed to run,\nfailed to read file"))
					})
				})

				Context("when the state cannot be unmarshaled", func() {
					BeforeEach(func() {
						unmarshalFunc := func(contents []byte, output interface{}) error {
							return errors.New("failed to unmarshal")
						}

						executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, unmarshalFunc, json.Marshal, ioutil.WriteFile)
					})

					It("returns an error", func() {
						_, err := executor.CreateEnv(createEnvInput)
						Expect(err).To(MatchError("the following errors occurred:\nfailed to run,\nfailed to unmarshal"))
					})
				})
			})

			It("fails when the state cannot be read", func() {
				readFile := func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read file")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, readFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.CreateEnv(createEnvInput)
				Expect(err).To(MatchError("failed to read file"))
			})

			It("fails when the state cannot be unmarshaled", func() {
				unmarshalFunc := func(contents []byte, output interface{}) error {
					return errors.New("failed to unmarshal")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, unmarshalFunc, json.Marshal, ioutil.WriteFile)
				_, err := executor.CreateEnv(createEnvInput)
				Expect(err).To(MatchError("failed to unmarshal"))
			})
		})
	})

	Describe("DeleteEnv", func() {
		var (
			cmd *fakes.BOSHCommand

			tempDir          string
			tempDirFunc      func(string, string) (string, error)
			tempDirCallCount int

			executor bosh.Executor

			deleteEnvInput bosh.DeleteEnvInput
			manifestPath   string
			variablesPath  string
			statePath      string
		)

		BeforeEach(func() {
			var err error

			cmd = &fakes.BOSHCommand{}
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDirFunc = func(prefix, dir string) (string, error) {
				tempDirCallCount++
				return tempDir, nil
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			deleteEnvInput = bosh.DeleteEnvInput{
				Manifest:  "some-manifest",
				Variables: "some-variables",
				State:     map[string]interface{}{},
			}

			manifestPath = fmt.Sprintf("%s/manifest.yml", tempDir)
			variablesPath = fmt.Sprintf("%s/variables.yml", tempDir)
			statePath = fmt.Sprintf("%s/state.json", tempDir)

			cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				return ioutil.WriteFile(statePath, []byte(`{"key": "value"}`), os.ModePerm)
			}
		})

		AfterEach(func() {
			tempDirCallCount = 0
		})

		It("deletes a bosh environment", func() {
			err := executor.DeleteEnv(deleteEnvInput)
			Expect(err).NotTo(HaveOccurred())

			Expect(tempDirCallCount).To(Equal(1))

			manifestContents, err := ioutil.ReadFile(manifestPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(manifestContents)).To(Equal("some-manifest"))

			variablesContents, err := ioutil.ReadFile(variablesPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(variablesContents)).To(Equal("some-variables"))

			writer, dir, args := cmd.RunArgsForCall(0)
			Expect(writer).To(Equal(os.Stdout))
			Expect(dir).To(Equal(tempDir))
			Expect(args).To(Equal([]string{
				"delete-env", manifestPath,
				"--vars-store", variablesPath,
				"--state", statePath,
			}))
		})

		Context("failure cases", func() {
			createEnvDeleteEnvFailureCases(func(executor bosh.Executor) error {
				deleteEnvInput := bosh.DeleteEnvInput{
					Manifest:  "some-manifest",
					Variables: "some-variables",
					State:     map[string]interface{}{},
				}
				return executor.DeleteEnv(deleteEnvInput)
			})

			Context("when command run fails", func() {
				BeforeEach(func() {
					cmd.RunReturnsOnCall(0, errors.New("failed to run"))
					executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

					cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
						ioutil.WriteFile(statePath, []byte(`{"partial": "state"}`), os.ModePerm)
						return errors.New("failed to run")
					}
				})

				It("returns a create env error with a valid bosh state", func() {
					expectedError := bosh.NewDeleteEnvError(map[string]interface{}{
						"partial": "state",
					}, errors.New("failed to run"))
					err := executor.DeleteEnv(deleteEnvInput)
					Expect(err).To(MatchError(expectedError))
				})

				Context("when the state cannot be read", func() {
					BeforeEach(func() {
						readFile := func(filename string) ([]byte, error) {
							return []byte{}, errors.New("failed to read file")
						}

						executor = bosh.NewExecutor(cmd, tempDirFunc, readFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
					})

					It("returns an error", func() {
						err := executor.DeleteEnv(deleteEnvInput)
						Expect(err).To(MatchError("the following errors occurred:\nfailed to run,\nfailed to read file"))
					})
				})
			})
		})
	})

	Describe("Version", func() {
		var (
			cmd              *fakes.BOSHCommand
			tempDir          string
			tempDirFunc      func(string, string) (string, error)
			tempDirCallCount int

			executor bosh.Executor
		)
		BeforeEach(func() {
			cmd = &fakes.BOSHCommand{}
			cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				stdout.Write([]byte("some-text version 2.0.24 some-other-text"))
				return nil
			}

			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDirFunc = func(prefix, dir string) (string, error) {
				tempDirCallCount++
				return tempDir, nil
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
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
			It("returns an error when the temporary directory cannot be created", func() {
				tempDirFunc = func(prefix, dir string) (string, error) {
					return "", errors.New("failed to create temp dir")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Version()
				Expect(err).To(MatchError("failed to create temp dir"))
			})

			It("returns an error when the run cmd fails", func() {
				cmd.RunReturns(errors.New("failed to run cmd"))
				_, err := executor.Version()
				Expect(err).To(MatchError("failed to run cmd"))
			})

			It("returns a bosh version error when the version cannot be parsed", func() {
				expectedError := bosh.NewBOSHVersionError(errors.New("BOSH version could not be parsed"))
				cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
					stdout.Write([]byte(""))
					return nil
				}
				_, err := executor.Version()
				Expect(err).To(Equal(expectedError))
			})
		})
	})
})
