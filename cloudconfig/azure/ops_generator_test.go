package azure_test

import (
	"errors"
	"os"
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
			"some-key":                    "some-value",
		}}

		var err error
		expectedOpsFile, err = os.ReadFile(filepath.Join("fixtures", "azure-ops.yml"))
		Expect(err).NotTo(HaveOccurred())

		opsGenerator = azure.NewOpsGenerator(terraformManager)
	})

	Describe("GenerateVars", func() {
		It("returns the contents for a cloud config vars file", func() {
			varsYAML, err := opsGenerator.GenerateVars(incomingState)

			Expect(err).NotTo(HaveOccurred())

			Expect(varsYAML).To(MatchYAML(`
bosh_network_name: some-virtual-network-name
bosh_subnet_name: some-subnet-name
bosh_default_security_group: some-security-group
some-key: some-value
`))

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
		})

		Context("with a cf load balancer", func() {
			BeforeEach(func() {
				incomingState.LB.Type = "cf"

				terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
					"cf_app_gateway_name": "some-app-gateway-name",
				}}
			})

			It("returns an ops file with the cloud properties for the LB", func() {
				ops, err := opsGenerator.GenerateVars(incomingState)
				Expect(err).ToNot(HaveOccurred())
				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
				Expect(ops).To(ContainSubstring(`cf_app_gateway_name: some-app-gateway-name`))
			})
		})

		Context("with a concourse load balancer", func() {
			BeforeEach(func() {
				incomingState.LB.Type = "concourse"

				terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
					"concourse_lb_name": "some-load-balancer-name",
				}}
			})

			It("returns an ops file with the cloud properties for the LB", func() {
				ops, err := opsGenerator.GenerateVars(incomingState)
				Expect(err).ToNot(HaveOccurred())
				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
				Expect(ops).To(ContainSubstring(`concourse_lb_name: some-load-balancer-name`))
			})
		})

		Context("failure cases", func() {
			Context("when terraform output provider fails to retrieve", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("banana")
				})

				It("returns an error", func() {
					_, err := opsGenerator.GenerateVars(storage.State{})
					Expect(err).To(MatchError("Get terraform outputs: banana"))
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
