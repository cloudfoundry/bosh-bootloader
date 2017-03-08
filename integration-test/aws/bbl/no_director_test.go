package integration_test

import (
	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("no director test", func() {
	var (
		bbl   actors.BBL
		aws   actors.AWS
		state integration.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadAWSConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "no-director-env")
		aws = actors.NewAWS(configuration)
		state = integration.NewState(configuration.StateFileDir)
	})

	It("successfully bbls up and destroys with no director", func() {
		var (
			stackName string
		)

		By("calling bbl up with the no-director flag", func() {
			bbl.Up(actors.AWSIAAS, []string{"--name", bbl.PredefinedEnvID(), "--no-director"})
		})

		By("checking that the stack exists", func() {
			stackName = state.StackName()

			Expect(aws.StackExists(stackName)).To(BeTrue())

			natInstanceID := aws.GetPhysicalID(stackName, "NATInstance")
			Expect(natInstanceID).NotTo(BeEmpty())

			tags := aws.GetEC2InstanceTags(natInstanceID)
			Expect(tags["bbl-env-id"]).To(Equal(bbl.PredefinedEnvID()))
		})

		By("checking that the bosh director does not exists", func() {
			directorAddress := bbl.DirectorAddress()
			Expect(directorAddress).To(Equal(""))
		})

		By("calling bbl destroy", func() {
			bbl.Destroy()
		})

		By("checking that the stack no longer exists", func() {
			Expect(aws.StackExists(stackName)).To(BeFalse())
		})
	})
})
