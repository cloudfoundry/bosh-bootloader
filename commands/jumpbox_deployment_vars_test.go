package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JumpboxDeploymentVars", func() {

	var (
		logger           *fakes.Logger
		boshManager      *fakes.BOSHManager
		stateValidator   *fakes.StateValidator
		terraformManager *fakes.TerraformManager

		jumpboxDeploymentVars commands.JumpboxDeploymentVars

		terraformOutputs terraform.Outputs
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		boshManager = &fakes.BOSHManager{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}

		boshManager.VersionCall.Returns.Version = "2.0.24"

		terraformOutputs = terraform.Outputs{
			Map: map[string]interface{}{"some-name": "some-output"},
		}
		terraformManager.GetOutputsCall.Returns.Outputs = terraformOutputs

		jumpboxDeploymentVars = commands.NewJumpboxDeploymentVars(logger, boshManager, stateValidator, terraformManager)
	})

	Describe("CheckFastFails", func() {
		Context("when the state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
			})

			It("returns an error", func() {
				err := jumpboxDeploymentVars.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to validate state"))
			})
		})

		Context("when the bosh installed has a version less than v2.0.24", func() {
			BeforeEach(func() {
				boshManager.VersionCall.Returns.Version = "1.9.0"
			})

			It("returns an error", func() {
				err := jumpboxDeploymentVars.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})

			Context("when the state has no director", func() {
				It("returns no error", func() {
					err := jumpboxDeploymentVars.CheckFastFails([]string{}, storage.State{
						NoDirector: true,
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("Execute", func() {
		It("calls out to bosh manager and prints the resulting information", func() {
			boshManager.GetJumpboxDeploymentVarsCall.Returns.Vars = "some-vars-yaml"

			err := jumpboxDeploymentVars.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshManager.GetJumpboxDeploymentVarsCall.CallCount).To(Equal(1))
			Expect(boshManager.GetJumpboxDeploymentVarsCall.Receives.TerraformOutputs).To(Equal(terraformOutputs))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("some-vars-yaml"))
		})

		Context("failure cases", func() {
			Context("when we fail to get deployment vars", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("coconut")
				})

				It("returns an error", func() {
					err := jumpboxDeploymentVars.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("get terraform outputs: coconut"))
				})
			})
		})
	})
})
