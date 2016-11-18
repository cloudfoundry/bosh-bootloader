package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Update LBs", func() {
	var (
		command                   commands.UpdateLBs
		incomingState             storage.State
		certFilePath              string
		keyFilePath               string
		chainFilePath             string
		certificateManager        *fakes.CertificateManager
		certificateValidator      *fakes.CertificateValidator
		availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
		infrastructureManager     *fakes.InfrastructureManager
		awsCredentialValidator    *fakes.AWSCredentialValidator
		boshClientProvider        *fakes.BOSHClientProvider
		boshClient                *fakes.BOSHClient
		logger                    *fakes.Logger
		guidGenerator             *fakes.GuidGenerator
		stateStore                *fakes.StateStore
		stateValidator            *fakes.StateValidator
	)

	var updateLBs = func(certificatePath, keyPath, chainPath string, state storage.State) error {
		return command.Execute([]string{
			"--cert", certificatePath,
			"--key", keyPath,
			"--chain", chainPath,
		}, state)
	}

	BeforeEach(func() {
		var err error

		certificateManager = &fakes.CertificateManager{}
		certificateValidator = &fakes.CertificateValidator{}
		availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
		infrastructureManager = &fakes.InfrastructureManager{}
		awsCredentialValidator = &fakes.AWSCredentialValidator{}
		logger = &fakes.Logger{}
		guidGenerator = &fakes.GuidGenerator{}
		stateStore = &fakes.StateStore{}
		stateValidator = &fakes.StateValidator{}
		boshClient = &fakes.BOSHClient{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		boshClientProvider.ClientCall.Returns.Client = boshClient

		availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"a", "b", "c"}
		certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
			Body: "some-old-certificate-contents",
			ARN:  "some-certificate-arn",
		}

		guidGenerator.GenerateCall.Returns.Output = "abcd"
		infrastructureManager.ExistsCall.Returns.Exists = true

		incomingState = storage.State{
			Stack: storage.Stack{
				LBType:          "concourse",
				CertificateName: "some-certificate-name",
			},
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		}

		certFilePath, err = testhelpers.WriteContentsToTempFile("some-certificate-contents")
		Expect(err).NotTo(HaveOccurred())

		keyFilePath, err = testhelpers.WriteContentsToTempFile("some-key-contents")
		Expect(err).NotTo(HaveOccurred())

		chainFilePath, err = testhelpers.WriteContentsToTempFile("some-chain-contents")
		Expect(err).NotTo(HaveOccurred())

		command = commands.NewUpdateLBs(awsCredentialValidator, certificateManager,
			availabilityZoneRetriever, infrastructureManager, boshClientProvider, logger, certificateValidator, guidGenerator,
			stateStore, stateValidator)
	})

	Describe("Execute", func() {
		It("returns an error when state validator fails", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			err := command.Execute([]string{}, storage.State{})

			Expect(stateValidator.ValidateCall.CallCount).To(Equal(1))
			Expect(err).To(MatchError("state validator failed"))
		})

		It("creates the new certificate with private key", func() {
			updateLBs(certFilePath, keyFilePath, "", storage.State{
				Stack: storage.Stack{
					LBType:          "cf",
					CertificateName: "some-old-certificate-name",
				},
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
			})

			Expect(logger.StepCall.Messages).To(ContainElement("uploading new certificate"))
			Expect(certificateManager.CreateCall.Receives.Certificate).To(Equal(certFilePath))
			Expect(certificateManager.CreateCall.Receives.PrivateKey).To(Equal(keyFilePath))
			Expect(certificateManager.CreateCall.Receives.CertificateName).To(Equal("cf-elb-cert-abcd"))
		})

		Context("when uploading with a chain", func() {
			It("creates the new certificate with private key and chain", func() {
				updateLBs(certFilePath, keyFilePath, chainFilePath, storage.State{
					Stack: storage.Stack{
						LBType:          "cf",
						CertificateName: "some-old-certificate-name",
					},
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
				})

				Expect(certificateManager.CreateCall.Receives.Certificate).To(Equal(certFilePath))
				Expect(certificateManager.CreateCall.Receives.PrivateKey).To(Equal(keyFilePath))
				Expect(certificateManager.CreateCall.Receives.Chain).To(Equal(chainFilePath))
			})
		})

		It("updates cloudformation with the new certificate", func() {
			updateLBs(certFilePath, keyFilePath, "", storage.State{
				Stack: storage.Stack{
					Name:   "some-stack",
					LBType: "concourse",
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
			})

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(certificateManager.DescribeCall.Receives.CertificateName).To(Equal("concourse-elb-cert-abcd-some-env-id-timestamp"))

			Expect(infrastructureManager.UpdateCall.Receives.KeyPairName).To(Equal("some-key-pair"))
			Expect(infrastructureManager.UpdateCall.Receives.NumberOfAvailabilityZones).To(Equal(3))
			Expect(infrastructureManager.UpdateCall.Receives.StackName).To(Equal("some-stack"))
			Expect(infrastructureManager.UpdateCall.Receives.LBType).To(Equal("concourse"))
			Expect(infrastructureManager.UpdateCall.Receives.LBCertificateARN).To(Equal("some-certificate-arn"))
			Expect(infrastructureManager.UpdateCall.Receives.EnvID).To(Equal("some-env-id-timestamp"))
		})

		It("names the loadbalancer without EnvID when EnvID is not set", func() {
			updateLBs(certFilePath, keyFilePath, "", storage.State{
				Stack: storage.Stack{
					Name:   "some-stack",
					LBType: "concourse",
				},
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				KeyPair: storage.KeyPair{
					Name: "some-key-pair",
				},
				EnvID: "",
			})

			Expect(certificateManager.DescribeCall.Receives.CertificateName).To(Equal("concourse-elb-cert-abcd"))
		})

		It("deletes the existing certificate and private key", func() {
			updateLBs(certFilePath, keyFilePath, "", storage.State{
				Stack: storage.Stack{
					LBType:          "cf",
					CertificateName: "some-certificate-name",
				},
			})

			Expect(logger.StepCall.Messages).To(ContainElement("deleting old certificate"))
			Expect(certificateManager.DeleteCall.Receives.CertificateName).To(Equal("some-certificate-name"))
		})

		It("checks if the bosh director exists", func() {
			err := updateLBs(certFilePath, keyFilePath, "", incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

			Expect(boshClient.InfoCall.CallCount).To(Equal(1))
		})

		Context("if the user hasn't bbl'd up yet", func() {
			It("returns an error if the stack does not exist", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false
				err := updateLBs(certFilePath, keyFilePath, "", storage.State{})
				Expect(err).To(MatchError(commands.BBLNotFound))
			})

			It("returns an error if the bosh director does not exist", func() {
				boshClient.InfoCall.Returns.Error = errors.New("director not found")

				err := updateLBs(certFilePath, keyFilePath, "", storage.State{
					Stack: storage.Stack{
						LBType:          "concourse",
						CertificateName: "some-certificate-name",
					},
				})
				Expect(err).To(MatchError(commands.BBLNotFound))
			})
		})

		It("returns an error if there is no lb", func() {
			err := updateLBs(certFilePath, keyFilePath, "", storage.State{
				Stack: storage.Stack{
					LBType: "none",
				},
			})
			Expect(err).To(MatchError(commands.LBNotFound))
		})

		It("does not update the certificate if the provided certificate is the same", func() {
			certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
				Body:  "\nsome-certificate-contents\n",
				Chain: "\nsome-chain-contents\n",
			}

			err := updateLBs(certFilePath, keyFilePath, chainFilePath, incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.PrintlnCall.Receives.Message).To(Equal("no updates are to be performed"))

			Expect(certificateManager.CreateCall.CallCount).To(Equal(0))
			Expect(certificateManager.DeleteCall.CallCount).To(Equal(0))
			Expect(infrastructureManager.UpdateCall.CallCount).To(Equal(0))
		})

		It("returns an error if the certificate is the same and the chain has changed", func() {
			certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
				Body: "\nsome-certificate-contents\n",
			}

			err := updateLBs(certFilePath, keyFilePath, chainFilePath, incomingState)
			Expect(err).To(MatchError("you cannot change the chain after the lb has been created, please delete and re-create the lb with the chain"))

			Expect(certificateManager.CreateCall.CallCount).To(Equal(0))
			Expect(certificateManager.DeleteCall.CallCount).To(Equal(0))
			Expect(infrastructureManager.UpdateCall.CallCount).To(Equal(0))
		})

		It("returns an error when the certificate validator fails", func() {
			certificateValidator.ValidateCall.Returns.Error = errors.New("failed to validate")
			err := command.Execute([]string{
				"--cert", "/path/to/cert",
				"--key", "/path/to/key",
				"--chain", "/path/to/chain",
			}, storage.State{
				Stack: storage.Stack{
					LBType: "concourse",
				},
			})

			Expect(err).To(MatchError("failed to validate"))
			Expect(certificateValidator.ValidateCall.Receives.Command).To(Equal("update-lbs"))
			Expect(certificateValidator.ValidateCall.Receives.CertificatePath).To(Equal("/path/to/cert"))
			Expect(certificateValidator.ValidateCall.Receives.KeyPath).To(Equal("/path/to/key"))
			Expect(certificateValidator.ValidateCall.Receives.ChainPath).To(Equal("/path/to/chain"))

			Expect(certificateManager.CreateCall.CallCount).To(Equal(0))
			Expect(certificateManager.DeleteCall.CallCount).To(Equal(0))
		})

		Context("when --skip-if-missing is provided", func() {
			It("no-ops when lb does not exist", func() {
				err := command.Execute([]string{
					"--cert", certFilePath,
					"--key", keyFilePath,
					"--skip-if-missing",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(infrastructureManager.UpdateCall.CallCount).To(Equal(0))
				Expect(certificateManager.CreateCall.CallCount).To(Equal(0))

				Expect(logger.PrintlnCall.Receives.Message).To(Equal(`no lb type exists, skipping...`))
			})

			DescribeTable("updates the lb if the lb exists",
				func(currentLBType string) {
					incomingState.Stack.LBType = currentLBType
					err := command.Execute([]string{
						"--cert", certFilePath,
						"--key", keyFilePath,
						"--skip-if-missing",
					}, incomingState)
					Expect(err).NotTo(HaveOccurred())

					Expect(infrastructureManager.UpdateCall.CallCount).To(Equal(1))
					Expect(certificateManager.CreateCall.CallCount).To(Equal(1))
				},
				Entry("when the current lb-type is 'cf'", "cf"),
				Entry("when the current lb-type is 'concourse'", "concourse"),
			)
		})

		Describe("state manipulation", func() {
			It("updates the state with the new certificate name", func() {
				err := updateLBs(certFilePath, keyFilePath, "", storage.State{
					Stack: storage.Stack{
						LBType:          "cf",
						CertificateName: "some-certificate-name",
					},
					EnvID: "some-env-timestamp",
				})
				Expect(err).NotTo(HaveOccurred())

				state := stateStore.SetCall.Receives.State
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
				Expect(state.Stack.CertificateName).To(Equal("cf-elb-cert-abcd-some-env-timestamp"))
			})
		})

		Describe("failure cases", func() {
			It("returns an error when the chain file cannot be opened", func() {
				certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
					Body: "some-certificate-contents",
				}

				err := updateLBs(certFilePath, keyFilePath, "/some/fake/path", storage.State{
					Stack: storage.Stack{
						LBType:          "cf",
						CertificateName: "some-certificate-name",
					},
				})

				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when aws credential validator fails", func() {
				awsCredentialValidator.ValidateCall.Returns.Error = errors.New("aws credentials validator failed")

				err := command.Execute([]string{}, storage.State{})

				Expect(err).To(MatchError("aws credentials validator failed"))
			})

			It("returns an error when the original certificate cannot be described", func() {
				certificateManager.DescribeCall.Stub = func(certificateName string) (iam.Certificate, error) {
					if certificateName == "some-certificate-name" {
						return iam.Certificate{}, errors.New("old certificate failed to describe")
					}

					return iam.Certificate{}, nil
				}

				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("old certificate failed to describe"))
			})

			It("returns an error when new certificate cannot be described", func() {
				certificateManager.DescribeCall.Stub = func(certificateName string) (iam.Certificate, error) {
					if certificateName == "concourse-elb-cert-abcd" {
						return iam.Certificate{}, errors.New("new certificate failed to describe")
					}

					return iam.Certificate{}, nil
				}

				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("new certificate failed to describe"))
			})

			It("returns an error when the certificate file cannot be read", func() {
				err := updateLBs("some-fake-file", keyFilePath, "", incomingState)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when the infrastructure manager fails to check the existance of a stack", func() {
				infrastructureManager.ExistsCall.Returns.Error = errors.New("failed to check for stack")
				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("failed to check for stack"))
			})

			It("returns an error when invalid flags are provided", func() {
				err := command.Execute([]string{
					"--invalid-flag",
				}, incomingState)

				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined")))
				Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(0))
			})

			It("returns an error when infrastructure update fails", func() {
				infrastructureManager.UpdateCall.Returns.Error = errors.New("failed to update stack")
				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("failed to update stack"))
			})

			It("returns an error when availability zone retriever fails", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("az retrieve failed")
				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("az retrieve failed"))
			})

			It("returns an error when certificate creation fails", func() {
				certificateManager.CreateCall.Returns.Error = errors.New("certificate creation failed")
				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("certificate creation failed"))
			})

			It("returns an error when certificate deletion fails", func() {
				certificateManager.DeleteCall.Returns.Error = errors.New("certificate deletion failed")
				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("certificate deletion failed"))
			})

			It("returns an error when a GUID cannot be generated", func() {
				guidGenerator.GenerateCall.Returns.Error = errors.New("Out of entropy in the universe")
				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("Out of entropy in the universe"))
			})

			It("returns an error when state cannot be set", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set state")}}
				err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("failed to set state"))
			})
		})
	})
})
