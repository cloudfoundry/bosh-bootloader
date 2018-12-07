package commands_test

import (
	"errors"
	"fmt"
	"math/rand"

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
		fakeLogger         *fakes.Logger
		fakeStateValidator *fakes.StateValidator
		terraformManager   *fakes.TerraformManager
	)

	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
		fakeStateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}
	})

	Describe("CheckFastFails", func() {
		Context("when the state validator fails", func() {
			BeforeEach(func() {
				fakeStateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			})

			It("returns an error", func() {
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, terraformManager, "")

				err := command.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("state validator failed"))
			})
		})
	})

	Describe("Execute", func() {
		Context("bbl manages the jumpbox", func() {
			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{
					Map: map[string]interface{}{"external_ip": "some-external-ip"},
				}
			})

			It("prints out the jumpbox information", func() {
				command := commands.NewStateQuery(fakeLogger, fakeStateValidator, terraformManager, "jumpbox address")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
				Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("some-external-ip"))
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
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, terraformManager, propertyName)

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

		Context("failure cases", func() {
			Context("when the terraform output provider fails", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get terraform output")
				})

				It("jumpbox-address returns an error", func() {
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, terraformManager, "jumpbox address")

					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("failed to get terraform output"))
				})
			})

			Context("when the state value is empty", func() {
				It("returns an error", func() {
					propertyName := fmt.Sprintf("%s-%d", "some-name", rand.Int())
					command := commands.NewStateQuery(fakeLogger, fakeStateValidator, terraformManager, propertyName)
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
