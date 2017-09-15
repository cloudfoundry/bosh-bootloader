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
		logger           *fakes.Logger
		boshManager      *fakes.BOSHManager
		stateValidator   *fakes.StateValidator
		terraformManager *fakes.TerraformManager

		boshDeploymentVars commands.BOSHDeploymentVars
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		boshManager = &fakes.BOSHManager{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}

		boshManager.VersionCall.Returns.Version = "2.0.24"

		terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{"some-name": "some-output"}

		boshDeploymentVars = commands.NewBOSHDeploymentVars(logger, boshManager, stateValidator, terraformManager)
	})

	Describe("CheckFastFails", func() {
		It("returns an error when the state validator fails", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
			err := boshDeploymentVars.CheckFastFails([]string{}, storage.State{})
			Expect(err).To(MatchError("failed to validate state"))
		})

		Context("when the bosh installed has a version less than v2.0.24", func() {
			BeforeEach(func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
			})

			It("returns an error", func() {
				err := boshDeploymentVars.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})

			Context("when the state has no director", func() {
				It("returns no error", func() {
					err := boshDeploymentVars.CheckFastFails([]string{}, storage.State{
						NoDirector: true,
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("Execute", func() {
		It("calls out to bosh manager and prints the resulting information", func() {
			boshManager.GetDirectorDeploymentVarsCall.Returns.Vars = "some-vars-yaml"

			err := boshDeploymentVars.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshManager.GetDirectorDeploymentVarsCall.CallCount).To(Equal(1))
			Expect(boshManager.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs).To(HaveKeyWithValue("some-name", "some-output"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("some-vars-yaml"))
		})

		Context("failure cases", func() {
			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("coconut")
			})
			It("returns an error when we fail to get deployment vars", func() {
				err := boshDeploymentVars.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("get terraform outputs: coconut"))
			})
		})
	})
})
