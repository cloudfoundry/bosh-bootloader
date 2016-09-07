package commands_test

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Destroy", func() {
	var (
		destroy                commands.Destroy
		boshDeleter            *fakes.BOSHDeleter
		stackManager           *fakes.StackManager
		infrastructureManager  *fakes.InfrastructureManager
		vpcStatusChecker       *fakes.VPCStatusChecker
		stringGenerator        *fakes.StringGenerator
		logger                 *fakes.Logger
		keyPairDeleter         *fakes.KeyPairDeleter
		certificateDeleter     *fakes.CertificateDeleter
		awsCredentialValidator *fakes.AWSCredentialValidator
		stateStore             *fakes.StateStore
		stdin                  *bytes.Buffer
	)

	BeforeEach(func() {
		stdin = bytes.NewBuffer([]byte{})
		logger = &fakes.Logger{}

		vpcStatusChecker = &fakes.VPCStatusChecker{}
		stackManager = &fakes.StackManager{}
		infrastructureManager = &fakes.InfrastructureManager{}
		boshDeleter = &fakes.BOSHDeleter{}
		keyPairDeleter = &fakes.KeyPairDeleter{}
		certificateDeleter = &fakes.CertificateDeleter{}
		stringGenerator = &fakes.StringGenerator{}
		awsCredentialValidator = &fakes.AWSCredentialValidator{}
		stateStore = &fakes.StateStore{}

		destroy = commands.NewDestroy(awsCredentialValidator, logger, stdin, boshDeleter,
			vpcStatusChecker, stackManager, stringGenerator, infrastructureManager,
			keyPairDeleter, certificateDeleter, stateStore)
	})

	Describe("Execute", func() {
		It("returns an error when aws credential validator fails", func() {
			awsCredentialValidator.ValidateCall.Returns.Error = errors.New("aws credentials validator failed")

			err := destroy.Execute([]string{}, storage.State{})

			Expect(err).To(MatchError("aws credentials validator failed"))
		})

		DescribeTable("prompting the user for confirmation",
			func(response string, proceed bool) {
				fmt.Fprintf(stdin, "%s\n", response)

				err := destroy.Execute([]string{}, storage.State{
					BOSH: storage.BOSH{
						DirectorName: "some-director",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete your infrastructure? This operation cannot be undone!"))

				if proceed {
					Expect(logger.StepCall.Messages).To(ContainElement("destroying BOSH director"))
					Expect(logger.StepCall.Messages).To(ContainElement("destroying AWS stack"))
					Expect(boshDeleter.DeleteCall.CallCount).To(Equal(1))
				} else {
					Expect(logger.StepCall.Receives.Message).To(Equal("exiting"))
					Expect(boshDeleter.DeleteCall.CallCount).To(Equal(0))
				}
			},
			Entry("responding with 'yes'", "yes", true),
			Entry("responding with 'y'", "y", true),
			Entry("responding with 'Yes'", "Yes", true),
			Entry("responding with 'Y'", "Y", true),
			Entry("responding with 'no'", "no", false),
			Entry("responding with 'n'", "n", false),
			Entry("responding with 'No'", "No", false),
			Entry("responding with 'N'", "N", false),
		)

		Context("when the --no-confirm flag is supplied", func() {
			DescribeTable("destroys without prompting the user for confirmation", func(flag string) {
				err := destroy.Execute([]string{flag}, storage.State{
					BOSH: storage.BOSH{
						DirectorName: "some-director",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(boshDeleter.DeleteCall.CallCount).To(Equal(1))
			},
				Entry("--no-confirm", "--no-confirm"),
				Entry("-n", "-n"),
			)
		})

		Describe("destroying the infrastructure", func() {
			var (
				state storage.State
			)

			BeforeEach(func() {
				stdin.Write([]byte("yes\n"))
				state = storage.State{
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-aws-region",
					},
					KeyPair: storage.KeyPair{
						Name:       "some-ec2-key-pair-name",
						PrivateKey: "some-private-key",
						PublicKey:  "some-public-key",
					},
					BOSH: storage.BOSH{
						DirectorUsername: "some-director-username",
						DirectorPassword: "some-director-password",
						State: map[string]interface{}{
							"key": "value",
						},
						Credentials: map[string]string{
							"some-username": "some-password",
						},
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
						Manifest:               "bosh-init-manifest",
					},
					Stack: storage.Stack{
						Name:            "some-stack-name",
						LBType:          "some-lb-type",
						CertificateName: "some-certificate-name",
					},
				}
			})

			It("fails fast if BOSH deployed VMs still exist in the VPC", func() {
				stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
					Name:   "some-stack-name",
					Status: "some-stack-status",
					Outputs: map[string]string{
						"VPCID": "some-vpc-id",
					},
				}
				vpcStatusChecker.ValidateSafeToDeleteCall.Returns.Error = errors.New("vpc some-vpc-id is not safe to delete")

				err := destroy.Execute([]string{}, state)
				Expect(err).To(MatchError("vpc some-vpc-id is not safe to delete"))

				Expect(vpcStatusChecker.ValidateSafeToDeleteCall.Receives.VPCID).To(Equal("some-vpc-id"))
			})

			It("invokes bosh-init delete", func() {
				stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
					Name:   "some-stack-name",
					Status: "some-stack-status",
					Outputs: map[string]string{
						"BOSHSubnet":              "some-subnet-id",
						"BOSHSubnetAZ":            "some-availability-zone",
						"BOSHEIP":                 "some-elastic-ip",
						"BOSHUserAccessKey":       "some-access-key-id",
						"BOSHUserSecretAccessKey": "some-secret-access-key",
						"BOSHSecurityGroup":       "some-security-group",
					},
				}

				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(stackManager.DescribeCall.Receives.StackName).To(Equal("some-stack-name"))

				Expect(boshDeleter.DeleteCall.Receives.BOSHInitManifest).To(Equal("bosh-init-manifest"))
				Expect(boshDeleter.DeleteCall.Receives.BOSHInitState).To(Equal(boshinit.State{"key": "value"}))
				Expect(boshDeleter.DeleteCall.Receives.EC2PrivateKey).To(Equal("some-private-key"))
			})

			It("deletes the stack", func() {
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(infrastructureManager.DeleteCall.Receives.StackName).To(Equal("some-stack-name"))
			})

			It("deletes the certificate", func() {
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(certificateDeleter.DeleteCall.Receives.CertificateName).To(Equal("some-certificate-name"))
				Expect(logger.StepCall.Messages).To(ContainElement("deleting certificate"))
			})

			It("doesn't call delete certificate if there is no certificate to delete", func() {
				state.Stack.CertificateName = ""
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(certificateDeleter.DeleteCall.CallCount).To(Equal(0))
			})

			It("deletes the keypair", func() {
				err := destroy.Execute([]string{}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairDeleter.DeleteCall.Receives.Name).To(Equal("some-ec2-key-pair-name"))
			})

			It("clears the state", func() {
				err := destroy.Execute([]string{}, state)

				Expect(err).NotTo(HaveOccurred())
				Expect(stateStore.SetCall.CallCount).To(Equal(4))
				Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{}))
			})

			Context("reentrance", func() {
				Context("when the stack fails to delete", func() {
					It("removes the bosh properties from state and returns an error", func() {
						infrastructureManager.DeleteCall.Returns.Error = errors.New("failed to delete stack")

						err := destroy.Execute([]string{}, state)
						Expect(err).To(MatchError("failed to delete stack"))

						Expect(stateStore.SetCall.CallCount).To(Equal(1))
						Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
							AWS: storage.AWS{
								AccessKeyID:     "some-access-key-id",
								SecretAccessKey: "some-secret-access-key",
								Region:          "some-aws-region",
							},
							KeyPair: storage.KeyPair{
								Name:       "some-ec2-key-pair-name",
								PrivateKey: "some-private-key",
								PublicKey:  "some-public-key",
							},
							BOSH: storage.BOSH{},
							Stack: storage.Stack{
								Name:            "some-stack-name",
								LBType:          "some-lb-type",
								CertificateName: "some-certificate-name",
							},
						}))
					})
				})

				Context("when there is no bosh to delete", func() {
					It("does not attempt to delete the bosh", func() {
						state.BOSH = storage.BOSH{}
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.PrintlnCall.Receives.Message).To(Equal("no BOSH director, skipping..."))
						Expect(boshDeleter.DeleteCall.CallCount).To(Equal(0))
					})
				})

				Context("when the certificate fails to delete", func() {
					It("removes the stack from the state and returns an error", func() {
						certificateDeleter.DeleteCall.Returns.Error = errors.New("failed to delete certificate")

						err := destroy.Execute([]string{}, state)
						Expect(err).To(MatchError("failed to delete certificate"))

						Expect(stateStore.SetCall.CallCount).To(Equal(2))
						Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
							AWS: storage.AWS{
								AccessKeyID:     "some-access-key-id",
								SecretAccessKey: "some-secret-access-key",
								Region:          "some-aws-region",
							},
							KeyPair: storage.KeyPair{
								Name:       "some-ec2-key-pair-name",
								PrivateKey: "some-private-key",
								PublicKey:  "some-public-key",
							},
							BOSH: storage.BOSH{},
							Stack: storage.Stack{
								Name:            "",
								LBType:          "",
								CertificateName: "some-certificate-name",
							},
						}))
					})
				})

				Context("when the keypair fails to delete", func() {
					It("removes the certificate from the state and returns an error", func() {
						keyPairDeleter.DeleteCall.Returns.Error = errors.New("failed to delete keypair")

						err := destroy.Execute([]string{}, state)
						Expect(err).To(MatchError("failed to delete keypair"))

						Expect(stateStore.SetCall.CallCount).To(Equal(3))
						Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
							AWS: storage.AWS{
								AccessKeyID:     "some-access-key-id",
								SecretAccessKey: "some-secret-access-key",
								Region:          "some-aws-region",
							},
							KeyPair: storage.KeyPair{
								Name:       "some-ec2-key-pair-name",
								PrivateKey: "some-private-key",
								PublicKey:  "some-public-key",
							},
							BOSH: storage.BOSH{},
							Stack: storage.Stack{
								Name:            "",
								LBType:          "",
								CertificateName: "",
							},
						}))
					})
				})

				Context("when there is no stack to delete", func() {
					BeforeEach(func() {
						stackManager.DescribeCall.Returns.Error = cloudformation.StackNotFound
					})

					It("does not validate the vpc", func() {
						state.Stack = storage.Stack{}
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(vpcStatusChecker.ValidateSafeToDeleteCall.CallCount).To(Equal(0))
					})

					It("does not attempt to delete the stack", func() {
						state.Stack = storage.Stack{}
						err := destroy.Execute([]string{}, state)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.PrintlnCall.Receives.Message).To(Equal("no AWS stack, skipping..."))
						Expect(infrastructureManager.DeleteCall.CallCount).To(Equal(0))
					})
				})
			})
		})

		Context("failure cases", func() {
			BeforeEach(func() {
				stdin.Write([]byte("yes\n"))
			})

			Context("when an invalid command line flag is supplied", func() {
				It("returns an error", func() {
					err := destroy.Execute([]string{"--invalid-flag"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -invalid-flag"))
				})
			})

			Context("when the bosh delete fails", func() {
				It("returns an error", func() {
					boshDeleter.DeleteCall.Returns.Error = errors.New("BOSH Delete Failed")

					err := destroy.Execute([]string{}, storage.State{
						BOSH: storage.BOSH{
							DirectorName: "some-director",
						},
					})
					Expect(err).To(MatchError("BOSH Delete Failed"))
				})
			})

			Context("when the stack manager cannot describe the stack", func() {
				It("returns an error", func() {
					stackManager.DescribeCall.Returns.Error = errors.New("cannot describe stack")

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("cannot describe stack"))
				})
			})

			Context("when failing to delete the stack", func() {
				It("returns an error", func() {
					infrastructureManager.DeleteCall.Returns.Error = errors.New("failed to delete stack")

					err := destroy.Execute([]string{}, storage.State{
						Stack: storage.Stack{
							Name: "some-stack-name",
						},
					})
					Expect(err).To(MatchError("failed to delete stack"))
				})
			})

			Context("when the keypair cannot be deleted", func() {
				It("returns an error", func() {
					keyPairDeleter.DeleteCall.Returns.Error = errors.New("failed to delete keypair")

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("failed to delete keypair"))
				})
			})

			Context("when the certificate cannot be deleted", func() {
				It("returns an error", func() {
					certificateDeleter.DeleteCall.Returns.Error = errors.New("failed to delete certificate")

					err := destroy.Execute([]string{}, storage.State{
						Stack: storage.Stack{
							CertificateName: "some-certificate",
						}})
					Expect(err).To(MatchError("failed to delete certificate"))
				})
			})

			Context("when state store fails to set the state before destroying infrastructure", func() {
				It("returns an error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set state")}}

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("failed to set state"))
				})
			})

			Context("when state store fails to set the state before destroying certificate", func() {
				It("returns an error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("failed to set state"))
				})
			})

			Context("when state store fails to set the state before destroying keypair", func() {
				It("returns an error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("failed to set state")}}

					err := destroy.Execute([]string{}, storage.State{
						Stack: storage.Stack{
							CertificateName: "some-certificate-name",
						},
					})
					Expect(err).To(MatchError("failed to set state"))
				})
			})

			Context("when the state fails to be set", func() {
				It("return an error", func() {
					stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("failed to set state")}}

					err := destroy.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("failed to set state"))
				})
			})
		})
	})
})
