package aws_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/aws"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputGenerator", func() {
	var (
		executor        *fakes.TerraformExecutor
		outputGenerator aws.OutputGenerator
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		executor.OutputsCall.Returns.Outputs = map[string]interface{}{
			"bosh_eip":                      "some-bosh-eip",
			"bosh_url":                      "some-bosh-url",
			"bosh_user_access_key":          "some-bosh-user-access-key",
			"bosh_user_secret_access_key":   "some-bosh-user-secret-access_key",
			"nat_eip":                       "some-nat-eip",
			"bosh_subnet_id":                "some-bosh-subnet-id",
			"bosh_subnet_availability_zone": "some-bosh-subnet-availability-zone",
			"bosh_security_group":           "some-bosh-security-group",
			"internal_security_group":       "some-internal-security-group",
			"internal_subnet_ids":           "some-internal-subnet-ids",
			"internal_subnet_cidrs":         "some-internal-subnet-cidrs",
			"vpc_id":                        "some-vpc-id",
		}

		outputGenerator = aws.NewOutputGenerator(executor)
	})

	Context("when no lb exists", func() {
		It("returns all terraform outputs except lb related outputs", func() {
			outputs, err := outputGenerator.Generate(storage.State{
				IAAS:    "aws",
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
				LB: storage.LB{
					Type:   "",
					Domain: "",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(executor.OutputsCall.Receives.TFState).To(Equal("some-tf-state"))

			Expect(outputs).To(Equal(map[string]interface{}{
				"az":                      "some-bosh-subnet-availability-zone",
				"external_ip":             "some-bosh-eip",
				"director_address":        "some-bosh-url",
				"access_key_id":           "some-bosh-user-access-key",
				"secret_access_key":       "some-bosh-user-secret-access_key",
				"subnet_id":               "some-bosh-subnet-id",
				"default_security_groups": "some-bosh-security-group",
				"internal_security_group": "some-internal-security-group",
				"internal_subnet_ids":     "some-internal-subnet-ids",
				"internal_subnet_cidrs":   "some-internal-subnet-cidrs",
			}))
		})
	})

	Context("failure cases", func() {
		Context("when the executor fails to retrieve the outputs", func() {
			It("returns an error", func() {
				executor.OutputsCall.Returns.Error = errors.New("no can do")

				_, err := outputGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("no can do"))
			})
		})
	})
})
