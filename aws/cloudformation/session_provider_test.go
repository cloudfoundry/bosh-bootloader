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

var _ = Describe("SessionProvider", func() {
	var provider cloudformation.SessionProvider

	BeforeEach(func() {
		provider = cloudformation.NewSessionProvider()
	})

	Describe("Session", func() {
		It("returns a Session with the provided configuration", func() {
			session, err := provider.Session(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-region",
				EndpointOverride: "some-endpoint-override",
			})
			Expect(err).NotTo(HaveOccurred())

			_, ok := session.(cloudformation.Session)
			Expect(ok).To(BeTrue())

			client, ok := session.(*awscloudformation.CloudFormation)
			Expect(ok).To(BeTrue())

			Expect(client.Config.Credentials).To(Equal(credentials.NewStaticCredentials("some-access-key-id", "some-secret-access-key", "")))
			Expect(client.Config.Region).To(Equal(goaws.String("some-region")))
			Expect(client.Config.Endpoint).To(Equal(goaws.String("some-endpoint-override")))
		})

		Context("failure cases", func() {
			It("returns an error when the credentials are not provided", func() {
				_, err := provider.Session(aws.Config{})
				Expect(err).To(MatchError("aws access key id must be provided"))
			})
		})
	})
})
