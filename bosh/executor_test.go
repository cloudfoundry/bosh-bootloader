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
			gcpCredentialsPath   string
			privateKeyPath       string
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
				IAAS:                  "aws",
				DirectorName:          "some-director-name",
				AccessKeyID:           "some-access-key-id",
				SecretAccessKey:       "some-secret-access-key",
				Region:                "some-region",
				AZ:                    "some-az",
				DefaultKeyName:        "some-key-name",
				DefaultSecurityGroups: []string{"some-security-group"},
				SubnetID:              "some-subnet",
				PrivateKey:            "some-ssh-key",
				ExternalIP:            "some-external-ip",
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: variablesYMLContents,
				OpsFile:   []byte("some-ops-file"),
			}

			gcpInterpolateInput = bosh.InterpolateInput{
				IAAS:         "gcp",
				DirectorName: "some-director-name",
				Zone:         "some-zone",
				Network:      "some-network",
				Subnetwork:   "some-subnetwork",
				Tags: []string{
					"some-internal-tag",
					"some-bosh-open-tag",
				},
				ProjectID:       "some-project-id",
				ExternalIP:      "some-external-ip",
				CredentialsJSON: `{"key":"value"}`,
				PrivateKey:      "some-ssh-key",
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: variablesYMLContents,
				OpsFile:   []byte("some-ops-file"),
			}

			gcpCredentialsPath = fmt.Sprintf("%s/gcp_credentials.json", tempDir)
			privateKeyPath = fmt.Sprintf("%s/private_key", tempDir)

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)
		})

		AfterEach(func() {
			tempDirCallCount = 0
		})

		Context("when iaas is gcp", func() {
			It("writes gcp credentials to a file in the temp dir", func() {
				_, err := executor.Interpolate(gcpInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				credentialsContents, err := ioutil.ReadFile(gcpCredentialsPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(credentialsContents)).To(Equal(`{"key":"value"}`))
			})
		})

		Context("when iaas is aws", func() {
			It("writes private key to a file in the temp dir", func() {
				_, err := executor.Interpolate(awsInterpolateInput)
				Expect(err).NotTo(HaveOccurred())

				privateKeyContents, err := ioutil.ReadFile(privateKeyPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(privateKeyContents)).To(Equal("some-ssh-key"))
			})
		})

		DescribeTable("generates a bosh manifest", func(interpolateInputFunc func() bosh.InterpolateInput, expectedIAASArgsFunc func() []string) {
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
				"-v", "internal_cidr=10.0.0.0/24",
				"-v", "internal_gw=10.0.0.1",
				"-v", "internal_ip=10.0.0.6",
				"-v", "external_ip=some-external-ip",
				"-v", "director_name=some-director-name",
			}, expectedIAASArgsFunc()...)

			Expect(cmd.RunCall.Receives.Args).To(Equal(expectedArgs))
			Expect(cmd.RunCall.Receives.Debug).To(Equal(true))

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
			}, func() []string {
				return []string{
					"-v", "access_key_id=some-access-key-id",
					"-v", "secret_access_key=some-secret-access-key",
					"-v", "region=some-region",
					"-v", "az=some-az",
					"-v", "default_key_name=some-key-name",
					"-v", "default_security_groups=[some-security-group]",
					"-v", "subnet_id=some-subnet",
					"--var-file", fmt.Sprintf("private_key=%s", privateKeyPath),
				}
			}),
			Entry("on gcp", func() bosh.InterpolateInput {
				return gcpInterpolateInput
			}, func() []string {
				return []string{
					"-v", "zone=some-zone",
					"-v", "network=some-network",
					"-v", "subnetwork=some-subnetwork",
					"-v", `tags=[some-internal-tag,some-bosh-open-tag]`,
					"-v", `project_id=some-project-id`,
					"--var-file", fmt.Sprintf("gcp_credentials_json=%s", gcpCredentialsPath),
				}
			}),
		)

		It("does not pass in false to run command on interpolate", func() {
			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, false)
			_, err := executor.Interpolate(awsInterpolateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd.RunCall.Receives.Debug).To(Equal(true))
		})

		Describe("failure cases", func() {
			It("fails when the temporary directory cannot be created", func() {
				tempDirFunc = func(prefix, dir string) (string, error) {
					return "", errors.New("failed to create temp dir")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)
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

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc, true)
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

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc, true)
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

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc, true)
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

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc, true)
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

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc, true)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("failed to write external ip not recommended Ops file"))
			})

			It("fails when trying to write the private key", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/private_key", tempDir) {
						return errors.New("failed to write private key")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc, true)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to write private key"))
			})

			It("fails when trying to write GCP credentials", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/gcp_credentials.json", tempDir) {
						return errors.New("failed to write GCP credentials")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc, true)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("failed to write GCP credentials"))
			})

			It("fails when trying to run command", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run command")

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to run command"))
			})

			It("fails when the variables file fails to be read", func() {
				readFileFunc := func(path string) ([]byte, error) {
					return []byte{}, errors.New("failed to read variables file")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, readFileFunc, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)
				_, err := executor.Interpolate(bosh.InterpolateInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to read variables file"))
			})

			It("fails when the variables fail to be unmarshaled", func() {
				unmarshalFunc := func(contents []byte, output interface{}) error {
					return errors.New("failed to unmarshal variables")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, unmarshalFunc, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)
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

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)

		})

		It("fails when the temporary directory cannot be created", func() {
			tempDirFunc = func(prefix, dir string) (string, error) {
				return "", errors.New("failed to create temp dir")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)
			err := callback(executor)
			Expect(err).To(MatchError("failed to create temp dir"))
		})

		It("fails when the state fails to marshal", func() {
			marshalFunc := func(input interface{}) ([]byte, error) {
				return []byte{}, errors.New("failed to marshal state")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, marshalFunc, ioutil.WriteFile, true)
			err := callback(executor)
			Expect(err).To(MatchError("failed to marshal state"))
		})

		It("fails when the state cannot be written to a file", func() {
			writeFile := func(filename string, contents []byte, mode os.FileMode) error {
				return errors.New("failed to write file")
			}

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFile, true)
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

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)

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
			Expect(cmd.RunCall.Receives.Debug).To(Equal(true))

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

			It("fails when the state cannot be read", func() {
				readFile := func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read file")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, readFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)
				_, err := executor.CreateEnv(createEnvInput)
				Expect(err).To(MatchError("failed to read file"))
			})

			It("fails when the state cannot be unmarshaled", func() {
				unmarshalFunc := func(contents []byte, output interface{}) error {
					return errors.New("failed to unmarshal")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, unmarshalFunc, json.Marshal, ioutil.WriteFile, true)
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

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile, true)

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
			Expect(cmd.RunCall.Receives.Debug).To(Equal(true))
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
		})
	})
})
