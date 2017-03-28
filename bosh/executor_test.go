package bosh_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	Describe("Interpolate", func() {
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
			var err error

			cmd = &fakes.BOSHCommand{}
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDirFunc = func(prefix, dir string) (string, error) {
				tempDirCallCount++
				return tempDir, nil
			}

			stateJSONContents = `{"key":"value"}`
			variablesYMLContents = "key: value"

			awsInterpolateInput = bosh.InterpolateInput{
				IAAS: "aws",
				DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-elastic-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-bosh-user-access-key
secret_access_key: some-bosh-user-secret-access-key
default_key_name: some-keypair-name
default_security_groups: [some-bosh-security-group]
region: some-region
private_key: |-
  some-private-key`,
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: variablesYMLContents,
				OpsFile:   []byte("some-ops-file"),
			}

			gcpInterpolateInput = bosh.InterpolateInput{
				IAAS: "gcp",
				DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags: [some-bosh-tag, some-internal-tag]
project_id: some-project-id
gcp_credentials_json: 'some-credential-json'`,
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: variablesYMLContents,
				OpsFile:   []byte("some-ops-file"),
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		AfterEach(func() {
			tempDirCallCount = 0
		})

		DescribeTable("generates a bosh manifest", func(interpolateInputFunc func() bosh.InterpolateInput) {
			cmd.RunCall.Stub = func(stdout io.Writer) {
				stdout.Write([]byte("some-manifest"))
			}

			interpolateOutput, err := executor.Interpolate(interpolateInputFunc())
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.CallCount).To(Equal(1))
			Expect(tempDirCallCount).To(Equal(1))

			expectedArgs := append([]string{
				"interpolate", fmt.Sprintf("%s/bosh.yml", tempDir),
				"--var-errs",
				"--var-errs-unused",
				"-o", fmt.Sprintf("%s/cpi.yml", tempDir),
				"-o", fmt.Sprintf("%s/external-ip-not-recommended.yml", tempDir),
				"-o", fmt.Sprintf("%s/user-ops-file.yml", tempDir),
				"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
				"--vars-file", fmt.Sprintf("%s/deployment-vars.yml", tempDir)})

			Expect(cmd.RunCall.Receives.Args).To(Equal(expectedArgs))

			opsFileContents, err := ioutil.ReadFile(fmt.Sprintf("%s/user-ops-file.yml", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(opsFileContents).To(Equal([]byte("some-ops-file")))

			Expect(interpolateOutput.Manifest).To(Equal("some-manifest"))
			Expect(interpolateOutput.Variables).To(Equal(map[interface{}]interface{}{
				"key": "value",
			}))
		},
			Entry("on aws", func() bosh.InterpolateInput {
				return awsInterpolateInput
			}),
			Entry("on gcp", func() bosh.InterpolateInput {
				return gcpInterpolateInput
			}),
		)

		It("does not pass in false to run command on interpolate", func() {
			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
			_, err := executor.Interpolate(awsInterpolateInput)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("failure cases", func() {
			It("fails when the temporary directory cannot be created", func() {
				tempDirFunc = func(prefix, dir string) (string, error) {
					return "", errors.New("failed to create temp dir")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Interpolate(gcpInterpolateInput)
				Expect(err).To(MatchError("failed to create temp dir"))
			})

			It("fails when the passed in variables cannot be written", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/variables.yml", tempDir) {
						return errors.New("failed to write variables")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Interpolate(gcpInterpolateInput)
				Expect(err).To(MatchError("failed to write variables"))
			})

			It("fails when trying to write the user ops file", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/user-ops-file.yml", tempDir) {
						return errors.New("failed to write user ops file")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Interpolate(bosh.InterpolateInput{})
				Expect(err).To(MatchError("failed to write user ops file"))
			})

			It("fails when trying to write the bosh manifest file", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/bosh.yml", tempDir) {
						return errors.New("failed to write bosh manifest")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Interpolate(bosh.InterpolateInput{})
				Expect(err).To(MatchError("failed to write bosh manifest"))
			})

			It("fails when trying to write the CPI Ops file", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/cpi.yml", tempDir) {
						return errors.New("failed to write CPI Ops file")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("failed to write CPI Ops file"))
			})

			It("fails when trying to write the external ip not recommended Ops file", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/external-ip-not-recommended.yml", tempDir) {
						return errors.New("failed to write external ip not recommended Ops file")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("failed to write external ip not recommended Ops file"))
			})

			It("fails when trying to write the deployment vars", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/deployment-vars.yml", tempDir) {
						return errors.New("failed to write deployment vars")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to write deployment vars"))
			})

			It("fails when trying to run command", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run command")

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to run command"))
			})

			It("fails when the variables file fails to be read", func() {
				readFileFunc := func(path string) ([]byte, error) {
					return []byte{}, errors.New("failed to read variables file")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, readFileFunc, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to read variables file"))
			})

			It("fails when the variables fail to be unmarshaled", func() {
				unmarshalFunc := func(contents []byte, output interface{}) error {
					return errors.New("failed to unmarshal variables")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, unmarshalFunc, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS:      "aws",
					Variables: variablesYMLContents,
				})
				Expect(err).To(MatchError("failed to unmarshal variables"))
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

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)

		})

		It("fails when the temporary directory cannot be created", func() {
			tempDirFunc = func(prefix, dir string) (string, error) {
				return "", errors.New("failed to create temp dir")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
			err := callback(executor)
			Expect(err).To(MatchError("failed to create temp dir"))
		})

		It("fails when the state fails to marshal", func() {
			marshalFunc := func(input interface{}) ([]byte, error) {
				return []byte{}, errors.New("failed to marshal state")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, marshalFunc, ioutil.WriteFile)
			err := callback(executor)
			Expect(err).To(MatchError("failed to marshal state"))
		})

		It("fails when the state cannot be written to a file", func() {
			writeFile := func(filename string, contents []byte, mode os.FileMode) error {
				return errors.New("failed to write file")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFile)
			err := callback(executor)
			Expect(err).To(MatchError("failed to write file"))
		})

		It("fails when the run command returns an error", func() {
			cmd.RunCall.Returns.Error = errors.New("failed to run")
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

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			createEnvInput = bosh.CreateEnvInput{
				Manifest:  "some-manifest",
				Variables: "some-variables",
				State:     map[string]interface{}{},
			}

			manifestPath = fmt.Sprintf("%s/manifest.yml", tempDir)
			variablesPath = fmt.Sprintf("%s/variables.yml", tempDir)
			statePath = fmt.Sprintf("%s/state.json", tempDir)

			cmd.RunCall.Stub = func(io.Writer) {
				err = ioutil.WriteFile(statePath, []byte(`{"key": "value"}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
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

			Expect(cmd.RunCall.Receives.Stdout).To(Equal(os.Stdout))
			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{
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
					cmd.RunCall.Returns.Error = errors.New("failed to run")
					executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
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

						executor = bosh.NewExecutor(cmd, tempDirFunc, readFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
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

						executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, unmarshalFunc, json.Marshal, ioutil.WriteFile)
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

				executor = bosh.NewExecutor(cmd, tempDirFunc, readFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.CreateEnv(createEnvInput)
				Expect(err).To(MatchError("failed to read file"))
			})

			It("fails when the state cannot be unmarshaled", func() {
				unmarshalFunc := func(contents []byte, output interface{}) error {
					return errors.New("failed to unmarshal")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, unmarshalFunc, json.Marshal, ioutil.WriteFile)
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

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			deleteEnvInput = bosh.DeleteEnvInput{
				Manifest:  "some-manifest",
				Variables: "some-variables",
				State:     map[string]interface{}{},
			}

			manifestPath = fmt.Sprintf("%s/manifest.yml", tempDir)
			variablesPath = fmt.Sprintf("%s/variables.yml", tempDir)
			statePath = fmt.Sprintf("%s/state.json", tempDir)

			cmd.RunCall.Stub = func(io.Writer) {
				err = ioutil.WriteFile(statePath, []byte(`{"key": "value"}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
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

			Expect(cmd.RunCall.Receives.Stdout).To(Equal(os.Stdout))
			Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{
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
					cmd.RunCall.Stub = func(io.Writer) {
						err := ioutil.WriteFile(statePath, []byte(`{"partial": "state"}`), os.ModePerm)
						Expect(err).NotTo(HaveOccurred())
					}

					cmd.RunCall.Returns.Error = errors.New("failed to run")
					executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
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

						executor = bosh.NewExecutor(cmd, tempDirFunc, readFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
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
			cmd.RunCall.Stub = func(stdout io.Writer) {
				stdout.Write([]byte("some-text version 2.0.0 some-other-text"))
			}

			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			tempDirFunc = func(prefix, dir string) (string, error) {
				tempDirCallCount++
				return tempDir, nil
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.Receives.Args).To(Equal([]string{"-v"}))
		})

		It("returns the correctly trimmed version", func() {
			version, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("2.0.0"))
		})

		Context("failure cases", func() {
			It("returns an error when the temporary directory cannot be created", func() {
				tempDirFunc = func(prefix, dir string) (string, error) {
					return "", errors.New("failed to create temp dir")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Version()
				Expect(err).To(MatchError("failed to create temp dir"))
			})

			It("returns an error when the run cmd fails", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run cmd")
				_, err := executor.Version()
				Expect(err).To(MatchError("failed to run cmd"))
			})

			It("returns an error when the version cannot be parsed", func() {
				cmd.RunCall.Stub = func(stdout io.Writer) {
					stdout.Write([]byte(""))
				}
				_, err := executor.Version()
				Expect(err).To(MatchError("BOSH version could not be parsed"))
			})
		})
	})
})
