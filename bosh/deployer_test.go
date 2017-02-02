package bosh_test

import (
	"encoding/json"
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

var _ = Describe("Deployer", func() {
	Describe("Deploy", func() {
		var (
			cmd *fakes.BOSHCommand

			tempDir          string
			tempDirFunc      func(string, string) (string, error)
			tempDirCallCount int

			deployer bosh.Deployer

			stateJSONContents    string
			variablesYMLContents string
			credentialsPath      string
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

			err = ioutil.WriteFile(fmt.Sprintf("%s/state.json", tempDir), []byte(stateJSONContents), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
			err = ioutil.WriteFile(fmt.Sprintf("%s/variables.yml", tempDir), []byte(variablesYMLContents), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
			credentialsPath = fmt.Sprintf("%s/credentials.json", tempDir)

			deployer = bosh.NewDeployer(cmd, tempDirFunc, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal)
		})

		AfterEach(func() {
			tempDirCallCount = 0
		})

		DescribeTable("deploys a bosh", func(deployInput bosh.DeployInput, iaasSpecificArgs []string) {
			deployOutput, err := deployer.Deploy(deployInput)
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCall.CallCount).To(Equal(1))
			Expect(tempDirCallCount).To(Equal(1))

			if deployInput.IAAS == "gcp" {
				iaasSpecificArgs = append(iaasSpecificArgs, []string{"--var-file", fmt.Sprintf("gcp_credentials_json=%s", credentialsPath)}...)
				credentialsContents, err := ioutil.ReadFile(credentialsPath)
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
				"-v", "director_name=some-director-name",
				"--var-file", fmt.Sprintf("private_key=%s", privateKeyPath),
			}, iaasSpecificArgs...)
			Expect(cmd.RunCall.Receives.Args).To(Equal(expectedArgs))
			privateKeyContents, err := ioutil.ReadFile(privateKeyPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(privateKeyContents)).To(Equal("some-ssh-key"))

			Expect(deployOutput).To(Equal(bosh.DeployOutput{
				BOSHState: map[string]interface{}{
					"key": "value",
				},
				Variables: map[string]interface{}{
					"key": "value",
				},
			}))
		},
			Entry("on aws", bosh.DeployInput{
				IAAS:                 "aws",
				DirectorName:         "some-director-name",
				AccessKeyID:          "some-access-key-id",
				SecretAccessKey:      "some-secret-access-key",
				Region:               "some-region",
				AZ:                   "some-az",
				DefaultKeyName:       "some-key-name",
				DefaultSecurityGroup: "some-security-group",
				SubnetID:             "some-subnet",
				PrivateKey:           "some-ssh-key",
			}, []string{
				"-v", "access_key_id=some-access-key-id",
				"-v", "secret_access_key=some-secret-access-key",
				"-v", "region=some-region",
				"-v", "az=some-az",
				"-v", "default_key_name=some-key-name",
				"-v", "default_security_group=some-security-group",
				"-v", "subnet_id=some-subnet",
			}),
			Entry("on gcp", bosh.DeployInput{
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
			}, []string{
				"-v", "zone=some-zone",
				"-v", "network=some-network",
				"-v", "subnetwork=some-subnetwork",
				"-v", `tags=[some-internal-tag,some-bosh-open-tag]`,
				"-v", `project_id=some-project-id`,
				"-v", `external_ip=some-external-ip`,
			}),
		)
	})
})
