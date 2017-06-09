package bosh_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

const (
	variablesYAML = `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`
)

var _ = Describe("Manager", func() {
	Describe("Create", func() {
		var (
			boshExecutor     *fakes.BOSHExecutor
			terraformManager *fakes.TerraformManager
			stackManager     *fakes.StackManager
			logger           *fakes.Logger
			socks5Proxy      *fakes.Socks5Proxy
			boshManager      bosh.Manager
			incomingGCPState storage.State
			incomingAWSState storage.State

			osUnsetenvKey string
			osSetenvKey   string
			osSetenvValue string
		)

		BeforeEach(func() {
			terraformManager = &fakes.TerraformManager{}
			stackManager = &fakes.StackManager{}
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, terraformManager, stackManager, logger, socks5Proxy)

			bosh.SetOSSetenv(func(key, value string) error {
				osSetenvKey = key
				osSetenvValue = value
				return nil
			})

			bosh.SetOSUnsetenv(func(key string) error {
				osUnsetenvKey = key
				return nil
			})

			terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
				"network_name":       "some-network",
				"subnetwork_name":    "some-subnetwork",
				"bosh_open_tag_name": "some-bosh-tag",
				"internal_tag_name":  "some-internal-tag",
				"external_ip":        "some-external-ip",
				"director_address":   "some-director-address",
				"jumpbox_url":        "some-jumpbox-url",
			}

			incomingGCPState = storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				KeyPair: storage.KeyPair{
					PrivateKey: "some-private-key",
				},
				GCP: storage.GCP{
					Zone:              "some-zone",
					ProjectID:         "some-project-id",
					ServiceAccountKey: "some-credential-json",
				},
				BOSH: storage.BOSH{
					State: map[string]interface{}{
						"some-key": "some-value",
					},
				},
				TFState: "some-tf-state",
				LB: storage.LB{
					Type: "cf",
				},
			}

			incomingAWSState = storage.State{
				IAAS:  "aws",
				EnvID: "some-env-id",
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
				},
				AWS: storage.AWS{
					Region: "some-region",
				},
				BOSH: storage.BOSH{
					State: map[string]interface{}{
						"some-key": "some-value",
					},
				},
				TFState: "some-tf-state",
				LB: storage.LB{
					Type: "cf",
				},
			}
		})

		AfterEach(func() {
			bosh.ResetOSSetenv()
		})

		It("logs bosh director status messages", func() {
			boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: variablesYAML,
			}

			_, err := boshManager.Create(incomingGCPState)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Messages).To(ContainSequence([]string{"creating bosh director", "created bosh director"}))
		})

		Context("when iaas is gcp", func() {
			It("queries values from terraform manager", func() {
				boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
					Manifest:  "some-manifest",
					Variables: variablesYAML,
				}

				_, err := boshManager.Create(incomingGCPState)
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingGCPState))
			})

			It("generates a bosh manifest", func() {
				boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
					Manifest:  "some-manifest",
					Variables: variablesYAML,
				}

				boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
					State: map[string]interface{}{
						"some-new-key": "some-new-value",
					},
				}

				incomingGCPState.BOSH.UserOpsFile = "some-ops-file"
				_, err := boshManager.Create(incomingGCPState)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshExecutor.CreateEnvCall.CallCount).To(Equal(1))
				Expect(boshExecutor.InterpolateCall.Receives.InterpolateInput).To(Equal(bosh.InterpolateInput{
					IAAS: "gcp",
					DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags: [some-bosh-tag, some-internal-tag]
project_id: some-project-id
gcp_credentials_json: 'some-credential-json'`,
					BOSHState: map[string]interface{}{
						"some-key": "some-value",
					},
					Variables: "",
					OpsFile:   "some-ops-file",
				}))

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(0))
				Expect(boshExecutor.JumpboxInterpolateCall.CallCount).To(Equal(0))
			})

			It("returns a state with a proper bosh state", func() {
				boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
					Manifest:  "some-manifest",
					Variables: variablesYAML,
				}

				boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
					State: map[string]interface{}{
						"some-new-key": "some-new-value",
					},
				}

				state, err := boshManager.Create(incomingGCPState)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					KeyPair: storage.KeyPair{
						PrivateKey: "some-private-key",
					},
					GCP: storage.GCP{
						Zone:              "some-zone",
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-credential-json",
					},
					BOSH: storage.BOSH{
						State: map[string]interface{}{
							"some-new-key": "some-new-value",
						},
						Variables:              variablesYAML,
						Manifest:               "some-manifest",
						DirectorName:           "bosh-some-env-id",
						DirectorAddress:        "some-director-address",
						DirectorUsername:       "admin",
						DirectorPassword:       "some-admin-password",
						DirectorSSLCA:          "some-ca",
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "cf",
					},
				}))
			})

			Context("when jumpbox enabled is true", func() {
				var jumpboxDeploymentVars string
				var deploymentVars string
				BeforeEach(func() {
					incomingGCPState = storage.State{
						IAAS:  "gcp",
						EnvID: "some-env-id",
						KeyPair: storage.KeyPair{
							PrivateKey: "some-private-key",
						},
						GCP: storage.GCP{
							Zone:              "some-zone",
							ProjectID:         "some-project-id",
							ServiceAccountKey: "some-credential-json",
						},
						Jumpbox: storage.Jumpbox{
							Enabled:   true,
							Variables: "jumpbox_ssh:\n  private_key: some-jumpbox-private-key",
							Manifest:  "name: jumpbox",
							State: map[string]interface{}{
								"some-key": "some-value",
							},
						},
						BOSH: storage.BOSH{
							State: map[string]interface{}{
								"some-key": "some-value",
							},
						},
						TFState: "some-tf-state",
						LB: storage.LB{
							Type: "cf",
						},
					}

					jumpboxDeploymentVars = `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.5
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags: [some-bosh-tag, some-internal-tag]
project_id: some-project-id
gcp_credentials_json: 'some-credential-json'`

					deploymentVars = `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags: [some-internal-tag]
project_id: some-project-id
gcp_credentials_json: 'some-credential-json'`

					boshExecutor.JumpboxInterpolateCall.Returns.Output = bosh.JumpboxInterpolateOutput{
						Manifest:  "name: jumpbox",
						Variables: "jumpbox_ssh:\n  private_key: some-jumpbox-private-key",
					}

					boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
						Manifest:  "some-manifest",
						Variables: variablesYAML,
					}
				})

				It("queries values from terraform manager", func() {
					boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
						State: map[string]interface{}{
							"some-key": "some-value",
						},
					}
					_, err := boshManager.Create(incomingGCPState)
					Expect(err).NotTo(HaveOccurred())

					Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingGCPState))
				})

				It("logs jumpbox status messages", func() {
					_, err := boshManager.Create(incomingGCPState)
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.StepCall.Messages).To(ContainSequence([]string{"creating jumpbox", "created jumpbox"}))
				})

				It("generates a jumpbox and bosh manifest", func() {
					_, err := boshManager.Create(incomingGCPState)
					Expect(err).NotTo(HaveOccurred())

					Expect(boshExecutor.InterpolateCall.Receives.InterpolateInput).To(Equal(bosh.InterpolateInput{
						IAAS: "gcp",
						JumpboxDeploymentVars: jumpboxDeploymentVars,
						DeploymentVars:        deploymentVars,
						BOSHState: map[string]interface{}{
							"some-key": "some-value",
						},
						Variables: "",
					}))
				})

				It("starts a socks5 proxy for the duration of creating the bosh director", func() {
					socks5ProxyAddr := "localhost:1234"
					socks5Proxy.AddrCall.Returns.Addr = socks5ProxyAddr

					_, err := boshManager.Create(incomingGCPState)
					Expect(err).NotTo(HaveOccurred())

					Expect(osUnsetenvKey).To(Equal("BOSH_ALL_PROXY"))
					Expect(socks5Proxy.StartCall.CallCount).To(Equal(1))
					Expect(socks5Proxy.StartCall.Receives.JumpboxPrivateKey).To(Equal("some-jumpbox-private-key"))
					Expect(socks5Proxy.StartCall.Receives.JumpboxExternalURL).To(Equal("some-jumpbox-url"))
					Expect(osSetenvKey).To(Equal("BOSH_ALL_PROXY"))
					Expect(osSetenvValue).To(Equal(fmt.Sprintf("socks5://%s", socks5ProxyAddr)))

					Expect(logger.StepCall.Messages).To(ContainSequence([]string{
						"creating jumpbox",
						"created jumpbox",
						"starting socks5 proxy to jumpbox",
						"creating bosh director",
						"created bosh director",
					}))
				})

				It("returns a bbl state with a proper jumpbox state", func() {
					boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
						State: map[string]interface{}{
							"some-new-key": "some-new-value",
						},
					}

					state, err := boshManager.Create(incomingGCPState)
					Expect(err).NotTo(HaveOccurred())

					Expect(state).To(Equal(storage.State{
						IAAS:  "gcp",
						EnvID: "some-env-id",
						KeyPair: storage.KeyPair{
							PrivateKey: "some-private-key",
						},
						GCP: storage.GCP{
							Zone:              "some-zone",
							ProjectID:         "some-project-id",
							ServiceAccountKey: "some-credential-json",
						},
						Jumpbox: storage.Jumpbox{
							Enabled:   true,
							Variables: "jumpbox_ssh:\n  private_key: some-jumpbox-private-key",
							Manifest:  "name: jumpbox",
							State: map[string]interface{}{
								"some-new-key": "some-new-value",
							},
						},
						BOSH: storage.BOSH{
							State: map[string]interface{}{
								"some-new-key": "some-new-value",
							},
							Variables:              variablesYAML,
							Manifest:               "some-manifest",
							DirectorName:           "bosh-some-env-id",
							DirectorAddress:        "https://10.0.0.6:25555",
							DirectorUsername:       "admin",
							DirectorPassword:       "some-admin-password",
							DirectorSSLCA:          "some-ca",
							DirectorSSLCertificate: "some-certificate",
							DirectorSSLPrivateKey:  "some-private-key",
						},
						TFState: "some-tf-state",
						LB: storage.LB{
							Type: "cf",
						},
					}))
				})

				Context("failure cases", func() {
					Context("when the jumpbox variables cannot be parsed", func() {
						BeforeEach(func() {
							boshExecutor.JumpboxInterpolateCall.Returns.Output.Variables = "%%%"
						})

						It("returns an error", func() {
							_, err := boshManager.Create(incomingGCPState)
							Expect(err).To(MatchError("yaml: could not find expected directive name"))
						})
					})

					It("returns an error when the socks5Proxy fails to start", func() {
						socks5Proxy.StartCall.Returns.Error = errors.New("failed to start socks5Proxy")
						_, err := boshManager.Create(incomingGCPState)
						Expect(err).To(MatchError("failed to start socks5Proxy"))
					})
				})
			})
		})

		Context("when iaas is aws", func() {
			Context("when cloudformation was used to create infrastructure", func() {
				BeforeEach(func() {
					stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
						Outputs: map[string]string{
							"BOSHSubnetAZ":            "some-bosh-subnet-az",
							"BOSHUserAccessKey":       "some-bosh-user-access-key",
							"BOSHUserSecretAccessKey": "some-bosh-user-secret-access-key",
							"BOSHSecurityGroup":       "some-bosh-security-group",
							"BOSHSubnet":              "some-bosh-subnet",
							"BOSHEIP":                 "some-bosh-elastic-ip",
							"BOSHURL":                 "some-bosh-url",
						},
					}
					incomingAWSState.TFState = ""
					incomingAWSState.Stack = storage.Stack{
						Name: "some-stack",
					}

					boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
						Manifest:  "some-manifest",
						Variables: variablesYAML,
					}

					boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
						State: map[string]interface{}{
							"some-new-key": "some-new-value",
						},
					}
				})

				It("generates a bosh manifest", func() {
					incomingAWSState.BOSH.UserOpsFile = "some-ops-file"
					_, err := boshManager.Create(incomingAWSState)
					Expect(err).NotTo(HaveOccurred())

					Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(0))
					Expect(stackManager.DescribeCall.CallCount).To(Equal(2))
					Expect(stackManager.DescribeCall.Receives.StackName).To(Equal("some-stack"))

					Expect(boshExecutor.InterpolateCall.Receives.InterpolateInput).To(Equal(bosh.InterpolateInput{
						IAAS: "aws",
						DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-elastic-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-bosh-user-access-key
secret_access_key: some-bosh-user-secret-access-key
default_key_name: some-keypair-name
default_security_groups: [some-bosh-security-group]
region: some-region
private_key: |-
  some-private-key`,
						BOSHState: map[string]interface{}{
							"some-key": "some-value",
						},
						Variables: "",
						OpsFile:   "some-ops-file",
					}))
				})
			})

			Context("when terraform was used to create infrastructure", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
						"az":                      "some-bosh-subnet-az",
						"access_key_id":           "some-bosh-user-access-key",
						"secret_access_key":       "some-bosh-user-secret-access-key",
						"default_security_groups": "some-bosh-security-group",
						"subnet_id":               "some-bosh-subnet",
						"external_ip":             "some-bosh-elastic-ip",
						"director_address":        "some-bosh-url",
					}

					boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
						Manifest:  "some-manifest",
						Variables: variablesYAML,
					}

					boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
						State: map[string]interface{}{
							"some-new-key": "some-new-value",
						},
					}
				})

				It("generates a bosh manifest", func() {
					incomingAWSState.BOSH.UserOpsFile = "some-ops-file"
					_, err := boshManager.Create(incomingAWSState)
					Expect(err).NotTo(HaveOccurred())

					Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(2))
					Expect(stackManager.DescribeCall.CallCount).To(Equal(0))
					Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingAWSState))

					Expect(boshExecutor.InterpolateCall.Receives.InterpolateInput).To(Equal(bosh.InterpolateInput{
						IAAS: "aws",
						DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-elastic-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-bosh-user-access-key
secret_access_key: some-bosh-user-secret-access-key
default_key_name: some-keypair-name
default_security_groups: [some-bosh-security-group]
region: some-region
private_key: |-
  some-private-key`,
						BOSHState: map[string]interface{}{
							"some-key": "some-value",
						},
						Variables: "",
						OpsFile:   "some-ops-file",
					}))
				})

				It("returns a state with a proper bosh state", func() {
					state, err := boshManager.Create(incomingAWSState)
					Expect(err).NotTo(HaveOccurred())

					Expect(state).To(Equal(storage.State{
						IAAS:  "aws",
						EnvID: "some-env-id",
						KeyPair: storage.KeyPair{
							Name:       "some-keypair-name",
							PrivateKey: "some-private-key",
						},
						AWS: storage.AWS{
							Region: "some-region",
						},
						BOSH: storage.BOSH{
							State: map[string]interface{}{
								"some-new-key": "some-new-value",
							},
							Variables:              variablesYAML,
							Manifest:               "some-manifest",
							DirectorName:           "bosh-some-env-id",
							DirectorAddress:        "some-bosh-url",
							DirectorUsername:       "admin",
							DirectorPassword:       "some-admin-password",
							DirectorSSLCA:          "some-ca",
							DirectorSSLCertificate: "some-certificate",
							DirectorSSLPrivateKey:  "some-private-key",
						},
						TFState: "some-tf-state",
						LB: storage.LB{
							Type: "cf",
						},
					}))
				})
			})
		})

		It("creates a bosh environment", func() {
			boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: variablesYAML,
			}

			_, err := boshManager.Create(incomingGCPState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshExecutor.CreateEnvCall.Receives.Input).To(Equal(bosh.CreateEnvInput{
				Manifest: "some-manifest",
				State: map[string]interface{}{
					"some-key": "some-value",
				},
				Variables: variablesYAML,
			}))
		})

		Context("failure cases", func() {
			It("returns an error when terraform output provider fails", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to output")
				_, err := boshManager.Create(incomingGCPState)
				Expect(err).To(MatchError("failed to output"))
			})

			It("returns an error when an invalid iaas is provided", func() {
				_, err := boshManager.Create(storage.State{})
				Expect(err).To(MatchError("A valid IAAS was not provided"))
			})

			It("returns an error when the executor's interpolate call fails", func() {
				boshExecutor.InterpolateCall.Returns.Error = errors.New("failed to interpolate")
				_, err := boshManager.Create(incomingGCPState)
				Expect(err).To(MatchError("failed to interpolate"))
			})

			It("returns an error when the executor's create env call fails with non create env error", func() {
				boshExecutor.CreateEnvCall.Returns.Error = errors.New("failed to create")
				_, err := boshManager.Create(incomingGCPState)
				Expect(err).To(MatchError("failed to create"))
			})

			Context("when interpolate outputs invalid yaml", func() {
				It("returns an error", func() {
					boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
						Manifest:  "some-manifest",
						Variables: "%%%",
					}

					_, err := boshManager.Create(incomingAWSState)
					Expect(err).To(MatchError("failed to get director outputs:\nyaml: could not find expected directive name"))
				})
			})

			Context("when the executor's create env call fails with create env error", func() {
				var (
					expectedError bosh.ManagerCreateError
					expectedState storage.State
				)

				BeforeEach(func() {
					boshState := map[string]interface{}{
						"partial": "bosh-state",
					}

					boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
						Manifest:  "some-manifest",
						Variables: variablesYAML,
					}

					createEnvError := bosh.NewCreateEnvError(boshState, errors.New("failed to create env"))
					boshExecutor.CreateEnvCall.Returns.Error = createEnvError

					expectedState = incomingAWSState
					expectedState.BOSH = storage.BOSH{
						Manifest:  "some-manifest",
						State:     boshState,
						Variables: variablesYAML,
					}
					expectedError = bosh.NewManagerCreateError(expectedState, createEnvError)
				})

				It("returns a bosh manager create error with a valid state", func() {
					_, err := boshManager.Create(incomingAWSState)
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when the stack manager returns an error", func() {
				BeforeEach(func() {
					stackManager.DescribeCall.Returns.Error = errors.New("stack manager describe failed")
				})

				It("returns the error", func() {
					_, err := boshManager.Create(storage.State{
						IAAS: "aws",
					})
					Expect(err).To(MatchError("stack manager describe failed"))
				})
			})
		})
	})

	Describe("Delete", func() {
		var (
			stackManager     *fakes.StackManager
			boshExecutor     *fakes.BOSHExecutor
			terraformManager *fakes.TerraformManager
			logger           *fakes.Logger
			socks5Proxy      *fakes.Socks5Proxy
			boshManager      bosh.Manager
		)

		BeforeEach(func() {
			terraformManager = &fakes.TerraformManager{}
			stackManager = &fakes.StackManager{}
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, terraformManager, stackManager, logger, socks5Proxy)
		})

		It("calls delete env", func() {
			err := boshManager.Delete(storage.State{
				BOSH: storage.BOSH{
					Manifest: "some-manifest",
					State: map[string]interface{}{
						"key": "value",
					},
					Variables: variablesYAML,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(boshExecutor.DeleteEnvCall.Receives.Input).To(Equal(bosh.DeleteEnvInput{
				Manifest: "some-manifest",
				State: map[string]interface{}{
					"key": "value",
				},
				Variables: variablesYAML,
			}))
		})

		Context("failure cases", func() {
			Context("when the executor's delete env call fails with delete env error", func() {
				var (
					incomingState storage.State
					expectedError bosh.ManagerDeleteError
					expectedState storage.State
				)

				BeforeEach(func() {
					incomingState = storage.State{
						IAAS: "aws",
						BOSH: storage.BOSH{
							Manifest: "some-manifest",
							State: map[string]interface{}{
								"key": "value",
							},
							Variables: variablesYAML,
						},
					}

					boshState := map[string]interface{}{
						"partial": "bosh-state",
					}
					deleteEnvError := bosh.NewDeleteEnvError(boshState, errors.New("failed to delete env"))
					boshExecutor.DeleteEnvCall.Returns.Error = deleteEnvError

					expectedState = incomingState
					expectedState.BOSH = storage.BOSH{
						Manifest:  "some-manifest",
						State:     boshState,
						Variables: variablesYAML,
					}
					expectedError = bosh.NewManagerDeleteError(expectedState, deleteEnvError)
				})

				It("returns a bosh manager delete error with a valid state", func() {
					err := boshManager.Delete(incomingState)
					Expect(err).To(MatchError(expectedError))
				})
			})

			It("returns an error when the delete env fails", func() {
				boshExecutor.DeleteEnvCall.Returns.Error = errors.New("failed to delete")
				err := boshManager.Delete(storage.State{})
				Expect(err).To(MatchError("failed to delete"))
			})
		})
	})

	Describe("GetDeploymentVars", func() {
		var (
			stackManager     *fakes.StackManager
			boshExecutor     *fakes.BOSHExecutor
			terraformManager *fakes.TerraformManager
			logger           *fakes.Logger
			socks5Proxy      *fakes.Socks5Proxy
			boshManager      bosh.Manager
		)

		BeforeEach(func() {
			terraformManager = &fakes.TerraformManager{}
			stackManager = &fakes.StackManager{}
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, terraformManager, stackManager, logger, socks5Proxy)
		})

		Context("gcp", func() {
			var (
				incomingState storage.State
			)

			BeforeEach(func() {
				incomingState = storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					KeyPair: storage.KeyPair{
						PrivateKey: "some-private-key",
					},
					GCP: storage.GCP{
						Zone:              "some-zone",
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-credential-json",
					},
					BOSH: storage.BOSH{
						State: map[string]interface{}{
							"some-key": "some-value",
						},
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "cf",
					},
				}

				terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"network_name":       "some-network",
					"subnetwork_name":    "some-subnetwork",
					"bosh_open_tag_name": "some-bosh-tag",
					"internal_tag_name":  "some-internal-tag",
					"external_ip":        "some-external-ip",
					"director_address":   "some-director-address",
				}
			})

			It("returns a correct yaml string of bosh deployment variables", func() {
				vars, err := boshManager.GetDeploymentVars(incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags: [some-bosh-tag, some-internal-tag]
project_id: some-project-id
gcp_credentials_json: 'some-credential-json'`))
			})
		})

		Context("aws", func() {
			var (
				incomingState storage.State
			)

			BeforeEach(func() {
				incomingState = storage.State{
					IAAS:  "aws",
					EnvID: "some-env-id",
					KeyPair: storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-private-key",
					},
					AWS: storage.AWS{
						Region: "some-region",
					},
					BOSH: storage.BOSH{
						State: map[string]interface{}{
							"some-key": "some-value",
						},
					},
					LB: storage.LB{
						Type: "cf",
					},
				}
			})

			Context("when terraform was used to standup infrastructure", func() {
				BeforeEach(func() {
					incomingState.TFState = "some-tf-state"
					terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
						"az":                      "some-bosh-subnet-az",
						"access_key_id":           "some-bosh-user-access-key",
						"secret_access_key":       "some-bosh-user-secret-access-key",
						"default_security_groups": "some-bosh-security-group",
						"subnet_id":               "some-bosh-subnet",
						"external_ip":             "some-bosh-elastic-ip",
						"director_address":        "some-bosh-url",
					}
				})

				It("returns a correct yaml string of bosh deployment variables", func() {
					vars, err := boshManager.GetDeploymentVars(incomingState)
					Expect(err).NotTo(HaveOccurred())
					Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-elastic-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-bosh-user-access-key
secret_access_key: some-bosh-user-secret-access-key
default_key_name: some-keypair-name
default_security_groups: [some-bosh-security-group]
region: some-region
private_key: |-
  some-private-key`))
				})
			})

			Context("when cloudformation was used to standup infrastructure", func() {
				BeforeEach(func() {
					incomingState.Stack.Name = "some-stack"
					stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
						Outputs: map[string]string{
							"BOSHSubnetAZ":            "some-bosh-subnet-az",
							"BOSHUserAccessKey":       "some-bosh-user-access-key",
							"BOSHUserSecretAccessKey": "some-bosh-user-secret-access-key",
							"BOSHSecurityGroup":       "some-bosh-security-group",
							"BOSHSubnet":              "some-bosh-subnet",
							"BOSHEIP":                 "some-bosh-elastic-ip",
							"BOSHURL":                 "some-bosh-url",
						},
					}
				})

				It("returns a correct yaml string of bosh deployment variables", func() {
					vars, err := boshManager.GetDeploymentVars(incomingState)
					Expect(err).NotTo(HaveOccurred())
					Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-elastic-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-bosh-user-access-key
secret_access_key: some-bosh-user-secret-access-key
default_key_name: some-keypair-name
default_security_groups: [some-bosh-security-group]
region: some-region
private_key: |-
  some-private-key`))
				})
			})

			Context("when the stack manager returns an error", func() {
				BeforeEach(func() {
					stackManager.DescribeCall.Returns.Error = errors.New("stack manager describe failed")
				})

				It("returns the error", func() {
					_, err := boshManager.GetDeploymentVars(storage.State{
						IAAS: "aws",
					})
					Expect(err).To(MatchError("stack manager describe failed"))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when the terraform output provider fails", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to output")
				_, err := boshManager.GetDeploymentVars(storage.State{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("failed to output"))
			})
		})
	})

	Describe("Version", func() {
		var (
			stackManager     *fakes.StackManager
			boshExecutor     *fakes.BOSHExecutor
			terraformManager *fakes.TerraformManager
			logger           *fakes.Logger
			socks5Proxy      *fakes.Socks5Proxy
			boshManager      bosh.Manager
		)

		BeforeEach(func() {
			terraformManager = &fakes.TerraformManager{}
			stackManager = &fakes.StackManager{}
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, terraformManager, stackManager, logger, socks5Proxy)

			boshExecutor.VersionCall.Returns.Version = "2.0.0"
		})

		It("calls out to bosh executor version", func() {
			version, err := boshManager.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(boshExecutor.VersionCall.CallCount).To(Equal(1))
			Expect(version).To(Equal("2.0.0"))
		})

		Context("when executor returns a bosh version error", func() {
			var expectedError bosh.BOSHVersionError
			BeforeEach(func() {
				boshExecutor.VersionCall.Returns.Version = ""
				expectedError = bosh.NewBOSHVersionError(errors.New("BOSH version could not be parsed"))
				boshExecutor.VersionCall.Returns.Error = expectedError
			})

			It("logs a warning and returns the error", func() {
				version, err := boshManager.Version()

				Expect(boshExecutor.VersionCall.CallCount).To(Equal(1))
				Expect(logger.PrintlnCall.CallCount).To(Equal(1))
				Expect(err).To(Equal(expectedError))
				Expect(version).To(Equal(""))
			})
		})

		It("returns an error when executor fails", func() {
			boshExecutor.VersionCall.Returns.Error = errors.New("failed to execute")
			_, err := boshManager.Version()
			Expect(err).To(MatchError("failed to execute"))
		})
	})
})
