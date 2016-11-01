package commands_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/boshinit"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/ssl"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Up", func() {
	Describe("Execute", func() {
		var (
			command                   commands.Up
			boshDeployer              *fakes.BOSHDeployer
			infrastructureManager     *fakes.InfrastructureManager
			keyPairSynchronizer       *fakes.KeyPairSynchronizer
			stringGenerator           *fakes.StringGenerator
			cloudConfigurator         *fakes.BoshCloudConfigurator
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			certificateDescriber      *fakes.CertificateDescriber
			awsCredentialValidator    *fakes.AWSCredentialValidator
			cloudConfigManager        *fakes.CloudConfigManager
			boshClientProvider        *fakes.BOSHClientProvider
			boshClient                *fakes.BOSHClient
			envIDGenerator            *fakes.EnvIDGenerator
			boshInitCredentials       map[string]string
			stateStore                *fakes.StateStore
			clientProvider            *fakes.ClientProvider
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

			boshDeployer = &fakes.BOSHDeployer{}
			boshDeployer.DeployCall.Returns.Output = boshinit.DeployOutput{
				DirectorSSLKeyPair: ssl.KeyPair{
					CA:          []byte("updated-ca"),
					Certificate: []byte("updated-certificate"),
					PrivateKey:  []byte("updated-private-key"),
				},
				BOSHInitState: boshinit.State{
					"updated-key": "updated-value",
				},
				BOSHInitManifest: "name: bosh",
			}

			stringGenerator = &fakes.StringGenerator{}
			stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
				return fmt.Sprintf("%s%s", prefix, "some-random-string"), nil
			}

			cloudConfigurator = &fakes.BoshCloudConfigurator{}
			cloudConfigManager = &fakes.CloudConfigManager{}

			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}

			certificateDescriber = &fakes.CertificateDescriber{}

			awsCredentialValidator = &fakes.AWSCredentialValidator{}

			boshClient = &fakes.BOSHClient{}
			boshClientProvider = &fakes.BOSHClientProvider{}

			boshClientProvider.ClientCall.Returns.Client = boshClient

			envIDGenerator = &fakes.EnvIDGenerator{}
			envIDGenerator.GenerateCall.Returns.EnvID = "bbl-lake-time:stamp"

			stateStore = &fakes.StateStore{}
			clientProvider = &fakes.ClientProvider{}

			command = commands.NewUp(
				awsCredentialValidator, infrastructureManager, keyPairSynchronizer, boshDeployer,
				stringGenerator, cloudConfigurator, availabilityZoneRetriever, certificateDescriber,
				cloudConfigManager, boshClientProvider, envIDGenerator, stateStore,
				clientProvider,
			)

			boshInitCredentials = map[string]string{
				"mbusUsername":              "some-mbus-username",
				"natsUsername":              "some-nats-username",
				"postgresUsername":          "some-postgres-username",
				"registryUsername":          "some-registry-username",
				"blobstoreDirectorUsername": "some-blobstore-director-username",
				"blobstoreAgentUsername":    "some-blobstore-agent-username",
				"hmUsername":                "some-hm-username",
				"mbusPassword":              "some-mbus-password",
				"natsPassword":              "some-nats-password",
				"postgresPassword":          "some-postgres-password",
				"registryPassword":          "some-registry-password",
				"blobstoreDirectorPassword": "some-blobstore-director-password",
				"blobstoreAgentPassword":    "some-blobstore-agent-password",
				"hmPassword":                "some-hm-password",
			}
		})

		It("returns an error when aws credential validator fails", func() {
			awsCredentialValidator.ValidateCall.Returns.Error = errors.New("failed to validate aws credentials")
			err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
			Expect(err).To(MatchError("failed to validate aws credentials"))
		})

		Context("when AWS creds are provided through environment variables", func() {
			BeforeEach(func() {
				os.Setenv("BBL_AWS_ACCESS_KEY_ID", "some-access-key")
				os.Setenv("BBL_AWS_SECRET_ACCESS_KEY", "some-access-secret")
				os.Setenv("BBL_AWS_REGION", "some-region")
			})
			AfterEach(func() {
				os.Setenv("BBL_AWS_ACCESS_KEY_ID", "")
				os.Setenv("BBL_AWS_SECRET_ACCESS_KEY", "")
				os.Setenv("BBL_AWS_REGION", "")
			})

			It("honors the environment variables to fetch the AWS creds", func() {
				err := command.Execute([]string{"--iaas", "aws"}, storage.State{
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
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(clientProvider.SetConfigCall.CallCount).To(Equal(1))
				Expect(clientProvider.SetConfigCall.Receives.Config).To(Equal(aws.Config{
					Region:          "some-region",
					SecretAccessKey: "some-access-secret",
					AccessKeyID:     "some-access-key",
				}))
				Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(0))
			})

			It("honors missing creds passed as arguments", func() {
				os.Setenv("BBL_AWS_ACCESS_KEY_ID", "")
				err := command.Execute([]string{
					"--iaas", "aws",
					"--aws-access-key-id", "access-key-from-arguments",
				}, storage.State{
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
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(clientProvider.SetConfigCall.CallCount).To(Equal(1))
				Expect(clientProvider.SetConfigCall.Receives.Config).To(Equal(aws.Config{
					Region:          "some-region",
					SecretAccessKey: "some-access-secret",
					AccessKeyID:     "access-key-from-arguments",
				}))
				Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(0))
			})
		})

		It("honors the cli flags", func() {
			err := command.Execute([]string{
				"--iaas", "aws",
				"--aws-access-key-id", "new-aws-access-key-id",
				"--aws-secret-access-key", "new-aws-secret-access-key",
				"--aws-region", "new-aws-region",
			}, storage.State{
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
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(clientProvider.SetConfigCall.CallCount).To(Equal(1))
			Expect(clientProvider.SetConfigCall.Receives.Config).To(Equal(aws.Config{
				Region:          "new-aws-region",
				SecretAccessKey: "new-aws-secret-access-key",
				AccessKeyID:     "new-aws-access-key-id",
			}))
			Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(0))
		})

		It("syncs the keypair", func() {
			err := command.Execute([]string{"--iaas", "aws"}, storage.State{
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
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(clientProvider.SetConfigCall.CallCount).To(Equal(0))
			Expect(awsCredentialValidator.ValidateCall.CallCount).To(Equal(1))

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

		It("generates an bbl-env-id", func() {
			incomingState := storage.State{
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
			}

			err := command.Execute([]string{"--iaas", "aws"}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(envIDGenerator.GenerateCall.CallCount).To(Equal(1))
		})

		It("creates/updates the stack with the given name", func() {
			incomingState := storage.State{
				AWS: storage.AWS{
					Region:          "some-aws-region",
					SecretAccessKey: "some-secret-access-key",
					AccessKeyID:     "some-access-key-id",
				},
			}

			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}

			err := command.Execute([]string{"--iaas", "aws"}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(infrastructureManager.CreateCall.Receives.StackName).To(Equal("stack-bbl-lake-time-stamp"))
			Expect(infrastructureManager.CreateCall.Receives.KeyPairName).To(Equal("keypair-bbl-lake-time:stamp"))
			Expect(infrastructureManager.CreateCall.Receives.NumberOfAvailabilityZones).To(Equal(1))
			Expect(infrastructureManager.CreateCall.Receives.EnvID).To(Equal("bbl-lake-time:stamp"))
			Expect(infrastructureManager.CreateCall.Returns.Error).To(BeNil())
		})

		It("deploys bosh", func() {
			infrastructureManager.ExistsCall.Returns.Exists = true

			incomingState := storage.State{
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}

			err := command.Execute([]string{"--iaas", "aws"}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshDeployer.DeployCall.Receives.Input).To(Equal(boshinit.DeployInput{
				DirectorName:     "bosh-bbl-lake-time:stamp",
				DirectorUsername: "user-some-random-string",
				DirectorPassword: "p-some-random-string",
				State:            map[string]interface{}{},
				InfrastructureConfiguration: boshinit.InfrastructureConfiguration{
					AWSRegion:        "some-aws-region",
					SubnetID:         "some-bosh-subnet",
					AvailabilityZone: "some-bosh-subnet-az",
					ElasticIP:        "some-bosh-elastic-ip",
					AccessKeyID:      "some-bosh-user-access-key",
					SecretAccessKey:  "some-bosh-user-secret-access-key",
					SecurityGroup:    "some-bosh-security-group",
				},
				SSLKeyPair: ssl.KeyPair{},
				EC2KeyPair: ec2.KeyPair{
					Name:       "some-keypair-name",
					PublicKey:  "some-public-key",
					PrivateKey: "some-private-key",
				},
			}))
		})

		Context("when there is an lb", func() {
			It("attaches the lb certificate to the lb type in cloudformation", func() {
				certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
					Name: "some-certificate-name",
					ARN:  "some-certificate-arn",
					Body: "some-certificate-body",
				}

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{
					Stack: storage.Stack{
						Name:            "some-stack-name",
						LBType:          "concourse",
						CertificateName: "some-certificate-name",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(infrastructureManager.CreateCall.Receives.LBCertificateARN).To(Equal("some-certificate-arn"))
			})
		})

		Describe("cloud configurator", func() {
			BeforeEach(func() {
				infrastructureManager.CreateCall.Stub = func(keyPairName string, numberOfAZs int, stackName, lbType, envID string) (cloudformation.Stack, error) {
					stack := cloudformation.Stack{
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
				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-bosh-url"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("user-some-random-string"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("p-some-random-string"))

				Expect(cloudConfigManager.UpdateCall.Receives.CloudConfigInput).To(Equal(cloudConfigInput))
				Expect(cloudConfigManager.UpdateCall.Receives.BOSHClient).To(Equal(boshClient))
			})

			Context("when no load balancer has been requested", func() {
				It("generates a cloud config", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}

					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})

					Expect(err).NotTo(HaveOccurred())
					Expect(cloudConfigurator.ConfigureCall.CallCount).To(Equal(1))
					Expect(cloudConfigurator.ConfigureCall.Receives.Stack).To(Equal(cloudformation.Stack{
						Name: "bbl-aws-some-random-string",
						Outputs: map[string]string{
							"BOSHSecurityGroup":       "some-bosh-security-group",
							"BOSHSubnet":              "some-bosh-subnet",
							"BOSHSubnetAZ":            "some-bosh-subnet-az",
							"BOSHEIP":                 "some-bosh-elastic-ip",
							"BOSHURL":                 "some-bosh-url",
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

					err := command.Execute([]string{"--iaas", "aws"}, storage.State{
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
							"BOSHURL":                  "some-bosh-url",
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

					err := command.Execute([]string{"--iaas", "aws"}, storage.State{
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
							"BOSHURL":                 "some-bosh-url",
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

					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).To(MatchError("error syncing key pair"))
					Expect(stateStore.SetCall.CallCount).To(Equal(1))
					Expect(stateStore.SetCall.Receives.State.KeyPair.Name).To(Equal("keypair-bbl-lake-time:stamp"))
				})
			})

			Context("when the availability zone retriever fails", func() {
				It("saves the public/private key and returns an error", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("availability zone retrieve failed")

					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).To(MatchError("availability zone retrieve failed"))
					Expect(stateStore.SetCall.CallCount).To(Equal(2))
					Expect(stateStore.SetCall.Receives.State.KeyPair.PrivateKey).To(Equal("some-private-key"))
					Expect(stateStore.SetCall.Receives.State.KeyPair.PublicKey).To(Equal("some-public-key"))
				})
			})

			Context("when the cloudformation fails", func() {
				It("saves the stack name and returns an error", func() {
					infrastructureManager.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).To(MatchError("infrastructure creation failed"))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives.State.Stack.Name).To(Equal("stack-bbl-lake-time-stamp"))
				})

				It("saves the private/public key and returns an error", func() {
					infrastructureManager.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).To(MatchError("infrastructure creation failed"))
					Expect(stateStore.SetCall.CallCount).To(Equal(3))
					Expect(stateStore.SetCall.Receives.State.KeyPair.PrivateKey).To(Equal("some-private-key"))
					Expect(stateStore.SetCall.Receives.State.KeyPair.PublicKey).To(Equal("some-public-key"))
				})
			})

			Context("when the bosh cloud config fails", func() {
				It("saves the bosh properties and returns an error", func() {
					cloudConfigManager.UpdateCall.Returns.Error = errors.New("cloud config update failed")

					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).To(MatchError("cloud config update failed"))
					Expect(stateStore.SetCall.CallCount).To(Equal(4))
					Expect(stateStore.SetCall.Receives.State.BOSH).To(Equal(storage.BOSH{
						DirectorName:           "bosh-bbl-lake-time:stamp",
						DirectorUsername:       "user-some-random-string",
						DirectorPassword:       "p-some-random-string",
						DirectorAddress:        "some-bosh-url",
						DirectorSSLCA:          "updated-ca",
						DirectorSSLCertificate: "updated-certificate",
						DirectorSSLPrivateKey:  "updated-private-key",
						State: boshinit.State{
							"updated-key": "updated-value",
						},
						Manifest: "name: bosh",
					}))
				})
			})
		})

		Describe("state manipulation", func() {
			Context("iaas", func() {
				Context("when the iaas does not exist in the state", func() {
					It("writes the iaas provided to the state", func() {
						err := command.Execute([]string{
							"--iaas", "gcp",
						}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
							IAAS: "gcp",
						}))
					})

					Context("when no iaas is provided", func() {
						It("returns an error", func() {
							err := command.Execute([]string{}, storage.State{})
							Expect(err).To(MatchError("--iaas [gcp,aws] must be provided"))
						})
					})
				})

				Context("when the iaas: gcp exists in the state", func() {
					var existingState storage.State

					BeforeEach(func() {
						existingState = storage.State{
							IAAS: "gcp",
						}
						err := command.Execute([]string{}, existingState)
						Expect(err).NotTo(HaveOccurred())
					})

					It("uses the state iaas", func() {
						Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
							IAAS: "gcp",
						}))
					})

					Context("when --iaas aws is provided", func() {
						var err error
						BeforeEach(func() {
							err = command.Execute([]string{"--iaas", "aws"}, existingState)
						})

						It("returns an error", func() {
							Expect(err).To(MatchError("the iaas provided must match the iaas in bbl-state.json"))
						})
					})
				})
			})

			Context("aws credentials", func() {
				Context("when the credentials do not exist", func() {
					It("saves the credentials", func() {
						err := command.Execute([]string{
							"--iaas", "aws",
							"--aws-access-key-id", "some-aws-access-key-id",
							"--aws-secret-access-key", "some-aws-secret-access-key",
							"--aws-region", "some-aws-region",
						}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives.State.AWS).To(Equal(storage.AWS{
							AccessKeyID:     "some-aws-access-key-id",
							SecretAccessKey: "some-aws-secret-access-key",
							Region:          "some-aws-region",
						}))
					})

					Context("failure cases", func() {
						It("returns an error when saving the state fails", func() {
							stateStore.SetCall.Returns = []fakes.SetCallReturn{
								{
									Error: errors.New("saving the state failed"),
								},
							}
							err := command.Execute([]string{
								"--iaas", "aws",
								"--aws-access-key-id", "some-aws-access-key-id",
								"--aws-secret-access-key", "some-aws-secret-access-key",
								"--aws-region", "some-aws-region",
							}, storage.State{})
							Expect(err).To(MatchError("saving the state failed"))
						})

						It("returns an error when parsing the flags fail", func() {
							err := command.Execute([]string{
								"--iaas", "aws",
								"--aws-access-key-id", "some-aws-access-key-id",
								"--aws-secret-access-key", "some-aws-secret-access-key",
								"--unknown-flag", "some-value",
								"--aws-region", "some-aws-region",
							}, storage.State{})
							Expect(err).To(MatchError("flag provided but not defined: -unknown-flag"))

						})
					})
				})
				Context("when the credentials do exist", func() {
					It("overrides the credentials when they're passed in", func() {
						err := command.Execute([]string{
							"--iaas", "aws",
							"--aws-access-key-id", "new-aws-access-key-id",
							"--aws-secret-access-key", "new-aws-secret-access-key",
							"--aws-region", "new-aws-region",
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
						err := command.Execute([]string{"--iaas", "aws"}, storage.State{
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

						err := command.Execute([]string{"--iaas", "aws"}, incomingState)
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

						err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
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
						incomingState := storage.State{}
						err := command.Execute([]string{"--iaas", "aws"}, incomingState)
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
						err := command.Execute([]string{"--iaas", "aws"}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.Stack.Name).To(Equal("some-other-stack-name"))
					})
				})
			})

			Context("env id", func() {
				Context("when the env id doesn't exist", func() {
					It("populates a new bbl env id", func() {
						envIDGenerator.GenerateCall.Returns.EnvID = "bbl-lake-time:stamp"

						err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(stateStore.SetCall.Receives.State.EnvID).To(Equal("bbl-lake-time:stamp"))
					})
				})

				Context("when the env id exists", func() {
					It("does not modify the state", func() {
						incomingState := storage.State{
							EnvID: "bbl-lake-time:stamp",
						}

						err := command.Execute([]string{"--iaas", "aws"}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.EnvID).To(Equal("bbl-lake-time:stamp"))
					})
				})
			})

			Describe("bosh", func() {
				BeforeEach(func() {
					infrastructureManager.ExistsCall.Returns.Exists = true
				})

				Context("boshinit manifest", func() {
					It("writes the boshinit manifest", func() {
						err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.BOSH.Manifest).To(ContainSubstring("name: bosh"))
					})

					It("writes the updated boshinit manifest", func() {
						boshDeployer.DeployCall.Returns.Output = boshinit.DeployOutput{
							BOSHInitManifest: "name: updated-bosh",
						}

						err := command.Execute([]string{"--iaas", "aws"}, storage.State{
							BOSH: storage.BOSH{
								Manifest: "name: bosh",
							},
						})
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.BOSH.Manifest).To(ContainSubstring("name: updated-bosh"))

					})
				})

				Context("bosh state", func() {
					It("writes the bosh state", func() {
						err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.BOSH.State).To(Equal(map[string]interface{}{
							"updated-key": "updated-value",
						}))
					})

					It("writes the updated boshinit manifest", func() {
						boshDeployer.DeployCall.Returns.Output = boshinit.DeployOutput{
							BOSHInitState: boshinit.State{
								"some-key":       "some-value",
								"some-other-key": "some-other-value",
							},
						}

						err := command.Execute([]string{"--iaas", "aws"}, storage.State{
							BOSH: storage.BOSH{
								Manifest: "name: bosh",
								State: boshinit.State{
									"some-key": "some-value",
								},
							},
						})
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.BOSH.State).To(Equal(map[string]interface{}{
							"some-key":       "some-value",
							"some-other-key": "some-other-value",
						}))
					})
				})

				It("writes the bosh director address", func() {
					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					state := stateStore.SetCall.Receives.State
					Expect(state.BOSH.DirectorAddress).To(ContainSubstring("some-bosh-url"))
				})

				It("writes the bosh director name", func() {
					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					state := stateStore.SetCall.Receives.State
					Expect(state.BOSH.DirectorName).To(ContainSubstring("bosh-bbl-lake-time:stamp"))
				})

				Context("when the bosh director ssl keypair exists", func() {
					It("returns the given state unmodified", func() {
						err := command.Execute([]string{"--iaas", "aws"}, storage.State{
							BOSH: storage.BOSH{
								DirectorSSLCA:          "some-ca",
								DirectorSSLCertificate: "some-certificate",
								DirectorSSLPrivateKey:  "some-private-key",
							},
						})
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.BOSH.DirectorSSLCA).To(Equal("some-ca"))
						Expect(state.BOSH.DirectorSSLCertificate).To(Equal("some-certificate"))
						Expect(state.BOSH.DirectorSSLPrivateKey).To(Equal("some-private-key"))
					})
				})

				Context("when the bosh director ssl keypair doesn't exist", func() {
					It("returns the state with a new key pair", func() {
						err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.BOSH.DirectorSSLCA).To(Equal("updated-ca"))
						Expect(state.BOSH.DirectorSSLCertificate).To(Equal("updated-certificate"))
						Expect(state.BOSH.DirectorSSLPrivateKey).To(Equal("updated-private-key"))
						Expect(state.BOSH.State).To(Equal(map[string]interface{}{
							"updated-key": "updated-value",
						}))
					})
				})

				Context("when there are no director credentials", func() {
					It("deploys with randomized director credentials", func() {
						err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(boshDeployer.DeployCall.Receives.Input.DirectorUsername).To(Equal("user-some-random-string"))
						Expect(boshDeployer.DeployCall.Receives.Input.DirectorPassword).To(Equal("p-some-random-string"))
						Expect(state.BOSH.DirectorPassword).To(Equal("p-some-random-string"))
					})
				})

				Context("when there are director credentials", func() {
					It("uses the old credentials", func() {
						incomingState := storage.State{
							BOSH: storage.BOSH{
								DirectorUsername: "some-director-username",
								DirectorPassword: "some-director-password",
							},
						}
						err := command.Execute([]string{"--iaas", "aws"}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(boshDeployer.DeployCall.Receives.Input.DirectorUsername).To(Equal("some-director-username"))
						Expect(boshDeployer.DeployCall.Receives.Input.DirectorPassword).To(Equal("some-director-password"))
					})
				})

				Context("when the bosh credentials don't exist", func() {
					It("returns the state with random credentials", func() {
						boshDeployer.DeployCall.Returns.Output = boshinit.DeployOutput{
							Credentials: boshInitCredentials,
						}

						err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						state := stateStore.SetCall.Receives.State
						Expect(state.BOSH.Credentials).To(Equal(boshInitCredentials))
					})

					Context("when the bosh credentials exist in the bbl state", func() {
						It("deploys with those credentials and returns the state with the same credentials", func() {
							boshDeployer.DeployCall.Returns.Output = boshinit.DeployOutput{
								Credentials: boshInitCredentials,
							}
							err := command.Execute([]string{"--iaas", "aws"}, storage.State{
								BOSH: storage.BOSH{Credentials: boshInitCredentials},
							})
							Expect(err).NotTo(HaveOccurred())

							state := stateStore.SetCall.Receives.State
							Expect(boshDeployer.DeployCall.Receives.Input.Credentials).To(Equal(boshInitCredentials))
							Expect(state.BOSH.Credentials).To(Equal(boshInitCredentials))
						})
					})
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when the certificate cannot be described", func() {
				certificateDescriber.DescribeCall.Returns.Error = errors.New("failed to describe")
				err := command.Execute([]string{"--iaas", "aws"}, storage.State{
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("failed to describe"))
			})

			It("returns an error when the cloud config cannot be uploaded", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update")
				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("failed to update"))
			})

			It("returns an error when the BOSH state exists, but the cloudformation stack does not", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{
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

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("error checking if stack exists"))
			})

			It("returns an error when infrastructure cannot be created", func() {
				infrastructureManager.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("infrastructure creation failed"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshDeployer.DeployCall.Returns.Error = errors.New("cannot deploy bosh")

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("cannot deploy bosh"))
			})

			It("returns an error when it cannot generate a string for the bosh director credentials", func() {
				stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
					if prefix != "bbl-aws-" {
						return "", errors.New("cannot generate string")
					}

					return "", nil
				}
				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("cannot generate string"))
			})

			It("returns an error when availability zones cannot be retrieved", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("availability zone could not be retrieved")

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("availability zone could not be retrieved"))
			})

			It("returns an error when env id generator fails", func() {
				envIDGenerator.GenerateCall.Returns.Error = errors.New("env id generation failed")

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("env id generation failed"))
			})

			It("returns an error when state store fails to set the state when gcp is set", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set state")}}

				err := command.Execute([]string{"--iaas", "gcp"}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before syncing the keypair", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set state")}}

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before retrieving availability zones", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set state")}}

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before creating the stack", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {errors.New("failed to set state")}}

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before updating the cloud config", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {errors.New("failed to set state")}}

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when state store fails to set the state before method exits", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {}, {}, {}, {errors.New("failed to set state")}}

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("failed to set state"))
			})

			It("returns an error when only some of the AWS parameters are provided", func() {
				err := command.Execute([]string{"--iaas", "aws", "--aws-access-key-id", "some-key-id", "--aws-region", "some-region"}, storage.State{})
				Expect(err).To(MatchError("AWS secret access key must be provided"))
			})

			It("returns an error when no AWS parameters are provided and the bbl-state AWS values are empty", func() {
				awsCredentialValidator.ValidateCall.Returns.Error = errors.New("AWS secret access key must be provided")

				err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
				Expect(err).To(MatchError("AWS secret access key must be provided"))
			})
		})
	})
})
