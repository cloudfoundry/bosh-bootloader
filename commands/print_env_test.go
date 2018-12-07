package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/renderers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
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
		rendererFactory  renderers.Factory
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
		credhubGetter.GetCertsCall.Returns.Certs = "-----BEGIN CERTIFICATE-----\nsome-credhub-certs\n-----END CERTIFICATE-----\n"
		credhubGetter.GetPasswordCall.Returns.Password = "some-credhub-password"

		fileIO = &fakes.FileIO{}

		state = storage.State{
			BOSH: storage.BOSH{
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				DirectorAddress:  "some-director-address",
				DirectorSSLCA:    "-----BEGIN CERTIFICATE-----\nsome-director-ca-cert\n-----END CERTIFICATE-----\n",
			},
			Jumpbox: storage.Jumpbox{
				URL: "some-magical-jumpbox-url:22",
			},
		}
		rendererFactory = renderers.NewFactory(&fakes.EnvGetter{Values: make(map[string]string)})
		printEnv = commands.NewPrintEnv(logger, stderrLogger, stateValidator, allProxyGetter, credhubGetter, terraformManager, fileIO, rendererFactory)
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
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CA_CERT='-----BEGIN CERTIFICATE-----\nsome-director-ca-cert\n-----END CERTIFICATE-----\n'"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=some-director-address"))

			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_SERVER=some-credhub-server"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_CA_CERT='-----BEGIN CERTIFICATE-----\nsome-credhub-certs\n-----END CERTIFICATE-----\n'"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_CLIENT=credhub-admin"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_SECRET=some-credhub-password"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export CREDHUB_PROXY=ipfs://some-domain-with?private_key=the-key-path"))

			Expect(logger.PrintlnCall.Messages).To(ContainElement(`export JUMPBOX_PRIVATE_KEY=the-key-path`))
			Expect(logger.PrintlnCall.Messages).To(ContainElement(`export BOSH_ALL_PROXY=ipfs://some-domain-with?private_key=the-key-path`))
		})

		Context("WhenPSModulePathIsSet", func() {
			It("prints powershell environment variables", func() {

				values := map[string]string{"PSModulePath": "something"}
				rendererFactory = renderers.NewFactory(&fakes.EnvGetter{Values: values})
				printEnv = commands.NewPrintEnv(logger, stderrLogger, stateValidator, allProxyGetter, credhubGetter, terraformManager, fileIO, rendererFactory)

				err := printEnv.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(allProxyGetter.GeneratePrivateKeyCall.CallCount).To(Equal(1))
				Expect(allProxyGetter.BoshAllProxyCall.Receives.JumpboxURL).To(Equal("some-magical-jumpbox-url:22"))
				Expect(allProxyGetter.BoshAllProxyCall.Receives.PrivateKey).To(Equal("the-key-path"))

				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:BOSH_CLIENT=\"some-director-username\""))
				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:BOSH_CLIENT_SECRET=\"some-director-password\""))
				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:BOSH_CA_CERT='\r\n-----BEGIN CERTIFICATE-----\nsome-director-ca-cert\n-----END CERTIFICATE-----\n'"))
				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:BOSH_ENVIRONMENT=\"some-director-address\""))

				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:CREDHUB_SERVER=\"some-credhub-server\""))
				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:CREDHUB_CA_CERT='\r\n-----BEGIN CERTIFICATE-----\nsome-credhub-certs\n-----END CERTIFICATE-----\n'"))
				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:CREDHUB_CLIENT=\"credhub-admin\""))
				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:CREDHUB_SECRET=\"some-credhub-password\""))
				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:CREDHUB_PROXY=\"ipfs://some-domain-with?private_key=the-key-path\""))

				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:JUMPBOX_PRIVATE_KEY=\"the-key-path\""))
				Expect(logger.PrintlnCall.Messages).To(ContainElement("$env:BOSH_ALL_PROXY=\"ipfs://some-domain-with?private_key=the-key-path\""))
			})
		})

		Context("failure cases", func() {
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
