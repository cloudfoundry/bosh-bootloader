package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Up", func() {
	var (
		command commands.Up

		fakeAWSUp *fakes.AWSUp
		fakeGCPUp *fakes.GCPUp
	)

	BeforeEach(func() {
		fakeAWSUp = &fakes.AWSUp{Name: "aws"}
		fakeGCPUp = &fakes.GCPUp{Name: "gcp"}

		command = commands.NewUp(fakeAWSUp, fakeGCPUp)
	})

	Describe("Execute", func() {
		Context("when state does not contain an iaas", func() {
			Context("when desired iaas is gcp", func() {
				It("executes the GCP up", func() {
					err := command.Execute([]string{"--iaas", "gcp"}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{}))
				})
			})

			Context("when desired iaas is aws", func() {
				It("executes the AWS up", func() {
					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeAWSUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeAWSUp.ExecuteCall.Receives.Args).To(Equal([]string{"--iaas", "aws"}))
					Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(storage.State{}))
				})
			})

			Context("when iaas is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("--iaas [gcp, aws] must be provided"))
				})
			})

			Context("when an invalid iaas is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--iaas", "bad-iaas"}, storage.State{})
					Expect(err).To(MatchError(`"bad-iaas" is invalid; supported values: [gcp, aws]`))
				})
			})

			Context("failure cases", func() {
				It("returns an error when the desired up command fails", func() {
					fakeAWSUp.ExecuteCall.Returns.Error = errors.New("failed execution")
					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).To(MatchError("failed execution"))
				})

				It("returns an error when undefined flags are passed", func() {
					err := command.Execute([]string{"--foo", "bar"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -foo"))
				})
			})
		})

		Context("when state contains an iaas", func() {
			Context("when iaas is AWS", func() {
				It("executes the AWS up", func() {
					err := command.Execute([]string{"--aws-access-key-id", "some-access-key-id"}, storage.State{IAAS: "aws"})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeAWSUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeAWSUp.ExecuteCall.Receives.Args).To(Equal([]string{"--aws-access-key-id", "some-access-key-id"}))
					Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(storage.State{IAAS: "aws"}))
				})
			})

			Context("when iaas is GCP", func() {
				It("executes the GCP up", func() {
					err := command.Execute([]string{}, storage.State{IAAS: "gcp"})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{IAAS: "gcp"}))
				})
			})

			Context("when iaas specified is different than the iaas in state", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--iaas", "aws"}, storage.State{IAAS: "gcp"})
					Expect(err).To(MatchError("the iaas provided must match the iaas in bbl-state.json"))
				})
			})
		})
	})
})
