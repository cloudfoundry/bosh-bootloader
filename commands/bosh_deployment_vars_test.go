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

var _ = Describe("BOSHDeploymentVars", func() {
	var (
		boshManager      *fakes.BOSHManager
		logger           *fakes.Logger
		stateValidator   *fakes.StateValidator
		terraformManager *fakes.TerraformManager

		boshDeploymentVars commands.BOSHDeploymentVars

		terraformOutputs terraform.Outputs
	)

	BeforeEach(func() {
		boshManager = &fakes.BOSHManager{}
		logger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}

		boshManager.VersionCall.Returns.Version = "2.0.24"
		boshManager.GetDirectorDeploymentVarsCall.Returns.Vars = "some-vars-yaml"
		terraformOutputs = terraform.Outputs{
			Map: map[string]interface{}{"some-name": "some-output"},
		}
		terraformManager.GetOutputsCall.Returns.Outputs = terraformOutputs

		boshDeploymentVars = commands.NewBOSHDeploymentVars(logger, boshManager, stateValidator, terraformManager)
	})

	Describe("CheckFastFails", func() {
		Context("when the state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
			})

			It("returns an error", func() {
				err := boshDeploymentVars.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to validate state"))
			})
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
			err := boshDeploymentVars.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(boshManager.GetDirectorDeploymentVarsCall.CallCount).To(Equal(1))
			Expect(boshManager.GetDirectorDeploymentVarsCall.Receives.TerraformOutputs).To(Equal(terraformOutputs))

			Expect(logger.PrintlnCall.Messages).To(ContainElement(`Deprecation warning: the bosh-deployment-vars command has been deprecated and will be removed in bbl v6.0.0. The bosh deployment vars are stored in the vars directory.`))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("some-vars-yaml"))
		})

		Context("failure cases", func() {
			Context("when we fail to get deployment vars", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("coconut")
				})

				It("returns an error", func() {
					err := boshDeploymentVars.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("get terraform outputs: coconut"))
				})
			})
		})
	})
})
