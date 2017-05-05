package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LBs", func() {
	var (
		lbsCommand commands.LBs

		gcpLBs         *fakes.GCPLBs
		awsLBs         *fakes.AWSLBs
		stateValidator *fakes.StateValidator
		logger         *fakes.Logger
	)

	BeforeEach(func() {
		gcpLBs = &fakes.GCPLBs{}
		awsLBs = &fakes.AWSLBs{}

		stateValidator = &fakes.StateValidator{}
		logger = &fakes.Logger{}

		lbsCommand = commands.NewLBs(gcpLBs, awsLBs, stateValidator, logger)
	})

	Describe("Execute", func() {
		Context("when bbl'd up on aws", func() {
			It("prints LB ips", func() {
				incomingState := storage.State{
					IAAS: "aws",
				}
				err := lbsCommand.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(awsLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("when bbl'd up on gcp", func() {
			It("prints LB ips", func() {
				incomingState := storage.State{
					IAAS: "gcp",
				}
				err := lbsCommand.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(gcpLBs.ExecuteCall.Receives.SubcommandFlags).To(Equal([]string{}))
				Expect(gcpLBs.ExecuteCall.Receives.State).To(Equal(incomingState))
			})
		})

		Context("failure cases", func() {
			It("returns an error when state validator fails", func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")

				err := lbsCommand.Execute([]string{}, storage.State{})

				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
				Expect(err).To(MatchError("state validator failed"))
			})

			It("returns an error when the AWSLBs fails", func() {
				awsLBs.ExecuteCall.Returns.Error = errors.New("something bad happened")

				err := lbsCommand.Execute([]string{}, storage.State{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("something bad happened"))
			})

			It("returns an error when the GCPLBs fails", func() {
				gcpLBs.ExecuteCall.Returns.Error = errors.New("something bad happened")

				err := lbsCommand.Execute([]string{}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
