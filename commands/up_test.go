package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
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
			command                        commands.Up
			boshDeployer                   *fakes.BOSHDeployer
			infrastructureManager          *fakes.InfrastructureManager
			keyPairSynchronizer            *fakes.KeyPairSynchronizer
			cloudFormationClient           *fakes.CloudFormationClient
			ec2Client                      *fakes.EC2Client
			elbClient                      *fakes.ELBClient
			iamClient                      *fakes.IAMClient
			clientProvider                 *fakes.ClientProvider
			stringGenerator                *fakes.StringGenerator
			cloudConfigurator              *fakes.BoshCloudConfigurator
			availabilityZoneRetriever      *fakes.AvailabilityZoneRetriever
			elbDescriber                   *fakes.ELBDescriber
			loadBalancerCertificateManager *fakes.LoadBalancerCertificateManager
			globalFlags                    commands.GlobalFlags
			boshInitCredentials            map[string]string
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
			elbClient = &fakes.ELBClient{}
			iamClient = &fakes.IAMClient{}

			clientProvider = &fakes.ClientProvider{}
			clientProvider.CloudFormationClientCall.Returns.Client = cloudFormationClient
			clientProvider.EC2ClientCall.Returns.Client = ec2Client
			clientProvider.ELBClientCall.Returns.Client = elbClient
			clientProvider.IAMClientCall.Returns.Client = iamClient

			stringGenerator = &fakes.StringGenerator{}
			stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
				return fmt.Sprintf("%s%s", prefix, "some-random-string"), nil
			}

			cloudConfigurator = &fakes.BoshCloudConfigurator{}

			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}

			elbDescriber = &fakes.ELBDescriber{}

			loadBalancerCertificateManager = &fakes.LoadBalancerCertificateManager{}
			loadBalancerCertificateManager.IsValidLBTypeCall.Returns.Result = true

			globalFlags = commands.GlobalFlags{
				EndpointOverride: "some-endpoint",
			}

			command = commands.NewUp(
				infrastructureManager, keyPairSynchronizer, clientProvider, boshDeployer,
				stringGenerator, cloudConfigurator, availabilityZoneRetriever, elbDescriber, loadBalancerCertificateManager,
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

		It("checks if the lb type flag is valid", func() {
			loadBalancerCertificateManager.IsValidLBTypeCall.Returns.Result = true
			_, err := command.Execute(commands.GlobalFlags{}, []string{"--lb-type", "concourse"}, storage.State{})
			Expect(loadBalancerCertificateManager.IsValidLBTypeCall.Receives.LBType).To(Equal("concourse"))
			Expect(err).NotTo(HaveOccurred())
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

		Context("when there is an lb type and lb certificate", func() {
			It("attaches the lb certificate to the lb type in cloudformation", func() {
				loadBalancerCertificateManager.CreateCall.Returns.Output = iam.CertificateCreateOutput{
					CertificateName: "some-certificate-name",
					CertificateARN:  "some-certificate-arn",
				}

				subcommandArgs := []string{
					"--lb-type", "concourse",
					"--cert", "some-certificate-file",
					"--key", "some-private-key-file",
				}

				_, err := command.Execute(globalFlags, subcommandArgs, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(infrastructureManager.CreateCall.Receives.LBCertificateARN).To(Equal("some-certificate-arn"))
			})
		})

		Context("when specifying an lb type that is not \"none\"", func() {
			Context("when cert and key are provided", func() {
				It("uploads the given cert and key", func() {
					subcommandArgs := []string{
						"--lb-type", "concourse",
						"--cert", "some-certificate-file",
						"--key", "some-private-key-file",
					}

					_, err := command.Execute(globalFlags, subcommandArgs, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(loadBalancerCertificateManager.CreateCall.Receives.Input).To(Equal(iam.CertificateCreateInput{
						CurrentCertificateName: "",
						CurrentLBType:          "",
						DesiredLBType:          "concourse",
						CertPath:               "some-certificate-file",
						KeyPath:                "some-private-key-file",
					}))
				})
			})
		})

		Context("when there are instances attached to an lb", func() {
			BeforeEach(func() {
				infrastructureManager.ExistsCall.Returns.Exists = true
			})

			It("should not check for ELBs if the stack does not exist", func() {
				infrastructureManager.ExistsCall.Returns.Exists = false
				infrastructureManager.DescribeCall.Returns.Error = cloudformation.StackNotFound

				_, err := command.Execute(globalFlags, []string{"--lb-type", "none"}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should not verify instances if the lb type has not changed", func() {
				incomingState := storage.State{
					Stack: storage.Stack{
						Name:   "some-stack-name",
						LBType: "concourse",
					},
				}

				infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
					Name:    "some-stack-name",
					Outputs: map[string]string{"ConcourseLoadBalancer": "some-load-balancer"},
				}

				_, err := command.Execute(globalFlags, []string{"--lb-type", "concourse"}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(elbDescriber.DescribeCall.CallCount).To(Equal(0))
			})

			Context("concourse", func() {
				It("should fast fail when ConcourseLoadBalancer has instances", func() {
					incomingState := storage.State{
						Stack: storage.Stack{
							Name:   "some-stack-name",
							LBType: "concourse",
						},
					}

					infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
						Name:    "some-stack-name",
						Outputs: map[string]string{"ConcourseLoadBalancer": "some-load-balancer"},
					}

					elbDescriber.DescribeCall.Returns.Instances = []string{"some-instance-1", "some-instance-2"}

					_, err := command.Execute(globalFlags, []string{"--lb-type", "none"}, incomingState)
					Expect(err).To(MatchError("Load balancer \"some-load-balancer\" cannot be deleted since it has attached instances: some-instance-1, some-instance-2"))
				})
			})

			Context("cf", func() {
				It("should fast fail when CFRouterLoadBalancer has instances", func() {
					incomingState := storage.State{
						Stack: storage.Stack{
							Name:   "some-stack-name",
							LBType: "cf",
						},
					}

					infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
						Name: "some-stack-name",
						Outputs: map[string]string{
							"CFRouterLoadBalancer":   "some-router-load-balancer",
							"CFSSHProxyLoadBalancer": "some-ssh-proxy-load-balancer",
						},
					}

					elbDescriber.DescribeCall.Stub = func(lbName string) ([]string, error) {
						if lbName == "some-router-load-balancer" {
							return []string{"some-instance-1", "some-instance-2"}, nil
						}

						return []string{}, nil
					}

					_, err := command.Execute(globalFlags, []string{"--lb-type", "none"}, incomingState)
					Expect(err).To(MatchError("Load balancer \"some-router-load-balancer\" cannot be deleted since it has attached instances: some-instance-1, some-instance-2"))
				})

				It("should fast fail when CFSSHProxyLoadBalancer has instances", func() {
					incomingState := storage.State{
						Stack: storage.Stack{
							Name:   "some-stack-name",
							LBType: "cf",
						},
					}

					infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
						Name: "some-stack-name",
						Outputs: map[string]string{
							"CFRouterLoadBalancer":   "some-router-load-balancer",
							"CFSSHProxyLoadBalancer": "some-ssh-proxy-load-balancer",
						},
					}

					elbDescriber.DescribeCall.Stub = func(lbName string) ([]string, error) {
						if lbName == "some-ssh-proxy-load-balancer" {
							return []string{"some-instance-1", "some-instance-2"}, nil
						}

						return []string{}, nil
					}

					_, err := command.Execute(globalFlags, []string{"--lb-type", "none"}, incomingState)
					Expect(err).To(MatchError("Load balancer \"some-ssh-proxy-load-balancer\" cannot be deleted since it has attached instances: some-instance-1, some-instance-2"))
				})
			})
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
					loadBalancerCertificateManager.CreateCall.Returns.Output = iam.CertificateCreateOutput{
						CertificateName: "some-certificate-name",
						CertificateARN:  "some-certificate-arn",
						LBType:          "concourse",
					}

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
					loadBalancerCertificateManager.CreateCall.Returns.Output = iam.CertificateCreateOutput{
						CertificateName: "some-certificate-name",
						CertificateARN:  "some-certificate-arn",
						LBType:          "cf",
					}

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
				It("populates the lb type from the load balancer certificate manager", func() {
					loadBalancerCertificateManager.CreateCall.Returns.Output = iam.CertificateCreateOutput{
						LBType: "cf",
					}

					state, err := command.Execute(globalFlags, []string{"--lb-type", ""}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(state.Stack.LBType).To(Equal("cf"))
				})

				Context("when the cert and key changes", func() {
					It("saves certificate name in state when certificate manager updates certificate", func() {
						incomingState := storage.State{
							Stack: storage.Stack{
								CertificateName: "some-certificate-name",
							},
						}
						loadBalancerCertificateManager.CreateCall.Returns.Output = iam.CertificateCreateOutput{
							CertificateName: "some-other-certificate-name",
						}
						state, err := command.Execute(globalFlags, []string{
							"--lb-type", "concourse",
							"--cert", "some-cert-path",
							"--key", "some-key-path",
						}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(state.Stack.CertificateName).To(Equal("some-other-certificate-name"))
						Expect(loadBalancerCertificateManager.CreateCall.Receives.Input).To(Equal(iam.CertificateCreateInput{
							CurrentCertificateName: "some-certificate-name",
							DesiredLBType:          "concourse",
							CertPath:               "some-cert-path",
							KeyPath:                "some-key-path",
						}))
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

				Context("boshinit manifest", func() {
					It("writes the boshinit manifest", func() {
						state, err := command.Execute(globalFlags, []string{}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(state.BOSH.Manifest).To(ContainSubstring("name: bosh"))
					})

					It("writes the updated boshinit manifest", func() {
						boshDeployer.DeployCall.Returns.Output = boshinit.DeployOutput{
							BOSHInitManifest: "name: updated-bosh",
						}

						state, err := command.Execute(globalFlags, []string{}, storage.State{
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
						state, err := command.Execute(globalFlags, []string{}, storage.State{})
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

						state, err := command.Execute(globalFlags, []string{}, storage.State{
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

			It("returns an error when an unknown lb-type is supplied", func() {
				loadBalancerCertificateManager.IsValidLBTypeCall.Returns.Result = false
				_, err := command.Execute(commands.GlobalFlags{}, []string{"--lb-type", "some-lb"}, storage.State{})
				Expect(loadBalancerCertificateManager.IsValidLBTypeCall.Receives.LBType).To(Equal("some-lb"))
				Expect(err).To(MatchError("Unknown lb-type \"some-lb\", supported lb-types are: concourse, cf or none"))
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

			It("returns an error when the IAM client can not be created", func() {
				clientProvider.IAMClientCall.Returns.Error = errors.New("error creating client")

				_, err := command.Execute(globalFlags, []string{
					"--lb-type", "cf",
					"--cert", "some-cert",
					"--key", "some-key",
				}, storage.State{})
				Expect(err).To(MatchError("error creating client"))
			})

			It("returns an error when the certificate cannot be uploaded", func() {
				loadBalancerCertificateManager.CreateCall.Returns.Error = errors.New("error uploading certificate")

				_, err := command.Execute(globalFlags, []string{
					"--lb-type", "cf",
					"--cert", "some-cert",
					"--key", "some-key",
				}, storage.State{})
				Expect(err).To(MatchError("error uploading certificate"))
			})

			It("returns an error when the ELB client can not be created", func() {
				infrastructureManager.ExistsCall.Returns.Exists = true
				clientProvider.ELBClientCall.Returns.Error = errors.New("error creating client")

				_, err := command.Execute(globalFlags, []string{"--lb-type", "concourse"}, storage.State{})
				Expect(err).To(MatchError("error creating client"))
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

			It("returns an error when infrastructure cannot be described", func() {
				infrastructureManager.ExistsCall.Returns.Exists = true
				infrastructureManager.DescribeCall.Returns.Error = errors.New("infrastructure describe failed")

				_, err := command.Execute(globalFlags, []string{"--lb-type", "concourse"}, storage.State{})
				Expect(err).To(MatchError("infrastructure describe failed"))
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
					stringGenerator, cloudConfigurator, availabilityZoneRetriever, elbDescriber, loadBalancerCertificateManager)

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

			It("returns an error when an elb cannot be described", func() {
				infrastructureManager.ExistsCall.Returns.Exists = true
				infrastructureManager.DescribeCall.Returns.Stack = cloudformation.Stack{
					Name: "some-stack-name",
					Outputs: map[string]string{
						"CFRouterLoadBalancer":   "some-router-load-balancer",
						"CFSSHProxyLoadBalancer": "some-ssh-proxy-load-balancer",
					},
				}

				elbDescriber.DescribeCall.Returns.Error = errors.New("elb cannot be described")

				_, err := command.Execute(globalFlags, []string{"--lb-type", "concourse"}, storage.State{})
				Expect(err).To(MatchError("elb cannot be described"))
			})
		})
	})
})
