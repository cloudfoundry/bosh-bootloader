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
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("TerraformOpsGenerator", func() {
	Describe("Generate", func() {
		var (
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			terraformManager          *fakes.TerraformManager
			opsGenerator              aws.TerraformOpsGenerator

			incomingState   storage.State
			expectedOpsYAML string
		)

		BeforeEach(func() {
			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
			terraformManager = &fakes.TerraformManager{}

			incomingState = storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region: "us-east-1",
				},
				TFState: "some-tf-state",
			}

			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{
				"us-east-1a",
				"us-east-1b",
				"us-east-1c",
			}

			terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
				"internal_subnet_cidrs": []interface{}{
					"10.0.16.0/20",
					"10.0.32.0/20",
					"10.0.48.0/20",
				},
				"internal_subnet_ids": []interface{}{
					"some-internal-subnet-ids-1",
					"some-internal-subnet-ids-2",
					"some-internal-subnet-ids-3",
				},
				"internal_security_group":              "some-internal-security-group",
				"cf_router_lb_name":                    "some-cf-router-lb-name",
				"cf_router_lb_internal_security_group": "some-cf-router-lb-internal-security-group",
				"cf_ssh_lb_name":                       "some-cf-ssh-lb-name",
				"cf_ssh_lb_internal_security_group":    "some-cf-ssh-lb-internal-security-group",
				"cf_tcp_lb_name":                       "some-cf-tcp-lb-name",
				"cf_tcp_lb_internal_security_group":    "some-cf-tcp-lb-internal-security-group",
				"concourse_lb_name":                    "some-concourse-lb-name",
				"concourse_lb_internal_security_group": "some-concourse-lb-internal-security-group",
			}

			opsGenerator = aws.NewTerraformOpsGenerator(availabilityZoneRetriever, terraformManager)
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

				Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("us-east-1"))
				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingState))

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

				Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("us-east-1"))
				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingState))

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

				Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("us-east-1"))
				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingState))

				Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOpsYAML))
			})
		})

		Context("failure cases", func() {
			It("returns an error when az retriever fails to retrieve", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("failed to retrieve")
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to retrieve"))
			})

			It("returns an error when the infrastructure manager fails to describe stack", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get outputs")
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to get outputs"))
			})

			It("returns an error when it fails to parse a cidr block", func() {
				terraformManager.GetOutputsCall.Returns.Outputs["internal_subnet_cidrs"] = []interface{}{
					"****",
				}
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError(`"****" cannot parse CIDR block`))
			})

			It("returns an error when ops fails to marshal", func() {
				aws.SetMarshal(func(interface{}) ([]byte, error) {
					return []byte{}, errors.New("failed to marshal")
				})
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to marshal"))
				aws.ResetMarshal()
			})

			DescribeTable("when an terraform output is missing", func(outputKey, lbType string) {
				delete(terraformManager.GetOutputsCall.Returns.Outputs, outputKey)
				_, err := opsGenerator.Generate(storage.State{
					LB: storage.LB{
						Type: lbType,
					},
				})
				Expect(err).To(MatchError(fmt.Sprintf("missing %s terraform output", outputKey)))
			},
				Entry("when internal_subnet_cidrs is missing", "internal_subnet_cidrs", ""),
				Entry("when internal_subnet_ids is missing", "internal_subnet_ids", ""),
				Entry("when internal_security_group is missing", "internal_security_group", ""),

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
})
