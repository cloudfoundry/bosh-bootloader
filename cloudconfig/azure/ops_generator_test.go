package azure_test

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/azure"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AzureOpsGenerator", func() {
	var (
		terraformManager *fakes.TerraformManager
		opsGenerator     azure.OpsGenerator

		incomingState   storage.State
		expectedOpsFile []byte
	)

	BeforeEach(func() {
		terraformManager = &fakes.TerraformManager{}

		incomingState = storage.State{
			IAAS: "azure",
		}

		terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
			"bosh_network_name":           "some-virtual-network-name",
			"bosh_subnet_name":            "some-subnet-name",
			"bosh_default_security_group": "some-security-group",
			"application_gateway":         "some-app-gateway-name",
			"some-key":                    "some-value",
		}}

		var err error
		expectedOpsFile, err = ioutil.ReadFile(filepath.Join("fixtures", "azure-ops.yml"))
		Expect(err).NotTo(HaveOccurred())

		opsGenerator = azure.NewOpsGenerator(terraformManager)
	})

	Describe("GenerateVars", func() {
		It("returns the contents for a cloud config vars file", func() {
			varsYAML, err := opsGenerator.GenerateVars(incomingState)

			Expect(err).NotTo(HaveOccurred())

			Expect(varsYAML).To(MatchYAML(`
az1_gateway: 10.0.16.1
az1_range: 10.0.16.0/20
az1_reserved_1: 10.0.16.2-10.0.16.3
az1_reserved_2: 10.0.31.255
az1_static: 10.0.31.190-10.0.31.254
az2_gateway: 10.0.32.1
az2_range: 10.0.32.0/20
az2_reserved_1: 10.0.32.2-10.0.32.3
az2_reserved_2: 10.0.47.255
az2_static: 10.0.47.190-10.0.47.254
az3_gateway: 10.0.48.1
az3_range: 10.0.48.0/20
az3_reserved_1: 10.0.48.2-10.0.48.3
az3_reserved_2: 10.0.63.255
az3_static: 10.0.63.190-10.0.63.254
bosh_network_name: some-virtual-network-name
bosh_subnet_name: some-subnet-name
bosh_default_security_group: some-security-group
application_gateway: some-app-gateway-name
some-key: some-value
`))

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
		})

		Context("with a load balancer", func() {
			BeforeEach(func() {
				incomingState.LB.Type = "cf"
			})

			It("returns an ops file with the cloud properties for the LB", func() {
				ops, err := opsGenerator.GenerateVars(incomingState)
				Expect(err).ToNot(HaveOccurred())
				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
				Expect(ops).To(ContainSubstring(`application_gateway: some-app-gateway-name`))
			})
		})

		Context("failure cases", func() {
			Context("when terraform output provider fails to retrieve", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to output")
				})

				It("returns an error", func() {
					_, err := opsGenerator.GenerateVars(storage.State{})
					Expect(err).To(MatchError("failed to output"))
				})
			})

			Context("when ops fail to marshal", func() {
				BeforeEach(func() {
					azure.SetMarshal(func(interface{}) ([]byte, error) {
						return []byte{}, errors.New("failed to marshal")
					})
				})

				AfterEach(func() {
					azure.ResetMarshal()
				})

				It("returns an error", func() {
					_, err := opsGenerator.GenerateVars(storage.State{})
					Expect(err).To(MatchError("failed to marshal"))
				})
			})
		})
	})

	Describe("Generate", func() {
		It("returns an ops file to transform the base cloud config into azure specific cloud config", func() {
			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(opsYAML).To(MatchYAML(expectedOpsFile))
		})

		Context("failure cases", func() {
			Context("when ops fail to marshal", func() {
				BeforeEach(func() {
					azure.SetMarshal(func(interface{}) ([]byte, error) {
						return []byte{}, errors.New("failed to marshal")
					})
				})

				AfterEach(func() {
					azure.ResetMarshal()
				})

				It("returns an error", func() {
					_, err := opsGenerator.Generate(storage.State{})
					Expect(err).To(MatchError("failed to marshal"))
				})
			})
		})
	})
})
