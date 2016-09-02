package integration_test

import (
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("keypair reentrant", func() {
	var (
		bbl actors.BBL
		aws actors.AWS
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration)
		aws = actors.NewAWS(configuration)
	})

	It("keeps the keypair name through failure to bbl up", func() {
		var envID string

		By("bbl-ing up with invalid AWS credentials", func() {
			bbl.UpWithInvalidAWSCredentials()
		})

		By("capturing env ID from failed attempt to bbl up", func() {
			envID = bbl.EnvID()
		})

		By("bbl-ing up with valid aws access key id and aws secret access key", func() {
			bbl.Up()
		})

		By("checking if the keypair name contains the env id from the failed attempt to bbl up", func() {
			expectedKeyName := fmt.Sprintf("keypair-%s", envID)

			keyPairs := aws.DescribeKeyPairs(expectedKeyName)

			Expect(keyPairs).To(HaveLen(1))
			Expect(*keyPairs[0].KeyName).To(Equal(expectedKeyName))
		})
	})

})
