package cloudstack_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/cloudstack"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpsGenerator", func() {
	var (
		terraformManager *fakes.TerraformManager
		opsGenerator     cloudstack.OpsGenerator

		incomingState storage.State
	)

	BeforeEach(func() {
		terraformManager = &fakes.TerraformManager{}

		incomingState = storage.State{
			IAAS:       "cloudstack",
			EnvID:      "myenv",
			CloudStack: storage.CloudStack{},
		}

		terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
			"internal_subnet_cidr_mapping": map[string]interface{}{
				"subnet1": "10.0.16.0/20",
				"subnet2": "10.0.48.0/20",
				"subnet3": "10.0.32.0/20",
			},
			"internal_subnet_gw_mapping": map[string]interface{}{
				"subnet1": "10.0.16.1",
				"subnet2": "10.0.48.1",
				"subnet3": "10.0.32.1",
			},
			"dns": []string{
				"8.8.8.8",
			},
		}}

		opsGenerator = cloudstack.NewOpsGenerator(terraformManager)
	})

	Describe("GenerateVars", func() {
		It("returns the contents for a cloud config vars file", func() {
			varsYAML, err := opsGenerator.GenerateVars(incomingState)

			Expect(err).NotTo(HaveOccurred())
			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(varsYAML).To(MatchYAML(`
cidr_subnet1: 10.0.16.0/20
cidr_subnet2: 10.0.48.0/20
cidr_subnet3: 10.0.32.0/20
dns:
- 8.8.8.8
gw_subnet1: 10.0.16.1
gw_subnet2: 10.0.48.1
gw_subnet3: 10.0.32.1
internal_subnet_cidr_mapping:
  subnet1: 10.0.16.0/20
  subnet2: 10.0.48.0/20
  subnet3: 10.0.32.0/20
internal_subnet_gw_mapping:
  subnet1: 10.0.16.1
  subnet2: 10.0.48.1
  subnet3: 10.0.32.1
reserved_1_subnet1: 10.0.16.2-10.0.16.6
reserved_1_subnet2: 10.0.48.2-10.0.48.6
reserved_1_subnet3: 10.0.32.2-10.0.32.6
static_subnet1: 10.0.31.55-10.0.31.254
static_subnet2: 10.0.63.55-10.0.63.254
static_subnet3: 10.0.47.55-10.0.47.254
`))
		})

		Context("failure cases", func() {
			Context("when the az subnet id map has a key not in the cidr map", func() {
				BeforeEach(func() {
					delete(terraformManager.GetOutputsCall.Returns.Outputs.Map, "internal_subnet_cidr_mapping")
				})
				It("returns an error", func() {
					_, err := opsGenerator.GenerateVars(incomingState)
					Expect(err.Error()).To(ContainSubstring("missing internal_subnet_cidr_mapping terraform output"))
				})
			})
			Context("when terraform fails to get outputs", func() {
				It("returns an error", func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("breadfruit")
					_, err := opsGenerator.GenerateVars(incomingState)
					Expect(err).To(MatchError("Get terraform outputs: breadfruit"))
				})
			})

			Context("when cidr block parsing fails", func() {
				It("returns an error", func() {
					terraformManager.GetOutputsCall.Returns.Outputs.Map["internal_subnet_cidr_mapping"] = map[string]interface{}{
						"subnet": "****",
					}
					_, err := opsGenerator.GenerateVars(incomingState)
					Expect(err).To(MatchError(`"****" cannot parse CIDR block`))
				})
			})

			DescribeTable("when a terraform output is missing", func(outputKey, lbType string) {
				delete(terraformManager.GetOutputsCall.Returns.Outputs.Map, outputKey)
				incomingState.LB.Type = lbType
				_, err := opsGenerator.GenerateVars(incomingState)
				Expect(err).To(MatchError(fmt.Sprintf("missing %s terraform output", outputKey)))
			},
				Entry("when dns is missing", "dns", ""),

				Entry("when internal_subnet_cidr_mapping is missing", "internal_subnet_cidr_mapping", ""),
				Entry("when internal_subnet_gw_mapping is missing", "internal_subnet_gw_mapping", ""),
			)
		})
	})

	Describe("Generate", func() {
		var expectedOpsYAML string
		BeforeEach(func() {

		})

		Context("when there is no iso-segment", func() {
			BeforeEach(func() {
				var err error
				baseOpsYAMLContents, err := ioutil.ReadFile(filepath.Join("fixtures", "ops-file.yml"))
				Expect(err).NotTo(HaveOccurred())
				expectedOpsYAML = string(baseOpsYAMLContents)
			})

			It("returns an ops file to transform base cloud config into cloudstack specific cloud config", func() {
				opsYAML, err := opsGenerator.Generate(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(opsYAML).To(MatchYAML(expectedOpsYAML))
			})
		})

		Context("when there is iso-segment", func() {
			BeforeEach(func() {
				baseOpsYAMLContents, err := ioutil.ReadFile(filepath.Join("fixtures", "ops-file.yml"))
				Expect(err).NotTo(HaveOccurred())
				lbsOpsYAMLContents, err := ioutil.ReadFile(filepath.Join("fixtures", "iso-segment.yml"))
				Expect(err).NotTo(HaveOccurred())
				expectedOpsYAML = strings.Join([]string{string(baseOpsYAMLContents), string(lbsOpsYAMLContents)}, "\n")
			})

			It("returns an ops file to transform base cloud config into cloudstack specific cloud config", func() {
				incomingState.CloudStack.IsoSegment = true
				opsYAML, err := opsGenerator.Generate(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(opsYAML).To(MatchYAML(expectedOpsYAML))
			})
		})

	})
})
