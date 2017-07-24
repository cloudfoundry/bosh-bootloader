package aws_test

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudFormationOpsGenerator", func() {
	Describe("Generate", func() {
		var (
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			infrastructureManager     *fakes.InfrastructureManager
			opsGenerator              aws.CloudFormationOpsGenerator

			incomingState   storage.State
			expectedOpsYAML []byte
		)

		BeforeEach(func() {
			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
			infrastructureManager = &fakes.InfrastructureManager{}

			incomingState = storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region: "us-east-1",
				},
				Stack: storage.Stack{
					Name: "some-stack",
				},
			}

			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"us-east-1a", "us-east-1b", "us-east-1c"}

			infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
				Outputs: map[string]string{
					"InternalSecurityGroup": "some-internal-security-group",
					"InternalSubnet1Name":   "some-internal-subnet-ids-1",
					"InternalSubnet1CIDR":   "10.0.16.0/20",
					"InternalSubnet2Name":   "some-internal-subnet-ids-2",
					"InternalSubnet2CIDR":   "10.0.32.0/20",
					"InternalSubnet3Name":   "some-internal-subnet-ids-3",
					"InternalSubnet3CIDR":   "10.0.48.0/20",
				},
			}

			var err error
			expectedOpsYAML, err = ioutil.ReadFile(filepath.Join("fixtures", "aws-ops.yml"))
			Expect(err).NotTo(HaveOccurred())

			opsGenerator = aws.NewCloudFormationOpsGenerator(availabilityZoneRetriever, infrastructureManager)
		})

		It("returns an ops file to transform base cloud config into aws specific cloud config", func() {
			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("us-east-1"))
			Expect(infrastructureManager.DescribeCall.Receives.StackName).To(Equal("some-stack"))

			Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOpsYAML))
		})

		DescribeTable("returns an ops file with additional vm extensions to support lb", func(lbType string, opsFile string, lbOutputs map[string]string) {
			incomingState.LB.Type = lbType

			expectedLBOpsFile, err := ioutil.ReadFile(filepath.Join("fixtures", opsFile))
			Expect(err).NotTo(HaveOccurred())

			expectedOps := strings.Join([]string{string(expectedOpsYAML), string(expectedLBOpsFile)}, "\n")

			for k, v := range lbOutputs {
				infrastructureManager.DescribeCall.Returns.Stack.Outputs[k] = v
			}
			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOps))
		},
			Entry("cf load balancer exists", "cf", "cloudformation-aws-cf-lb-ops.yml", map[string]string{
				"CFRouterLoadBalancer":                "some-cf-router-lb",
				"CFRouterInternalSecurityGroup":       "some-cf-router-internal-security-group",
				"CFSSHProxyLoadBalancer":              "some-cf-ssh-proxy-lb",
				"CFSSHProxyInternalSecurityGroup":     "some-cf-ssh-proxy-internal-security-group",
				"CFnick-da-gawdLoadBalancer":          "some-cf-nick-da-gawd-lb",
				"CFnick-da-gawdInternalSecurityGroup": "some-cf-nick-da-gawd-internal-security-group",
			}),
			Entry("concourse load balancer exists", "concourse", "aws-concourse-lb-ops.yml", map[string]string{
				"ConcourseLoadBalancer":          "some-concourse-lb-name",
				"ConcourseInternalSecurityGroup": "some-concourse-lb-internal-security-group",
			}),
		)

		Context("failure cases", func() {
			It("returns an error when az retriever fails to retrieve", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("failed to retrieve")
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to retrieve"))
			})

			It("returns an error when the infrastructure manager fails to describe stack", func() {
				infrastructureManager.DescribeCall.Returns.Error = errors.New("failed to describe")
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to describe"))
			})

			It("returns an error when it fails to parse a cidr block", func() {
				infrastructureManager.DescribeCall.Returns.Stack.Outputs["InternalSubnet1CIDR"] = "****"
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
		})
	})
})
