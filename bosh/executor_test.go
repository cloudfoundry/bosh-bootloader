package bosh_test

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Describe("Execute", func() {
		var (
			cmd *fakes.BOSHCommand

			tempDir          string
			tempDirFunc      func(string, string) (string, error)
			tempDirCallCount int

			executor bosh.Executor

			stateJSONContents    string
			variablesYMLContents string
			gcpCredentialsPath   string
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

			gcpCredentialsPath = fmt.Sprintf("%s/gcp_credentials.json", tempDir)

			executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		AfterEach(func() {
			tempDirCallCount = 0
		})

		DescribeTable("deploys a bosh", func(deployInput bosh.ExecutorInput, iaasSpecificArgs []string) {
			deployOutput, err := executor.Execute(deployInput)
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.CallCount).To(Equal(1))
			Expect(tempDirCallCount).To(Equal(1))

			if deployInput.IAAS == "gcp" {
				iaasSpecificArgs = append(iaasSpecificArgs, []string{"--var-file", fmt.Sprintf("gcp_credentials_json=%s", gcpCredentialsPath)}...)
				credentialsContents, err := ioutil.ReadFile(gcpCredentialsPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(credentialsContents)).To(Equal(`{"key":"value"}`))
			}

			privateKeyPath := fmt.Sprintf("%s/private_key", tempDir)

			expectedArgs := append([]string{
				"create-env", fmt.Sprintf("%s/bosh.yml", tempDir),
				"--state", fmt.Sprintf("%s/state.json", tempDir),
				"-o", fmt.Sprintf("%s/cpi.yml", tempDir),
				"-o", fmt.Sprintf("%s/external-ip-not-recommended.yml", tempDir),
				"--vars-store", fmt.Sprintf("%s/variables.yml", tempDir),
				"-v", "internal_cidr=10.0.0.0/24",
				"-v", "internal_gw=10.0.0.1",
				"-v", "internal_ip=10.0.0.6",
				"-v", "external_ip=some-external-ip",
				"-v", "director_name=some-director-name",
				"--var-file", fmt.Sprintf("private_key=%s", privateKeyPath),
			}, iaasSpecificArgs...)

			Expect(cmd.RunCall.Receives.Args).To(Equal(expectedArgs))
			privateKeyContents, err := ioutil.ReadFile(privateKeyPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(privateKeyContents)).To(Equal("some-ssh-key"))

			Expect(deployOutput).To(Equal(bosh.ExecutorOutput{
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: map[string]interface{}{
					"admin_password": "some-admin-password",
					"director_ssl": map[interface{}]interface{}{
						"certificate": "some-certificate",
						"private_key": "some-private-key",
						"ca":          "some-ca",
					},
				},
			}))
		},
			Entry("on aws", bosh.ExecutorInput{
				IAAS:                  "aws",
				Command:               "create-env",
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
				Variables: `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`,
			}, []string{
				"-v", "access_key_id=some-access-key-id",
				"-v", "secret_access_key=some-secret-access-key",
				"-v", "region=some-region",
				"-v", "az=some-az",
				"-v", "default_key_name=some-key-name",
				"-v", "default_security_groups=[some-security-group]",
				"-v", "subnet_id=some-subnet",
			}),
			Entry("on gcp", bosh.ExecutorInput{
				IAAS:         "gcp",
				Command:      "create-env",
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
				Variables: `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`,
			}, []string{
				"-v", "zone=some-zone",
				"-v", "network=some-network",
				"-v", "subnetwork=some-subnetwork",
				"-v", `tags=[some-internal-tag,some-bosh-open-tag]`,
				"-v", `project_id=some-project-id`,
			}),
		)

		DescribeTable("deletes a bosh", func(deleteInput bosh.ExecutorInput, iaasSpecificArgs []string) {
			stateFile := fmt.Sprintf("%s/state.json", tempDir)
			variablesFile := fmt.Sprintf("%s/variables.yml", tempDir)
			cmd.RunCall.Stub = func() {
				err := os.Remove(stateFile)
				Expect(err).NotTo(HaveOccurred())

				err = os.Remove(variablesFile)
				Expect(err).NotTo(HaveOccurred())
			}

			_, err := executor.Execute(deleteInput)
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.CallCount).To(Equal(1))
			Expect(tempDirCallCount).To(Equal(1))

			if deleteInput.IAAS == "gcp" {
				iaasSpecificArgs = append(iaasSpecificArgs, []string{"--var-file", fmt.Sprintf("gcp_credentials_json=%s", gcpCredentialsPath)}...)
				credentialsContents, err := ioutil.ReadFile(gcpCredentialsPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(credentialsContents)).To(Equal(`{"key":"value"}`))
			}

			privateKeyPath := fmt.Sprintf("%s/private_key", tempDir)

			expectedArgs := append([]string{
				"delete-env", fmt.Sprintf("%s/bosh.yml", tempDir),
				"--state", stateFile,
				"-o", fmt.Sprintf("%s/cpi.yml", tempDir),
				"-o", fmt.Sprintf("%s/external-ip-not-recommended.yml", tempDir),
				"--vars-store", variablesFile,
				"-v", "internal_cidr=10.0.0.0/24",
				"-v", "internal_gw=10.0.0.1",
				"-v", "internal_ip=10.0.0.6",
				"-v", "external_ip=some-external-ip",
				"-v", "director_name=some-director-name",
				"--var-file", fmt.Sprintf("private_key=%s", privateKeyPath),
			}, iaasSpecificArgs...)

			Expect(cmd.RunCall.Receives.Args).To(Equal(expectedArgs))
			privateKeyContents, err := ioutil.ReadFile(privateKeyPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(privateKeyContents)).To(Equal("some-ssh-key"))
		},
			Entry("on aws", bosh.ExecutorInput{
				IAAS:                  "aws",
				Command:               "delete-env",
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
				Variables: `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`,
			}, []string{
				"-v", "access_key_id=some-access-key-id",
				"-v", "secret_access_key=some-secret-access-key",
				"-v", "region=some-region",
				"-v", "az=some-az",
				"-v", "default_key_name=some-key-name",
				"-v", "default_security_groups=[some-security-group]",
				"-v", "subnet_id=some-subnet",
			}),
			Entry("on gcp", bosh.ExecutorInput{
				IAAS:         "gcp",
				Command:      "delete-env",
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
				Variables: `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`,
			}, []string{
				"-v", "zone=some-zone",
				"-v", "network=some-network",
				"-v", "subnetwork=some-subnetwork",
				"-v", `tags=[some-internal-tag,some-bosh-open-tag]`,
				"-v", `project_id=some-project-id`,
			}),
		)
		Describe("failure cases", func() {
			It("fails when the temporary directory cannot be created", func() {
				tempDirFunc = func(prefix, dir string) (string, error) {
					return "", errors.New("failed to create temp dir")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Execute(bosh.ExecutorInput{})
				Expect(err).To(MatchError("failed to create temp dir"))
			})

			It("fails when the bosh state cannot be marshaled", func() {
				marshalJSONFunc := func(boshState interface{}) ([]byte, error) {
					return []byte{}, errors.New("failed to marshal bosh state")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, marshalJSONFunc, ioutil.WriteFile)
				_, err := executor.Execute(bosh.ExecutorInput{
					BOSHState: map[string]interface{}{},
				})
				Expect(err).To(MatchError("failed to marshal bosh state"))
			})

			It("fails when the passed in bosh state cannot be written", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					return errors.New("failed to write bosh state")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Execute(bosh.ExecutorInput{
					BOSHState: map[string]interface{}{},
				})
				Expect(err).To(MatchError("failed to write bosh state"))
			})

			It("fails when the passed in variables cannot be written", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/variables.yml", tempDir) {
						return errors.New("failed to write variables")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Execute(bosh.ExecutorInput{
					Variables: "some-vars",
				})
				Expect(err).To(MatchError("failed to write variables"))
			})

			It("fails when trying to write the bosh manifest file", func() {
				writeFileFunc := func(path string, contents []byte, fileMode os.FileMode) error {
					if path == fmt.Sprintf("%s/bosh.yml", tempDir) {
						return errors.New("failed to write bosh manifest")
					}
					return nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Execute(bosh.ExecutorInput{})
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
				_, err := executor.Execute(bosh.ExecutorInput{
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
				_, err := executor.Execute(bosh.ExecutorInput{
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

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Execute(bosh.ExecutorInput{
					IAAS: "gcp",
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

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, writeFileFunc)
				_, err := executor.Execute(bosh.ExecutorInput{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("failed to write GCP credentials"))
			})

			It("fails when trying to run command", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run command")

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Execute(bosh.ExecutorInput{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to run command"))
			})

			It("fails when the variables file fails to be read", func() {
				readFileFunc := func(path string) ([]byte, error) {
					return []byte{}, errors.New("failed to read variables file")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, readFileFunc, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Execute(bosh.ExecutorInput{
					IAAS:    "aws",
					Command: "create-env",
				})
				Expect(err).To(MatchError("failed to read variables file"))
			})

			It("fails when the variables fail to be unmarshaled", func() {
				unmarshalFunc := func(contents []byte, output interface{}) error {
					return errors.New("failed to unmarshal variables")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, unmarshalFunc, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Execute(bosh.ExecutorInput{
					IAAS:    "aws",
					Command: "create-env",
					Variables: `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`,
				})
				Expect(err).To(MatchError("failed to unmarshal variables"))
			})

			It("fails when the state file fails to be read", func() {
				readFileFunc := func(path string) ([]byte, error) {
					if path == fmt.Sprintf("%s/state.json", tempDir) {
						return []byte{}, errors.New("failed to read state file")
					}
					return []byte{}, nil
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, readFileFunc, yaml.Unmarshal, json.Unmarshal, json.Marshal, ioutil.WriteFile)
				_, err := executor.Execute(bosh.ExecutorInput{
					IAAS:    "aws",
					Command: "create-env",
				})
				Expect(err).To(MatchError("failed to read state file"))
			})

			It("fails when the state file fails to be unmarshaled", func() {
				unmarshalFunc := func(contents []byte, output interface{}) error {
					return errors.New("failed to unmarshal state file")
				}

				executor = bosh.NewExecutor(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, unmarshalFunc, json.Marshal, ioutil.WriteFile)
				_, err := executor.Execute(bosh.ExecutorInput{
					IAAS:    "aws",
					Command: "create-env",
					Variables: `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`,
					BOSHState: map[string]interface{}{
						"key": "value",
					},
				})
				Expect(err).To(MatchError("failed to unmarshal state file"))
			})
		})
	})
})
