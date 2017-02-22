package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSUp", func() {
	Describe("Execute", func() {
		var (
			command                   commands.AWSUp
			boshManager               *fakes.BOSHManager
			infrastructureManager     *fakes.InfrastructureManager
			keyPairSynchronizer       *fakes.KeyPairSynchronizer
			cloudConfigurator         *fakes.BoshCloudConfigurator
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			certificateDescriber      *fakes.CertificateDescriber
			credentialValidator       *fakes.CredentialValidator
			cloudConfigManager        *fakes.CloudConfigManager
			boshClientProvider        *fakes.BOSHClientProvider
			boshClient                *fakes.BOSHClient
			stateStore                *fakes.StateStore
			clientProvider            *fakes.ClientProvider
			envIDManager              *fakes.EnvIDManager
		)

		BeforeEach(func() {
			keyPairSynchronizer = &fakes.KeyPairSynchronizer{}
			keyPairSynchronizer.SyncCall.Returns.KeyPair = ec2.KeyPair{
				Name:       "keypair-bbl-lake-time:stamp",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}

			infrastructureManager = &fakes.InfrastructureManager{}
			infrastructureManager.CreateCall.Returns.Stack = cloudformation.Stack{
				Name: "bbl-aws-some-random-string",
				Outputs: map[string]string{
					"BOSHSubnet":              "some-bosh-subnet",
					"BOSHSubnetAZ":            "some-bosh-subnet-az",
					"BOSHEIP":                 "some-bosh-elastic-ip",
					"BOSHURL":                 "some-bosh-url",
					"BOSHUserAccessKey":       "some-bosh-user-access-key",
					"BOSHUserSecretAccessKey": "some-bosh-user-secret-access-key",
					"BOSHSecurityGroup":       "some-bosh-security-group",
				},
			}

			boshManager = &fakes.BOSHManager{}
			boshManager.CreateCall.Returns.State = storage.State{
				BOSH: storage.BOSH{
					DirectorName:           "bosh-bbl-lake-time:stamp",
					DirectorUsername:       "admin",
					DirectorPassword:       "some-admin-password",
					DirectorAddress:        "some-director-address",
					DirectorSSLCA:          "some-ca",
					DirectorSSLCertificate: "some-certificate",
					DirectorSSLPrivateKey:  "some-private-key",
					State: map[string]interface{}{
						"new-key": "new-value",
					},
					Variables: variablesYAML,
					Manifest:  "some-bosh-manifest",
				},
			}

			cloudConfigurator = &fakes.BoshCloudConfigurator{}
			cloudConfigManager = &fakes.CloudConfigManager{}

			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}

			certificateDescriber = &fakes.CertificateDescriber{}

			credentialValidator = &fakes.CredentialValidator{}

			boshClient = &fakes.BOSHClient{}
			boshClientProvider = &fakes.BOSHClientProvider{}

			boshClientProvider.ClientCall.Returns.Client = boshClient

			stateStore = &fakes.StateStore{}
			clientProvider = &fakes.ClientProvider{}

			envIDManager = &fakes.EnvIDManager{}
			envIDManager.SyncCall.Returns.EnvID = "bbl-lake-time-stamp"

			command = commands.NewAWSUp(
				credentialValidator, infrastructureManager, keyPairSynchronizer, boshManager,
				cloudConfigurator, availabilityZoneRetriever, certificateDescriber,
				cloudConfigManager, boshClientProvider, stateStore,
				clientProvider, envIDManager,
			)
		})

		It("returns an error when aws credential validator fails", func() {
			credentialValidator.ValidateAWSCall.Returns.Error = errors.New("failed to validate aws credentials")
			err := command.Execute(commands.AWSUpConfig{}, storage.State{})
			Expect(err).To(MatchError("failed to validate aws credentials"))
		})

		It("retrieves a client with the provided credentials", func() {
			err := command.Execute(commands.AWSUpConfig{
				AccessKeyID:     "new-aws-access-key-id",
				SecretAccessKey: "new-aws-secret-access-key",
				Region:          "new-aws-region",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(clientProvider.SetConfigCall.CallCount).To(Equal(1))
			Expect(clientProvider.SetConfigCall.Receives.Config).To(Equal(aws.Config{
				Region:          "new-aws-region",
				SecretAccessKey: "new-aws-secret-access-key",
				AccessKeyID:     "new-aws-access-key-id",
			}))
			Expect(credentialValidator.ValidateAWSCall.CallCount).To(Equal(0))
		})

		It("calls the env id manager and saves the env id", func() {
			err := command.Execute(commands.AWSUpConfig{
				AccessKeyID:     "new-aws-access-key-id",
				SecretAccessKey: "new-aws-secret-access-key",
				Region:          "new-aws-region",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.Receives.State.EnvID).To(Equal("bbl-lake-time-stamp"))
		})

		Context("when a name is passed in for env-id", func() {
			It("passes that name in for the env id manager to use", func() {
				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "new-aws-access-key-id",
					SecretAccessKey: "new-aws-secret-access-key",
					Region:          "new-aws-region",
					Name:            "some-other-env-id",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
				Expect(envIDManager.SyncCall.Receives.Name).To(Equal("some-other-env-id"))
			})
		})

		It("syncs the keypair", func() {
			err := command.Execute(commands.AWSUpConfig{}, storage.State{
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				EnvID: "bbl-lake-time-stamp",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(clientProvider.SetConfigCall.CallCount).To(Equal(0))
			Expect(credentialValidator.ValidateAWSCall.CallCount).To(Equal(1))

			Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}))

			actualState := stateStore.SetCall.Receives.State
			Expect(actualState.KeyPair).To(Equal(storage.KeyPair{
				Name:       "some-keypair-name",
				PublicKey:  "some-public-key",
				PrivateKey: "some-private-key",
			}))
		})

		It("creates/updates the stack with the given name", func() {
			incomingState := storage.State{
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
				EnvID: "bbl-lake-time-stamp",
			}

			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}

			err := command.Execute(commands.AWSUpConfig{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(infrastructureManager.CreateCall.Receives.StackName).To(Equal("stack-bbl-lake-time-stamp"))
			Expect(infrastructureManager.CreateCall.Receives.KeyPairName).To(Equal("keypair-bbl-lake-time-stamp"))
			Expect(infrastructureManager.CreateCall.Receives.AZs).To(Equal([]string{"some-retrieved-az"}))
			Expect(infrastructureManager.CreateCall.Receives.EnvID).To(Equal("bbl-lake-time-stamp"))
			Expect(infrastructureManager.CreateCall.Returns.Error).To(BeNil())
		})

		It("deploys bosh", func() {
			incomingState := storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				Stack: storage.Stack{
					Name: "some-stack-name",
				},
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				EnvID: "bbl-lake-time-stamp",
			}

			err := command.Execute(commands.AWSUpConfig{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshManager.CreateCall.Receives.State).To(Equal(incomingState))
		})

		Context("when ops file are passed in via --ops-file flag", func() {
			It("passes the ops file contents to the bosh manager", func() {
				opsFile, err := ioutil.TempFile("", "ops-file")
				Expect(err).NotTo(HaveOccurred())

				opsFilePath := opsFile.Name()
				opsFileContents := "some-ops-file-contents"
				err = ioutil.WriteFile(opsFilePath, []byte(opsFileContents), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
					OpsFilePath:     opsFilePath,
				}, storage.State{
					EnvID: "bbl-lake-time-stamp",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateCall.Receives.OpsFile).To(Equal([]byte("some-ops-file-contents")))
			})
		})

		Context("when bosh az is provided via --aws-bosh-az flag", func() {
			It("passes the bosh az to the infrastructure manager", func() {
				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
					BOSHAZ:          "some-bosh-az",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(infrastructureManager.CreateCall.Receives.BOSHAZ).To(Equal("some-bosh-az"))
			})

			Context("when a stack exists and the aws-bosh-az is provided and different", func() {
				It("returns an error message", func() {
					err := command.Execute(commands.AWSUpConfig{
						AccessKeyID:     "some-aws-access-key-id",
						SecretAccessKey: "some-aws-secret-access-key",
						Region:          "some-aws-region",
						BOSHAZ:          "other-bosh-az",
					}, storage.State{
						Stack: storage.Stack{
							Name:   "some-stack",
							BOSHAZ: "some-bosh-az",
						},
					})
					Expect(err).To(MatchError("The --aws-bosh-az cannot be changed for existing environments."))
				})
			})
		})

		Context("when there is an lb", func() {
			It("attaches the lb certificate to the lb type in cloudformation", func() {
				certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
					Name: "some-certificate-name",
					ARN:  "some-certificate-arn",
					Body: "some-certificate-body",
				}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{
					Stack: storage.Stack{
						Name:            "some-stack-name",
						LBType:          "concourse",
						CertificateName: "some-certificate-name",
					},
					EnvID: "bbl-lake-time-stamp",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(infrastructureManager.CreateCall.Receives.LBCertificateARN).To(Equal("some-certificate-arn"))
			})
		})

		Describe("cloud configurator", func() {
			BeforeEach(func() {
				infrastructureManager.CreateCall.Stub = func(keyPairName string, azs []string, stackName, lbType, envID string) (cloudformation.Stack, error) {
					stack := cloudformation.Stack{
						Name: "bbl-aws-some-random-string",
						Outputs: map[string]string{
							"BOSHSubnet":              "some-bosh-subnet",
							"BOSHSubnetAZ":            "some-bosh-subnet-az",
							"BOSHEIP":                 "some-bosh-elastic-ip",
							"BOSHURL":                 "some-director-address",
							"BOSHUserAccessKey":       "some-bosh-user-access-key",
							"BOSHUserSecretAccessKey": "some-bosh-user-secret-access-key",
							"BOSHSecurityGroup":       "some-bosh-security-group",
						},
					}

					switch lbType {
					case "concourse":
						stack.Outputs["ConcourseLoadBalancer"] = "some-lb-name"
						stack.Outputs["ConcourseLoadBalancerURL"] = "some-lb-url"
					case "cf":
						stack.Outputs["RouterLB"] = "some-router-lb-name"
						stack.Outputs["RouterLBURL"] = "some-router-lb-url"
						stack.Outputs["SSHProxyLB"] = "some-ssh-proxy-lb-name"
						stack.Outputs["SSHProxyLBURL"] = "some-ssh-proxy-lb-url"
					default:
					}

					return stack, nil
				}
			})

			It("upload the cloud config to the director", func() {
				cloudConfigInput := bosh.CloudConfigInput{
					AZs: []string{"az1", "az2", "az3"},
				}

				cloudConfigurator.ConfigureCall.Returns.CloudConfigInput = cloudConfigInput
				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("admin"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-admin-password"))

				Expect(cloudConfigManager.UpdateCall.Receives.CloudConfigInput).To(Equal(cloudConfigInput))
				Expect(cloudConfigManager.UpdateCall.Receives.BOSHClient).To(Equal(boshClient))
			})

			Context("when no load balancer has been requested", func() {
				It("generates a cloud config", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}

					err := command.Execute(commands.AWSUpConfig{}, storage.State{})

					Expect(err).NotTo(HaveOccurred())
					Expect(cloudConfigurator.ConfigureCall.CallCount).To(Equal(1))
					Expect(cloudConfigurator.ConfigureCall.Receives.Stack).To(Equal(cloudformation.Stack{
						Name: "bbl-aws-some-random-string",
						Outputs: map[string]string{
							"BOSHSecurityGroup":       "some-bosh-security-group",
							"BOSHSubnet":              "some-bosh-subnet",
							"BOSHSubnetAZ":            "some-bosh-subnet-az",
							"BOSHEIP":                 "some-bosh-elastic-ip",
							"BOSHURL":                 "some-director-address",
							"BOSHUserAccessKey":       "some-bosh-user-access-key",
							"BOSHUserSecretAccessKey": "some-bosh-user-secret-access-key",
						},
					}))
					Expect(cloudConfigurator.ConfigureCall.Receives.AZs).To(ConsistOf("some-retrieved-az"))
					Expect(certificateDescriber.DescribeCall.CallCount).To(Equal(0))
				})
			})

			Context("when the load balancer type is concourse", func() {
				It("generates a cloud config", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}
					certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
						Name: "some-certificate-name",
						ARN:  "some-certificate-arn",
						Body: "some-certificate-body",
					}

					err := command.Execute(commands.AWSUpConfig{}, storage.State{
						Stack: storage.Stack{
							LBType:          "concourse",
							CertificateName: "some-certificate-name",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudConfigurator.ConfigureCall.CallCount).To(Equal(1))
					Expect(cloudConfigurator.ConfigureCall.Receives.Stack).To(Equal(cloudformation.Stack{
						Name: "bbl-aws-some-random-string",
						Outputs: map[string]string{
							"BOSHSecurityGroup":        "some-bosh-security-group",
							"BOSHSubnet":               "some-bosh-subnet",
							"BOSHSubnetAZ":             "some-bosh-subnet-az",
							"BOSHEIP":                  "some-bosh-elastic-ip",
							"BOSHURL":                  "some-director-address",
							"BOSHUserAccessKey":        "some-bosh-user-access-key",
							"BOSHUserSecretAccessKey":  "some-bosh-user-secret-access-key",
							"ConcourseLoadBalancerURL": "some-lb-url",
							"ConcourseLoadBalancer":    "some-lb-name",
						},
					}))

					Expect(cloudConfigurator.ConfigureCall.Receives.AZs).To(ConsistOf("some-retrieved-az"))
				})
			})

			Context("when the load balancer type is cf", func() {
				It("generates a cloud config", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}
					certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
						Name: "some-certificate-name",
						ARN:  "some-certificate-arn",
						Body: "some-certificate-body",
					}

					err := command.Execute(commands.AWSUpConfig{}, storage.State{
						Stack: storage.Stack{
							LBType:          "cf",
							CertificateName: "some-certificate-name",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudConfigurator.ConfigureCall.CallCount).To(Equal(1))
					Expect(cloudConfigurator.ConfigureCall.Receives.Stack).To(Equal(cloudformation.Stack{
						Name: "bbl-aws-some-random-string",
						Outputs: map[string]string{
							"BOSHSecurityGroup":       "some-bosh-security-group",
							"BOSHSubnet":              "some-bosh-subnet",
							"BOSHSubnetAZ":            "some-bosh-subnet-az",
							"BOSHEIP":                 "some-bosh-elastic-ip",
							"BOSHURL":                 "some-director-address",
							"BOSHUserAccessKey":       "some-bosh-user-access-key",
							"BOSHUserSecretAccessKey": "some-bosh-user-secret-access-key",
							"RouterLBURL":             "some-router-lb-url",
							"RouterLB":                "some-router-lb-name",
							"SSHProxyLBURL":           "some-ssh-proxy-lb-url",
							"SSHProxyLB":              "some-ssh-proxy-lb-name",
						},
					}))

					Expect(cloudConfigurator.ConfigureCall.Receives.AZs).To(ConsistOf("some-retrieved-az"))
				})
			})
		})

		Describe("reentrant", func() {
			Context("when the key pair fails to sync", func() {
				It("saves the keypair name and returns an error", func() {
					keyPairSynchronizer.SyncCall.Returns.Error = errors.New("error syncing key pair")

					err := command.Execute(commands.AWSUpConfig{}, storage.State{})
					Expect(err).To(MatchError("error syncing key pair"))
					Expect(stateStore.SetCall.CallCount).To(Equal(1))
					Expect(stateStore.SetCall.Receives.State.KeyPair.Name).To(Equal("keypair-bbl-lake-time-stamp"))
				})
			})

			Context("when the availability zone retriever fails", func() {
				It("saves the public/private key and returns an error", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("availability zone retrieve failed")

					err := command.Execute(commands.AWSUpConfig{}, storage.State{
						EnvID: "bbl-lake-time:stamp",
					})
					Expect(err).To(MatchError("availability zone retrieve failed"))
					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives.State.KeyPair.PrivateKey).To(Equal("some-private-key"))
					Expect(stateStore.SetCall.Receives.State.KeyPair.PublicKey).To(Equal("some-public-key"))
				})
			})

			Context("when the cloudformation fails", func() {
				It("saves the stack name and bosh az and returns an error", func() {
					infrastructureManager.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

					err := command.Execute(commands.AWSUpConfig{
						BOSHAZ: "some-bosh-az",
					}, storage.State{
						EnvID: "bbl-lake-time-stamp",
					})
					Expect(err).To(MatchError("infrastructure creation failed"))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives.State.Stack.Name).To(Equal("stack-bbl-lake-time-stamp"))
					Expect(stateStore.SetCall.Receives.State.Stack.BOSHAZ).To(Equal("some-bosh-az"))
				})

				It("saves the private/public key and returns an error", func() {
					infrastructureManager.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

					err := command.Execute(commands.AWSUpConfig{}, storage.State{})
					Expect(err).To(MatchError("infrastructure creation failed"))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives.State.KeyPair.PrivateKey).To(Equal("some-private-key"))
					Expect(stateStore.SetCall.Receives.State.KeyPair.PublicKey).To(Equal("some-public-key"))
				})
			})

			Context("when the bosh cloud config fails", func() {
				It("saves the bosh properties and returns an error", func() {
					cloudConfigManager.UpdateCall.Returns.Error = errors.New("cloud config update failed")

					err := command.Execute(commands.AWSUpConfig{}, storage.State{
						EnvID: "bbl-lake-time-stamp",
					})
					Expect(err).To(MatchError("cloud config update failed"))
					Expect(stateStore.SetCall.CallCount).To(Equal(4))
					Expect(stateStore.SetCall.Receives.State.BOSH).To(Equal(storage.BOSH{
						DirectorName:           "bosh-bbl-lake-time:stamp",
						DirectorUsername:       "admin",
						DirectorPassword:       "some-admin-password",
						DirectorAddress:        "some-director-address",
						DirectorSSLCA:          "some-ca",
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
						State: map[string]interface{}{
							"new-key": "new-value",
						},
						Variables: variablesYAML,
						Manifest:  "some-bosh-manifest",
					}))
				})
			})
		})

		Describe("state manipulation", func() {
			Context("iaas", func() {
				It("writes iaas aws to state", func() {
					err := command.Execute(commands.AWSUpConfig{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(stateStore.SetCall.Receives.State.IAAS).To(Equal("aws"))
				})
			})

			Context("aws credentials", func() {
				Context("when the credentials do not exist", func() {
					It("saves the credentials", func() {
						err := command.Execute(commands.AWSUpConfig{
							AccessKeyID:     "some-aws-access-key-id",
							SecretAccessKey: "some-aws-secret-access-key",
							Region:          "some-aws-region",
						}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives.State.AWS).To(Equal(storage.AWS{
							AccessKeyID:     "some-aws-access-key-id",
							SecretAccessKey: "some-aws-secret-access-key",
							Region:          "some-aws-region",
						}))
					})
				})
				Context("when the credentials do exist", func() {
					It("overrides the credentials when they're passed in", func() {
						err := command.Execute(commands.AWSUpConfig{
							AccessKeyID:     "new-aws-access-key-id",
							SecretAccessKey: "new-aws-secret-access-key",
							Region:          "new-aws-region",
						}, storage.State{
							AWS: storage.AWS{
								AccessKeyID:     "old-aws-access-key-id",
								SecretAccessKey: "old-aws-secret-access-key",
								Region:          "old-aws-region",
							},
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives.State.AWS).To(Equal(storage.AWS{
							AccessKeyID:     "new-aws-access-key-id",
							SecretAccessKey: "new-aws-secret-access-key",
							Region:          "new-aws-region",
						}))
					})

					It("does not override the credentials when they're not passed in", func() {
						err := command.Execute(commands.AWSUpConfig{}, storage.State{
							AWS: storage.AWS{
								AccessKeyID:     "aws-access-key-id",
								SecretAccessKey: "aws-secret-access-key",
								Region:          "aws-region",
							},
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives.State.AWS).To(Equal(storage.AWS{
							AccessKeyID:     "aws-access-key-id",
							SecretAccessKey: "aws-secret-access-key",
							Region:          "aws-region",
						}))
					})
				})
			})

			Context("aws keypair", func() {
				Context("when the keypair exists", func() {
					It("saves the given state unmodified", func() {
						keyPairSynchronizer.SyncCall.Returns.KeyPair = ec2.KeyPair{
							Name:       "some-existing-keypair",
							PrivateKey: "some-private-key",
							PublicKey:  "some-public-key",
						}

						incomingState := storage.State{
							KeyPair: storage.KeyPair{
								Name:       "some-existing-keypair",
								PrivateKey: "some-private-key",
								PublicKey:  "some-public-key",
							},
						}

						err := command.Execute(commands.AWSUpConfig{}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
							Name:       "some-existing-keypair",
							PrivateKey: "some-private-key",
							PublicKey:  "some-public-key",
						}))

						Expect(stateStore.SetCall.Receives.State.KeyPair).To(Equal(incomingState.KeyPair))
					})
				})

				Context("when the keypair doesn't exist", func() {
					It("saves the state with a new key pair", func() {
						keyPairSynchronizer.SyncCall.Returns.KeyPair = ec2.KeyPair{
							Name:       "keypair-bbl-lake-time:stamp",
							PrivateKey: "some-private-key",
							PublicKey:  "some-public-key",
						}

						envIDManager.SyncCall.Returns.EnvID = "bbl-lake-time:stamp"

						err := command.Execute(commands.AWSUpConfig{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
							Name: "keypair-bbl-lake-time:stamp",
						}))

						actualState := stateStore.SetCall.Receives.State
						Expect(actualState.KeyPair).To(Equal(storage.KeyPair{
							Name:       "keypair-bbl-lake-time:stamp",
							PrivateKey: "some-private-key",
							PublicKey:  "some-public-key",
						}))
					})
				})
			})

			Context("cloudformation", func() {
				Context("when the stack name doesn't exist", func() {
					It("populates a new stack name", func() {
						incomingState := storage.State{
							EnvID: "bbl-lake-time-stamp",
						}
						err := command.Execute(commands.AWSUpConfig{}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.Stack.Name).To(Equal("stack-bbl-lake-time-stamp"))
					})
				})

				Context("when the stack name exists", func() {
					It("does not modify the state", func() {
						incomingState := storage.State{
							Stack: storage.Stack{
								Name: "some-other-stack-name",
							},
						}
						err := command.Execute(commands.AWSUpConfig{}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.Stack.Name).To(Equal("some-other-stack-name"))
					})
				})
			})

			Describe("bosh", func() {
				BeforeEach(func() {
					infrastructureManager.ExistsCall.Returns.Exists = true
				})

				Context("bosh state", func() {
					It("writes the bosh state", func() {
						err := command.Execute(commands.AWSUpConfig{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives.State.BOSH).To(Equal(storage.BOSH{
							DirectorName:           "bosh-bbl-lake-time:stamp",
							DirectorUsername:       "admin",
							DirectorPassword:       "some-admin-password",
							DirectorAddress:        "some-director-address",
							DirectorSSLCA:          "some-ca",
							DirectorSSLCertificate: "some-certificate",
							DirectorSSLPrivateKey:  "some-private-key",
							State: map[string]interface{}{
								"new-key": "new-value",
							},
							Variables: variablesYAML,
							Manifest:  "some-bosh-manifest",
						}))
					})
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when the env id manager fails", func() {
				envIDManager.SyncCall.Returns.Error = errors.New("env id manager failed")

				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
				}, storage.State{})
				Expect(err).To(MatchError("env id manager failed"))

			})

			It("returns an error when saving the state fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{
					{
						Error: errors.New("saving the state failed"),
					},
				}
				err := command.Execute(commands.AWSUpConfig{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
				}, storage.State{})
				Expect(err).To(MatchError("saving the state failed"))
			})

			It("returns an error when the certificate cannot be described", func() {
				certificateDescriber.DescribeCall.Returns.Error = errors.New("failed to describe")
				err := command.Execute(commands.AWSUpConfig{}, storage.State{
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("failed to describe"))
			})

			It("returns an error when the cloud config cannot be uploaded", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update")
				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to update"))
			})

			It("returns an error when the BOSH state exists, but the cloudformation stack does not", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false

				err := command.Execute(commands.AWSUpConfig{}, storage.State{
					AWS: storage.AWS{
						Region: "some-aws-region",
					},
					BOSH: storage.BOSH{
						DirectorAddress: "some-director-address",
					},
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
				})

				Expect(infrastructureManager.ExistsCall.Receives.StackName).To(Equal("some-stack-name"))

				Expect(err).To(MatchError("Found BOSH data in state directory, " +
					"but Cloud Formation stack \"some-stack-name\" cannot be found for region \"some-aws-region\" and given " +
					"AWS credentials. bbl cannot safely proceed. Open an issue on GitHub at " +
					"https://github.com/cloudfoundry/bosh-bootloader/issues/new if you need assistance."))

				Expect(infrastructureManager.CreateCall.CallCount).To(Equal(0))
			})

			It("returns an error when checking if the infrastructure exists fails", func() {
				infrastructureManager.ExistsCall.Returns.Error = errors.New("error checking if stack exists")

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("error checking if stack exists"))
			})

			It("returns an error when infrastructure cannot be created", func() {
				infrastructureManager.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("infrastructure creation failed"))
			})

			It("returns an error when the ops file cannot be read", func() {
				err := command.Execute(commands.AWSUpConfig{
					OpsFilePath: "some/fake/path",
				}, storage.State{})
				Expect(err).To(MatchError("open some/fake/path: no such file or directory"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshManager.CreateCall.Returns.Error = errors.New("cannot deploy bosh")

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("cannot deploy bosh"))
			})

			It("returns an error when availability zones cannot be retrieved", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("availability zone could not be retrieved")

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("availability zone could not be retrieved"))
			})

			It("returns an error when state store fails to set the state before syncing the keypair", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set state")}}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before retrieving availability zones", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before creating the stack", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("failed to set state")}}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before updating the cloud config", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("failed to set state")}}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before method exits", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {}, {errors.New("failed to set state")}}

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when only some of the AWS parameters are provided", func() {
				err := command.Execute(commands.AWSUpConfig{AccessKeyID: "some-key-id", Region: "some-region"}, storage.State{})
				Expect(err).To(MatchError("AWS secret access key must be provided"))
			})

			It("returns an error when no AWS parameters are provided and the bbl-state AWS values are empty", func() {
				credentialValidator.ValidateAWSCall.Returns.Error = errors.New("AWS secret access key must be provided")

				err := command.Execute(commands.AWSUpConfig{}, storage.State{})
				Expect(err).To(MatchError("AWS secret access key must be provided"))
			})
		})
	})
})
