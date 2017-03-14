package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrintEnv", func() {
	var (
		logger                  *fakes.Logger
		stateValidator          *fakes.StateValidator
		terraformOutputProvider *fakes.TerraformOutputProvider
		infrastructureManager   *fakes.InfrastructureManager
		printEnv                commands.PrintEnv
		state                   storage.State
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}
		terraformOutputProvider = &fakes.TerraformOutputProvider{}
		infrastructureManager = &fakes.InfrastructureManager{}

		state = storage.State{
			BOSH: storage.BOSH{
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				DirectorAddress:  "some-director-address",
				DirectorSSLCA:    "some-director-ca-cert",
			},
		}

		printEnv = commands.NewPrintEnv(logger, stateValidator, terraformOutputProvider, infrastructureManager)
	})

	It("prints the correct environment variables for the bosh cli", func() {
		err := printEnv.Execute([]string{}, state)
		Expect(err).NotTo(HaveOccurred())
		Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CLIENT=some-director-username"))
		Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CLIENT_SECRET=some-director-password"))
		Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CA_CERT='some-director-ca-cert'"))
		Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=some-director-address"))
	})

	Context("when print-env is called on a bbl env with no director", func() {
		Context("aws", func() {
			It("prints only the BOSH_ENVIRONMENT", func() {
				infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
					Outputs: map[string]string{
						"BOSHEIP": "some-external-ip",
					},
				}

				err := printEnv.Execute([]string{}, storage.State{
					IAAS:       "aws",
					NoDirector: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=https://some-external-ip:25555"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CLIENT=some-director-username"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CLIENT_SECRET=some-director-password"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CA_CERT='some-director-ca-cert'"))
			})
		})
		Context("gcp", func() {
			It("prints only the BOSH_ENVIRONMENT", func() {
				terraformOutputProvider.GetCall.Returns.Outputs = terraform.Outputs{
					ExternalIP: "some-external-ip",
				}

				err := printEnv.Execute([]string{}, storage.State{
					IAAS:       "gcp",
					NoDirector: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=https://some-external-ip:25555"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CLIENT=some-director-username"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CLIENT_SECRET=some-director-password"))
				Expect(logger.PrintlnCall.Messages).NotTo(ContainElement("export BOSH_CA_CERT='some-director-ca-cert'"))
			})
		})
	})

	Context("failure cases", func() {
		It("returns an error when the state does not exist", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("failed to validate state")
			err := printEnv.Execute([]string{}, storage.State{})
			Expect(err).To(MatchError("failed to validate state"))
		})

		It("returns an error when the terraform outputter fails", func() {
			terraformOutputProvider.GetCall.Returns.Error = errors.New("failed to get terraform output")
			err := printEnv.Execute([]string{}, storage.State{
				IAAS:       "gcp",
				NoDirector: true,
			})
			Expect(err).To(MatchError("failed to get terraform output"))
		})

		It("returns an error when the infrastructure manager fails to describe stack", func() {
			infrastructureManager.DescribeCall.Returns.Error = errors.New("failed to describe stack")
			err := printEnv.Execute([]string{}, storage.State{
				IAAS:       "aws",
				NoDirector: true,
			})
			Expect(err).To(MatchError("failed to describe stack"))
		})

		It("returns an error when the external ip cannot be found for a given IAAS", func() {
			err := printEnv.Execute([]string{}, storage.State{
				IAAS:       "lol",
				NoDirector: true,
			})
			Expect(err).To(MatchError("Could not find external IP for given IAAS"))
		})
	})
})
