package bosh_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	Describe("CreateDirector", func() {
		var (
			boshExecutor     *fakes.BOSHExecutor
			logger           *fakes.Logger
			socks5Proxy      *fakes.Socks5Proxy
			boshManager      *bosh.Manager
			incomingGCPState storage.State
			terraformOutputs map[string]interface{}

			osUnsetenvKey string
			osSetenvKey   string
			osSetenvValue string
		)

		BeforeEach(func() {
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy)

			bosh.SetOSSetenv(func(key, value string) error {
				osSetenvKey = key
				osSetenvValue = value
				return nil
			})

			bosh.SetOSUnsetenv(func(key string) error {
				osUnsetenvKey = key
				return nil
			})

			terraformOutputs = map[string]interface{}{
				"network_name":           "some-network",
				"subnetwork_name":        "some-subnetwork",
				"bosh_open_tag_name":     "some-jumpbox-tag",
				"bosh_director_tag_name": "some-director-tag",
				"internal_tag_name":      "some-internal-tag",
				"external_ip":            "some-external-ip",
				"director_address":       "some-director-address",
				"jumpbox_url":            "some-jumpbox-url",
			}

			incomingGCPState = storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				GCP: storage.GCP{
					Zone:              "some-zone",
					ProjectID:         "some-project-id",
					ServiceAccountKey: "some-credential-json",
				},
				BOSH: storage.BOSH{
					State: map[string]interface{}{
						"some-key": "some-value",
					},
					UserOpsFile: "some-yaml",
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
			boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: variablesYAML,
			}

			_, err := boshManager.CreateDirector(incomingGCPState, terraformOutputs)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{"creating bosh director", "created bosh director"}))
		})

		Context("when iaas is gcp", func() {
			It("generates a bosh manifest", func() {
				boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
					Manifest:  "some-manifest",
					Variables: variablesYAML,
				}

				boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
					State: map[string]interface{}{
						"some-new-key": "some-new-value",
					},
				}

				incomingGCPState.BOSH.UserOpsFile = "some-ops-file"
				_, err := boshManager.CreateDirector(incomingGCPState, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshExecutor.CreateEnvCall.CallCount).To(Equal(1))
				Expect(boshExecutor.DirectorInterpolateCall.Receives.InterpolateInput).To(Equal(bosh.InterpolateInput{
					IAAS: "gcp",
					DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags:
- some-director-tag
- some-jumpbox-tag
project_id: some-project-id
gcp_credentials_json: some-credential-json
`,
					Variables: "",
					OpsFile:   "some-ops-file",
				}))

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(0))
				Expect(boshExecutor.JumpboxInterpolateCall.CallCount).To(Equal(0))
			})

			It("returns a state with a proper bosh state", func() {
				boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
					Manifest:  "some-manifest",
					Variables: variablesYAML,
				}

				boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
					State: map[string]interface{}{
						"some-new-key": "some-new-value",
					},
				}

				state, err := boshManager.CreateDirector(incomingGCPState, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
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
						UserOpsFile:            "some-yaml",
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "cf",
					},
				}))
			})
		})

		Context("when iaas is aws", func() {
			incomingAWSState := storage.State{
				IAAS:  "aws",
				EnvID: "some-env-id",
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				},
				BOSH: storage.BOSH{
					State: map[string]interface{}{
						"some-key": "some-value",
					},
					UserOpsFile: "some-yaml",
				},
				TFState: "some-tf-state",
				LB: storage.LB{
					Type: "cf",
				},
			}
			Context("when terraform was used to create infrastructure", func() {
				BeforeEach(func() {
					terraformOutputs = map[string]interface{}{
						"bosh_iam_instance_profile":     "some-bosh-iam-instance-profile",
						"bosh_subnet_availability_zone": "some-bosh-subnet-az",
						"bosh_security_group":           "some-bosh-security-group",
						"bosh_subnet_id":                "some-bosh-subnet",
						"external_ip":                   "some-bosh-external-ip",
						"director_address":              "some-director-address",
						"bosh_vms_key_name":             "some-keypair-name",
						"bosh_vms_private_key":          "some-private-key",
						"kms_key_arn":                   "some-kms-arn",
					}

					boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
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
					_, err := boshManager.CreateDirector(incomingAWSState, terraformOutputs)
					Expect(err).NotTo(HaveOccurred())

					Expect(boshExecutor.DirectorInterpolateCall.Receives.InterpolateInput).To(Equal(bosh.InterpolateInput{
						IAAS: "aws",
						DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-external-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-access-key-id
secret_access_key: some-secret-access-key
iam_instance_profile: some-bosh-iam-instance-profile
default_key_name: some-keypair-name
default_security_groups:
- some-bosh-security-group
region: some-region
private_key: some-private-key
kms_key_arn: some-kms-arn
`,
						Variables: "",
						OpsFile:   "some-yaml",
					}))
				})

				It("returns a state with a proper bosh state", func() {
					state, err := boshManager.CreateDirector(incomingAWSState, terraformOutputs)
					Expect(err).NotTo(HaveOccurred())

					Expect(state).To(Equal(storage.State{
						IAAS:  "aws",
						EnvID: "some-env-id",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-region",
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
							UserOpsFile:            "some-yaml",
						},
						TFState: "some-tf-state",
						LB: storage.LB{
							Type: "cf",
						},
					}))
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

					boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
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
					_, err := boshManager.CreateDirector(incomingAWSState, terraformOutputs)
					Expect(err).To(MatchError(expectedError))
				})
			})
		})

		It("creates a bosh environment", func() {
			boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: variablesYAML,
			}

			_, err := boshManager.CreateDirector(incomingGCPState, terraformOutputs)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshExecutor.CreateEnvCall.Receives.Input).To(Equal(bosh.CreateEnvInput{
				Manifest: "some-manifest",
				State: map[string]interface{}{
					"some-key": "some-value",
				},
				Variables: variablesYAML,
			}))
		})

		Context("when an error occurs", func() {
			It("returns an error when the executor's interpolate call fails", func() {
				boshExecutor.DirectorInterpolateCall.Returns.Error = errors.New("failed to interpolate")

				_, err := boshManager.CreateDirector(incomingGCPState, terraformOutputs)
				Expect(err).To(MatchError("failed to interpolate"))
			})

			It("returns an error when the executor's create env call fails with non create env error", func() {
				boshExecutor.CreateEnvCall.Returns.Error = errors.New("failed to create")

				_, err := boshManager.CreateDirector(incomingGCPState, terraformOutputs)
				Expect(err).To(MatchError("failed to create"))
			})

			Context("when interpolate outputs invalid yaml", func() {
				It("returns an error", func() {
					boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
						Manifest:  "some-manifest",
						Variables: "%%%",
					}

					_, err := boshManager.CreateDirector(storage.State{IAAS: "aws"}, terraformOutputs)
					Expect(err).To(MatchError("failed to get director outputs:\nyaml: could not find expected directive name"))
				})
			})
		})
	})

	Describe("CreateJumpbox", func() {
		var (
			boshExecutor     *fakes.BOSHExecutor
			logger           *fakes.Logger
			socks5Proxy      *fakes.Socks5Proxy
			boshManager      *bosh.Manager
			incomingGCPState storage.State
			terraformOutputs map[string]interface{}

			jumpboxDeploymentVars string
			deploymentVars        string

			osUnsetenvKey string
			osSetenvKey   string
			osSetenvValue string
		)

		BeforeEach(func() {
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy)

			bosh.SetOSSetenv(func(key, value string) error {
				osSetenvKey = key
				osSetenvValue = value
				return nil
			})

			bosh.SetOSUnsetenv(func(key string) error {
				osUnsetenvKey = key
				return nil
			})

			terraformOutputs = map[string]interface{}{
				"network_name":           "some-network",
				"subnetwork_name":        "some-subnetwork",
				"bosh_open_tag_name":     "some-jumpbox-tag",
				"bosh_director_tag_name": "some-director-tag",
				"internal_tag_name":      "some-internal-tag",
				"external_ip":            "some-external-ip",
				"director_address":       "some-director-address",
				"jumpbox_url":            "some-jumpbox-url",
			}

			incomingGCPState = storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
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
tags:
- some-jumpbox-tag
project_id: some-project-id
gcp_credentials_json: some-credential-json
`

			deploymentVars = `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags:
- some-director-tag
project_id: some-project-id
gcp_credentials_json: some-credential-json
`

			boshExecutor.JumpboxInterpolateCall.Returns.Output = bosh.JumpboxInterpolateOutput{
				Manifest:  "name: jumpbox",
				Variables: "jumpbox_ssh:\n  private_key: some-jumpbox-private-key",
			}

			boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: variablesYAML,
			}
		})

		AfterEach(func() {
			bosh.ResetOSSetenv()
		})

		It("logs jumpbox status messages", func() {
			_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{"creating jumpbox", "created jumpbox"}))
		})

		It("starts a socks5 proxy for the duration of creating the bosh director", func() {
			socks5ProxyAddr := "localhost:1234"
			socks5Proxy.AddrCall.Returns.Addr = socks5ProxyAddr

			_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
			Expect(err).NotTo(HaveOccurred())

			Expect(osUnsetenvKey).To(Equal("BOSH_ALL_PROXY"))
			Expect(socks5Proxy.StartCall.CallCount).To(Equal(1))
			Expect(socks5Proxy.StartCall.Receives.JumpboxPrivateKey).To(Equal("some-jumpbox-private-key"))
			Expect(socks5Proxy.StartCall.Receives.JumpboxExternalURL).To(Equal("some-jumpbox-url"))
			Expect(osSetenvKey).To(Equal("BOSH_ALL_PROXY"))
			Expect(osSetenvValue).To(Equal(fmt.Sprintf("socks5://%s", socks5ProxyAddr)))

			Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
				"creating jumpbox",
				"created jumpbox",
				"starting socks5 proxy to jumpbox",
			}))
		})

		Context("when bosh director is created after jumpbox", func() {
			It("generates a jumpbox and bosh manifest", func() {
				afterJumpboxState, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				_, err = boshManager.CreateDirector(afterJumpboxState, terraformOutputs)

				Expect(boshExecutor.DirectorInterpolateCall.Receives.InterpolateInput).To(Equal(bosh.InterpolateInput{
					IAAS: "gcp",
					JumpboxDeploymentVars: jumpboxDeploymentVars,
					DeploymentVars:        deploymentVars,
					Variables:             "",
				}))
			})

			It("returns a bbl state with a proper jumpbox state", func() {
				boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
					State: map[string]interface{}{
						"some-new-key": "some-new-value",
					},
				}

				afterJumpboxState, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				state, err := boshManager.CreateDirector(afterJumpboxState, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						Zone:              "some-zone",
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-credential-json",
					},
					Jumpbox: storage.Jumpbox{
						Enabled:   true,
						URL:       "some-jumpbox-url",
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
		})

		Context("when an error occurs", func() {
			Context("when the jumpbox variables cannot be parsed", func() {
				BeforeEach(func() {
					boshExecutor.JumpboxInterpolateCall.Returns.Output.Variables = "%%%"
				})

				It("returns an error", func() {
					_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
					Expect(err).To(MatchError("jumpbox key: yaml: could not find expected directive name"))
				})
			})

			Context("when create env returns a typed error", func() {
				BeforeEach(func() {
					boshState := make(map[string]interface{})
					boshState["foo"] = "bar"
					boshExecutor.CreateEnvCall.Returns.Error = bosh.NewCreateEnvError(boshState, errors.New("apple"))
				})

				It("returns an error", func() {
					_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
					Expect(err).To(MatchError("create env error: apple"))
				})
			})

			Context("when create env returns an untyped error", func() {
				BeforeEach(func() {
					boshExecutor.CreateEnvCall.Returns.Error = errors.New("banana")
				})

				It("returns an error", func() {
					_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
					Expect(err).To(MatchError("create env: banana"))
				})
			})

			It("returns an error when the socks5Proxy fails to start", func() {
				socks5Proxy.StartCall.Returns.Error = errors.New("coconut")

				_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
				Expect(err).To(MatchError("start proxy: coconut"))
			})
		})
	})

	Describe("DeleteJumpbox", func() {
		var (
			boshExecutor *fakes.BOSHExecutor
			logger       *fakes.Logger
			socks5Proxy  *fakes.Socks5Proxy
			boshManager  *bosh.Manager

			vars string
		)

		BeforeEach(func() {
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy)

			vars = `jumpbox_ssh:
  private_key: some-private-key
  public_key: some-private-key
`
		})

		It("calls delete env", func() {
			boshExecutor.JumpboxInterpolateCall.Returns.Output = bosh.JumpboxInterpolateOutput{
				Manifest:  "some-manifest",
				Variables: vars,
			}

			err := boshManager.DeleteJumpbox(storage.State{
				IAAS: "gcp",
				Jumpbox: storage.Jumpbox{
					Enabled:  true,
					Manifest: "some-manifest",
					State: map[string]interface{}{
						"key": "value",
					},
					Variables: vars,
				},
			}, map[string]interface{}{"jumpbox_ssh": "nick-da-quick"})
			Expect(err).NotTo(HaveOccurred())
			Expect(boshExecutor.DeleteEnvCall.Receives.Input).To(Equal(bosh.DeleteEnvInput{
				Manifest: "some-manifest",
				State: map[string]interface{}{
					"key": "value",
				},
				Variables: vars,
			}))
		})

		Context("when an error occurs", func() {
			Context("when the executor's delete env call fails with delete env error", func() {
				var (
					incomingState storage.State
					expectedError bosh.ManagerDeleteError
					expectedState storage.State
				)

				BeforeEach(func() {
					incomingState = storage.State{
						IAAS: "gcp",
						Jumpbox: storage.Jumpbox{
							Enabled:  true,
							Manifest: "some-manifest",
							State: map[string]interface{}{
								"key": "value",
							},
							Variables: vars,
						},
					}

					jumpboxState := map[string]interface{}{
						"partial": "jumpbox-state",
					}
					deleteEnvError := bosh.NewDeleteEnvError(jumpboxState, errors.New("failed to delete env"))
					boshExecutor.DeleteEnvCall.Returns.Error = deleteEnvError

					expectedState = incomingState
					expectedState.Jumpbox.State = jumpboxState
					expectedError = bosh.NewManagerDeleteError(expectedState, deleteEnvError)
				})

				It("returns a bosh manager delete error with a valid state", func() {
					err := boshManager.DeleteJumpbox(incomingState, map[string]interface{}{
						"director_address": "nick-da-quick",
					})
					Expect(err).To(MatchError(expectedError))
				})
			})

			It("returns an error when the delete env fails", func() {
				boshExecutor.DeleteEnvCall.Returns.Error = errors.New("failed to delete")

				err := boshManager.DeleteJumpbox(storage.State{
					IAAS: "gcp",
					Jumpbox: storage.Jumpbox{
						Enabled: true,
					},
				}, map[string]interface{}{"director_address": "nick-da-quick"})
				Expect(err).To(MatchError("failed to delete"))
			})
		})
	})

	Describe("Delete", func() {
		var (
			boshExecutor *fakes.BOSHExecutor
			logger       *fakes.Logger
			socks5Proxy  *fakes.Socks5Proxy
			boshManager  *bosh.Manager

			osSetenvKey   string
			osSetenvValue string
		)

		BeforeEach(func() {
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy)

			bosh.SetOSSetenv(func(key, value string) error {
				osSetenvKey = key
				osSetenvValue = value
				return nil
			})
		})

		It("calls delete env", func() {
			boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: variablesYAML,
			}

			err := boshManager.Delete(storage.State{
				IAAS: "aws",
				BOSH: storage.BOSH{
					Manifest: "some-manifest",
					State: map[string]interface{}{
						"key": "value",
					},
					Variables: variablesYAML,
				},
			}, map[string]interface{}{"director_address": "nick-da-quick"})
			Expect(err).NotTo(HaveOccurred())
			Expect(boshExecutor.DeleteEnvCall.Receives.Input).To(Equal(bosh.DeleteEnvInput{
				Manifest: "some-manifest",
				State: map[string]interface{}{
					"key": "value",
				},
				Variables: variablesYAML,
			}))
		})

		Context("when a jumbox deployment exists", func() {
			It("starts a socks5 proxy and gets the jumpbox deployment vars", func() {
				socks5ProxyAddr := "localhost:1234"
				socks5Proxy.AddrCall.Returns.Addr = socks5ProxyAddr

				boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
					Manifest:  "some-manifest",
					Variables: variablesYAML,
				}

				err := boshManager.Delete(storage.State{
					IAAS: "gcp",
					Jumpbox: storage.Jumpbox{
						Enabled:   true,
						Variables: "jumpbox_ssh:\n  private_key: some-jumpbox-private-key",
						Manifest:  "name: jumpbox",
						State: map[string]interface{}{
							"some-key": "some-value",
						},
						URL: "some-jumpbox-url",
					},
					BOSH: storage.BOSH{
						Manifest: "some-manifest",
						State: map[string]interface{}{
							"key": "value",
						},
						Variables: variablesYAML,
					},
				}, map[string]interface{}{"director_address": "nick-da-quick"})
				Expect(err).NotTo(HaveOccurred())

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(1))
				Expect(socks5Proxy.StartCall.Receives.JumpboxPrivateKey).To(Equal("some-jumpbox-private-key"))
				Expect(socks5Proxy.StartCall.Receives.JumpboxExternalURL).To(Equal("some-jumpbox-url"))
				Expect(osSetenvKey).To(Equal("BOSH_ALL_PROXY"))
				Expect(osSetenvValue).To(Equal(fmt.Sprintf("socks5://%s", socks5ProxyAddr)))

				Expect(boshExecutor.DeleteEnvCall.Receives.Input).To(Equal(bosh.DeleteEnvInput{
					Manifest: "some-manifest",
					State: map[string]interface{}{
						"key": "value",
					},
					Variables: variablesYAML,
				}))
			})
		})

		Context("when an error occurs", func() {
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
					err := boshManager.Delete(incomingState, map[string]interface{}{"director_address": "nick-da-quick"})
					Expect(err).To(MatchError(expectedError))
				})
			})

			It("returns an error when the delete env fails", func() {
				boshExecutor.DeleteEnvCall.Returns.Error = errors.New("failed to delete")

				err := boshManager.Delete(storage.State{
					IAAS: "aws",
				}, map[string]interface{}{"director_address": "nick-da-quick"})
				Expect(err).To(MatchError("failed to delete"))
			})

			Context("when a jumpbox deployment exists", func() {
				It("returns an error when the socks5Proxy fails to start", func() {
					socks5Proxy.StartCall.Returns.Error = errors.New("failed to start socks5Proxy")

					err := boshManager.Delete(storage.State{
						IAAS: "gcp",
						Jumpbox: storage.Jumpbox{
							Enabled:   true,
							Variables: "jumpbox_ssh:\n  private_key: some-jumpbox-private-key",
						},
					}, map[string]interface{}{"director_address": "nick-da-quick"})
					Expect(err).To(MatchError("failed to start socks5Proxy"))
				})
			})
		})
	})

	Describe("GetJumpboxDeploymentVars", func() {
		var (
			boshExecutor *fakes.BOSHExecutor
			logger       *fakes.Logger
			socks5Proxy  *fakes.Socks5Proxy
			boshManager  *bosh.Manager
		)

		BeforeEach(func() {
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy)
		})

		Context("aws", func() {
			var incomingState storage.State
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS: "aws",
					Jumpbox: storage.Jumpbox{
						Enabled: true,
					},
					EnvID: "some-env-id",
					AWS: storage.AWS{
						Region:          "some-region",
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "cf",
					},
				}
			})

			It("returns a correct yaml string of bosh deployment variables", func() {
				vars := boshManager.GetJumpboxDeploymentVars(incomingState, map[string]interface{}{
					"network_name":                  "some-network",
					"bosh_subnet_id":                "some-subnetwork",
					"bosh_subnet_availability_zone": "some-zone",
					"bosh_iam_instance_profile":     "some-instance-profile",
					"bosh_vms_key_name":             "some-key-name",
					"bosh_vms_private_key":          "some-private-key",
					"jumpbox_security_group":        "some-security-group",
					"external_ip":                   "some-external-ip",
				})
				Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.5
director_name: bosh-some-env-id
external_ip: some-external-ip
az: some-zone
subnet_id: some-subnetwork
access_key_id: some-access-key-id
secret_access_key: some-secret-access-key
iam_instance_profile: some-instance-profile
default_key_name: some-key-name
default_security_groups:
- some-security-group
region: some-region
private_key: some-private-key
`))
			})
		})

		Context("gcp", func() {
			var incomingState storage.State
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					Jumpbox: storage.Jumpbox{
						Enabled: true,
					},
					GCP: storage.GCP{
						Zone:              "some-zone",
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-credential-json",
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "cf",
					},
				}
			})

			It("returns a correct yaml string of bosh deployment variables", func() {
				vars := boshManager.GetJumpboxDeploymentVars(incomingState, map[string]interface{}{
					"network_name":       "some-network",
					"subnetwork_name":    "some-subnetwork",
					"bosh_open_tag_name": "some-jumpbox-tag",
					"external_ip":        "some-external-ip",
				})
				Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.5
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags:
- some-jumpbox-tag
project_id: some-project-id
gcp_credentials_json: some-credential-json
`))
			})
		})
	})

	Describe("GetDeploymentVars", func() {
		var (
			boshExecutor *fakes.BOSHExecutor
			logger       *fakes.Logger
			socks5Proxy  *fakes.Socks5Proxy
			boshManager  *bosh.Manager
		)

		BeforeEach(func() {
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy)
		})

		Context("gcp", func() {
			var incomingState storage.State
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
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
			})
			It("returns a correct yaml string of bosh deployment variables", func() {
				vars := boshManager.GetDeploymentVars(incomingState, map[string]interface{}{
					"network_name":           "some-network",
					"subnetwork_name":        "some-subnetwork",
					"bosh_open_tag_name":     "some-jumpbox-tag",
					"bosh_director_tag_name": "some-director-tag",
					"internal_tag_name":      "some-internal-tag",
					"external_ip":            "some-external-ip",
					"director_address":       "some-director-address",
				})
				Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags:
- some-director-tag
- some-jumpbox-tag
project_id: some-project-id
gcp_credentials_json: some-credential-json
`))
			})

			Context("when using a jumpbox", func() {
				BeforeEach(func() {
					incomingState.Jumpbox.Enabled = true
				})
				It("returns a correct yaml string of bosh deployment variables", func() {
					vars := boshManager.GetDeploymentVars(incomingState, map[string]interface{}{
						"network_name":           "some-network",
						"subnetwork_name":        "some-subnetwork",
						"bosh_open_tag_name":     "some-jumpbox-tag",
						"bosh_director_tag_name": "some-director-tag",
						"internal_tag_name":      "some-internal-tag",
						"external_ip":            "some-external-ip",
						"director_address":       "some-director-address",
					})
					Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags:
- some-director-tag
project_id: some-project-id
gcp_credentials_json: some-credential-json
`))
				})
			})

			Context("when terraform outputs are missing", func() {
				Context("gcp jumpbox", func() {
					BeforeEach(func() {
						incomingState.Jumpbox.Enabled = true
					})
					It("returns valid yaml", func() {
						vars := boshManager.GetDeploymentVars(incomingState, map[string]interface{}{})
						Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
zone: some-zone
tags:
- ""
project_id: some-project-id
gcp_credentials_json: some-credential-json
`))
					})
				})

				Context("gcp", func() {
					It("returns valid yaml", func() {
						vars := boshManager.GetDeploymentVars(incomingState, map[string]interface{}{})
						Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
zone: some-zone
tags:
- ""
- ""
project_id: some-project-id
gcp_credentials_json: some-credential-json
`))
					})
				})
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
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
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
				})

				It("returns a correct yaml string of bosh deployment variables", func() {
					vars := boshManager.GetDeploymentVars(incomingState, map[string]interface{}{
						"bosh_iam_instance_profile":     "some-bosh-iam-instance-profile",
						"bosh_subnet_availability_zone": "some-bosh-subnet-az",
						"bosh_security_group":           "some-bosh-security-group",
						"bosh_subnet_id":                "some-bosh-subnet",
						"bosh_vms_key_name":             "some-keypair-name",
						"bosh_vms_private_key":          "some-private-key",
						"external_ip":                   "some-bosh-external-ip",
						"director_address":              "some-director-address",
						"kms_key_arn":                   "some-kms-arn",
					})
					Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-external-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-access-key-id
secret_access_key: some-secret-access-key
iam_instance_profile: some-bosh-iam-instance-profile
default_key_name: some-keypair-name
default_security_groups:
- some-bosh-security-group
region: some-region
private_key: some-private-key
kms_key_arn: some-kms-arn
`))
				})

				Context("when using a jumpbox", func() {
					BeforeEach(func() {
						incomingState.TFState = "some-tf-state"
						incomingState.Jumpbox.Enabled = true
					})

					It("returns a correct yaml string of bosh deployment variables", func() {
						vars := boshManager.GetDeploymentVars(incomingState, map[string]interface{}{
							"bosh_iam_instance_profile":     "some-bosh-iam-instance-profile",
							"bosh_subnet_availability_zone": "some-bosh-subnet-az",
							"bosh_security_group":           "some-bosh-security-group",
							"bosh_subnet_id":                "some-bosh-subnet",
							"bosh_vms_key_name":             "some-keypair-name",
							"bosh_vms_private_key":          "some-private-key",
							"external_ip":                   "some-bosh-external-ip",
							"director_address":              "some-director-address",
							"kms_key_arn":                   "some-kms-arn",
						})
						Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-access-key-id
secret_access_key: some-secret-access-key
iam_instance_profile: some-bosh-iam-instance-profile
default_key_name: some-keypair-name
default_security_groups:
- some-bosh-security-group
region: some-region
private_key: some-private-key
kms_key_arn: some-kms-arn
`))
					})
				})

				Context("when terraform outputs are missing", func() {
					It("returns valid yaml", func() {
						vars := boshManager.GetDeploymentVars(incomingState, map[string]interface{}{})
						Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
access_key_id: some-access-key-id
secret_access_key: some-secret-access-key
default_security_groups:
- ""
region: some-region
`))
					})
				})
			})
		})
	})

	Describe("Version", func() {
		var (
			boshExecutor *fakes.BOSHExecutor
			logger       *fakes.Logger
			socks5Proxy  *fakes.Socks5Proxy
			boshManager  *bosh.Manager
		)

		BeforeEach(func() {
			boshExecutor = &fakes.BOSHExecutor{}
			logger = &fakes.Logger{}
			socks5Proxy = &fakes.Socks5Proxy{}
			boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy)

			boshExecutor.VersionCall.Returns.Version = "2.0.24"
		})

		It("calls out to bosh executor version", func() {
			version, err := boshManager.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(boshExecutor.VersionCall.CallCount).To(Equal(1))
			Expect(version).To(Equal("2.0.24"))
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
