package aws_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpsGenerator", func() {
	var (
		terraformManager          *fakes.TerraformManager
		availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
		opsGenerator              aws.OpsGenerator

		incomingState storage.State
	)

	BeforeEach(func() {
		terraformManager = &fakes.TerraformManager{}
		availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}

		incomingState = storage.State{
			IAAS: "aws",
			AWS: storage.AWS{
				Region: "us-east-1",
			},
		}

		terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
			"internal_security_group":              "some-internal-security-group",
			"cf_router_lb_name":                    "some-cf-router-lb-name",
			"cf_router_lb_internal_security_group": "some-cf-router-lb-internal-security-group",
			"cf_ssh_lb_name":                       "some-cf-ssh-lb-name",
			"cf_ssh_lb_internal_security_group":    "some-cf-ssh-lb-internal-security-group",
			"cf_tcp_lb_name":                       "some-cf-tcp-lb-name",
			"cf_tcp_lb_internal_security_group":    "some-cf-tcp-lb-internal-security-group",
			"concourse_lb_name":                    "some-concourse-lb-name",
			"concourse_lb_internal_security_group": "some-concourse-lb-internal-security-group",
			"internal_az_subnet_id_mapping": map[string]interface{}{
				"us-east-1c": "some-internal-subnet-ids-3",
				"us-east-1a": "some-internal-subnet-ids-1",
				"us-east-1b": "some-internal-subnet-ids-2",
			},
			"internal_az_subnet_cidr_mapping": map[string]interface{}{
				"us-east-1a": "10.0.16.0/20",
				"us-east-1c": "10.0.48.0/20",
				"us-east-1b": "10.0.32.0/20",
			},
		}}

		opsGenerator = aws.NewOpsGenerator(terraformManager, availabilityZoneRetriever)
	})
	Describe("GenerateVars", func() {
		It("returns the contents for a cloud config vars file", func() {
			varsYAML, err := opsGenerator.GenerateVars(incomingState)

			Expect(err).NotTo(HaveOccurred())
			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(varsYAML).To(gomegamatchers.MatchYAML(`
az1_name: us-east-1a
az1_gateway: 10.0.16.1
az1_range: 10.0.16.0/20
az1_reserved_1: 10.0.16.2-10.0.16.3
az1_reserved_2: 10.0.31.255
az1_static: 10.0.31.190-10.0.31.254
az1_subnet: some-internal-subnet-ids-1
az2_name: us-east-1b
az2_gateway: 10.0.32.1
az2_range: 10.0.32.0/20
az2_reserved_1: 10.0.32.2-10.0.32.3
az2_reserved_2: 10.0.47.255
az2_static: 10.0.47.190-10.0.47.254
az2_subnet: some-internal-subnet-ids-2
az3_gateway: 10.0.48.1
az3_name: us-east-1c
az3_range: 10.0.48.0/20
az3_reserved_1: 10.0.48.2-10.0.48.3
az3_reserved_2: 10.0.63.255
az3_static: 10.0.63.190-10.0.63.254
az3_subnet: some-internal-subnet-ids-3
internal_security_group: some-internal-security-group
cf_router_lb_name: some-cf-router-lb-name
cf_router_lb_internal_security_group: some-cf-router-lb-internal-security-group
cf_ssh_lb_name: some-cf-ssh-lb-name
cf_ssh_lb_internal_security_group: some-cf-ssh-lb-internal-security-group
cf_tcp_lb_name: some-cf-tcp-lb-name
cf_tcp_lb_internal_security_group: some-cf-tcp-lb-internal-security-group
concourse_lb_name: some-concourse-lb-name
concourse_lb_internal_security_group: some-concourse-lb-internal-security-group
`))
		})

		Context("failure cases", func() {
			Context("when the az subnet id map has a key not in the cidr map", func() {
				BeforeEach(func() {
					delete(terraformManager.GetOutputsCall.Returns.Outputs.Map["internal_az_subnet_cidr_mapping"].(map[string]interface{}), "us-east-1a")
				})
				It("returns an error", func() {
					_, err := opsGenerator.GenerateVars(incomingState)
					Expect(err).To(MatchError("missing AZ in terraform output: internal_az_subnet_cidr_mapping"))
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
					terraformManager.GetOutputsCall.Returns.Outputs.Map["internal_az_subnet_cidr_mapping"] = map[string]interface{}{
						"us-east-1a": "****",
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
				Entry("when internal_security_group is missing", "internal_security_group", ""),

				Entry("when internal_az_subnet_id_mapping is missing", "internal_az_subnet_id_mapping", "cf"),
				Entry("when internal_az_subnet_cidr_mapping is missing", "internal_az_subnet_cidr_mapping", "cf"),
				Entry("when cf_router_lb_name is missing", "cf_router_lb_name", "cf"),
				Entry("when cf_router_lb_internal_security_group is missing", "cf_router_lb_internal_security_group", "cf"),
				Entry("when cf_ssh_lb_name is missing", "cf_ssh_lb_name", "cf"),
				Entry("when cf_ssh_lb_internal_security_group is missing", "cf_ssh_lb_internal_security_group", "cf"),
				Entry("when cf_tcp_lb_name", "cf_tcp_lb_name", "cf"),
				Entry("when cf_tcp_lb_internal_security_group is missing", "cf_tcp_lb_internal_security_group", "cf"),

				Entry("when concourse_lb_name is missing", "concourse_lb_name", "concourse"),
				Entry("when concourse_lb_internal_security_group is missing", "concourse_lb_internal_security_group", "concourse"),
			)
		})
	})

	Describe("Generate", func() {
		var expectedOpsYAML string
		BeforeEach(func() {
			availabilityZoneRetriever.RetrieveAvailabilityZonesCall.Returns.AZs = []string{"us-east-1a", "us-east-1b", "us-east-1c"}
		})

		Context("when there are no lbs", func() {
			BeforeEach(func() {
				var err error
				baseOpsYAMLContents, err := ioutil.ReadFile(filepath.Join("fixtures", "aws-ops.yml"))
				Expect(err).NotTo(HaveOccurred())
				expectedOpsYAML = string(baseOpsYAMLContents)
			})

			It("returns an ops file to transform base cloud config into aws specific cloud config", func() {
				opsYAML, err := opsGenerator.Generate(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOpsYAML))
			})
		})

		Context("when there are cf lbs", func() {
			BeforeEach(func() {
				baseOpsYAMLContents, err := ioutil.ReadFile(filepath.Join("fixtures", "aws-ops.yml"))
				Expect(err).NotTo(HaveOccurred())
				lbsOpsYAMLContents, err := ioutil.ReadFile(filepath.Join("fixtures", "terraform-aws-cf-lb-ops.yml"))
				Expect(err).NotTo(HaveOccurred())
				expectedOpsYAML = strings.Join([]string{string(baseOpsYAMLContents), string(lbsOpsYAMLContents)}, "\n")
			})

			It("returns an ops file to transform base cloud config into aws specific cloud config", func() {
				incomingState.LB.Type = "cf"
				opsYAML, err := opsGenerator.Generate(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOpsYAML))
			})
		})

		Context("when there is a concourse lb", func() {
			BeforeEach(func() {
				baseOpsYAMLContents, err := ioutil.ReadFile(filepath.Join("fixtures", "aws-ops.yml"))
				Expect(err).NotTo(HaveOccurred())
				lbsOpsYAMLContents, err := ioutil.ReadFile(filepath.Join("fixtures", "aws-concourse-lb-ops.yml"))
				Expect(err).NotTo(HaveOccurred())
				expectedOpsYAML = strings.Join([]string{string(baseOpsYAMLContents), string(lbsOpsYAMLContents)}, "\n")
			})

			It("returns an ops file to transform base cloud config into aws specific cloud config", func() {
				incomingState.LB.Type = "concourse"
				opsYAML, err := opsGenerator.Generate(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOpsYAML))
			})
		})

		Context("when an error occurs", func() {
			Context("when ops fails to marshal", func() {
				It("returns an error", func() {
					aws.SetMarshal(func(interface{}) ([]byte, error) {
						return []byte{}, errors.New("failed to marshal")
					})
					_, err := opsGenerator.Generate(storage.State{})
					Expect(err).To(MatchError("failed to marshal"))
					aws.ResetMarshal()
				})
			})
		})
	})
})
