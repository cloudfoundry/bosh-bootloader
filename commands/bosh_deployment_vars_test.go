package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BOSHDeploymentVars", func() {

	var (
		logger      *fakes.Logger
		boshManager *fakes.BOSHManager

		boshDeploymentVars commands.BOSHDeploymentVars
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		boshManager = &fakes.BOSHManager{}

		boshDeploymentVars = commands.NewBOSHDeploymentVars(logger, boshManager)
	})

	It("calls out to bosh manager and prints the resulting information", func() {
		boshManager.GetDeploymentVarsCall.Returns.Vars = "some-vars-yaml"
		err := boshDeploymentVars.Execute([]string{}, storage.State{})
		Expect(err).NotTo(HaveOccurred())
		Expect(boshManager.GetDeploymentVarsCall.CallCount).To(Equal(1))
		Expect(logger.PrintlnCall.Messages).To(ContainElement("some-vars-yaml"))
	})

	Context("failure cases", func() {
		It("returns an error when we fail to get deployment vars", func() {
			boshManager.GetDeploymentVarsCall.Returns.Error = errors.New("failed to get deployment vars")
			err := boshDeploymentVars.Execute([]string{}, storage.State{})
			Expect(err).To(MatchError("failed to get deployment vars"))
		})
	})
})
