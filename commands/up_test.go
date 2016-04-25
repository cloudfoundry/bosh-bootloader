package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
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
			cloudFormationClient      *fakes.CloudFormationClient
			ec2Client                 *fakes.EC2Client
			clientProvider            *fakes.ClientProvider
			stringGenerator           *fakes.StringGenerator
			cloudConfigurator         *fakes.BoshCloudConfigurator
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			globalFlags               commands.GlobalFlags
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
					Certificate: []byte("updated-certificate"),
					PrivateKey:  []byte("updated-private-key"),
				},
				BOSHInitState: boshinit.State{
					"updated-key": "updated-value",
				},
				BOSHInitManifest: "name: bosh",
			}

			cloudFormationClient = &fakes.CloudFormationClient{}
			ec2Client = &fakes.EC2Client{}

			clientProvider = &fakes.ClientProvider{}
			clientProvider.CloudFormationClientCall.Returns.Client = cloudFormationClient
			clientProvider.EC2ClientCall.Returns.Client = ec2Client

			stringGenerator = &fakes.StringGenerator{}
			stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
				return fmt.Sprintf("%s%s", prefix, "some-random-string"), nil
			}

			cloudConfigurator = &fakes.BoshCloudConfigurator{}

			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}

			globalFlags = commands.GlobalFlags{
				EndpointOverride: "some-endpoint",
			}

			command = commands.NewUp(
				infrastructureManager, keyPairSynchronizer, clientProvider, boshDeployer,
				stringGenerator, cloudConfigurator, availabilityZoneRetriever)

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
				"redisPassword":             "some-redis-password",
				"postgresPassword":          "some-postgres-password",
				"registryPassword":          "some-registry-password",
				"blobstoreDirectorPassword": "some-blobstore-director-password",
				"blobstoreAgentPassword":    "some-blobstore-agent-password",
				"hmPassword":                "some-hm-password",
			}
		})

		It("syncs the keypair", func() {
			state, err := command.Execute(globalFlags, []string{}, storage.State{
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

			Expect(clientProvider.EC2ClientCall.Receives.Config).To(Equal(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint",
			}))
			Expect(keyPairSynchronizer.SyncCall.Receives.EC2Client).To(Equal(ec2Client))
			Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}))

			Expect(state.KeyPair).To(Equal(storage.KeyPair{
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
			}

			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}
			var stackNameWasGenerated bool

			stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
				if prefix == "bbl-aws-" {
					stackNameWasGenerated = true
				}
				return prefix + "some-random-string", nil
			}

			_, err := command.Execute(globalFlags, []string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(clientProvider.CloudFormationClientCall.Receives.Config).To(Equal(aws.Config{
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint",
			}))

			Expect(stackNameWasGenerated).To(BeTrue())

			Expect(infrastructureManager.CreateCall.Receives.StackName).To(Equal("bbl-aws-some-random-string"))
			Expect(infrastructureManager.CreateCall.Receives.KeyPairName).To(Equal("some-keypair-name"))
			Expect(infrastructureManager.CreateCall.Receives.NumberOfAvailabilityZones).To(Equal(1))
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
				BOSH: storage.BOSH{
					DirectorSSLCertificate: "some-certificate",
					DirectorSSLPrivateKey:  "some-private-key",
					State: map[string]interface{}{
						"key": "value",
					},
				},
			}

			_, err := command.Execute(globalFlags, []string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshDeployer.DeployCall.Receives.Input).To(Equal(boshinit.DeployInput{
				DirectorUsername: "user-some-random-string",
				DirectorPassword: "p-some-random-string",
				State: boshinit.State{
					"key": "value",
				},
				InfrastructureConfiguration: boshinit.InfrastructureConfiguration{
					AWSRegion:        "some-aws-region",
					SubnetID:         "some-bosh-subnet",
					AvailabilityZone: "some-bosh-subnet-az",
					ElasticIP:        "some-bosh-elastic-ip",
					AccessKeyID:      "some-bosh-user-access-key",
					SecretAccessKey:  "some-bosh-user-secret-access-key",
					SecurityGroup:    "some-bosh-security-group",
				},
				SSLKeyPair: ssl.KeyPair{
					Certificate: []byte("some-certificate"),
					PrivateKey:  []byte("some-private-key"),
				},
				EC2KeyPair: ec2.KeyPair{
					Name:       "some-keypair-name",
					PublicKey:  "some-public-key",
					PrivateKey: "some-private-key",
				},
			}))
		})

		Context("when there is no keypair", func() {
			It("syncs with an empty keypair", func() {
				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairSynchronizer.SyncCall.Receives.EC2Client).To(Equal(ec2Client))
				Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{}))
			})
		})

		Describe("cloud configurator", func() {
			BeforeEach(func() {
				infrastructureManager.CreateCall.Stub = func(keyPairName string, numberOfAZs int, stackName string, lbType string, client cloudformation.Client) (cloudformation.Stack, error) {
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

			Context("when no load balancer has been requested", func() {
				It("generates a cloud config", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}

					_, err := command.Execute(globalFlags, []string{}, storage.State{})

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
				})
			})

			Context("when the load balancer type is concourse", func() {
				It("generates a cloud config", func() {
					availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{"some-retrieved-az"}

					_, err := command.Execute(globalFlags, []string{"--lb-type", "concourse"}, storage.State{})
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

					_, err := command.Execute(globalFlags, []string{"--lb-type", "cf"}, storage.State{})
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

			Context("lb type", func() {
				Context("when the lb type does not exist", func() {
					It("populates the lb type", func() {
						state, err := command.Execute(globalFlags, []string{"--lb-type", ""}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(state.Stack.LBType).To(Equal("none"))
					})
				})

				Context("when the lb type exists", func() {
					It("does not change the lb type when no lb type has been specified", func() {
						incomingState := storage.State{
							Stack: storage.Stack{
								Name:   "some-stack-name",
								LBType: "concourse",
							},
						}

						state, err := command.Execute(globalFlags, []string{"--lb-type", ""}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(state.Stack.LBType).To(Equal("concourse"))
					})

					It("updates the state when an lb type has been specified", func() {
						incomingState := storage.State{
							Stack: storage.Stack{
								Name:   "some-stack-name",
								LBType: "cf",
							},
						}

						state, err := command.Execute(globalFlags, []string{"--lb-type", "concourse"}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(state.Stack.LBType).To(Equal("concourse"))
					})
				})
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
						state, err := command.Execute(globalFlags, []string{}, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state.KeyPair).To(Equal(incomingState.KeyPair))
					})
				})

				Context("when the keypair doesn't exist", func() {
					It("returns the state with a new key pair", func() {
						state, err := command.Execute(globalFlags, []string{}, storage.State{})
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
						state, err := command.Execute(globalFlags, []string{}, incomingState)
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
						state, err := command.Execute(globalFlags, []string{}, incomingState)
						Expect(err).NotTo(HaveOccurred())
						Expect(state.Stack.Name).To(Equal("some-other-stack-name"))
					})
				})
			})

			Describe("bosh", func() {
				BeforeEach(func() {
					infrastructureManager.ExistsCall.Returns.Exists = true
				})

				It("writes the boshinit manifest", func() {
					state, err := command.Execute(globalFlags, []string{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(state.BOSH.Manifest).To(ContainSubstring("name: bosh"))
				})

				It("writes the bosh director address", func() {
					state, err := command.Execute(globalFlags, []string{}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(state.BOSH.DirectorAddress).To(ContainSubstring("some-bosh-url"))
				})

				Context("when the bosh director ssl keypair exists", func() {
					It("returns the given state unmodified", func() {
						state, err := command.Execute(globalFlags, []string{}, storage.State{
							BOSH: storage.BOSH{
								DirectorSSLCertificate: "some-certificate",
								DirectorSSLPrivateKey:  "some-private-key",
							},
						})
						Expect(err).NotTo(HaveOccurred())
						Expect(state.BOSH.DirectorSSLCertificate).To(Equal("some-certificate"))
						Expect(state.BOSH.DirectorSSLPrivateKey).To(Equal("some-private-key"))
					})
				})

				Context("when the bosh director ssl keypair doesn't exist", func() {
					It("returns the state with a new key pair", func() {
						state, err := command.Execute(globalFlags, []string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(state.BOSH.DirectorSSLCertificate).To(Equal("updated-certificate"))
						Expect(state.BOSH.DirectorSSLPrivateKey).To(Equal("updated-private-key"))
						Expect(state.BOSH.State).To(Equal(map[string]interface{}{
							"updated-key": "updated-value",
						}))
					})
				})

				Context("when there are no director credentials", func() {
					It("deploys with randomized director credentials", func() {
						state, err := command.Execute(globalFlags, []string{}, storage.State{})
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
						_, err := command.Execute(globalFlags, []string{}, incomingState)
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

						state, err := command.Execute(globalFlags, []string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())
						Expect(state.BOSH.Credentials).To(Equal(boshInitCredentials))
					})

					Context("when the bosh credentials exist in the state.json", func() {
						It("deploys with those credentials and returns the state with the same credentials", func() {
							state, err := command.Execute(globalFlags, []string{}, storage.State{
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
			Context("when an invalid command line flag is supplied", func() {
				It("returns an error", func() {
					_, err := command.Execute(commands.GlobalFlags{}, []string{"--invalid-flag"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -invalid-flag"))
				})
			})

			It("returns an error when the BOSH state exists, but the cloudformation stack does not", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false

				_, err := command.Execute(globalFlags, []string{}, storage.State{
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

				Expect(infrastructureManager.ExistsCall.Receives.Client).To(Equal(cloudFormationClient))
				Expect(infrastructureManager.ExistsCall.Receives.StackName).To(Equal("some-stack-name"))

				Expect(err).To(MatchError("Found BOSH data in state directory, " +
					"but Cloud Formation stack \"some-stack-name\" cannot be found for region \"some-aws-region\" and given " +
					"AWS credentials. bbl cannot safely proceed. Open an issue on GitHub at " +
					"https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you need assistance."))

				Expect(infrastructureManager.CreateCall.CallCount).To(Equal(0))
			})

			It("returns an error when checking if the infrastructure exists fails", func() {
				infrastructureManager.ExistsCall.Returns.Error = errors.New("error checking if stack exists")

				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("error checking if stack exists"))
			})

			It("returns an error when the cloudformation client can not be created", func() {
				clientProvider.CloudFormationClientCall.Returns.Error = errors.New("error creating client")

				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("error creating client"))
			})

			It("returns an error when the ec2 client can not be created", func() {
				clientProvider.EC2ClientCall.Returns.Error = errors.New("error creating client")

				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("error creating client"))
			})

			It("returns an error when the key pair fails to sync", func() {
				keyPairSynchronizer.SyncCall.Returns.Error = errors.New("error syncing key pair")

				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("error syncing key pair"))
			})

			It("returns an error when infrastructure cannot be created", func() {
				infrastructureManager.CreateCall.Returns.Error = errors.New("infrastructure creation failed")

				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("infrastructure creation failed"))
			})

			It("returns an error when the cloud config cannot be configured", func() {
				cloudConfigurator.ConfigureCall.Returns.Error = errors.New("bosh cloud configuration failed")

				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("bosh cloud configuration failed"))
			})

			It("returns an error when bosh cannot be deployed", func() {
				boshDeployer := &fakes.BOSHDeployer{}
				boshDeployer.DeployCall.Returns.Error = errors.New("cannot deploy bosh")
				command = commands.NewUp(
					infrastructureManager, keyPairSynchronizer, clientProvider, boshDeployer,
					stringGenerator, cloudConfigurator, availabilityZoneRetriever)

				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("cannot deploy bosh"))
			})

			It("returns an error when it cannot generate a string for the stack name", func() {
				stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
					if prefix == "bbl-aws-" {
						return "", errors.New("cannot generate string")
					}

					return "", nil
				}
				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("cannot generate string"))
			})

			It("returns an error when it cannot generate a string for the bosh director credentials", func() {
				stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
					if prefix != "bbl-aws-" {
						return "", errors.New("cannot generate string")
					}

					return "", nil
				}
				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("cannot generate string"))
			})

			It("returns an error when availability zones cannot be retrieved", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("availability zone could not be retrieved")

				_, err := command.Execute(globalFlags, []string{}, storage.State{})
				Expect(err).To(MatchError("availability zone could not be retrieved"))
			})
		})
	})
})
