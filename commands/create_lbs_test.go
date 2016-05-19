package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create LBs", func() {
	Describe("Execute", func() {
		var (
			command                   commands.CreateLBs
			certificateManager        *fakes.CertificateManager
			infrastructureManager     *fakes.InfrastructureManager
			boshClient                *fakes.BOSHClient
			boshClientProvider        *fakes.BOSHClientProvider
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			boshCloudConfigurator     *fakes.BoshCloudConfigurator
			awsCredentialValidator    *fakes.AWSCredentialValidator
			incomingState             storage.State
		)

		BeforeEach(func() {
			certificateManager = &fakes.CertificateManager{}
			infrastructureManager = &fakes.InfrastructureManager{}
			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
			boshCloudConfigurator = &fakes.BoshCloudConfigurator{}
			boshClient = &fakes.BOSHClient{}
			boshClientProvider = &fakes.BOSHClientProvider{}
			awsCredentialValidator = &fakes.AWSCredentialValidator{}

			boshClientProvider.ClientCall.Returns.Client = boshClient

			infrastructureManager.ExistsCall.Returns.Exists = true

			incomingState = storage.State{
				Stack: storage.Stack{
					Name: "some-stack",
				},
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				KeyPair: storage.KeyPair{
					Name: "some-key-pair",
				},
				BOSH: storage.BOSH{
					DirectorAddress:  "some-director-address",
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}

			command = commands.NewCreateLBs(awsCredentialValidator, certificateManager, infrastructureManager,
				availabilityZoneRetriever, boshClientProvider, boshCloudConfigurator)
		})

		It("returns an error if aws credential validator fails", func() {
			awsCredentialValidator.ValidateCall.Returns.Error = errors.New("failed to validate aws credentials")
			_, err := command.Execute([]string{}, storage.State{})
			Expect(err).To(MatchError("failed to validate aws credentials"))
		})

		It("uploads a cert and key", func() {
			_, err := command.Execute([]string{
				"--type", "concourse",
				"--cert", "temp/some-cert.crt",
				"--key", "temp/some-key.key",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateManager.CreateCall.Receives.Certificate).To(Equal("temp/some-cert.crt"))
			Expect(certificateManager.CreateCall.Receives.PrivateKey).To(Equal("temp/some-key.key"))
		})

		It("creates a load balancer in cloudformation with certificate", func() {
			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"a", "b", "c"}
			certificateManager.CreateCall.Returns.CertificateName = "some-certificate-name"
			certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
				ARN: "some-certificate-arn",
			}
			_, err := command.Execute([]string{
				"--type", "concourse",
				"--cert", "temp/some-cert.crt",
				"--key", "temp/some-key.key",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(certificateManager.DescribeCall.Receives.CertificateName).To(Equal("some-certificate-name"))

			Expect(infrastructureManager.UpdateCall.Receives.KeyPairName).To(Equal("some-key-pair"))
			Expect(infrastructureManager.UpdateCall.Receives.NumberOfAvailabilityZones).To(Equal(3))
			Expect(infrastructureManager.UpdateCall.Receives.StackName).To(Equal("some-stack"))
			Expect(infrastructureManager.UpdateCall.Receives.LBType).To(Equal("concourse"))
			Expect(infrastructureManager.UpdateCall.Receives.LBCertificateARN).To(Equal("some-certificate-arn"))
		})

		It("updates the cloud config with lb type", func() {
			infrastructureManager.UpdateCall.Returns.Stack = cloudformation.Stack{
				Name: "some-stack",
			}
			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"a", "b", "c"}
			_, err := command.Execute([]string{
				"--type", "concourse",
				"--cert", "temp/some-cert.crt",
				"--key", "temp/some-key.key",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshCloudConfigurator.ConfigureCall.Receives.Stack).To(Equal(cloudformation.Stack{
				Name: "some-stack",
			}))
			Expect(boshCloudConfigurator.ConfigureCall.Receives.AZs).To(Equal([]string{"a", "b", "c"}))
			Expect(boshCloudConfigurator.ConfigureCall.Receives.Client).To(Equal(boshClient))
		})

		Context("invalid lb type", func() {
			It("returns an error", func() {
				_, err := command.Execute([]string{
					"--type", "some-invalid-lb",
					"--cert", "temp/some-cert.crt",
					"--key", "temp/some-key.key",
				}, storage.State{})
				Expect(err).To(MatchError("\"some-invalid-lb\" is not a valid lb type, valid lb types are: concourse and cf"))
			})
		})

		Context("fast fail if the stack or BOSH director does not exist", func() {
			It("returns an error when the stack does not exist", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false

				_, err := command.Execute([]string{
					"--type", "concourse",
					"--cert", "temp/some-cert.crt",
					"--key", "temp/some-key.key",
				}, incomingState)

				Expect(infrastructureManager.ExistsCall.Receives.StackName).To(Equal("some-stack"))

				Expect(err).To(MatchError(commands.BBLNotFound))
			})

			It("returns an error when the BOSH director does not exist", func() {
				boshClient.InfoCall.Returns.Error = errors.New("director not found")
				infrastructureManager.ExistsCall.Returns.Exists = true

				_, err := command.Execute([]string{
					"--type", "concourse",
					"--cert", "temp/some-cert.crt",
					"--key", "temp/some-key.key",
				}, incomingState)

				Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

				Expect(boshClient.InfoCall.CallCount).To(Equal(1))

				Expect(err).To(MatchError(commands.BBLNotFound))
			})
		})

		Context("state manipulation", func() {
			It("returns a state with new certificate name and lb type", func() {
				certificateManager.CreateCall.Returns.CertificateName = "some-certificate-name"
				state, err := command.Execute([]string{
					"--type", "concourse",
					"--cert", "temp/some-cert.crt",
					"--key", "temp/some-key.key",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(state.Stack.CertificateName).To(Equal("some-certificate-name"))
				Expect(state.Stack.LBType).To(Equal("concourse"))
			})

		})

		Context("failure cases", func() {
			DescribeTable("returns an error when an lb already exists",
				func(newLbType, oldLbType string) {
					_, err := command.Execute([]string{"--type", newLbType}, storage.State{
						Stack: storage.Stack{
							LBType: oldLbType,
						},
					})
					Expect(err).To(MatchError(fmt.Sprintf("bbl already has a %s load balancer attached, please remove the previous load balancer before attaching a new one", oldLbType)))
				},
				Entry("when the previous lb type is concourse", "concourse", "cf"),
				Entry("when the previous lb type is cf", "cf", "concourse"),
			)

			It("returns an error when the infrastructure manager fails to check the existance of a stack", func() {
				infrastructureManager.ExistsCall.Returns.Error = errors.New("failed to check for stack")
				_, err := command.Execute([]string{"--type", "concourse"}, storage.State{})
				Expect(err).To(MatchError("failed to check for stack"))
			})

			Context("when an invalid command line flag is supplied", func() {
				It("returns an error", func() {
					_, err := command.Execute([]string{"--invalid-flag"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -invalid-flag"))
				})
			})

			Context("when availability zone retriever fails", func() {
				It("returns an error", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("failed to retrieve azs")

					_, err := command.Execute([]string{"--type", "concourse"}, storage.State{})
					Expect(err).To(MatchError("failed to retrieve azs"))
				})
			})

			Context("when update infrastructure manager fails", func() {
				It("returns an error", func() {
					infrastructureManager.UpdateCall.Returns.Error = errors.New("failed to update infrastructure")

					_, err := command.Execute([]string{"--type", "concourse"}, storage.State{})
					Expect(err).To(MatchError("failed to update infrastructure"))
				})
			})

			Context("when certificate manager fails to create a certificate", func() {
				It("returns an error", func() {
					certificateManager.CreateCall.Returns.Error = errors.New("failed to create cert")

					_, err := command.Execute([]string{"--type", "concourse"}, storage.State{})
					Expect(err).To(MatchError("failed to create cert"))
				})
			})

			Context("when bosh cloud configurator fails to configure", func() {
				It("returns an error", func() {
					boshCloudConfigurator.ConfigureCall.Returns.Error = errors.New("failed to configure")

					_, err := command.Execute([]string{"--type", "concourse"}, storage.State{})
					Expect(err).To(MatchError("failed to configure"))
				})
			})
		})
	})
})
