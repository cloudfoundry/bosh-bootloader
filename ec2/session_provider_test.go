package ec2_test

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ec2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SessionProvider", func() {
	var provider ec2.SessionProvider

	BeforeEach(func() {
		provider = ec2.NewSessionProvider()
	})

	Describe("Session", func() {
		It("returns a Session with the provided configuration", func() {
			session := provider.Session(ec2.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-region",
				EndpointOverride: "some-endpoint-override",
			})

			_, ok := session.(ec2.Session)
			Expect(ok).To(BeTrue())

			client, ok := session.(*awsec2.EC2)
			Expect(ok).To(BeTrue())

			Expect(client.Config.Credentials).To(Equal(credentials.NewStaticCredentials("some-access-key-id", "some-secret-access-key", "")))
			Expect(client.Config.Region).To(Equal(goaws.String("some-region")))
			Expect(client.Config.Endpoint).To(Equal(goaws.String("some-endpoint-override")))
		})
	})
})
