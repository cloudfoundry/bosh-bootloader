package acceptance_test

import (
	// "fmt"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = FDescribe("up test", func() {
	var (
		azure  actors.Azure
		bbl    actors.BBL
		config acceptance.Config
		state  acceptance.State
	)

	BeforeEach(func() {
		var err error
		config, err = acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		azure = actors.NewAzure(config)
		bbl = actors.NewBBL(config.StateFileDir, pathToBBL, config, "azure-env")
		state = acceptance.NewState(config.StateFileDir)
	})

	AfterEach(func() {
		session := bbl.Down()
		Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
	})

	// It("creates the resource group", func() {
	// 	session := bbl.Up(config.IAAS, []string{"--name", bbl.PredefinedEnvID(), "--no-director"})
	// 	Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

	// 	exists, err := azure.GetResourceGroup(fmt.Sprintf("%s-bosh", bbl.PredefinedEnvID()))
	// 	Expect(err).NotTo(HaveOccurred())
	// 	Expect(exists).To(BeTrue())
	// })

	FIt("creates the director", func() {
		session := bbl.Up(config.IAAS, []string{"--name", bbl.PredefinedEnvID()})
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

	})
})
