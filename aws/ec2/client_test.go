package ec2_test

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"

	goaws "github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe("NewClient", func() {
		It("returns a Client with the provided configuration", func() {
			client := ec2.NewClient(aws.Config{
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				Region:          "some-region",
			})

			_, ok := client.(ec2.Client)
			Expect(ok).To(BeTrue())

			ec2Client, ok := client.(*awsec2.EC2)
			Expect(ok).To(BeTrue())

			Expect(ec2Client.Config.Credentials).To(Equal(credentials.NewStaticCredentials("some-access-key-id", "some-secret-access-key", "")))
			Expect(ec2Client.Config.Region).To(Equal(goaws.String("some-region")))
		})
	})
})
