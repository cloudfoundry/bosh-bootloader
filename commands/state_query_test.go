package commands_test

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateQuery", func() {
	var (
		fakeLogger           *fakes.Logger
		fakeStateValidator   *fakes.StateValidator
		fakeTerraformManager *fakes.TerraformManager
	)

	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
		fakeStateValidator = &fakes.StateValidator{}
		fakeTerraformManager = &fakes.TerraformManager{}
	})

	Describe("CheckFastFails", func() {
		Context("when the state validator fails", func() {
			BeforeEach(func() {
				fakeStateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			})

			It("returns an error", func() {
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformManager, "")

				err := command.CheckFastFails([]string{}, storage.State{})

				Expect(err).To(MatchError("state validator failed"))
			})
		})

		Context("bbl does not manage the bosh director", func() {
			var state storage.State

			BeforeEach(func() {
				state = storage.State{
					EnvID:      "some-env-id",
					NoDirector: true,
				}
			})

			DescribeTable("prints out the director information",
				func(propertyName string) {
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformManager, propertyName)

					err := command.CheckFastFails([]string{}, state)
					Expect(err).To(MatchError("Error BBL does not manage this director."))
				},
				Entry("director-username", "director username"),
				Entry("director-password", "director password"),
				Entry("director-ssl-ca", "director ca cert"),
			)
		})
	})

	Describe("Execute", func() {
		Context("bbl manages the jumpbox", func() {
			var state storage.State

			BeforeEach(func() {
				state = storage.State{
					Jumpbox: storage.Jumpbox{
						URL: "some-jumpbox-url",
					},
				}
			})

			It("prints out the jumpbox information", func() {
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformManager, "jumpbox address")

				err := command.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("some-jumpbox-url"))
			})
		})

		Context("bbl manages the bosh director", func() {
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
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformManager, propertyName)

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
					EnvID:      "some-env-id",
					NoDirector: true,
				}
			})

			It("prints the env id", func() {
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformManager, "environment id")

				err := command.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("some-env-id"))
			})

			It("prints the eip as the director-address", func() {
				fakeTerraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"external_ip": "some-external-ip",
				}

				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformManager, "director address")
				err := command.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("https://some-external-ip:25555"))
			})
		})

		Context("failure cases", func() {
			Context("when the terraform output provider fails", func() {
				BeforeEach(func() {
					fakeTerraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get terraform output")
				})

				It("returns an error", func() {
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformManager, "director address")

					err := command.Execute([]string{}, storage.State{
						IAAS:       "gcp",
						NoDirector: true,
					})

					Expect(err).To(MatchError("failed to get terraform output"))
				})
			})

			Context("when the state value is empty", func() {
				It("returns an error", func() {
					propertyName := fmt.Sprintf("%s-%d", "some-name", rand.Int())
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, fakeTerraformManager, propertyName)
					err := command.Execute([]string{}, storage.State{
						BOSH: storage.BOSH{},
					})
					Expect(err).To(MatchError(fmt.Sprintf("Could not retrieve %s, please make sure you are targeting the proper state dir.", propertyName)))

					Expect(fakeLogger.PrintlnCall.CallCount).To(Equal(0))
				})
			})
		})
	})
})
