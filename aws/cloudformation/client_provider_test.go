package cloudformation_test

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClientProvider", func() {
	var provider cloudformation.ClientProvider

	BeforeEach(func() {
		provider = cloudformation.NewClientProvider()
	})

	Describe("Client", func() {
		It("returns a Client with the provided configuration", func() {
			client, err := provider.Client(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-region",
				EndpointOverride: "some-endpoint-override",
			})
			Expect(err).NotTo(HaveOccurred())

			_, ok := client.(cloudformation.Client)
			Expect(ok).To(BeTrue())

			cloudformationClient, ok := client.(*awscloudformation.CloudFormation)
			Expect(ok).To(BeTrue())

			Expect(cloudformationClient.Config.Credentials).To(Equal(credentials.NewStaticCredentials("some-access-key-id", "some-secret-access-key", "")))
			Expect(cloudformationClient.Config.Region).To(Equal(goaws.String("some-region")))
			Expect(cloudformationClient.Config.Endpoint).To(Equal(goaws.String("some-endpoint-override")))
		})

		Context("failure cases", func() {
			It("returns an error when the credentials are not provided", func() {
				_, err := provider.Client(aws.Config{})
				Expect(err).To(MatchError("aws access key id must be provided"))
			})
		})
	})
})
