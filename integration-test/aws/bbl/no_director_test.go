package integration_test

import (
	"fmt"

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
		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "no-director-env")
		aws = actors.NewAWS(configuration)
		state = integration.NewState(configuration.StateFileDir)
	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	It("successfully standups up a no director infrastructure", func() {
		By("calling bbl up with the no-director flag", func() {
			bbl.Up(actors.AWSIAAS, []string{"--name", bbl.PredefinedEnvID(), "--no-director"})
		})

		By("checking that an instance exists", func() {
			instances := aws.Instances(fmt.Sprintf("%s-vpc", bbl.PredefinedEnvID()))
			Expect(instances).To(HaveLen(1))
			Expect(instances).To(Equal([]string{fmt.Sprintf("%s-nat", bbl.PredefinedEnvID())}))

			tags := aws.GetEC2InstanceTags(fmt.Sprintf("%s-nat", bbl.PredefinedEnvID()))
			Expect(tags["EnvID"]).To(Equal(bbl.PredefinedEnvID()))
		})

		By("checking that director details are not printed", func() {
			directorUsername := bbl.DirectorUsername()
			Expect(directorUsername).To(Equal(""))
		})
	})
})
