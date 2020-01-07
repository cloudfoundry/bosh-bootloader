package aws_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/aws"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InputGenerator", func() {
	var (
		awsClient *fakes.AWSClient

		inputGenerator aws.InputGenerator
	)

	BeforeEach(func() {
		awsClient = &fakes.AWSClient{}
		awsClient.RetrieveAZsCall.Returns.AZs = []string{"z1", "z2", "z3"}

		inputGenerator = aws.NewInputGenerator(awsClient)
	})

	Describe("Generate", func() {
		Context("when env-id is greater than 18 characters", func() {
			It("creates a short env-id with truncated env_id and sha1sum[0:7]", func() {
				inputs, err := inputGenerator.Generate(storage.State{
					EnvID: "some-env-id-that-is-pretty-long",
					AWS: storage.AWS{
						Region: "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(awsClient.RetrieveAZsCall.Receives.Region).To(Equal("some-region"))

				Expect(inputs["env_id"]).To(Equal("some-env-id-that-is-pretty-long"))
				Expect(inputs["short_env_id"]).To(Equal("some-env-i-1fc794e"))
			})
		})

		It("receives BBL state and returns a map of terraform variables", func() {
			inputs, err := inputGenerator.Generate(storage.State{
				EnvID: "some-env-id",
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				LB: storage.LB{
					Type: "concourse",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(awsClient.RetrieveAZsCall.Receives.Region).To(Equal("some-region"))

			Expect(inputs).To(Equal(map[string]interface{}{
				"env_id":             "some-env-id",
				"short_env_id":       "some-env-id",
				"region":             "some-region",
				"availability_zones": []string{"z1", "z2", "z3"},
			}))
		})

		Context("when a cf lb exists", func() {
			var state storage.State

			BeforeEach(func() {
				state = storage.State{
					IAAS:  "aws",
					EnvID: "some-env-id",
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
					LB: storage.LB{
						Type: "cf",
						Cert: "some-cert",
						Key:  "some-key",
					},
				}
			})

			It("returns a map with additional cf load balancer inputs", func() {
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(awsClient.RetrieveAZsCall.Receives.Region).To(Equal("some-region"))

				Expect(inputs).To(Equal(map[string]interface{}{
					"env_id":                      "some-env-id",
					"short_env_id":                "some-env-id",
					"region":                      "some-region",
					"availability_zones":          []string{"z1", "z2", "z3"},
					"ssl_certificate":             "some-cert",
					"ssl_certificate_private_key": "some-key",
				}))
			})

			Context("when a domain name is supplied", func() {
				BeforeEach(func() {
					state.LB.Domain = "some-domain"
					awsClient.RetrieveDNSCall.Returns.DNS = "zone-id"
				})

				It("returns a map with additional domain input", func() {
					inputs, err := inputGenerator.Generate(state)
					Expect(err).NotTo(HaveOccurred())

					Expect(awsClient.RetrieveAZsCall.Receives.Region).To(Equal("some-region"))
					Expect(awsClient.RetrieveDNSCall.Receives.URL).To(Equal("some-domain"))

					Expect(inputs).To(Equal(map[string]interface{}{
						"env_id":                      "some-env-id",
						"short_env_id":                "some-env-id",
						"region":                      "some-region",
						"availability_zones":          []string{"z1", "z2", "z3"},
						"ssl_certificate":             "some-cert",
						"ssl_certificate_private_key": "some-key",
						"system_domain":               "some-domain",
						"parent_zone":                 "zone-id",
					}))
				})
			})
		})

		Context("failure cases", func() {
			Context("when the availability zone retriever fails", func() {
				It("returns an error", func() {
					awsClient.RetrieveAZsCall.Returns.Error = errors.New("failed to get zones")

					_, err := inputGenerator.Generate(storage.State{})
					Expect(err).To(MatchError("failed to get zones"))
				})
			})
		})
	})

	Describe("Credentials", func() {
		It("returns the access key and secret key", func() {
			state := storage.State{
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
			}

			credentials := inputGenerator.Credentials(state)

			Expect(credentials).To(Equal(map[string]string{
				"access_key": "some-access-key-id",
				"secret_key": "some-secret-access-key",
			}))
		})
	})
})
