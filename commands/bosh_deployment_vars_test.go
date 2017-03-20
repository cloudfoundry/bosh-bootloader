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
		boshManager.VersionCall.Returns.Version = "2.0.0"

		boshDeploymentVars = commands.NewBOSHDeploymentVars(logger, boshManager)
	})

	It("calls out to bosh manager and prints the resulting information", func() {
		boshManager.GetDeploymentVarsCall.Returns.Vars = "some-vars-yaml"
		err := boshDeploymentVars.Execute([]string{}, storage.State{})
		Expect(err).NotTo(HaveOccurred())
		Expect(boshManager.GetDeploymentVarsCall.CallCount).To(Equal(1))
		Expect(logger.PrintlnCall.Messages).To(ContainElement("some-vars-yaml"))
	})

	It("runs successfully if the version is less than 2.0.0 but the state has no director", func() {
		boshManager.VersionCall.Returns.Version = "1.9.9"

		err := boshDeploymentVars.Execute([]string{}, storage.State{
			NoDirector: true,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("failure cases", func() {
		It("returns an error when we fail to get deployment vars", func() {
			boshManager.GetDeploymentVarsCall.Returns.Error = errors.New("failed to get deployment vars")
			err := boshDeploymentVars.Execute([]string{}, storage.State{})
			Expect(err).To(MatchError("failed to get deployment vars"))
		})

		It("fast fails if the bosh installed is less than v2.0.0", func() {
			boshManager.VersionCall.Returns.Version = "1.9.9"

			err := boshDeploymentVars.Execute([]string{}, storage.State{})
			Expect(err).To(MatchError("BOSH version must be at least v2.0.0"))
		})
	})
})
