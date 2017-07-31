package aws_test

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/cloudfoundry/bosh-bootloader/aws"
)

var _ = Describe("Config", func() {
	Describe("ClientConfig", func() {
		It("returns an AWS config which is consumable by AWS client functions", func() {
			config := aws.Config{
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: " some-secret-access-key",
				Region:          "some-region",
			}

			awsConfig := &goaws.Config{
				Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
				Region:      goaws.String(config.Region),
			}

			Expect(config.ClientConfig()).To(Equal(awsConfig))
		})
	})
})
