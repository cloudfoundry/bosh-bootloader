package commands_test

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateQuery", func() {
	var (
		fakeLogger                  *fakes.Logger
		fakeStateValidator          *fakes.StateValidator
		fakeTerraformOutputProvider *fakes.TerraformOutputProvider
		fakeInfrastructureManager   *fakes.InfrastructureManager
	)

	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
		fakeStateValidator = &fakes.StateValidator{}
		fakeTerraformOutputProvider = &fakes.TerraformOutputProvider{}
		fakeInfrastructureManager = &fakes.InfrastructureManager{}
	})

	Describe("Execute", func() {
		Context("bbl does manage the bosh director", func() {
			var state storage.State

			BeforeEach(func() {
				state = storage.State{
					BOSH: storage.BOSH{
						DirectorAddress:  "some-director-address",
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
						DirectorSSLCA:    "some-director-ssl-ca",
					},
				}
			})

			DescribeTable("prints out the director information",
				func(propertyName, expectedOutput string) {
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, propertyName)

					err := command.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal(expectedOutput))
				},
				Entry("director-address", "director address", "some-director-address"),
				Entry("director-username", "director username", "some-director-username"),
				Entry("director-password", "director password", "some-director-password"),
				Entry("director-ssl-ca", "director ca cert", "some-director-ssl-ca"),
			)
		})

		Context("bbl does not manage the bosh director", func() {
			var state storage.State

			BeforeEach(func() {
				state = storage.State{
					NoDirector: true,
				}
			})

			DescribeTable("prints out the director information",
				func(propertyName string) {
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, propertyName)

					err := command.Execute([]string{}, state)
					Expect(err).To(MatchError("Error BBL does not manage this director."))
				},
				Entry("director-username", "director username"),
				Entry("director-password", "director password"),
				Entry("director-ssl-ca", "director ca cert"),
			)

			Context("gcp", func() {
				It("prints the eip as the director-address", func() {
					fakeTerraformOutputProvider.GetCall.Returns.Outputs = terraform.Outputs{
						ExternalIP: "some-external-ip",
					}

					state.IAAS = "gcp"

					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, "director address")
					err := command.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("https://some-external-ip:25555"))
				})
			})

			Context("aws", func() {
				It("prints the eip as the director-address", func() {
					fakeInfrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
						Outputs: map[string]string{
							"BOSHEIP": "some-external-ip",
						},
					}

					state.IAAS = "aws"

					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, "director address")
					err := command.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("https://some-external-ip:25555"))
				})

			})
		})

		Context("failure cases", func() {
			It("returns an error when the state validator fails", func() {
				fakeStateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, "")

				err := command.Execute([]string{}, storage.State{
					BOSH: storage.BOSH{},
				})

				Expect(err).To(MatchError("state validator failed"))
			})

			It("returns an error when the terraform output provider fails", func() {
				fakeTerraformOutputProvider.GetCall.Returns.Error = errors.New("failed to get terraform output")
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, "director address")

				err := command.Execute([]string{}, storage.State{
					IAAS:       "gcp",
					NoDirector: true,
				})

				Expect(err).To(MatchError("failed to get terraform output"))
			})

			It("returns an error when the infrastructure manager fails", func() {
				fakeInfrastructureManager.DescribeCall.Returns.Error = errors.New("failed to describe stack")
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, "director address")

				err := command.Execute([]string{}, storage.State{
					IAAS:       "aws",
					NoDirector: true,
				})

				Expect(err).To(MatchError("failed to describe stack"))
			})

			It("returns an error when an external ip cannot be found", func() {
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, "director address")

				err := command.Execute([]string{}, storage.State{
					IAAS:       "lol",
					NoDirector: true,
				})

				Expect(err).To(MatchError("Could not find external IP for given IAAS"))
			})

			It("returns an error when the state value is empty", func() {
				propertyName := fmt.Sprintf("%s-%d", "some-name", rand.Int())
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformOutputProvider, fakeInfrastructureManager, propertyName)
				err := command.Execute([]string{}, storage.State{
					BOSH: storage.BOSH{},
				})
				Expect(err).To(MatchError(fmt.Sprintf("Could not retrieve %s, please make sure you are targeting the proper state dir.", propertyName)))

				Expect(fakeLogger.PrintlnCall.CallCount).To(Equal(0))
			})
		})
	})
})
