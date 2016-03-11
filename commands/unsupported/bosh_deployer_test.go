package unsupported_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BoshDeployer", func() {
	var (
		stackManager         *fakes.StackManager
		manifestBuilder      *fakes.BOSHInitManifestBuilder
		cloudFormationClient *fakes.CloudFormationClient

		boshDeployer unsupported.BOSHDeployer
	)

	BeforeEach(func() {
		stackManager = &fakes.StackManager{}
		manifestBuilder = &fakes.BOSHInitManifestBuilder{}

		boshDeployer = unsupported.NewBOSHDeployer(stackManager, manifestBuilder)
	})

	Describe("Deploy", func() {
		It("deploys bosh and returns a key pair", func() {
			stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
				Outputs: map[string]string{
					"BOSHSubnet":              "subnet-12345",
					"BOSHSubnetAZ":            "some-az",
					"BOSHEIP":                 "some-elastic-ip",
					"BOSHUserAccessKey":       "some-access-key-id",
					"BOSHUserSecretAccessKey": "some-secret-access-key",
					"BOSHSecurityGroup":       "some-security-group",
				},
			}
			manifestBuilder.BuildCall.Returns.Properties = boshinit.ManifestProperties{
				SSLKeyPair: ssl.KeyPair{
					Certificate: []byte("updated-certificate"),
					PrivateKey:  []byte("updated-private-key"),
				},
			}

			keyPair, err := boshDeployer.Deploy(cloudFormationClient, "some-aws-region", "some-keypair-name",
				ssl.KeyPair{
					Certificate: []byte("some-certificate"),
					PrivateKey:  []byte("some-private-key"),
				})
			Expect(err).NotTo(HaveOccurred())

			Expect(stackManager.DescribeCall.Receives.Client).To(Equal(cloudFormationClient))
			Expect(stackManager.DescribeCall.Receives.StackName).To(Equal("concourse"))
			Expect(manifestBuilder.BuildCall.Receives.Properties).To(Equal(boshinit.ManifestProperties{
				SubnetID:         "subnet-12345",
				AvailabilityZone: "some-az",
				ElasticIP:        "some-elastic-ip",
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				DefaultKeyName:   "some-keypair-name",
				Region:           "some-aws-region",
				SecurityGroup:    "some-security-group",
				SSLKeyPair: ssl.KeyPair{
					Certificate: []byte("some-certificate"),
					PrivateKey:  []byte("some-private-key"),
				},
			}))
			Expect(keyPair).To(Equal(ssl.KeyPair{
				Certificate: []byte("updated-certificate"),
				PrivateKey:  []byte("updated-private-key"),
			}))
		})

		Context("failure cases", func() {
			Context("when the stack cannot be described", func() {
				It("returns an error", func() {
					stackManager.DescribeCall.Returns.Error = errors.New("failed to describe stack")

					_, err := boshDeployer.Deploy(cloudFormationClient, "some-aws-region", "some-keypair-name", ssl.KeyPair{})
					Expect(err).To(MatchError("failed to describe stack"))
				})
			})

			Context("when the manifest cannot be built", func() {
				It("returns an error", func() {
					manifestBuilder.BuildCall.Returns.Error = errors.New("failed to build manifest")

					_, err := boshDeployer.Deploy(cloudFormationClient, "some-aws-region", "some-keypair-name", ssl.KeyPair{})
					Expect(err).To(MatchError("failed to build manifest"))
				})
			})
		})
	})
})
