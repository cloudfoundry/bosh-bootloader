package commands_test

import (
	"errors"
	"io/ioutil"
	"strings"

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
		stateValidator   *fakes.StateValidator
		terraformManager *fakes.TerraformManager
		sshKeyGetter     *fakes.SSHKeyGetter
		printEnv         commands.PrintEnv
		state            storage.State
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		stateValidator = &fakes.StateValidator{}
		terraformManager = &fakes.TerraformManager{}
		sshKeyGetter = &fakes.SSHKeyGetter{}
		sshKeyGetter.GetCall.Returns.PrivateKey = "some-private-key"

		state = storage.State{
			BOSH: storage.BOSH{
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				DirectorAddress:  "some-director-address",
				DirectorSSLCA:    "some-director-ca-cert",
			},
			Jumpbox: storage.Jumpbox{
				URL: "some-magical-jumpbox-url",
			},
		}

		printEnv = commands.NewPrintEnv(logger, stateValidator, sshKeyGetter, terraformManager)
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

			Expect(sshKeyGetter.GetCall.Receives.Deployment).To(Equal("jumpbox"))

			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CLIENT=some-director-username"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CLIENT_SECRET=some-director-password"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_CA_CERT='some-director-ca-cert'"))
			Expect(logger.PrintlnCall.Messages).To(ContainElement("export BOSH_ENVIRONMENT=some-director-address"))

			Expect(logger.PrintlnCall.Messages).To(ContainElement(MatchRegexp(`export BOSH_ALL_PROXY=socks5://localhost:\d+`)))
			Expect(logger.PrintlnCall.Messages).To(ContainElement(MatchRegexp(`JUMPBOX_PRIVATE_KEY=.*[/\\]bosh_jumpbox_private.key`)))
			Expect(logger.PrintlnCall.Messages).To(ContainElement(MatchRegexp(`ssh -f -N -o StrictHostKeyChecking=no -o ServerAliveInterval=300 -D \d+ jumpbox@some-magical-jumpbox-url -i \$JUMPBOX_PRIVATE_KEY`)))
		})

		It("writes private key to file in temp dir", func() {
			err := printEnv.Execute([]string{}, state)
			Expect(err).NotTo(HaveOccurred())

			for _, line := range logger.PrintlnCall.Messages {
				if strings.HasPrefix(line, "export JUMPBOX_PRIVATE_KEY=") {
					privateKeyFilename := strings.TrimPrefix(line, "export JUMPBOX_PRIVATE_KEY=")

					privateKey, err := ioutil.ReadFile(privateKeyFilename)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(privateKey)).To(Equal("some-private-key"))
				}
			}
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

			Context("when ssh key getter fails", func() {
				BeforeEach(func() {
					sshKeyGetter.GetCall.Returns.Error = errors.New("papaya")
				})

				It("returns an error", func() {
					err := printEnv.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("papaya"))
				})
			})
		})
	})
})
