package elb_test

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/elb"

	goaws "github.com/aws/aws-sdk-go/aws"
	awselb "github.com/aws/aws-sdk-go/service/elb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe("NewClient", func() {
		It("returns a Client with the provided configuration", func() {
			client := elb.NewClient(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-region",
				EndpointOverride: "some-endpoint-override",
			})

			_, ok := client.(elb.Client)
			Expect(ok).To(BeTrue())

			elbClient, ok := client.(*awselb.ELB)
			Expect(ok).To(BeTrue())

			Expect(elbClient.Config.Credentials).To(Equal(credentials.NewStaticCredentials("some-access-key-id", "some-secret-access-key", "")))
			Expect(elbClient.Config.Region).To(Equal(goaws.String("some-region")))
			Expect(elbClient.Config.Endpoint).To(Equal(goaws.String("some-endpoint-override")))
		})
	})
})
