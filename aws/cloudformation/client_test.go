package cloudformation_test

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"

	goaws "github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe("NewClient", func() {
		It("returns a Client with the provided configuration", func() {
			client := cloudformation.NewClient(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-region",
				EndpointOverride: "some-endpoint-override",
			})

			_, ok := client.(cloudformation.Client)
			Expect(ok).To(BeTrue())

			cloudformationClient, ok := client.(*awscloudformation.CloudFormation)
			Expect(ok).To(BeTrue())

			Expect(cloudformationClient.Config.Credentials).To(Equal(credentials.NewStaticCredentials("some-access-key-id", "some-secret-access-key", "")))
			Expect(cloudformationClient.Config.Region).To(Equal(goaws.String("some-region")))
			Expect(cloudformationClient.Config.Endpoint).To(Equal(goaws.String("some-endpoint-override")))
		})
	})
})
