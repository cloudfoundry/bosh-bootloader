package commands_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWS Create LBs", func() {
	Describe("Execute", func() {
		var (
			command                   commands.AWSCreateLBs
			certificateManager        *fakes.CertificateManager
			infrastructureManager     *fakes.InfrastructureManager
			boshClient                *fakes.BOSHClient
			boshClientProvider        *fakes.BOSHClientProvider
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			credentialValidator       *fakes.CredentialValidator
			logger                    *fakes.Logger
			cloudConfigManager        *fakes.CloudConfigManager
			certificateValidator      *fakes.CertificateValidator
			guidGenerator             *fakes.GuidGenerator
			stateStore                *fakes.StateStore
			incomingState             storage.State
		)

		BeforeEach(func() {
			certificateManager = &fakes.CertificateManager{}
			infrastructureManager = &fakes.InfrastructureManager{}
			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
			boshClient = &fakes.BOSHClient{}
			boshClientProvider = &fakes.BOSHClientProvider{}
			credentialValidator = &fakes.CredentialValidator{}
			logger = &fakes.Logger{}
			cloudConfigManager = &fakes.CloudConfigManager{}
			certificateValidator = &fakes.CertificateValidator{}
			guidGenerator = &fakes.GuidGenerator{}
			stateStore = &fakes.StateStore{}

			boshClientProvider.ClientCall.Returns.Client = boshClient

			infrastructureManager.ExistsCall.Returns.Exists = true

			guidGenerator.GenerateCall.Returns.Output = "abcd"

			incomingState = storage.State{
				Stack: storage.Stack{
					Name:   "some-stack",
					BOSHAZ: "some-bosh-az",
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
				EnvID: "some-env-id-timestamp",
			}

			command = commands.NewAWSCreateLBs(logger, credentialValidator, certificateManager, infrastructureManager,
				availabilityZoneRetriever, boshClientProvider, cloudConfigManager, certificateValidator, guidGenerator,
				stateStore)
		})

		It("returns an error if credential validator fails", func() {
			credentialValidator.ValidateCall.Returns.Error = errors.New("failed to validate aws credentials")
			err := command.Execute(commands.AWSCreateLBsConfig{}, storage.State{})
			Expect(err).To(MatchError("failed to validate aws credentials"))
		})

		It("uploads a cert and key", func() {
			err := command.Execute(commands.AWSCreateLBsConfig{
				LBType:   "concourse",
				CertPath: "temp/some-cert.crt",
				KeyPath:  "temp/some-key.key",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateManager.CreateCall.Receives.Certificate).To(Equal("temp/some-cert.crt"))
			Expect(certificateManager.CreateCall.Receives.PrivateKey).To(Equal("temp/some-key.key"))
			Expect(certificateManager.CreateCall.Receives.CertificateName).To(Equal("concourse-elb-cert-abcd-some-env-id-timestamp"))
			Expect(logger.StepCall.Messages).To(ContainElement("uploading certificate"))

		})

		It("uploads a cert and key with chain", func() {
			err := command.Execute(commands.AWSCreateLBsConfig{
				LBType:    "concourse",
				CertPath:  "temp/some-cert.crt",
				KeyPath:   "temp/some-key.key",
				ChainPath: "temp/some-chain.crt",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateManager.CreateCall.Receives.Chain).To(Equal("temp/some-chain.crt"))

			Expect(certificateValidator.ValidateCall.Receives.Command).To(Equal("create-lbs"))
			Expect(certificateValidator.ValidateCall.Receives.CertificatePath).To(Equal("temp/some-cert.crt"))
			Expect(certificateValidator.ValidateCall.Receives.KeyPath).To(Equal("temp/some-key.key"))
			Expect(certificateValidator.ValidateCall.Receives.ChainPath).To(Equal("temp/some-chain.crt"))
		})

		It("creates a load balancer in cloudformation with certificate", func() {
			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"a", "b", "c"}
			certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
				ARN: "some-certificate-arn",
			}

			err := command.Execute(commands.AWSCreateLBsConfig{
				LBType:   "concourse",
				CertPath: "temp/some-cert.crt",
				KeyPath:  "temp/some-key.key",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(certificateManager.DescribeCall.Receives.CertificateName).To(Equal("concourse-elb-cert-abcd-some-env-id-timestamp"))

			Expect(infrastructureManager.UpdateCall.Receives.KeyPairName).To(Equal("some-key-pair"))
			Expect(infrastructureManager.UpdateCall.Receives.AZs).To(Equal([]string{"a", "b", "c"}))
			Expect(infrastructureManager.UpdateCall.Receives.StackName).To(Equal("some-stack"))
			Expect(infrastructureManager.UpdateCall.Receives.LBType).To(Equal("concourse"))
			Expect(infrastructureManager.UpdateCall.Receives.LBCertificateARN).To(Equal("some-certificate-arn"))
			Expect(infrastructureManager.UpdateCall.Receives.EnvID).To(Equal("some-env-id-timestamp"))
			Expect(infrastructureManager.UpdateCall.Receives.BOSHAZ).To(Equal("some-bosh-az"))
		})

		It("names the loadbalancer without EnvID when EnvID is not set", func() {
			incomingState.EnvID = ""

			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"a", "b", "c"}
			certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
				ARN: "some-certificate-arn",
			}

			err := command.Execute(commands.AWSCreateLBsConfig{
				LBType:   "concourse",
				CertPath: "temp/some-cert.crt",
				KeyPath:  "temp/some-key.key",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateManager.DescribeCall.Receives.CertificateName).To(Equal("concourse-elb-cert-abcd"))

		})

		Context("when the bbl environment has a BOSH director", func() {
			It("updates the cloud config with a state that has lb type", func() {
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "temp/some-cert.crt",
					KeyPath:  "temp/some-key.key",
				}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.Receives.State.Stack.LBType).To(Equal("concourse"))
			})
		})

		Context("when the bbl environment does not have a BOSH director", func() {
			It("does not create a BOSH client", func() {
				incomingState = storage.State{
					NoDirector: true,
					Stack: storage.Stack{
						Name:   "some-stack",
						BOSHAZ: "some-bosh-az",
					},
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
					KeyPair: storage.KeyPair{
						Name: "some-key-pair",
					},
					EnvID: "some-env-id-timestamp",
				}
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "temp/some-cert.crt",
					KeyPath:  "temp/some-key.key",
				}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshClientProvider.ClientCall.CallCount).To(Equal(0))
			})

			It("does not call cloudConfigManager", func() {
				incomingState = storage.State{
					NoDirector: true,
					Stack: storage.Stack{
						Name:   "some-stack",
						BOSHAZ: "some-bosh-az",
					},
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
					KeyPair: storage.KeyPair{
						Name: "some-key-pair",
					},
					EnvID: "some-env-id-timestamp",
				}
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "temp/some-cert.crt",
					KeyPath:  "temp/some-key.key",
				}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(cloudConfigManager.UpdateCall.CallCount).To(Equal(0))
			})
		})

		Context("when --skip-if-exists is provided", func() {
			It("no-ops when lb exists", func() {
				incomingState.Stack.LBType = "cf"
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:       "concourse",
					CertPath:     "temp/some-cert.crt",
					KeyPath:      "temp/some-key.key",
					SkipIfExists: true,
				}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(infrastructureManager.UpdateCall.CallCount).To(Equal(0))
				Expect(certificateManager.CreateCall.CallCount).To(Equal(0))

				Expect(logger.PrintlnCall.Receives.Message).To(Equal(`lb type "cf" exists, skipping...`))
			})

			DescribeTable("creates the lb if the lb does not exist",
				func(currentLBType string) {
					incomingState.Stack.LBType = currentLBType
					err := command.Execute(commands.AWSCreateLBsConfig{
						LBType:       "concourse",
						CertPath:     "temp/some-cert.crt",
						KeyPath:      "temp/some-key.key",
						SkipIfExists: true,
					}, incomingState)
					Expect(err).NotTo(HaveOccurred())

					Expect(infrastructureManager.UpdateCall.CallCount).To(Equal(1))
					Expect(certificateManager.CreateCall.CallCount).To(Equal(1))
				},
				Entry("when the current lb-type is 'none'", "none"),
				Entry("when the current lb-type is ''", ""),
			)
		})

		Context("invalid lb type", func() {
			It("returns an error", func() {
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "some-invalid-lb",
					CertPath: "temp/some-cert.crt",
					KeyPath:  "temp/some-key.key",
				}, incomingState)
				Expect(err).To(MatchError("\"some-invalid-lb\" is not a valid lb type, valid lb types are: concourse and cf"))
			})

			It("returns a helpful error when no lb type is provided", func() {
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "",
					CertPath: "temp/some-cert.crt",
					KeyPath:  "temp/some-key.key",
				}, incomingState)
				Expect(err).To(MatchError("--type is a required flag"))
			})
		})

		Context("fast fail if the stack or BOSH director does not exist", func() {
			It("returns an error when the stack does not exist", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false

				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "temp/some-cert.crt",
					KeyPath:  "temp/some-key.key",
				}, incomingState)

				Expect(infrastructureManager.ExistsCall.Receives.StackName).To(Equal("some-stack"))

				Expect(err).To(MatchError(commands.BBLNotFound))
			})

			It("returns an error when the BOSH director does not exist", func() {
				boshClient.InfoCall.Returns.Error = errors.New("director not found")
				infrastructureManager.ExistsCall.Returns.Exists = true

				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "temp/some-cert.crt",
					KeyPath:  "temp/some-key.key",
				}, incomingState)

				Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

				Expect(boshClient.InfoCall.CallCount).To(Equal(1))

				Expect(err).To(MatchError(commands.BBLNotFound))
			})
		})

		Context("state manipulation", func() {
			Context("when the env id does not exist", func() {
				It("saves state with new certificate name and lb type", func() {
					err := command.Execute(commands.AWSCreateLBsConfig{
						LBType:   "concourse",
						CertPath: "temp/some-cert.crt",
						KeyPath:  "temp/some-key.key",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(stateStore.SetCall.CallCount).To(Equal(1))
					state := stateStore.SetCall.Receives[0].State
					Expect(state.Stack.CertificateName).To(Equal("concourse-elb-cert-abcd"))
					Expect(state.Stack.LBType).To(Equal("concourse"))
				})
			})

			Context("when the env id exists", func() {
				It("saves state with new certificate name and lb type", func() {
					err := command.Execute(commands.AWSCreateLBsConfig{
						LBType:   "concourse",
						CertPath: "temp/some-cert.crt",
						KeyPath:  "temp/some-key.key",
					}, storage.State{
						EnvID: "some-env-id-timestamp",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(stateStore.SetCall.CallCount).To(Equal(1))
					state := stateStore.SetCall.Receives[0].State
					Expect(state.Stack.CertificateName).To(Equal("concourse-elb-cert-abcd-some-env-id-timestamp"))
					Expect(state.Stack.LBType).To(Equal("concourse"))
				})
			})
		})

		Context("required args", func() {
			It("returns an error when certificate validator fails for cert and key", func() {
				certificateValidator.ValidateCall.Returns.Error = errors.New("failed to validate")
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "/path/to/cert",
					KeyPath:  "/path/to/key",
				}, storage.State{})

				Expect(err).To(MatchError("failed to validate"))
				Expect(certificateValidator.ValidateCall.Receives.Command).To(Equal("create-lbs"))
				Expect(certificateValidator.ValidateCall.Receives.CertificatePath).To(Equal("/path/to/cert"))
				Expect(certificateValidator.ValidateCall.Receives.KeyPath).To(Equal("/path/to/key"))
				Expect(certificateValidator.ValidateCall.Receives.ChainPath).To(Equal(""))

				Expect(certificateManager.CreateCall.CallCount).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			DescribeTable("returns an error when an lb already exists",
				func(newLbType, oldLbType string) {
					err := command.Execute(commands.AWSCreateLBsConfig{
						LBType:   "concourse",
						CertPath: "/path/to/cert",
						KeyPath:  "/path/to/key",
					}, storage.State{
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
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "/path/to/cert",
					KeyPath:  "/path/to/key",
				}, storage.State{})
				Expect(err).To(MatchError("failed to check for stack"))
			})

			Context("when availability zone retriever fails", func() {
				It("returns an error", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("failed to retrieve azs")

					err := command.Execute(commands.AWSCreateLBsConfig{
						LBType:   "concourse",
						CertPath: "/path/to/cert",
						KeyPath:  "/path/to/key",
					}, storage.State{})
					Expect(err).To(MatchError("failed to retrieve azs"))
				})
			})

			Context("when update infrastructure manager fails", func() {
				It("returns an error", func() {
					infrastructureManager.UpdateCall.Returns.Error = errors.New("failed to update infrastructure")

					err := command.Execute(commands.AWSCreateLBsConfig{
						LBType:   "concourse",
						CertPath: "/path/to/cert",
						KeyPath:  "/path/to/key",
					}, storage.State{})
					Expect(err).To(MatchError("failed to update infrastructure"))
				})
			})

			Context("when certificate manager fails to create a certificate", func() {
				It("returns an error", func() {
					certificateManager.CreateCall.Returns.Error = errors.New("failed to create cert")

					err := command.Execute(commands.AWSCreateLBsConfig{
						LBType:   "concourse",
						CertPath: "/path/to/cert",
						KeyPath:  "/path/to/key",
					}, storage.State{})
					Expect(err).To(MatchError("failed to create cert"))
				})
			})

			Context("when cloud config manager update fails", func() {
				It("returns an error", func() {
					cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update cloud config")

					err := command.Execute(commands.AWSCreateLBsConfig{
						LBType:   "concourse",
						CertPath: "/path/to/cert",
						KeyPath:  "/path/to/key",
					}, storage.State{})
					Expect(err).To(MatchError("failed to update cloud config"))
				})
			})

			It("returns an error when a GUID cannot be generated", func() {
				guidGenerator.GenerateCall.Returns.Error = errors.New("Out of entropy in the universe")
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "/path/to/cert",
					KeyPath:  "/path/to/key",
				}, storage.State{})
				Expect(err).To(MatchError("Out of entropy in the universe"))
			})

			It("returns an error when the state fails to save", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to save state")}}
				err := command.Execute(commands.AWSCreateLBsConfig{
					LBType:   "concourse",
					CertPath: "/path/to/cert",
					KeyPath:  "/path/to/key",
				}, storage.State{})
				Expect(err).To(MatchError("failed to save state"))
			})
		})
	})
})
