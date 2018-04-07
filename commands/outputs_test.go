package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Outputs", func() {
	var (
		logger         *fakes.Logger
		stateStore     *fakes.StateStore
		carto          *fakes.Cartographer
		stateValidator *fakes.StateValidator
		outputsCommand commands.Outputs
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		carto = &fakes.Cartographer{}
		stateStore = &fakes.StateStore{}
		stateValidator = &fakes.StateValidator{}
		outputsCommand = commands.NewOutputs(logger, carto, stateStore, stateValidator)
	})

	Describe("CheckFastFails", func() {
		Context("when state validation fails", func() {
			BeforeEach(func() {
				stateValidator.ValidateCall.Returns.Error = errors.New("state validation failed")
			})

			It("returns an error", func() {
				err := outputsCommand.CheckFastFails([]string{}, storage.State{})
				Expect(err).To(MatchError("state validation failed"))
				Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
			})
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			carto.YmlizeCall.Returns.Yml = `external: address
firewall: |-
  cidr
  make sure we quote multiline strings`
		})

		It("prints the terraform outputs", func() {
			err := outputsCommand.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.PrintfCall.Receives.Message).To(ContainSubstring("external: address\nfirewall: |-\n  cidr\n  make sure we quote multiline strings"))
		})

		Context("when cartographer fails to ymlize", func() {
			BeforeEach(func() {
				carto.YmlizeCall.Returns.Error = errors.New("tangelo")
			})

			It("returns an error", func() {
				err := outputsCommand.Execute([]string{}, storage.State{})

				Expect(err).To(MatchError("tangelo"))
			})
		})
	})
})
