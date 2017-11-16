package bosh_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	var (
		boshExecutor *fakes.BOSHExecutor
		logger       *fakes.Logger
		socks5Proxy  *fakes.Socks5Proxy
		stateStore   *fakes.StateStore
		sshKeyGetter *fakes.SSHKeyGetter

		boshManager      *bosh.Manager
		terraformOutputs terraform.Outputs
		boshVars         string

		osUnsetenvKey string
		osSetenvKey   string
		osSetenvValue string
	)

	BeforeEach(func() {
		boshExecutor = &fakes.BOSHExecutor{}
		logger = &fakes.Logger{}
		socks5Proxy = &fakes.Socks5Proxy{}
		sshKeyGetter = &fakes.SSHKeyGetter{}
		sshKeyGetter.GetCall.Returns.PrivateKey = "some-jumpbox-private-key"

		stateStore = &fakes.StateStore{}
		stateStore.GetVarsDirCall.Returns.Directory = "some-bbl-vars-dir"
		stateStore.GetStateDirCall.Returns.Directory = "some-state-dir"
		stateStore.GetDirectorDeploymentDirCall.Returns.Directory = "some-director-deployment-dir"
		stateStore.GetJumpboxDeploymentDirCall.Returns.Directory = "some-jumpbox-deployment-dir"

		boshManager = bosh.NewManager(boshExecutor, logger, socks5Proxy, stateStore, sshKeyGetter)

		boshVars = `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`

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

	Describe("Director set-up", func() {
		var state storage.State
		BeforeEach(func() {
			state = storage.State{
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

			boshExecutor.CreateEnvCall.Returns.Variables = boshVars
		})

		Describe("InitializeDirector", func() {
			It("Calls DirectorCreateEnvArgs", func() {
				err := boshManager.InitializeDirector(state)
				Expect(err).NotTo(HaveOccurred())
				Expect(boshExecutor.DirectorCreateEnvArgsCall.Receives.InterpolateInput.VarsDir).To(Equal("some-bbl-vars-dir"))
				Expect(boshExecutor.DirectorCreateEnvArgsCall.Receives.InterpolateInput.StateDir).To(Equal("some-state-dir"))
				Expect(boshExecutor.DirectorCreateEnvArgsCall.Receives.InterpolateInput.DeploymentDir).To(Equal("some-director-deployment-dir"))
				Expect(boshExecutor.JumpboxCreateEnvArgsCall.CallCount).To(Equal(0))

				Expect(boshExecutor.CreateEnvCall.CallCount).To(Equal(0))
			})

			Context("when create env args fails", func() {
				BeforeEach(func() {
					boshExecutor.DirectorCreateEnvArgsCall.Returns.Error = errors.New("failed to interpolate")
				})

				It("returns an error", func() {
					err := boshManager.InitializeDirector(storage.State{})
					Expect(err).To(MatchError("failed to interpolate"))
				})
			})

			Context("when get vars dir fails", func() {
				It("returns an error", func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("pineapple")

					err := boshManager.InitializeDirector(storage.State{})
					Expect(err).To(MatchError("Get vars dir: pineapple"))
				})
			})

			Context("when get deployment dir fails", func() {
				It("returns an error", func() {
					stateStore.GetDirectorDeploymentDirCall.Returns.Error = errors.New("pineapple")

					err := boshManager.InitializeDirector(storage.State{})
					Expect(err).To(MatchError("Get deployment dir: pineapple"))
				})
			})
		})

		Describe("CreateDirector", func() {
			BeforeEach(func() {
				terraformOutputs = terraform.Outputs{Map: map[string]interface{}{
					"internal_cidr":          "10.2.0.0/24",
					"network_name":           "some-network",
					"subnetwork_name":        "some-subnetwork",
					"bosh_open_tag_name":     "some-jumpbox-tag",
					"jumpbox_tag_name":       "some-jumpbox-fw-tag",
					"bosh_director_tag_name": "some-director-tag",
					"internal_tag_name":      "some-internal-tag",
					"external_ip":            "some-external-ip",
					"director_address":       "some-director-address",
					"jumpbox_url":            "some-jumpbox-url",
				}}
			})

			It("generates a bosh manifest", func() {
				stateWithDirector, err := boshManager.CreateDirector(state, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{"creating bosh director", "created bosh director"}))

				Expect(boshExecutor.CreateEnvCall.Receives.Input.DeploymentVars).To(MatchYAML(`---
internal_cidr: 10.2.0.0/24
internal_gw: 10.2.0.1
internal_ip: 10.2.0.6
director_name: bosh-some-env-id
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags:
- some-director-tag
project_id: some-project-id
gcp_credentials_json: some-credential-json
`))
				Expect(boshExecutor.CreateEnvCall.CallCount).To(Equal(1))
				Expect(boshExecutor.CreateEnvCall.Receives.Input.Deployment).To(Equal("director"))
				Expect(boshExecutor.CreateEnvCall.Receives.Input.VarsDir).To(Equal("some-bbl-vars-dir"))
				Expect(boshExecutor.CreateEnvCall.Receives.Input.StateDir).To(Equal("some-state-dir"))

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(0))

				Expect(stateWithDirector.BOSH).To(Equal(storage.BOSH{
					DirectorName:           "bosh-some-env-id",
					DirectorAddress:        "https://10.2.0.6:25555",
					DirectorUsername:       "admin",
					DirectorPassword:       "some-admin-password",
					DirectorSSLCA:          "some-ca",
					DirectorSSLCertificate: "some-certificate",
					DirectorSSLPrivateKey:  "some-private-key",
					UserOpsFile:            "some-ops-file",
					State:                  nil,
				}))
			})

			Context("when an error occurs", func() {
				Context("when get vars dir fails", func() {
					It("returns an error", func() {
						stateStore.GetVarsDirCall.Returns.Error = errors.New("pineapple")

						_, err := boshManager.CreateDirector(storage.State{}, terraformOutputs)
						Expect(err).To(MatchError("Get vars dir: pineapple"))
					})
				})

				Context("when the executor's create env call fails with non create env error", func() {
					BeforeEach(func() {
						boshExecutor.CreateEnvCall.Returns.Error = errors.New("lychee")
					})

					It("returns an error", func() {
						_, err := boshManager.CreateDirector(storage.State{}, terraformOutputs)
						Expect(err).To(MatchError("Create director env: lychee"))
					})
				})
			})
		})
	})

	Describe("Jumpbox set-up", func() {
		var (
			deploymentVars string
			createEnvArgs  []string
			state          storage.State
		)

		BeforeEach(func() {
			terraformOutputs = terraform.Outputs{Map: map[string]interface{}{
				"internal_cidr":          "10.0.0.0/24",
				"network_name":           "some-network",
				"subnetwork_name":        "some-subnetwork",
				"bosh_open_tag_name":     "some-jumpbox-tag",
				"jumpbox_tag_name":       "some-jumpbox-fw-tag",
				"bosh_director_tag_name": "some-director-tag",
				"internal_tag_name":      "some-internal-tag",
				"external_ip":            "some-external-ip",
				"director_address":       "some-director-address",
				"jumpbox_url":            "some-jumpbox-url",
			}}

			state = storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				GCP: storage.GCP{
					Zone:              "some-zone",
					ProjectID:         "some-project-id",
					ServiceAccountKey: "some-credential-json",
				},
				Jumpbox: storage.Jumpbox{
					Manifest: "name: jumpbox",
					State: map[string]interface{}{
						"some-key": "some-value",
					},
				},
			}

			deploymentVars = `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.5
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags:
- some-jumpbox-tag
- some-jumpbox-fw-tag
project_id: some-project-id
gcp_credentials_json: some-credential-json
`

			createEnvArgs = []string{"bosh", "create-env", "/path/to/manifest.yml", "etc"}
			boshExecutor.CreateEnvCall.Returns.Variables = "jumpbox_ssh:\n  private_key: some-jumpbox-private-key"
		})

		AfterEach(func() {
			bosh.ResetOSSetenv()
		})

		Describe("InitializeJumpbox", func() {
			It("calls JumpboxCreateEnvArgsCall appropriately", func() {
				err := boshManager.InitializeJumpbox(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshExecutor.JumpboxCreateEnvArgsCall.Receives.InterpolateInput.DeploymentDir).To(Equal("some-jumpbox-deployment-dir"))
				Expect(boshExecutor.JumpboxCreateEnvArgsCall.Receives.InterpolateInput.VarsDir).To(Equal("some-bbl-vars-dir"))
				Expect(boshExecutor.JumpboxCreateEnvArgsCall.Receives.InterpolateInput.StateDir).To(Equal("some-state-dir"))
			})

			Context("when an error occurs", func() {
				Context("when get vars dir fails", func() {
					It("returns an error", func() {
						stateStore.GetVarsDirCall.Returns.Error = errors.New("kiwi")

						err := boshManager.InitializeJumpbox(state)
						Expect(err).To(MatchError("Get vars dir: kiwi"))
					})
				})

				Context("when get deployment dir fails", func() {
					It("returns an error", func() {
						stateStore.GetJumpboxDeploymentDirCall.Returns.Error = errors.New("kiwi")

						err := boshManager.InitializeJumpbox(state)
						Expect(err).To(MatchError("Get deployment dir: kiwi"))
					})
				})
			})
		})

		Describe("CreateJumpbox", func() {
			It("starts a socks5 proxy for the duration of creating the bosh director", func() {
				socks5ProxyAddr := "localhost:1234"
				socks5Proxy.AddrCall.Returns.Addr = socks5ProxyAddr

				_, err := boshManager.CreateJumpbox(state, terraformOutputs)
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

			It("returns a bbl state with bosh and jumpbox deployment values", func() {
				state, err := boshManager.CreateJumpbox(state, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshExecutor.CreateEnvCall.Receives.Input.VarsDir).To(Equal("some-bbl-vars-dir"))
				Expect(boshExecutor.CreateEnvCall.Receives.Input.StateDir).To(Equal("some-state-dir"))
				Expect(boshExecutor.CreateEnvCall.Receives.Input.Deployment).To(Equal("jumpbox"))
				Expect(boshExecutor.CreateEnvCall.Receives.Input.DeploymentVars).To(Equal(deploymentVars))

				Expect(state).To(Equal(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					GCP: storage.GCP{
						Zone:              "some-zone",
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-credential-json",
					},
					Jumpbox: storage.Jumpbox{
						URL:   "some-jumpbox-url",
						State: nil,
					},
				}))
			})

			Context("when an error occurs", func() {
				Context("when geting the jumpbox key fails", func() {
					BeforeEach(func() {
						sshKeyGetter.GetCall.Returns.Error = errors.New("soursop")
					})

					It("returns an error", func() {
						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("jumpbox key: soursop"))
					})
				})

				Context("when get vars dir fails", func() {
					It("returns an error", func() {
						stateStore.GetVarsDirCall.Returns.Error = errors.New("kiwi")

						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("Get vars dir: kiwi"))
					})
				})

				Context("when create env returns a typed error", func() {
					It("returns an error", func() {
						boshExecutor.CreateEnvCall.Returns.Error = bosh.NewCreateEnvError(map[string]interface{}{"foo": "bar"}, errors.New("apple"))

						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("Create jumpbox env: apple"))
					})
				})

				Context("when create env returns an untyped error", func() {
					It("returns an error", func() {
						boshExecutor.CreateEnvCall.Returns.Error = errors.New("banana")

						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("Create jumpbox env: banana"))
					})
				})

				Context("when the socks5 proxy fails to start", func() {
					It("returns an error", func() {
						socks5Proxy.StartCall.Returns.Error = errors.New("coconut")

						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("Start proxy: coconut"))
					})
				})

				Context("when the socks5 proxy fails to return an address", func() {
					It("returns an error", func() {
						socks5Proxy.AddrCall.Returns.Error = errors.New("mango")

						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("Get proxy address: mango"))
					})
				})
			})
		})
	})

	Describe("DeleteJumpbox", func() {
		var (
			vars          string
			incomingState storage.State
			jumpboxState  map[string]interface{}
		)
		BeforeEach(func() {
			vars = `jumpbox_ssh:
  private_key: some-private-key
  public_key: some-private-key
`

			jumpboxState = map[string]interface{}{
				"key": "value",
			}

			incomingState = storage.State{
				IAAS: "some-iaas",
				BOSH: storage.BOSH{
					Variables: "some-bosh-vars",
				},
				Jumpbox: storage.Jumpbox{
					Manifest: "some-manifest",
					State:    jumpboxState,
				},
			}
		})

		It("calls delete env", func() {
			err := boshManager.DeleteJumpbox(incomingState, terraform.Outputs{})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.Input).To(Equal(bosh.CreateEnvInput{
				Deployment:     "jumpbox",
				StateDir:       "some-state-dir",
				VarsDir:        "some-bbl-vars-dir",
				DeploymentVars: "internal_cidr: 10.0.0.0/24\ninternal_gw: 10.0.0.1\ninternal_ip: 10.0.0.5\ndirector_name: bosh-\n",
			}))

			Expect(boshExecutor.DeleteEnvCall.Receives.Input.Deployment).To(Equal("jumpbox"))
			Expect(boshExecutor.DeleteEnvCall.Receives.Input.VarsDir).To(Equal("some-bbl-vars-dir"))
			Expect(boshExecutor.DeleteEnvCall.Receives.Input.StateDir).To(Equal("some-state-dir"))
		})

		Context("when an error occurs", func() {
			Context("when the executor's delete env call fails with delete env error", func() {
				var expectedError bosh.ManagerDeleteError

				BeforeEach(func() {
					jumpboxState := map[string]interface{}{
						"partial": "jumpbox-state",
					}
					deleteEnvError := bosh.NewDeleteEnvError(jumpboxState, errors.New("failed to delete env"))
					boshExecutor.DeleteEnvCall.Returns.Error = deleteEnvError

					expectedState := incomingState
					expectedState.Jumpbox.State = jumpboxState
					expectedError = bosh.NewManagerDeleteError(expectedState, deleteEnvError)
				})

				It("returns a bosh manager delete error with a valid state", func() {
					err := boshManager.DeleteJumpbox(incomingState, terraform.Outputs{})
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when the delete env fails", func() {
				BeforeEach(func() {
					boshExecutor.DeleteEnvCall.Returns.Error = errors.New("passionfruit")
				})

				It("returns an error", func() {
					err := boshManager.DeleteJumpbox(storage.State{}, terraform.Outputs{})
					Expect(err).To(MatchError("Delete jumpbox env: passionfruit"))
				})
			})
		})
	})

	Describe("DeleteDirector", func() {
		var varsDir string

		BeforeEach(func() {
			var err error
			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			stateStore.GetVarsDirCall.Returns.Directory = varsDir
			vars := "jumpbox_ssh:\n  private_key: some-jumpbox-private-key\n  public_key: some-jumpbox-public-key\n"
			err = ioutil.WriteFile(filepath.Join(varsDir, "jumpbox-variables.yml"), []byte(vars), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())
		})

		It("calls delete env", func() {
			socks5ProxyAddr := "localhost:1234"
			socks5Proxy.AddrCall.Returns.Addr = socks5ProxyAddr

			err := boshManager.DeleteDirector(storage.State{
				Jumpbox: storage.Jumpbox{
					URL: "some-jumpbox-url",
				},
				BOSH: storage.BOSH{
					Manifest: "some-manifest",
					State: map[string]interface{}{
						"key": "value",
					},
					Variables:   boshVars,
					UserOpsFile: "some-ops-file",
				},
			}, terraform.Outputs{})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.Input).To(Equal(bosh.CreateEnvInput{
				Deployment:     "director",
				StateDir:       "some-state-dir",
				VarsDir:        varsDir,
				DeploymentVars: "internal_cidr: 10.0.0.0/24\ninternal_gw: 10.0.0.1\ninternal_ip: 10.0.0.6\ndirector_name: bosh-\n",
			}))

			Expect(socks5Proxy.StartCall.CallCount).To(Equal(1))
			Expect(socks5Proxy.StartCall.Receives.JumpboxPrivateKey).To(Equal("some-jumpbox-private-key"))
			Expect(socks5Proxy.StartCall.Receives.JumpboxExternalURL).To(Equal("some-jumpbox-url"))

			Expect(osSetenvKey).To(Equal("BOSH_ALL_PROXY"))
			Expect(osSetenvValue).To(Equal(fmt.Sprintf("socks5://%s", socks5ProxyAddr)))

			Expect(boshExecutor.DeleteEnvCall.Receives.Input).To(Equal(bosh.DeleteEnvInput{
				Deployment: "director",
				StateDir:   "some-state-dir",
				VarsDir:    varsDir,
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
					},
					Jumpbox: storage.Jumpbox{},
				}
			})

			Context("when getting the jumpbox key fails", func() {
				BeforeEach(func() {
					sshKeyGetter.GetCall.Returns.Error = errors.New("rambutan")
				})

				It("returns an error", func() {
					err := boshManager.DeleteDirector(incomingState, terraform.Outputs{})
					Expect(err).To(MatchError("Delete bosh director: rambutan"))
				})
			})

			Context("when the executor's delete env call fails with delete env error", func() {
				It("returns a bosh manager delete error with a valid state", func() {
					boshState := map[string]interface{}{"partial": "bosh-state"}
					deleteEnvError := bosh.NewDeleteEnvError(boshState, errors.New("failed to delete env"))
					boshExecutor.DeleteEnvCall.Returns.Error = deleteEnvError

					expectedState := incomingState
					expectedState.BOSH = storage.BOSH{
						Manifest: "some-manifest",
						State:    boshState,
					}
					expectedError := bosh.NewManagerDeleteError(expectedState, deleteEnvError)
					err := boshManager.DeleteDirector(incomingState, terraform.Outputs{})
					Expect(err).To(MatchError(expectedError))
				})
			})

			Context("when the delete env fails", func() {
				BeforeEach(func() {
					boshExecutor.DeleteEnvCall.Returns.Error = errors.New("coconut")
				})

				It("returns an error", func() {
					err := boshManager.DeleteDirector(incomingState, terraform.Outputs{})
					Expect(err).To(MatchError("Delete director env: coconut"))
				})
			})

			Context("when a jumpbox deployment exists", func() {
				Context("when the socks5Proxy fails to start", func() {
					BeforeEach(func() {
						socks5Proxy.StartCall.Returns.Error = errors.New("failed to start socks5Proxy")
					})

					It("returns an error", func() {
						err := boshManager.DeleteDirector(incomingState, terraform.Outputs{})
						Expect(err).To(MatchError("Start socks5 proxy: failed to start socks5Proxy"))
					})
				})

				Context("when the socks5Proxy fails to return an address", func() {
					BeforeEach(func() {
						socks5Proxy.AddrCall.Returns.Error = errors.New("plum")
					})

					It("returns an error", func() {
						err := boshManager.DeleteDirector(incomingState, terraform.Outputs{})
						Expect(err).To(MatchError("Get proxy address: plum"))
					})
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
				vars := boshManager.GetJumpboxDeploymentVars(incomingState, terraform.Outputs{Map: map[string]interface{}{
					"internal_cidr":                 "10.0.0.0/24",
					"network_name":                  "some-network",
					"bosh_subnet_id":                "some-subnetwork",
					"bosh_subnet_availability_zone": "some-zone",
					"bosh_iam_instance_profile":     "some-instance-profile",
					"bosh_vms_key_name":             "some-key-name",
					"bosh_vms_private_key":          "some-private-key",
					"jumpbox_security_group":        "some-security-group",
					"external_ip":                   "some-external-ip",
				}})
				Expect(vars).To(MatchYAML(`---
internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.5
director_name: bosh-some-env-id
external_ip: some-external-ip
private_key: some-private-key
az: some-zone
subnet_id: some-subnetwork
access_key_id: some-access-key-id
secret_access_key: some-secret-access-key
iam_instance_profile: some-instance-profile
default_key_name: some-key-name
default_security_groups:
- some-security-group
region: some-region
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
				vars := boshManager.GetJumpboxDeploymentVars(incomingState, terraform.Outputs{Map: map[string]interface{}{
					"internal_cidr":      "10.1.0.0/24",
					"network_name":       "some-network",
					"subnetwork_name":    "some-subnetwork",
					"bosh_open_tag_name": "some-jumpbox-tag",
					"jumpbox_tag_name":   "some-jumpbox-fw-tag",
					"external_ip":        "some-external-ip",
				}})
				Expect(vars).To(MatchYAML(`---
internal_cidr: 10.1.0.0/24
internal_gw: 10.1.0.1
internal_ip: 10.1.0.5
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags:
- some-jumpbox-tag
- some-jumpbox-fw-tag
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
				vars := boshManager.GetDirectorDeploymentVars(incomingState, terraform.Outputs{Map: map[string]interface{}{
					"internal_cidr":             "10.0.1.0/24",
					"network_name":              "some-network",
					"subnetwork_name":           "some-subnetwork",
					"bosh_open_tag_name":        "some-jumpbox-tag",
					"jumpbox_tag_name":          "some-jumpbox-fw-tag",
					"bosh_director_tag_name":    "some-director-tag",
					"internal_tag_name":         "some-internal-tag",
					"bosh_director_external_ip": "some-external-ip",
					"director_address":          "some-director-address",
				}})
				Expect(vars).To(MatchYAML(`---
internal_cidr: 10.0.1.0/24
internal_gw: 10.0.1.1
internal_ip: 10.0.1.6
director_name: bosh-some-env-id
external_ip: some-external-ip
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
					vars := boshManager.GetDirectorDeploymentVars(incomingState, terraform.Outputs{})
					Expect(vars).To(MatchYAML(`---
internal_cidr: 10.0.0.0/24
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
					vars := boshManager.GetDirectorDeploymentVars(incomingState, terraform.Outputs{Map: map[string]interface{}{
						"internal_cidr":                 "10.0.2.0/26",
						"bosh_iam_instance_profile":     "some-bosh-iam-instance-profile",
						"bosh_subnet_availability_zone": "some-bosh-subnet-az",
						"bosh_security_group":           "some-bosh-security-group",
						"bosh_subnet_id":                "some-bosh-subnet",
						"bosh_vms_key_name":             "some-keypair-name",
						"bosh_vms_private_key":          "some-private-key",
						"external_ip":                   "some-bosh-external-ip",
						"director_address":              "some-director-address",
						"kms_key_arn":                   "some-kms-arn",
					}})
					Expect(vars).To(MatchYAML(`---
internal_cidr: 10.0.2.0/26
internal_gw: 10.0.2.1
internal_ip: 10.0.2.6
director_name: bosh-some-env-id
private_key: some-private-key
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-access-key-id
secret_access_key: some-secret-access-key
iam_instance_profile: some-bosh-iam-instance-profile
default_key_name: some-keypair-name
default_security_groups:
- some-bosh-security-group
region: some-region
kms_key_arn: some-kms-arn
`))
				})
			})

			Context("when terraform outputs are missing", func() {
				It("returns valid yaml", func() {
					vars := boshManager.GetDirectorDeploymentVars(incomingState, terraform.Outputs{})
					Expect(vars).To(MatchYAML(`---
internal_cidr: 10.0.0.0/24
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

		Context("when executor fails", func() {
			BeforeEach(func() {
				boshExecutor.VersionCall.Returns.Error = errors.New("failed to execute")
			})

			It("returns an error", func() {
				_, err := boshManager.Version()
				Expect(err).To(MatchError("failed to execute"))
			})
		})
	})
})
