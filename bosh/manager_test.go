package bosh_test

import (
	"errors"
	"fmt"
	"io/ioutil"

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
				},
			}

			boshExecutor.CreateEnvCall.Returns.Variables = boshVars
		})

		Describe("InitializeDirector", func() {
			It("Calls PlanDirector", func() {
				err := boshManager.InitializeDirector(state)
				Expect(err).NotTo(HaveOccurred())
				Expect(boshExecutor.PlanDirectorCall.Receives.DirInput.VarsDir).To(Equal("some-bbl-vars-dir"))
				Expect(boshExecutor.PlanDirectorCall.Receives.DirInput.StateDir).To(Equal("some-state-dir"))
				Expect(boshExecutor.PlanDirectorCall.Receives.DeploymentDir).To(Equal("some-director-deployment-dir"))
				Expect(boshExecutor.PlanJumpboxCall.CallCount).To(Equal(0))

				Expect(boshExecutor.CreateEnvCall.CallCount).To(Equal(0))
			})

			Context("when create env args fails", func() {
				BeforeEach(func() {
					boshExecutor.PlanDirectorCall.Returns.Error = errors.New("failed to interpolate")
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
					"internal_cidr": "10.2.0.0/24",
					"some-key":      "some-value",
					"tags":          []interface{}{"some-tag", "some-other-tag"},
				}}
			})

			It("calls create env on the bosh executor with the expected arguments", func() {
				stateWithDirector, err := boshManager.CreateDirector(state, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{"creating bosh director", "created bosh director"}))

				Expect(boshExecutor.CreateEnvCall.CallCount).To(Equal(1))
				Expect(boshExecutor.CreateEnvCall.Receives.DirInput.Deployment).To(Equal("director"))
				Expect(boshExecutor.CreateEnvCall.Receives.DirInput.VarsDir).To(Equal("some-bbl-vars-dir"))
				Expect(boshExecutor.CreateEnvCall.Receives.DirInput.StateDir).To(Equal("some-state-dir"))

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(0))

				Expect(stateWithDirector.BOSH).To(Equal(storage.BOSH{
					DirectorName:           "bosh-some-env-id",
					DirectorAddress:        "https://10.2.0.6:25555",
					DirectorUsername:       "admin",
					DirectorPassword:       "some-admin-password",
					DirectorSSLCA:          "some-ca",
					DirectorSSLCertificate: "some-certificate",
					DirectorSSLPrivateKey:  "some-private-key",
					State: nil,
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

				Context("when the executor's create env call fails", func() {
					BeforeEach(func() {
						boshExecutor.CreateEnvCall.Returns.Error = errors.New("lychee")
					})

					It("returns an error", func() {
						_, err := boshManager.CreateDirector(storage.State{}, terraformOutputs)
						Expect(err).To(BeAssignableToTypeOf(bosh.ManagerCreateError{}))
						Expect(err).To(MatchError("lychee"))
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
				"internal_cidr": "10.0.0.0/24",
				"some-key":      "some-value",
				"jumpbox_url":   "some-jumpbox-url",
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
			It("calls PlanJumpboxCall appropriately", func() {
				err := boshManager.InitializeJumpbox(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshExecutor.PlanJumpboxCall.Receives.DeploymentDir).To(Equal("some-jumpbox-deployment-dir"))
				Expect(boshExecutor.PlanJumpboxCall.Receives.DirInput.VarsDir).To(Equal("some-bbl-vars-dir"))
				Expect(boshExecutor.PlanJumpboxCall.Receives.DirInput.StateDir).To(Equal("some-state-dir"))
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

				Expect(boshExecutor.CreateEnvCall.Receives.DirInput.VarsDir).To(Equal("some-bbl-vars-dir"))
				Expect(boshExecutor.CreateEnvCall.Receives.DirInput.StateDir).To(Equal("some-state-dir"))
				Expect(boshExecutor.CreateEnvCall.Receives.DirInput.Deployment).To(Equal("jumpbox"))

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

				Context("when create env returns an error", func() {
					It("returns a ManagerCreateError", func() {
						boshExecutor.CreateEnvCall.Returns.Error = errors.New("banana")

						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(BeAssignableToTypeOf(bosh.ManagerCreateError{}))
						Expect(err).To(MatchError("banana"))
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
			err := boshManager.DeleteJumpbox(incomingState, terraform.Outputs{Map: map[string]interface{}{
				"some-key": "some-value",
			}})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.DirInput.Deployment).To(Equal("jumpbox"))
			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.DirInput.StateDir).To(Equal("some-state-dir"))
			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.DirInput.VarsDir).To(Equal("some-bbl-vars-dir"))
			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.DeploymentVars).To(MatchYAML("some-key: some-value"))

			Expect(boshExecutor.DeleteEnvCall.Receives.DirInput.Deployment).To(Equal("jumpbox"))
			Expect(boshExecutor.DeleteEnvCall.Receives.DirInput.VarsDir).To(Equal("some-bbl-vars-dir"))
			Expect(boshExecutor.DeleteEnvCall.Receives.DirInput.StateDir).To(Equal("some-state-dir"))
		})

		Context("when state.Jumpbox is empty", func() {
			It("does not attempt to delete the bosh director", func() {
				err := boshManager.DeleteJumpbox(storage.State{}, terraform.Outputs{Map: map[string]interface{}{}})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshExecutor.DeleteEnvCall.CallCount).To(Equal(0))
			})
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
					err := boshManager.DeleteJumpbox(incomingState, terraform.Outputs{})
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
					Variables: boshVars,
				},
			}, terraform.Outputs{Map: map[string]interface{}{"some-key": "some-value"}})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.DirInput.Deployment).To(Equal("director"))
			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.DirInput.StateDir).To(Equal("some-state-dir"))
			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.DirInput.VarsDir).To(Equal(varsDir))
			Expect(boshExecutor.WriteDeploymentVarsCall.Receives.DeploymentVars).To(MatchYAML("some-key: some-value"))

			Expect(socks5Proxy.StartCall.CallCount).To(Equal(1))
			Expect(socks5Proxy.StartCall.Receives.JumpboxPrivateKey).To(Equal("some-jumpbox-private-key"))
			Expect(socks5Proxy.StartCall.Receives.JumpboxExternalURL).To(Equal("some-jumpbox-url"))

			Expect(osSetenvKey).To(Equal("BOSH_ALL_PROXY"))
			Expect(osSetenvValue).To(Equal(fmt.Sprintf("socks5://%s", socks5ProxyAddr)))

			Expect(boshExecutor.DeleteEnvCall.Receives.DirInput).To(Equal(bosh.DirInput{
				Deployment: "director",
				StateDir:   "some-state-dir",
				VarsDir:    varsDir,
			}))
		})

		Context("when state.BOSH is empty", func() {
			It("does not attempt to delete the bosh director", func() {
				err := boshManager.DeleteDirector(storage.State{}, terraform.Outputs{Map: map[string]interface{}{}})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshExecutor.DeleteEnvCall.CallCount).To(Equal(0))
			})
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
		var incomingState storage.State
		BeforeEach(func() {
			incomingState = storage.State{}
		})
		It("removes the jumpbox__ prefix from variable names", func() {
			vars := boshManager.GetJumpboxDeploymentVars(incomingState, terraform.Outputs{Map: map[string]interface{}{
				"some-key":      "some-value",
				"director__key": "some-director-value",
				"jumpbox__key":  "some-jumpbox-value",
				"key":           "some-ignored-value",
			}})
			Expect(vars).To(MatchYAML(`---
some-key: some-value
key: some-jumpbox-value
`))
		})
	})

	Describe("GetDirectorDeploymentVars", func() {
		var incomingState storage.State
		BeforeEach(func() {
			incomingState = storage.State{}
		})
		It("removes the director__ prefix from variable names", func() {
			vars := boshManager.GetDirectorDeploymentVars(incomingState, terraform.Outputs{Map: map[string]interface{}{
				"some-key":      "some-value",
				"director__key": "some-director-value",
				"jumpbox__key":  "some-jumpbox-value",
				"key":           "some-ignored-value",
			}})
			Expect(vars).To(MatchYAML(`---
some-key: some-value
key: some-director-value
`))
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
