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
			"nat_ssh_key_pair_name":  "some-key-pair-name",
			"access_key":             "some-access-key-id",
			"secret_key":             "some-secret-access-key",
			"region":                 "some-region",
			"bosh_availability_zone": "some-zone",
			"availability_zones":     `["z1","z2","z3"]`,
		}))
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
