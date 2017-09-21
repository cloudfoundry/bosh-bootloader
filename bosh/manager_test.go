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

var _ = Describe("Manager", func() {
	var (
		boshExecutor     *fakes.BOSHExecutor
		logger           *fakes.Logger
		socks5Proxy      *fakes.Socks5Proxy
		boshManager      *bosh.Manager
		terraformOutputs map[string]interface{}
		jumpboxVars      string
		boshVars         string

		osUnsetenvKey string
		osSetenvKey   string
		osSetenvValue string
	)

	BeforeEach(func() {
		boshExecutor = &fakes.BOSHExecutor{}
		logger = &fakes.Logger{}
		socks5Proxy = &fakes.Socks5Proxy{}
		boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy)

		boshVars = `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`
		jumpboxVars = "jumpbox_ssh:\n  private_key: some-jumpbox-private-key"

		bosh.SetOSSetenv(func(key, value string) error {
			osSetenvKey = key
			osSetenvValue = value
			return nil
		})
		bosh.SetOSUnsetenv(func(key string) error {
			osUnsetenvKey = key
			return nil
		})
	})

	AfterEach(func() {
		bosh.ResetOSSetenv()
	})

	Describe("CreateDirector", func() {
		BeforeEach(func() {
			boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: boshVars,
			}
			boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
				State: map[string]interface{}{"some-new-key": "some-new-value"}}
		})

		Context("gcp", func() {
			var incomingGCPState storage.State
			BeforeEach(func() {
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
						UserOpsFile: "some-ops-file",
					},
				}
			})

			It("generates a bosh manifest", func() {
				stateWithDirector, err := boshManager.CreateDirector(incomingGCPState, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{"creating bosh director", "created bosh director"}))

				Expect(boshExecutor.CreateEnvCall.CallCount).To(Equal(1))
				Expect(boshExecutor.DirectorInterpolateCall.Receives.InterpolateInput.DirectorDeploymentVars).To(Equal(`internal_cidr: 10.0.0.0/24
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

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(0))
				Expect(boshExecutor.JumpboxInterpolateCall.CallCount).To(Equal(0))
				Expect(stateWithDirector.BOSH).To(Equal(storage.BOSH{
					State: map[string]interface{}{
						"some-new-key": "some-new-value",
					},
					Variables:              boshVars,
					Manifest:               "some-manifest",
					DirectorName:           "bosh-some-env-id",
					DirectorAddress:        "https://10.0.0.6:25555",
					DirectorUsername:       "admin",
					DirectorPassword:       "some-admin-password",
					DirectorSSLCA:          "some-ca",
					DirectorSSLCertificate: "some-certificate",
					DirectorSSLPrivateKey:  "some-private-key",
					UserOpsFile:            "some-ops-file",
				}))
			})
		})

		Context("aws", func() {
			var incomingAWSState storage.State
			Context("when terraform was used to create infrastructure", func() {
				BeforeEach(func() {
					incomingAWSState = storage.State{
						IAAS:  "aws",
						EnvID: "some-env-id",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-region",
						},
						BOSH: storage.BOSH{
							State:       map[string]interface{}{"some-key": "some-value"},
							UserOpsFile: "some-yaml",
						},
					}
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
				})

				It("generates a bosh manifest", func() {
					stateWithDirector, err := boshManager.CreateDirector(incomingAWSState, terraformOutputs)
					Expect(err).NotTo(HaveOccurred())

					Expect(boshExecutor.DirectorInterpolateCall.Receives.InterpolateInput.DirectorDeploymentVars).To(Equal(`internal_cidr: 10.0.0.0/24
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
					Expect(stateWithDirector.BOSH).To(Equal(storage.BOSH{
						State: map[string]interface{}{
							"some-new-key": "some-new-value",
						},
						Variables:              boshVars,
						Manifest:               "some-manifest",
						DirectorName:           "bosh-some-env-id",
						DirectorAddress:        "https://10.0.0.6:25555",
						DirectorUsername:       "admin",
						DirectorPassword:       "some-admin-password",
						DirectorSSLCA:          "some-ca",
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
						UserOpsFile:            "some-yaml",
					}))
				})
			})

			Context("when the executor's create env call fails with create env error", func() {
				var expectedError bosh.ManagerCreateError

				BeforeEach(func() {
					boshState := map[string]interface{}{"partial": "bosh-state"}
					createEnvError := bosh.NewCreateEnvError(boshState, errors.New("failed to create env"))
					boshExecutor.CreateEnvCall.Returns.Error = createEnvError

					expectedState := storage.State{}
					expectedState.BOSH = storage.BOSH{
						Manifest:  "some-manifest",
						State:     boshState,
						Variables: boshVars,
					}
					expectedError = bosh.NewManagerCreateError(expectedState, createEnvError)
				})

				It("returns a bosh manager create error with a valid state", func() {
					_, err := boshManager.CreateDirector(storage.State{}, terraformOutputs)
					Expect(err).To(MatchError(expectedError))
				})
			})
		})

		Context("when an error occurs", func() {
			It("returns an error when the executor's interpolate call fails", func() {
				boshExecutor.DirectorInterpolateCall.Returns.Error = errors.New("failed to interpolate")

				_, err := boshManager.CreateDirector(storage.State{}, terraformOutputs)
				Expect(err).To(MatchError("failed to interpolate"))
			})

			It("returns an error when the executor's create env call fails with non create env error", func() {
				boshExecutor.CreateEnvCall.Returns.Error = errors.New("failed to create")

				_, err := boshManager.CreateDirector(storage.State{}, terraformOutputs)
				Expect(err).To(MatchError("failed to create"))
			})

			It("returns an error when interpolate outputs invalid yaml", func() {
				boshExecutor.DirectorInterpolateCall.Returns.Output.Variables = "%%%"

				_, err := boshManager.CreateDirector(storage.State{}, terraformOutputs)
				Expect(err).To(MatchError("failed to get director outputs:\nyaml: could not find expected directive name"))
			})
		})
	})

	Describe("CreateJumpbox", func() {
		var (
			jumpboxDeploymentVars string
			deploymentVars        string
			incomingGCPState      storage.State
		)

		BeforeEach(func() {
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
				Variables: boshVars,
			}
		})

		AfterEach(func() {
			bosh.ResetOSSetenv()
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
				"started proxy",
			}))
		})

		Context("when bosh director is created after jumpbox", func() {
			It("returns a bbl state with bosh and jumpbox deployment values", func() {
				boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
					State: map[string]interface{}{
						"some-new-key": "some-new-value",
					},
				}

				afterJumpboxState, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				state, err := boshManager.CreateDirector(afterJumpboxState, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshExecutor.DirectorInterpolateCall.Receives.InterpolateInput.JumpboxDeploymentVars).To(Equal(jumpboxDeploymentVars))
				Expect(boshExecutor.DirectorInterpolateCall.Receives.InterpolateInput.DirectorDeploymentVars).To(Equal(deploymentVars))

				Expect(state).To(Equal(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						Zone:              "some-zone",
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-credential-json",
					},
					Jumpbox: storage.Jumpbox{
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
						Variables:              boshVars,
						Manifest:               "some-manifest",
						DirectorName:           "bosh-some-env-id",
						DirectorAddress:        "https://10.0.0.6:25555",
						DirectorUsername:       "admin",
						DirectorPassword:       "some-admin-password",
						DirectorSSLCA:          "some-ca",
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
					},
				}))
			})
		})

		Context("when an error occurs", func() {
			Context("when the jumpbox variables cannot be parsed", func() {
				It("returns an error", func() {
					boshExecutor.JumpboxInterpolateCall.Returns.Output.Variables = "%%%"

					_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
					Expect(err).To(MatchError("jumpbox key: yaml: could not find expected directive name"))
				})
			})

			Context("when create env returns a typed error", func() {
				It("returns an error", func() {
					boshExecutor.CreateEnvCall.Returns.Error = bosh.NewCreateEnvError(map[string]interface{}{"foo": "bar"}, errors.New("apple"))

					_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
					Expect(err).To(MatchError("create env error: apple"))
				})
			})

			Context("when create env returns an untyped error", func() {
				It("returns an error", func() {
					boshExecutor.CreateEnvCall.Returns.Error = errors.New("banana")

					_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
					Expect(err).To(MatchError("create env: banana"))
				})
			})

			Context("when the socks5 proxy fails to start", func() {
				It("returns an error", func() {
					socks5Proxy.StartCall.Returns.Error = errors.New("coconut")

					_, err := boshManager.CreateJumpbox(incomingGCPState, terraformOutputs)
					Expect(err).To(MatchError("start proxy: coconut"))
				})
			})
		})
	})

	Describe("DeleteJumpbox", func() {
		var (
			vars          string
			incomingState storage.State
		)
		BeforeEach(func() {
			vars = `jumpbox_ssh:
  private_key: some-private-key
  public_key: some-private-key
`
			incomingState = storage.State{
				Jumpbox: storage.Jumpbox{
					Manifest: "some-manifest",
					State: map[string]interface{}{
						"key": "value",
					},
					Variables: vars,
				},
			}
		})

		It("calls delete env", func() {
			boshExecutor.JumpboxInterpolateCall.Returns.Output = bosh.JumpboxInterpolateOutput{
				Manifest:  "some-manifest",
				Variables: vars,
			}

			err := boshManager.DeleteJumpbox(incomingState, map[string]interface{}{"jumpbox_ssh": "nick-da-quick"})
			Expect(err).NotTo(HaveOccurred())
			Expect(boshExecutor.DeleteEnvCall.Receives.Input.Variables).To(Equal(vars))
		})

		Context("when an error occurs", func() {
			Context("when the executor's delete env call fails with delete env error", func() {
				var (
					expectedError bosh.ManagerDeleteError
					expectedState storage.State
				)

				BeforeEach(func() {
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
					Jumpbox: storage.Jumpbox{},
				}, map[string]interface{}{"director_address": "nick-da-quick"})
				Expect(err).To(MatchError("failed to delete"))
			})
		})
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			boshExecutor.DirectorInterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: boshVars,
			}
		})
		It("calls delete env", func() {
			socks5ProxyAddr := "localhost:1234"
			socks5Proxy.AddrCall.Returns.Addr = socks5ProxyAddr

			err := boshManager.Delete(storage.State{
				Jumpbox: storage.Jumpbox{
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
					Variables: boshVars,
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
				Variables: boshVars,
			}))
		})

		Context("when an error occurs", func() {
			var incomingState storage.State

			BeforeEach(func() {
				incomingState = storage.State{
					BOSH: storage.BOSH{
						Manifest: "some-manifest",
						State: map[string]interface{}{
							"key": "value",
						},
						Variables: boshVars,
					},
					Jumpbox: storage.Jumpbox{
						Variables: jumpboxVars,
					},
				}
			})
			Context("when the executor's delete env call fails with delete env error", func() {
				It("returns a bosh manager delete error with a valid state", func() {
					boshState := map[string]interface{}{"partial": "bosh-state"}
					deleteEnvError := bosh.NewDeleteEnvError(boshState, errors.New("failed to delete env"))
					boshExecutor.DeleteEnvCall.Returns.Error = deleteEnvError

					expectedState := incomingState
					expectedState.BOSH = storage.BOSH{
						Manifest:  "some-manifest",
						State:     boshState,
						Variables: boshVars,
					}
					expectedError := bosh.NewManagerDeleteError(expectedState, deleteEnvError)
					err := boshManager.Delete(incomingState, map[string]interface{}{"director_address": "nick-da-quick"})
					Expect(err).To(MatchError(expectedError))
				})
			})

			It("returns an error when the delete env fails", func() {
				boshExecutor.DeleteEnvCall.Returns.Error = errors.New("failed to delete")

				err := boshManager.Delete(incomingState, map[string]interface{}{"director_address": "nick-da-quick"})
				Expect(err).To(MatchError("failed to delete"))
			})

			Context("when a jumpbox deployment exists", func() {
				It("returns an error when the socks5Proxy fails to start", func() {
					socks5Proxy.StartCall.Returns.Error = errors.New("failed to start socks5Proxy")

					err := boshManager.Delete(incomingState, map[string]interface{}{"director_address": "nick-da-quick"})
					Expect(err).To(MatchError("failed to start socks5Proxy"))
				})
			})
		})
	})

	Describe("GetJumpboxDeploymentVars", func() {
		Context("aws", func() {
			var incomingState storage.State
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS:    "aws",
					Jumpbox: storage.Jumpbox{},
					EnvID:   "some-env-id",
					AWS: storage.AWS{
						Region:          "some-region",
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
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
					IAAS:    "gcp",
					EnvID:   "some-env-id",
					Jumpbox: storage.Jumpbox{},
					GCP: storage.GCP{
						Zone:              "some-zone",
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-credential-json",
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

	Describe("GetDirectorDeploymentVars", func() {
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
				}
			})
			It("returns a correct yaml string of bosh deployment variables", func() {
				vars := boshManager.GetDirectorDeploymentVars(incomingState, map[string]interface{}{
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

			Context("when terraform outputs are missing", func() {
				It("returns valid yaml", func() {
					vars := boshManager.GetDirectorDeploymentVars(incomingState, map[string]interface{}{})
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
		})

		Context("aws", func() {
			var incomingState storage.State

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
						State: map[string]interface{}{"some-key": "some-value"},
					},
				}
			})

			Context("when terraform was used to standup infrastructure", func() {
				It("returns a correct yaml string of bosh deployment variables", func() {
					vars := boshManager.GetDirectorDeploymentVars(incomingState, map[string]interface{}{
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
					vars := boshManager.GetDirectorDeploymentVars(incomingState, map[string]interface{}{})
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

	Describe("Version", func() {
		BeforeEach(func() {
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
