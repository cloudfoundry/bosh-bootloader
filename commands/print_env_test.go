package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrintEnv", func() {
	var (
		logger           *fakes.Logger
		stateValidator   *fakes.StateValidator
		terraformManager *fakes.TerraformManager
		printEnv         commands.PrintEnv
		state            storage.State
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}

		state = storage.State{
			BOSH: storage.BOSH{
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				DirectorAddress:  "some-director-address",
				DirectorSSLCA:    "some-director-ca-cert",
			},
		}

		printEnv = commands.NewPrintEnv(logger, stateValidator, terraformManager)
	})

	Describe("CheckFastFails", func() {
		It("returns an error when the state does not exist", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
			err := printEnv.CheckFastFails([]string{}, storage.State{})
			Expect(err).To(MatchError("failed to validate state"))
		})
	})

	Describe("Execute", func() {
		It("prints the correct environment variables for the bosh cli", func() {
			err := printEnv.Execute([]string{}, state)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CLIENT=some-director-username"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CLIENT_SECRET=some-director-password"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CA_CERT='some-director-ca-cert'"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=some-director-address"))
		})

		Context("when print-env is called on a bbl env with no director", func() {
			It("prints only the BOSH_ENVIRONMENT", func() {
				terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"external_ip": "some-external-ip",
				}

				err := printEnv.Execute([]string{}, storage.State{
					NoDirector: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=https://some-external-ip:25555"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CLIENT=some-director-username"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CLIENT_SECRET=some-director-password"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CA_CERT='some-director-ca-cert'"))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the terraform outputter fails", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get terraform output")
				err := printEnv.Execute([]string{}, storage.State{
					NoDirector: true,
				})
				Expect(err).To(MatchError("failed to get terraform output"))
			})
		})
	})
})
