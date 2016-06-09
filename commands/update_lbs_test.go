package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

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
	)

	var updateLBs = func(certificatePath, keyPath, chainPath string, state storage.State) (storage.State, error) {
		return command.Execute([]string{
			"--cert", certificatePath,
			"--key", keyPath,
			"--chain", chainPath,
		}, state)
	}

	var temporaryFileContaining = func(fileContents string) string {
		temporaryFile, err := ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(temporaryFile.Name(), []byte(fileContents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		return temporaryFile.Name()
	}

	BeforeEach(func() {
		certificateManager = &fakes.CertificateManager{}
		certificateValidator = &fakes.CertificateValidator{}
		availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
		infrastructureManager = &fakes.InfrastructureManager{}
		awsCredentialValidator = &fakes.AWSCredentialValidator{}
		logger = &fakes.Logger{}

		boshClient = &fakes.BOSHClient{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		boshClientProvider.ClientCall.Returns.Client = boshClient

		availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"a", "b", "c"}
		certificateManager.CreateCall.Returns.CertificateName = "some-certificate-name"
		certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
			Body: "some-old-certificate-contents",
			ARN:  "some-certificate-arn",
		}

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

		certFilePath = temporaryFileContaining("some-certificate-contents")
		keyFilePath = temporaryFileContaining("some-key-contents")
		chainFilePath = temporaryFileContaining("some-chain-contents")

		command = commands.NewUpdateLBs(awsCredentialValidator, certificateManager,
			availabilityZoneRetriever, infrastructureManager, boshClientProvider, logger, certificateValidator)
	})

	Describe("Execute", func() {
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
			})

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("some-region"))

			Expect(certificateManager.DescribeCall.Receives.CertificateName).To(Equal("some-certificate-name"))

			Expect(infrastructureManager.UpdateCall.Receives.KeyPairName).To(Equal("some-key-pair"))
			Expect(infrastructureManager.UpdateCall.Receives.NumberOfAvailabilityZones).To(Equal(3))
			Expect(infrastructureManager.UpdateCall.Receives.StackName).To(Equal("some-stack"))
			Expect(infrastructureManager.UpdateCall.Receives.LBType).To(Equal("concourse"))
			Expect(infrastructureManager.UpdateCall.Receives.LBCertificateARN).To(Equal("some-certificate-arn"))
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
			_, err := updateLBs(certFilePath, keyFilePath, "", incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

			Expect(boshClient.InfoCall.CallCount).To(Equal(1))
		})

		Context("if the user hasn't bbl'd up yet", func() {
			It("returns an error if the stack does not exist", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false
				_, err := updateLBs(certFilePath, keyFilePath, "", storage.State{})
				Expect(err).To(MatchError(commands.BBLNotFound))
			})

			It("returns an error if the bosh director does not exist", func() {
				boshClient.InfoCall.Returns.Error = errors.New("director not found")

				_, err := updateLBs(certFilePath, keyFilePath, "", storage.State{
					Stack: storage.Stack{
						LBType:          "concourse",
						CertificateName: "some-certificate-name",
					},
				})
				Expect(err).To(MatchError(commands.BBLNotFound))
			})
		})

		It("returns an error if there is no lb", func() {
			_, err := updateLBs(certFilePath, keyFilePath, "", storage.State{
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

			_, err := updateLBs(certFilePath, keyFilePath, chainFilePath, incomingState)
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

			_, err := updateLBs(certFilePath, keyFilePath, chainFilePath, incomingState)
			Expect(err).To(MatchError("you cannot change the chain after the lb has been created, please delete and re-create the lb with the chain"))

			Expect(certificateManager.CreateCall.CallCount).To(Equal(0))
			Expect(certificateManager.DeleteCall.CallCount).To(Equal(0))
			Expect(infrastructureManager.UpdateCall.CallCount).To(Equal(0))
		})

		It("returns an error when the certificate validator fails", func() {
			certificateValidator.ValidateCall.Returns.Error = errors.New("failed to validate")
			_, err := command.Execute([]string{
				"--cert", "/path/to/cert",
				"--key", "/path/to/key",
				"--chain", "/path/to/chain",
			}, storage.State{
				Stack: storage.Stack{
					LBType: "concourse",
				},
			})

			Expect(err).To(MatchError("failed to validate"))
			Expect(certificateValidator.ValidateCall.Receives.Command).To(Equal("unsupported-update-lbs"))
			Expect(certificateValidator.ValidateCall.Receives.CertificatePath).To(Equal("/path/to/cert"))
			Expect(certificateValidator.ValidateCall.Receives.KeyPath).To(Equal("/path/to/key"))
			Expect(certificateValidator.ValidateCall.Receives.ChainPath).To(Equal("/path/to/chain"))

			Expect(certificateManager.CreateCall.CallCount).To(Equal(0))
			Expect(certificateManager.DeleteCall.CallCount).To(Equal(0))
		})

		Context("when --skip-if-missing is provided", func() {
			It("no-ops when lb does not exist", func() {
				_, err := command.Execute([]string{
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
					_, err := command.Execute([]string{
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
				certificateManager.CreateCall.Returns.CertificateName = "some-new-certificate-name"

				state, err := updateLBs(certFilePath, keyFilePath, "", storage.State{
					Stack: storage.Stack{
						LBType:          "cf",
						CertificateName: "some-certificate-name",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(state.Stack.CertificateName).To(Equal("some-new-certificate-name"))
			})
		})

		Describe("failure cases", func() {
			It("returns an error when the chain file cannot be opened", func() {
				certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
					Body: "some-certificate-contents",
				}

				_, err := updateLBs(certFilePath, keyFilePath, "/some/fake/path", storage.State{
					Stack: storage.Stack{
						LBType:          "cf",
						CertificateName: "some-certificate-name",
					},
				})

				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when aws credential validator fails", func() {
				awsCredentialValidator.ValidateCall.Returns.Error = errors.New("aws credentials validator failed")

				_, err := command.Execute([]string{}, storage.State{})

				Expect(err).To(MatchError("aws credentials validator failed"))
			})

			It("returns an error when the original certificate cannot be described", func() {
				certificateManager.DescribeCall.Stub = func(certificateName string) (iam.Certificate, error) {
					if certificateName == "some-certificate-name" {
						return iam.Certificate{}, errors.New("old certificate failed to describe")
					}

					return iam.Certificate{}, nil
				}

				_, err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("old certificate failed to describe"))
			})

			It("returns an error when new certificate cannot be described", func() {
				certificateManager.CreateCall.Returns.CertificateName = "some-new-certificate-name"

				certificateManager.DescribeCall.Stub = func(certificateName string) (iam.Certificate, error) {
					if certificateName == "some-new-certificate-name" {
						return iam.Certificate{}, errors.New("new certificate failed to describe")
					}

					return iam.Certificate{}, nil
				}

				_, err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("new certificate failed to describe"))
			})

			It("returns an error when the certificate file cannot be read", func() {
				_, err := updateLBs("some-fake-file", keyFilePath, "", incomingState)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when the infrastructure manager fails to check the existance of a stack", func() {
				infrastructureManager.ExistsCall.Returns.Error = errors.New("failed to check for stack")
				_, err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("failed to check for stack"))
			})

			It("returns an error when invalid flags are provided", func() {
				_, err := command.Execute([]string{
					"--invalid-flag",
				}, incomingState)

				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined")))
			})

			It("returns an error when infrastructure update fails", func() {
				infrastructureManager.UpdateCall.Returns.Error = errors.New("failed to update stack")
				_, err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("failed to update stack"))
			})

			It("returns an error when availability zone retriever fails", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("az retrieve failed")
				_, err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("az retrieve failed"))
			})

			It("returns an error when certificate creation fails", func() {
				certificateManager.CreateCall.Returns.Error = errors.New("certificate creation failed")
				_, err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("certificate creation failed"))
			})

			It("returns an error when certificate deletion fails", func() {
				certificateManager.DeleteCall.Returns.Error = errors.New("certificate deletion failed")
				_, err := updateLBs(certFilePath, keyFilePath, "", incomingState)
				Expect(err).To(MatchError("certificate deletion failed"))
			})
		})
	})
})
