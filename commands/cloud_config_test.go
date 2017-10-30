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

		cloudConfigManager.InterpolateCall.Returns.CloudConfig = "some-cloud-config"
		cloudConfigManager.IsPresentCloudConfigCall.Returns.IsPresent = true
		cloudConfigManager.IsPresentCloudConfigVarsCall.Returns.IsPresent = true

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
		Context("when the state validator fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
			})

			It("returns an error", func() {
				err := cloudConfig.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to validate state"))
			})
		})
	})

	Describe("Execute", func() {
		It("prints the cloud configuration for the bbl environment", func() {
			err := cloudConfig.Execute([]string{}, state)
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudConfigManager.InitializeCall.CallCount).To(Equal(0))
			Expect(cloudConfigManager.InterpolateCall.CallCount).To(Equal(1))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("some-cloud-config"))
		})

		Context("when there are no cloud config files", func() {
			BeforeEach(func() {
				cloudConfigManager.IsPresentCloudConfigCall.Returns.IsPresent = false
			})

			It("initializes the cloud config", func() {
				err := cloudConfig.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.InitializeCall.CallCount).To(Equal(1))
				Expect(cloudConfigManager.InitializeCall.Receives.State).To(Equal(state))
			})
		})

		Context("when there is no cloud config vars file", func() {
			BeforeEach(func() {
				cloudConfigManager.IsPresentCloudConfigVarsCall.Returns.IsPresent = false
			})

			It("generates the cloud config vars", func() {
				err := cloudConfig.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.GenerateVarsCall.CallCount).To(Equal(1))
				Expect(cloudConfigManager.GenerateVarsCall.Receives.State).To(Equal(state))
			})
		})

		Context("failure cases", func() {
			Context("when the cloud config manager fails to initialize", func() {
				BeforeEach(func() {
					cloudConfigManager.IsPresentCloudConfigCall.Returns.IsPresent = false
					cloudConfigManager.InitializeCall.Returns.Error = errors.New("guava")
				})

				It("returns an error", func() {
					err := cloudConfig.Execute([]string{}, state)
					Expect(err).To(MatchError("guava"))
				})
			})

			Context("when the cloud config manager fails to generate vars", func() {
				BeforeEach(func() {
					cloudConfigManager.IsPresentCloudConfigVarsCall.Returns.IsPresent = false
					cloudConfigManager.GenerateVarsCall.Returns.Error = errors.New("papaya")
				})

				It("returns an error", func() {
					err := cloudConfig.Execute([]string{}, state)
					Expect(err).To(MatchError("papaya"))
				})
			})

			Context("when the cloud config manager fails to generate", func() {
				BeforeEach(func() {
					cloudConfigManager.InterpolateCall.Returns.Error = errors.New("mango")
				})

				It("returns an error", func() {
					err := cloudConfig.Execute([]string{}, state)
					Expect(err).To(MatchError("mango"))
				})
			})
		})
	})
})
