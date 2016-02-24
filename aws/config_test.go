package aws_test

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
)

var _ = Describe("Config", func() {
	Context("ValidateCredentials", func() {
		It("validates that the credentials have been set", func() {
			config := aws.Config{
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: " some-secret-access-key",
				Region:          "some-region",
			}

			err := config.ValidateCredentials()
			Expect(err).NotTo(HaveOccurred())
		})
		Context("failure cases", func() {
			It("returns an error when the access key id is missing", func() {
				config := aws.Config{
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				}

				Expect(config.ValidateCredentials()).To(MatchError("aws access key id must be provided"))
			})

			It("returns an error when the secret access key is missing", func() {
				config := aws.Config{
					AccessKeyID: "some-access-key-id",
					Region:      "some-region",
				}

				Expect(config.ValidateCredentials()).To(MatchError("aws secret access key must be provided"))
			})

			It("returns an error when the region is missing", func() {
				config := aws.Config{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
				}

				Expect(config.ValidateCredentials()).To(MatchError("aws region must be provided"))
			})
		})
	})

	Context("SessionConfig", func() {
		It("returns an AWS config which is consumable by AWS session functions", func() {
			config := aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  " some-secret-access-key",
				Region:           "some-region",
				EndpointOverride: "some-endpoint-override",
			}

			awsConfig := &goaws.Config{
				Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
				Region:      goaws.String(config.Region),
				Endpoint:    goaws.String(config.EndpointOverride),
			}

			Expect(config.SessionConfig()).To(Equal(awsConfig))
		})
	})
})
