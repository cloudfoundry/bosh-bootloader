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

var _ = Describe("PrintEnv", func() {
	var (
		logger           *fakes.Logger
		stderrLogger     *fakes.Logger
		stateValidator   *fakes.StateValidator
		terraformManager *fakes.TerraformManager
		allProxyGetter   *fakes.AllProxyGetter
		credhubGetter    *fakes.CredhubGetter
		fileIO           *fakes.FileIO
		printEnv         commands.PrintEnv
		state            storage.State
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		stderrLogger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}
		allProxyGetter = &fakes.AllProxyGetter{}
		allProxyGetter.GeneratePrivateKeyCall.Returns.PrivateKey = "the-key-path"
		allProxyGetter.BoshAllProxyCall.Returns.URL = "ipfs://some-domain-with?private_key=the-key-path"
		credhubGetter = &fakes.CredhubGetter{}
		credhubGetter.GetServerCall.Returns.Server = "some-credhub-server"
		credhubGetter.GetCertsCall.Returns.Certs = "some-credhub-certs"
		credhubGetter.GetPasswordCall.Returns.Password = "some-credhub-password"

		fileIO = &fakes.FileIO{}

		state = storage.State{
			BOSH: storage.BOSH{
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				DirectorAddress:  "some-director-address",
				DirectorSSLCA:    "some-director-ca-cert",
			},
			Jumpbox: storage.Jumpbox{
				URL: "some-magical-jumpbox-url:22",
			},
		}

		printEnv = commands.NewPrintEnv(logger, stderrLogger, stateValidator, allProxyGetter, credhubGetter, terraformManager, fileIO)
	})
	Describe("CheckFastFails", func() {
		Context("when the state does not exist", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
			})

			It("returns an error", func() {
				err := printEnv.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to validate state"))
			})
		})
	})

	Describe("Execute", func() {
		It("prints the correct environment variables for the bosh cli", func() {
			err := printEnv.Execute([]string{}, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(allProxyGetter.GeneratePrivateKeyCall.CallCount).To(Equal(1))
			Expect(allProxyGetter.BoshAllProxyCall.Receives.JumpboxURL).To(Equal("some-magical-jumpbox-url:22"))
			Expect(allProxyGetter.BoshAllProxyCall.Receives.PrivateKey).To(Equal("the-key-path"))

			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CLIENT=some-director-username"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CLIENT_SECRET=some-director-password"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CA_CERT='some-director-ca-cert'"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=some-director-address"))

			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_SERVER=some-credhub-server"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_CA_CERT='some-credhub-certs'"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_USER=credhub-cli"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_PASSWORD=some-credhub-password"))

			Expect(logger.PrintlnCall.Messages).To(ContainElement(`export JUMPBOX_PRIVATE_KEY=the-key-path`))
			Expect(logger.PrintlnCall.Messages).To(ContainElement(`export BOSH_ALL_PROXY=ipfs://some-domain-with?private_key=the-key-path`))
		})

		Context("when there is no director", func() {
			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{
					Map: map[string]interface{}{"external_ip": "some-external-ip"},
				}
			})

			It("prints only the BOSH_ENVIRONMENT", func() {
				err := printEnv.Execute([]string{}, storage.State{
					NoDirector: true,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))

				Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=https://some-external-ip:25555"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CLIENT=some-director-username"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CLIENT_SECRET=some-director-password"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CA_CERT='some-director-ca-cert'"))
			})
		})

		Context("failure cases", func() {
			Context("when terraform manager get outputs fails", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get terraform output")
				})

				It("returns an error", func() {
					err := printEnv.Execute([]string{}, storage.State{
						NoDirector: true,
					})
					Expect(err).To(MatchError("failed to get terraform output"))
				})
			})

			Context("when the allproxy getter fails to get a private key", func() {
				BeforeEach(func() {
					allProxyGetter.GeneratePrivateKeyCall.Returns.Error = errors.New("papaya")
				})

				It("returns an error", func() {
					err := printEnv.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("papaya"))
				})
			})

			Context("when credhub getter fails to get the password", func() {
				BeforeEach(func() {
					credhubGetter.GetPasswordCall.Returns.Error = errors.New("fig")
				})

				It("logs a warning and prints the other information", func() {
					err := printEnv.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())
					Expect(stderrLogger.PrintlnCall.Messages).To(ContainElement("No credhub password found."))
					Expect(logger.PrintlnCall.Messages).To(ContainElement(MatchRegexp(`export JUMPBOX_PRIVATE_KEY=`)))
				})
			})

			Context("when credhub getter fails to get the server", func() {
				BeforeEach(func() {
					credhubGetter.GetServerCall.Returns.Error = errors.New("starfruit")
				})

				It("logs a warning and prints the other information", func() {
					err := printEnv.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())
					Expect(stderrLogger.PrintlnCall.Messages).To(ContainElement("No credhub server found."))
					Expect(logger.PrintlnCall.Messages).To(ContainElement(MatchRegexp(`export JUMPBOX_PRIVATE_KEY=`)))
				})
			})

			Context("when credhub getter fails to get the certs", func() {
				BeforeEach(func() {
					credhubGetter.GetCertsCall.Returns.Error = errors.New("kiwi")
				})

				It("logs a warning and prints the other information", func() {
					err := printEnv.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())
					Expect(stderrLogger.PrintlnCall.Messages).To(ContainElement("No credhub certs found."))
					Expect(logger.PrintlnCall.Messages).To(ContainElement(MatchRegexp(`export JUMPBOX_PRIVATE_KEY=`)))
				})
			})
		})
	})
})
