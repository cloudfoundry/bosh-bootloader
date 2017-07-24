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
		availabilityZoneRetriever *fakes.AvailabilityZoneRetriever

		inputGenerator aws.InputGenerator
	)

	BeforeEach(func() {
		availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
		availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"z1", "z2", "z3"}

		inputGenerator = aws.NewInputGenerator(availabilityZoneRetriever)
	})

	Context("when env-id is greater than 18 characters", func() {
		It("creates a short env-id with truncated env_id and sha1sum[0:7]", func() {
			inputs, err := inputGenerator.Generate(storage.State{
				IAAS:    "aws",
				EnvID:   "some-env-id-that-is-pretty-long",
				TFState: "some-tf-state",
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				KeyPair: storage.KeyPair{
					Name: "some-key-pair-name",
				},
				Stack: storage.Stack{
					BOSHAZ: "some-zone",
				},
				LB: storage.LB{
					Type:   "",
					Domain: "",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(inputs["env_id"]).To(Equal("some-env-id-that-is-pretty-long"))
			Expect(inputs["short_env_id"]).To(Equal("some-env-i-1fc794e"))
		})
	})

	Context("when no lbs exist", func() {
		It("receives BBL state and returns a map of terraform variables", func() {
			inputs, err := inputGenerator.Generate(storage.State{
				IAAS:    "aws",
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				KeyPair: storage.KeyPair{
					Name: "some-key-pair-name",
				},
				Stack: storage.Stack{
					BOSHAZ: "some-zone",
				},
				LB: storage.LB{
					Type:   "",
					Domain: "",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(inputs).To(Equal(map[string]string{
				"env_id":                 "some-env-id",
				"short_env_id":           "some-env-id",
				"nat_ssh_key_pair_name":  "some-key-pair-name",
				"access_key":             "some-access-key-id",
				"secret_key":             "some-secret-access-key",
				"region":                 "some-region",
				"bosh_availability_zone": "some-zone",
				"availability_zones":     `["z1","z2","z3"]`,
			}))
		})
	})

	Context("when a cf lb exists", func() {
		var (
			state storage.State
		)

		BeforeEach(func() {
			state = storage.State{
				IAAS:    "aws",
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				KeyPair: storage.KeyPair{
					Name: "some-key-pair-name",
				},
				Stack: storage.Stack{
					BOSHAZ:          "some-zone",
					CertificateName: "some-certificate-name",
				},
				LB: storage.LB{
					Type:  "cf",
					Cert:  "some-cert",
					Chain: "some-chain",
					Key:   "some-key",
				},
			}
		})

		It("returns a map with additional cf load balancer inputs", func() {
			inputs, err := inputGenerator.Generate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(inputs).To(Equal(map[string]string{
				"env_id":                      "some-env-id",
				"short_env_id":                "some-env-id",
				"nat_ssh_key_pair_name":       "some-key-pair-name",
				"access_key":                  "some-access-key-id",
				"secret_key":                  "some-secret-access-key",
				"region":                      "some-region",
				"bosh_availability_zone":      "some-zone",
				"availability_zones":          `["z1","z2","z3"]`,
				"ssl_certificate_name_prefix": "",
				"ssl_certificate_name":        "some-certificate-name",
			}))
		})

		Context("when a domain name is supplied", func() {
			BeforeEach(func() {
				state.LB.Domain = "some-domain"
			})

			It("returns a map with additional domain input", func() {
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

				Expect(inputs).To(Equal(map[string]string{
					"env_id":                      "some-env-id",
					"short_env_id":                "some-env-id",
					"nat_ssh_key_pair_name":       "some-key-pair-name",
					"access_key":                  "some-access-key-id",
					"secret_key":                  "some-secret-access-key",
					"region":                      "some-region",
					"bosh_availability_zone":      "some-zone",
					"availability_zones":          `["z1","z2","z3"]`,
					"ssl_certificate_name":        "some-certificate-name",
					"ssl_certificate_name_prefix": "",
					"system_domain":               "some-domain",
				}))
			})
		})
	})

	Context("when a concourse lb exists", func() {
		var (
			state storage.State
		)

		BeforeEach(func() {
			state = storage.State{
				IAAS:    "aws",
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				KeyPair: storage.KeyPair{
					Name: "some-key-pair-name",
				},
				Stack: storage.Stack{
					BOSHAZ: "some-zone",
				},
				LB: storage.LB{
					Type:  "concourse",
					Cert:  "some-cert",
					Chain: "some-chain",
					Key:   "some-key",
				},
			}
		})

		It("returns a map with additional concourse load balancer inputs", func() {
			inputs, err := inputGenerator.Generate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(inputs).To(Equal(map[string]string{
				"env_id":                      "some-env-id",
				"short_env_id":                "some-env-id",
				"nat_ssh_key_pair_name":       "some-key-pair-name",
				"access_key":                  "some-access-key-id",
				"secret_key":                  "some-secret-access-key",
				"region":                      "some-region",
				"bosh_availability_zone":      "some-zone",
				"availability_zones":          `["z1","z2","z3"]`,
				"ssl_certificate":             "some-cert",
				"ssl_certificate_chain":       "some-chain",
				"ssl_certificate_private_key": "some-key",
				"ssl_certificate_name":        "",
				"ssl_certificate_name_prefix": "some-env-id",
			}))
		})
	})

	Context("failure cases", func() {
		Context("when the availability zone retriever fails", func() {
			It("returns an error", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("failed to get zones")

				_, err := inputGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to get zones"))
			})
		})

		Context("when the azs failed to marshal", func() {
			BeforeEach(func() {
				aws.SetJSONMarshal(func(interface{}) ([]byte, error) {
					return []byte{}, errors.New("failed to marshal")
				})
			})

			AfterEach(func() {
				aws.ResetJSONMarshal()
			})

			It("returns an error", func() {
				_, err := inputGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to marshal"))
			})
		})
	})
})
