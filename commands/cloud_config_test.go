package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudConfig", func() {
	var (
		logger             *fakes.Logger
		stateValidator     *fakes.StateValidator
		cloudConfig        commands.CloudConfig
		state              storage.State
		cloudConfigManager *fakes.CloudConfigManager
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}
		cloudConfigManager = &fakes.CloudConfigManager{}

		cloudConfigManager.GenerateCall.Returns.CloudConfig = "some-cloud-config"

		state = storage.State{
			BOSH: storage.BOSH{
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				DirectorAddress:  "some-director-address",
				DirectorSSLCA:    "some-director-ca-cert",
			},
		}

		cloudConfig = commands.NewCloudConfig(logger, stateValidator, cloudConfigManager)
	})

	Describe("CheckFastFails", func() {
		It("returns no error", func() {
			err := cloudConfig.CheckFastFails([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Execute", func() {
		It("prints the cloud configuration for the bbl environment", func() {
			err := cloudConfig.Execute([]string{}, state)
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudConfigManager.GenerateCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.GenerateCall.Receives.State).To(Equal(state))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("some-cloud-config"))
		})

		Context("failure cases", func() {
			It("returns an error when the cloud config manager fails to generate", func() {
				cloudConfigManager.GenerateCall.Returns.Error = errors.New("failed to generate cloud configuration")
				err := cloudConfig.Execute([]string{}, state)
				Expect(err).To(MatchError("failed to generate cloud configuration"))
			})

			It("returns an error when the state validator fails", func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
				err := cloudConfig.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to validate state"))
			})
		})
	})
})
