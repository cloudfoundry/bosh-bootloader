package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWS Update LBs", func() {
	var (
		awsCreateLBs  *fakes.AWSCreateLBs
		command       commands.AWSUpdateLBs
		incomingState storage.State
	)

	BeforeEach(func() {
		awsCreateLBs = &fakes.AWSCreateLBs{}

		incomingState = storage.State{
			IAAS:    "aws",
			TFState: "some-tf-state",
			LB: storage.LB{
				Type:   "cf",
				Cert:   "some-cert",
				Key:    "some-key",
				Domain: "some-domain",
			},
		}

		command = commands.NewAWSUpdateLBs(awsCreateLBs)
	})

	Describe("Execute", func() {
		It("calls out to AWS Create LBs", func() {
			config := commands.AWSCreateLBsConfig{
				CertPath: "some-cert-path",
				KeyPath:  "some-key-path",
				LBType:   "cf",
				Domain:   "some-domain",
			}
			err := command.Execute(config, incomingState)

			Expect(err).NotTo(HaveOccurred())
			Expect(awsCreateLBs.ExecuteCall.CallCount).To(Equal(1))
			Expect(awsCreateLBs.ExecuteCall.Receives.Config).To(Equal(commands.AWSCreateLBsConfig{
				CertPath: "some-cert-path",
				KeyPath:  "some-key-path",
				LBType:   "cf",
				Domain:   "some-domain",
			}))
			Expect(awsCreateLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
		})

		Context("when config does not contain system domain", func() {
			It("passes system domain from the state", func() {
				config := commands.AWSCreateLBsConfig{
					CertPath: "some-cert-path",
					KeyPath:  "some-key-path",
					LBType:   "cf",
					Domain:   "",
				}
				err := command.Execute(config, incomingState)

				Expect(err).NotTo(HaveOccurred())
				Expect(awsCreateLBs.ExecuteCall.CallCount).To(Equal(1))
				Expect(awsCreateLBs.ExecuteCall.Receives.Config).To(Equal(commands.AWSCreateLBsConfig{
					CertPath: "some-cert-path",
					KeyPath:  "some-key-path",
					LBType:   "cf",
					Domain:   "some-domain",
				}))
				Expect(awsCreateLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when config does not contain lb type", func() {
			It("passes lb type from the state", func() {
				config := commands.AWSCreateLBsConfig{
					CertPath: "some-cert-path",
					KeyPath:  "some-key-path",
					LBType:   "",
					Domain:   "some-domain",
				}
				err := command.Execute(config, incomingState)

				Expect(err).NotTo(HaveOccurred())
				Expect(awsCreateLBs.ExecuteCall.CallCount).To(Equal(1))
				Expect(awsCreateLBs.ExecuteCall.Receives.Config).To(Equal(commands.AWSCreateLBsConfig{
					CertPath: "some-cert-path",
					KeyPath:  "some-key-path",
					LBType:   "cf",
					Domain:   "some-domain",
				}))
				Expect(awsCreateLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when create lbs fails", func() {
			BeforeEach(func() {
				awsCreateLBs.ExecuteCall.Returns.Error = errors.New("fig")
			})
			It("returns an error", func() {
				err := command.Execute(commands.AWSCreateLBsConfig{}, incomingState)
				Expect(err).To(MatchError("fig"))
			})
		})
	})
})
