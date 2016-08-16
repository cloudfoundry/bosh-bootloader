package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

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
		)

		BeforeEach(func() {
			keyPairSynchronizer = &fakes.KeyPairSynchronizer{}
			keyPairSynchronizer.SyncCall.Returns.KeyPair = ec2.KeyPair{
				Name:       "some-keypair-name",
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
			envIDGenerator.GenerateCall.Returns.EnvID = "bbl-lake-timestamp"

			command = commands.NewUp(
				awsCredentialValidator, infrastructureManager, keyPairSynchronizer, boshDeployer,
				stringGenerator, cloudConfigurator, availabilityZoneRetriever, certificateDescriber,
				cloudConfigManager, boshClientProvider, envIDGenerator,
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

		It("returns an error if aws credential validator fails", func() {
			awsCredentialValidator.ValidateCall.Returns.Error = errors.New("failed to validate aws credentials")
			_, err := command.Execute([]string{}, storage.State{})
			Expect(err).To(MatchError("failed to validate aws credentials"))
		})

		It("syncs the keypair", func() {
			state, err := command.Execute([]string{}, storage.State{
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

			Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}))
			Expect(keyPairSynchronizer.SyncCall.Receives.EnvID).To(Equal("bbl-lake-timestamp"))

			Expect(state.KeyPair).To(Equal(storage.KeyPair{
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

			_, err := command.Execute([]string{}, incomingState)
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
			var stackNameWasGenerated bool

			stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
				if prefix == "bbl-aws-" {
					stackNameWasGenerated = true
				}
				return prefix + "some-random-string", nil
			}

			_, err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(stackNameWasGenerated).To(BeTrue())

			Expect(infrastructureManager.CreateCall.Receives.StackName).To(Equal("bbl-aws-some-random-string"))
			Expect(infrastructureManager.CreateCall.Receives.KeyPairName).To(Equal("some-keypair-name"))
			Expect(infrastructureManager.CreateCall.Receives.NumberOfAvailabilityZones).To(Equal(1))
			Expect(infrastructureManager.CreateCall.Receives.EnvID).To(Equal("bbl-lake-timestamp"))
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

			_, err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshDeployer.DeployCall.Receives.Input).To(Equal(boshinit.DeployInput{
				DirectorName:     "bosh-bbl-lake-timestamp",
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

				_, err := command.Execute([]string{}, storage.State{
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

		Context("when there is no keypair", func() {
			It("syncs with an empty keypair", func() {
				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{}))
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
				_, err := command.Execute([]string{}, storage.State{})
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

					_, err := command.Execute([]string{}, storage.State{})

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

					_, err := command.Execute([]string{}, storage.State{
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

					_, err := command.Execute([]string{}, storage.State{
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

		Describe("state manipulation", func() {
			BeforeEach(func() {
				keyPairSynchronizer.SyncCall.Returns.KeyPair = ec2.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				}
			})

			Context("aws keypair", func() {
				Context("when the keypair exists", func() {
					It("returns the given state unmodified", func() {
						incomingState := storage.State{
							KeyPair: storage.KeyPair{
								Name:       "some-keypair-name",
								PrivateKey: "some-private-key",
								PublicKey:  "some-public-key",
							},
						}
						state, err := command.Execute([]string{}, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state.KeyPair).To(Equal(incomingState.KeyPair))
					})
				})

				Context("when the keypair doesn't exist", func() {
					It("returns the state with a new key pair", func() {
						state, err := command.Execute([]string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(state.KeyPair).To(Equal(storage.KeyPair{
							Name:       "some-keypair-name",
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
						state, err := command.Execute([]string{}, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state.Stack.Name).To(Equal("bbl-aws-some-random-string"))
					})
				})

				Context("when the stack name exists", func() {
					It("does not modify the state", func() {
						incomingState := storage.State{
							Stack: storage.Stack{
								Name: "some-other-stack-name",
							},
						}
						state, err := command.Execute([]string{}, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state.Stack.Name).To(Equal("some-other-stack-name"))
					})
				})
			})

			Context("env id", func() {
				Context("when the env id doesn't exist", func() {
					It("populates a new bbl env id", func() {
						envIDGenerator.GenerateCall.Returns.EnvID = "bbl-lake-timestamp"

						state, err := command.Execute([]string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())
						Expect(state.EnvID).To(Equal("bbl-lake-timestamp"))
					})
				})

				Context("when the env id exists", func() {
					It("does not modify the state", func() {
						incomingState := storage.State{
							EnvID: "bbl-lake-timestamp",
						}

						state, err := command.Execute([]string{}, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state.EnvID).To(Equal("bbl-lake-timestamp"))
					})
				})
			})

			Describe("bosh", func() {
				BeforeEach(func() {
					infrastructureManager.ExistsCall.Returns.Exists = true
				})

				Context("boshinit manifest", func() {
					It("writes the boshinit manifest", func() {
						state, err := command.Execute([]string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(state.BOSH.Manifest).To(ContainSubstring("name: bosh"))
					})

					It("writes the updated boshinit manifest", func() {
						boshDeployer.DeployCall.Returns.Output = boshinit.DeployOutput{
							BOSHInitManifest: "name: updated-bosh",
						}

						state, err := command.Execute([]string{}, storage.State{
							BOSH: storage.BOSH{
								Manifest: "name: bosh",
							},
						})

						Expect(err).NotTo(HaveOccurred())
						Expect(state.BOSH.Manifest).To(ContainSubstring("name: updated-bosh"))

					})
				})

				Context("bosh state", func() {
					It("writes the bosh state", func() {
						state, err := command.Execute([]string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

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

						state, err := command.Execute([]string{}, storage.State{
							BOSH: storage.BOSH{
								Manifest: "name: bosh",
								State: boshinit.State{
									"some-key": "some-value",
								},
							},
						})

						Expect(err).NotTo(HaveOccurred())
						Expect(state.BOSH.State).To(Equal(map[string]interface{}{
							"some-key":       "some-value",
							"some-other-key": "some-other-value",
						}))
					})
				})

				It("writes the bosh director address", func() {
					state, err := command.Execute([]string{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(state.BOSH.DirectorAddress).To(ContainSubstring("some-bosh-url"))
				})

				It("writes the bosh director name", func() {
					state, err := command.Execute([]string{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(state.BOSH.DirectorName).To(ContainSubstring("bosh-bbl-lake-timestamp"))
				})

				Context("when the bosh director ssl keypair exists", func() {
					It("returns the given state unmodified", func() {
						state, err := command.Execute([]string{}, storage.State{
							BOSH: storage.BOSH{
								DirectorSSLCA:          "some-ca",
								DirectorSSLCertificate: "some-certificate",
								DirectorSSLPrivateKey:  "some-private-key",
							},
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(state.BOSH.DirectorSSLCA).To(Equal("some-ca"))
						Expect(state.BOSH.DirectorSSLCertificate).To(Equal("some-certificate"))
						Expect(state.BOSH.DirectorSSLPrivateKey).To(Equal("some-private-key"))
					})
				})

				Context("when the bosh director ssl keypair doesn't exist", func() {
					It("returns the state with a new key pair", func() {
						state, err := command.Execute([]string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

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
						state, err := command.Execute([]string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

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
						_, err := command.Execute([]string{}, incomingState)
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

						state, err := command.Execute([]string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())
						Expect(state.BOSH.Credentials).To(Equal(boshInitCredentials))
					})

					Context("when the bosh credentials exist in the state.json", func() {
						It("deploys with those credentials and returns the state with the same credentials", func() {
							state, err := command.Execute([]string{}, storage.State{
								BOSH: storage.BOSH{Credentials: boshInitCredentials},
							})

							Expect(err).NotTo(HaveOccurred())
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
				_, err := command.Execute([]string{}, storage.State{
					Stack: storage.Stack{
						LBType: "concourse",
					},
				})
				Expect(err).To(MatchError("failed to describe"))
			})

			It("returns an error when the cloud config cannot be uploaded", func() {
				cloudConfigManager.UpdateCall.Returns.Error = errors.New("failed to update")
				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to update"))
			})

			It("returns an error when the BOSH state exists, but the cloudformation stack does not", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false

				_, err := command.Execute([]string{}, storage.State{
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
					"https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you need assistance."))

				Expect(infrastructureManager.CreateCall.CallCount).To(Equal(0))
			})

			It("returns an error when checking if the infrastructure exists fails", func() {
				infrastructureManager.ExistsCall.Returns.Error = errors.New("error checking if stack exists")

				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("error checking if stack exists"))
			})

			It("returns an error when the key pair fails to sync", func() {
				keyPairSynchronizer.SyncCall.Returns.Error = errors.New("error syncing key pair")

				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("error syncing key pair"))
			})

			It("returns an error when infrastructure cannot be created", func() {
				infrastructureManager.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("infrastructure creation failed"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshDeployer.DeployCall.Returns.Error = errors.New("cannot deploy bosh")

				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("cannot deploy bosh"))
			})

			It("returns an error when it cannot generate a string for the stack name", func() {
				stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
					if prefix == "bbl-aws-" {
						return "", errors.New("cannot generate string")
					}

					return "", nil
				}
				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("cannot generate string"))
			})

			It("returns an error when it cannot generate a string for the bosh director credentials", func() {
				stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
					if prefix != "bbl-aws-" {
						return "", errors.New("cannot generate string")
					}

					return "", nil
				}
				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("cannot generate string"))
			})

			It("returns an error when availability zones cannot be retrieved", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("availability zone could not be retrieved")

				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("availability zone could not be retrieved"))
			})

			It("returns an error when env id generator fails", func() {
				envIDGenerator.GenerateCall.Returns.Error = errors.New("env id generation failed")

				_, err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("env id generation failed"))
			})
		})
	})
})
