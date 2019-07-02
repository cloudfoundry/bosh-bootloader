package bosh_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("Manager", func() {
	var (
		boshExecutor    *fakes.BOSHExecutor
		logger          *fakes.Logger
		stateStore      *fakes.StateStore
		sshKeyGetter    *fakes.SSHKeyGetter
		fs              *fakes.FileIO
		boshCLIProvider *fakes.BOSHCLIProvider
		boshCLI         *fakes.BOSHCLI

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
		sshKeyGetter = &fakes.SSHKeyGetter{}
		sshKeyGetter.GetCall.Returns.PrivateKey = "some-jumpbox-private-key"
		fs = &fakes.FileIO{}
		boshCLIProvider = &fakes.BOSHCLIProvider{}
		stateStore = &fakes.StateStore{}
		stateStore.GetVarsDirCall.Returns.Directory = "some-bbl-vars-dir"
		stateStore.GetStateDirCall.Returns.Directory = "some-state-dir"
		stateStore.GetDirectorDeploymentDirCall.Returns.Directory = "some-director-deployment-dir"
		stateStore.GetJumpboxDeploymentDirCall.Returns.Directory = "some-jumpbox-deployment-dir"

		boshManager = bosh.NewManager(boshExecutor, logger, stateStore, sshKeyGetter, fs, boshCLIProvider)

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
					Expect(err).To(MatchError("pineapple"))
				})
			})

			Context("when get deployment dir fails", func() {
				It("returns an error", func() {
					stateStore.GetDirectorDeploymentDirCall.Returns.Error = errors.New("pineapple")

					err := boshManager.InitializeDirector(storage.State{})
					Expect(err).To(MatchError("pineapple"))
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

				Expect(stateWithDirector.BOSH).To(Equal(storage.BOSH{
					DirectorName:           "bosh-some-env-id",
					DirectorAddress:        "https://10.2.0.6:25555",
					DirectorUsername:       "admin",
					DirectorPassword:       "some-admin-password",
					DirectorSSLCA:          "some-ca",
					DirectorSSLCertificate: "some-certificate",
					DirectorSSLPrivateKey:  "some-private-key",
					State:                  nil,
				}))
			})

			Context("when an error occurs", func() {
				Context("when get vars dir fails", func() {
					It("returns an error", func() {
						stateStore.GetVarsDirCall.Returns.Error = errors.New("pineapple")

						_, err := boshManager.CreateDirector(storage.State{}, terraformOutputs)
						Expect(err).To(MatchError("pineapple"))
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
			state storage.State
		)

		BeforeEach(func() {
			terraformOutputs = terraform.Outputs{Map: map[string]interface{}{
				"internal_cidr": "10.0.0.0/24",
				"some-key":      "some-value",
				"jumpbox_url":   "some-jumpbox-url:22",
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

			boshExecutor.CreateEnvCall.Returns.Variables = "jumpbox_ssh:\n  private_key: some-jumpbox-private-key"
			fs.TempDirCall.Returns.Name = "/fake/file/bosh-jumpbox"
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
						Expect(err).To(MatchError("kiwi"))
					})
				})

				Context("when get deployment dir fails", func() {
					It("returns an error", func() {
						stateStore.GetJumpboxDeploymentDirCall.Returns.Error = errors.New("kiwi")

						err := boshManager.InitializeJumpbox(state)
						Expect(err).To(MatchError("kiwi"))
					})
				})
			})
		})

		Describe("CreateJumpbox", func() {
			It("sets BOSH_ALL_PROXY to start an ssh tunnel and socks5 proxy to create the director", func() {
				_, err := boshManager.CreateJumpbox(state, terraformOutputs)
				Expect(err).NotTo(HaveOccurred())

				Expect(osUnsetenvKey).To(Equal("BOSH_ALL_PROXY"))
				Expect(osSetenvKey).To(Equal("BOSH_ALL_PROXY"))
				Expect(osSetenvValue).To(Equal("ssh+socks5://jumpbox@some-jumpbox-url:22?private-key=/fake/file/bosh-jumpbox/bosh_jumpbox_private.key"))

				Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
					"creating jumpbox",
					"created jumpbox",
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
						URL:   "some-jumpbox-url:22",
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
						Expect(err).To(MatchError("Get jumpbox private key: soursop"))
					})
				})

				Context("when get vars dir fails", func() {
					It("returns an error", func() {
						stateStore.GetVarsDirCall.Returns.Error = errors.New("kiwi")

						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("kiwi"))
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

				Context("when creating a temp directory fails", func() {
					BeforeEach(func() {
						fs.TempDirCall.Returns.Error = errors.New("fig")
					})

					It("returns a helpful error", func() {
						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("Create temp dir for jumpbox private key: fig"))
					})
				})

				Context("when writing the jumpbox private key fails", func() {
					BeforeEach(func() {
						fs.WriteFileCall.Returns = []fakes.WriteFileReturn{{errors.New("starfruit")}}
					})

					It("returns a helpful error", func() {
						_, err := boshManager.CreateJumpbox(state, terraformOutputs)
						Expect(err).To(MatchError("Write jumpbox private key: starfruit"))
					})
				})
			})
		})
	})

	Describe("CleanUpDirector", func() {
		BeforeEach(func() {
			boshCLI = &fakes.BOSHCLI{}
			boshCLIProvider.AuthenticatedCLICall.Returns.AuthenticatedCLI = boshCLI
			boshCLIProvider.AuthenticatedCLICall.Returns.Error = nil
		})

		Context("when there is no bosh director", func() {
			var state storage.State

			BeforeEach(func() {
				state = storage.State{}
			})

			It("does nothing", func() {
				boshManager.CleanUpDirector(state)

				Expect(logger.StepCall.CallCount).To(Equal(0))
				Expect(boshCLIProvider.AuthenticatedCLICall.CallCount).To(Equal(0))
			})
		})

		Context("when there is a bosh director", func() {
			var state storage.State

			BeforeEach(func() {
				state = storage.State{
					Jumpbox: storage.Jumpbox{
						URL: "some-jumpbox-url:22",
					},
					BOSH: storage.BOSH{
						DirectorAddress:  "director-address",
						DirectorUsername: "director-username",
						DirectorPassword: "director-password",
						DirectorSSLCA:    "ca-cert",
					},
				}
			})

			It("authenticates the CLI", func() {
				err := boshManager.CleanUpDirector(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshCLIProvider.AuthenticatedCLICall.CallCount).To(Equal(1))
				Expect(boshCLIProvider.AuthenticatedCLICall.Receives.Jumpbox).To(Equal(state.Jumpbox))
				Expect(boshCLIProvider.AuthenticatedCLICall.Receives.Stderr).To(Equal(os.Stderr))
				Expect(boshCLIProvider.AuthenticatedCLICall.Receives.DirectorAddress).To(Equal(state.BOSH.DirectorAddress))
				Expect(boshCLIProvider.AuthenticatedCLICall.Receives.DirectorUsername).To(Equal(state.BOSH.DirectorUsername))
				Expect(boshCLIProvider.AuthenticatedCLICall.Receives.DirectorPassword).To(Equal(state.BOSH.DirectorPassword))
				Expect(boshCLIProvider.AuthenticatedCLICall.Receives.DirectorCACert).To(Equal(state.BOSH.DirectorSSLCA))
			})

			It("runs bosh clean-up --all", func() {
				err := boshManager.CleanUpDirector(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshCLI.RunCall.CallCount).To(Equal(1))
				Expect(boshCLI.RunCall.Receives.Args).To(Equal([]string{"clean-up", "--all"}))
			})

			It("logs that it is cleaning up the director", func() {
				err := boshManager.CleanUpDirector(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.StepCall.Messages).To(gomegamatchers.ContainSequence([]string{
					"cleaning up director resources",
				}))
			})

			Context("when an error occurs", func() {
				Context("when the cli fails to authenticate", func() {
					BeforeEach(func() {
						boshCLIProvider.AuthenticatedCLICall.Returns.Error = errors.New("failed to authenticate cli")
					})

					It("returns an error", func() {
						err := boshManager.CleanUpDirector(state)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("failed to authenticate cli"))

						Expect(boshCLIProvider.AuthenticatedCLICall.CallCount).To(Equal(1))
					})
				})

				Context("when bosh clean-up --all fails", func() {
					BeforeEach(func() {
						boshCLI.RunCall.Returns.Error = errors.New("failed to run bosh clean-up")
					})

					It("returns an error", func() {
						err := boshManager.CleanUpDirector(state)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("failed to run bosh clean-up"))

						Expect(boshCLI.RunCall.CallCount).To(Equal(1))
					})
				})
			})
		})
	})

	Describe("DeleteJumpbox", func() {
		var (
			incomingState storage.State
			jumpboxState  map[string]interface{}
		)
		BeforeEach(func() {
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

		Context("when the executor's delete env call fails with delete env error", func() {
			var expectedError bosh.ManagerDeleteError

			BeforeEach(func() {
				deleteEnvError := errors.New("failed to delete env")
				boshExecutor.DeleteEnvCall.Returns.Error = deleteEnvError
				expectedError = bosh.NewManagerDeleteError(incomingState, deleteEnvError)
			})

			It("returns a bosh manager delete error with a valid state", func() {
				err := boshManager.DeleteJumpbox(incomingState, terraform.Outputs{})
				Expect(err).To(MatchError(expectedError))
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

			fs.TempDirCall.Returns.Name = "/fake/file/bosh-jumpbox"
		})

		It("calls delete env", func() {
			err := boshManager.DeleteDirector(storage.State{
				Jumpbox: storage.Jumpbox{
					URL: "some-jumpbox-url:22",
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

			Expect(osSetenvKey).To(Equal("BOSH_ALL_PROXY"))
			Expect(osSetenvValue).To(Equal("ssh+socks5://jumpbox@some-jumpbox-url:22?private-key=/fake/file/bosh-jumpbox/bosh_jumpbox_private.key"))

			Expect(boshExecutor.DeleteEnvCall.Receives.DirInput).To(Equal(bosh.DirInput{
				Deployment: "director",
				StateDir:   "some-state-dir",
				VarsDir:    varsDir,
			}))
		})

		Context("when an error occurs", func() {
			var state storage.State

			BeforeEach(func() {
				state = storage.State{
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
					err := boshManager.DeleteDirector(state, terraform.Outputs{})
					Expect(err).To(MatchError("Get jumpbox private key: rambutan"))
				})
			})

			Context("when creating a temp directory fails", func() {
				BeforeEach(func() {
					fs.TempDirCall.Returns.Error = errors.New("fig")
				})

				It("returns a helpful error", func() {
					err := boshManager.DeleteDirector(state, terraformOutputs)
					Expect(err).To(MatchError("Create temp dir for jumpbox private key: fig"))
				})
			})

			Context("when writing the jumpbox private key fails", func() {
				BeforeEach(func() {
					fs.WriteFileCall.Returns = []fakes.WriteFileReturn{{errors.New("starfruit")}}
				})

				It("returns a helpful error", func() {
					err := boshManager.DeleteDirector(state, terraformOutputs)
					Expect(err).To(MatchError("Write jumpbox private key: starfruit"))
				})
			})

			Context("when the executor's delete env call fails with delete env error", func() {
				It("returns a bosh manager delete error", func() {
					deleteEnvError := errors.New("Run bosh delete-env director: some error")
					boshExecutor.DeleteEnvCall.Returns.Error = deleteEnvError

					expectedError := bosh.NewManagerDeleteError(state, deleteEnvError)

					err := boshManager.DeleteDirector(state, terraform.Outputs{})
					Expect(err).To(MatchError(expectedError))
				})
			})
		})
	})

	Describe("GetJumpboxDeploymentVars", func() {
		It("removes the jumpbox__ prefix from variable names", func() {
			vars := boshManager.GetJumpboxDeploymentVars(storage.State{}, terraform.Outputs{Map: map[string]interface{}{
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
		It("removes the director__ prefix from variable names", func() {
			vars := boshManager.GetDirectorDeploymentVars(storage.State{}, terraform.Outputs{Map: map[string]interface{}{
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
			boshExecutor.VersionCall.Returns.Version = "1.1.1"
		})

		It("calls out to bosh executor version", func() {
			version, err := boshManager.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(boshExecutor.VersionCall.CallCount).To(Equal(1))
			Expect(version).To(Equal("1.1.1"))
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
