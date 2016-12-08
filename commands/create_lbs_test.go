package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-lbs", func() {
	var (
		command          commands.CreateLBs
		fakeAWSCreateLBs *fakes.AWSCreateLBs
		fakeGCPCreateLBs *fakes.GCPCreateLBs
	)

	BeforeEach(func() {
		fakeAWSCreateLBs = &fakes.AWSCreateLBs{}
		fakeGCPCreateLBs = &fakes.GCPCreateLBs{}

		command = commands.NewCreateLBs(fakeAWSCreateLBs, fakeGCPCreateLBs)
	})

	Describe("Execute", func() {
		It("creates a GCP lb type if the iaas if GCP", func() {
			err := command.Execute([]string{
				"--type", "concourse",
			}, storage.State{
				IAAS: "gcp",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeGCPCreateLBs.ExecuteCall.Receives.Args).Should(Equal([]string{"--type", "concourse"}))
		})

		It("creates an AWS lb type if the iaas is AWS", func() {
			err := command.Execute([]string{
				"--type", "concourse",
				"--cert", "my-cert",
				"--key", "my-key",
				"--chain", "my-chain",
				"--skip-if-exists", "true",
			}, storage.State{
				IAAS: "aws",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeAWSCreateLBs.ExecuteCall.Receives.Config).Should(Equal(commands.AWSCreateLBsConfig{
				LBType:       "concourse",
				CertPath:     "my-cert",
				KeyPath:      "my-key",
				ChainPath:    "my-chain",
				SkipIfExists: true,
			}))
		})

		Context("failure cases", func() {
			It("returns an error when an invalid command line flag is supplied", func() {
				err := command.Execute([]string{"--invalid-flag"}, storage.State{})
				Expect(err).To(MatchError("flag provided but not defined: -invalid-flag"))
			})

			It("returns an error when the AWSCreateLBs fails", func() {
				fakeAWSCreateLBs.ExecuteCall.Returns.Error = errors.New("something bad happened")

				err := command.Execute([]string{"some-aws-args"}, storage.State{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("something bad happened"))
			})

			It("returns an error when the GCPCreateLBs fails", func() {
				fakeGCPCreateLBs.ExecuteCall.Returns.Error = errors.New("something bad happened")

				err := command.Execute([]string{"some-gcp-args"}, storage.State{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
