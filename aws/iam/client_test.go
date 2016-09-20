package iam_test

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"

	goaws "github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe("NewClient", func() {
		It("returns a Client with the provided configuration", func() {
			client := iam.NewClient(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-region",
				EndpointOverride: "some-endpoint-override",
			})

			_, ok := client.(iam.Client)
			Expect(ok).To(BeTrue())

			iamClient, ok := client.(*awsiam.IAM)
			Expect(ok).To(BeTrue())

			Expect(iamClient.Config.Credentials).To(Equal(credentials.NewStaticCredentials("some-access-key-id", "some-secret-access-key", "")))
			Expect(iamClient.Config.Region).To(Equal(goaws.String("some-region")))
			Expect(iamClient.Config.Endpoint).To(Equal(goaws.String("some-endpoint-override")))
		})
	})
})
