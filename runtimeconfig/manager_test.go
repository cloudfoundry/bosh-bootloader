package runtimeconfig_test

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/runtimeconfig"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	var (
		logger        *fakes.Logger
		configUpdater *fakes.BOSHConfigUpdater
		manager       runtimeconfig.Manager
		incomingState storage.State

		fileIO      = &fakes.FileIO{}
		dirProvider *fakes.DirProvider
	)
	BeforeEach(func() {
		logger = &fakes.Logger{}
		dirProvider = &fakes.DirProvider{}
		fileIO = &fakes.FileIO{}
		configUpdater = &fakes.BOSHConfigUpdater{}
		incomingState = storage.State{
			IAAS: "gcp",
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		}
		manager = runtimeconfig.NewManager(logger, dirProvider, configUpdater, fileIO)
		dirProvider.GetDirectorDeploymentDirCall.Returns.Dir = "some-bosh-deployment-dir"
		dirProvider.GetRuntimeConfigDirCall.Returns.Dir = "some-runtime-config-dir"
	})

	Describe("Initialize", func() {
		It("returns a dns runtime config", func() {
			fileIO.ReadFileCall.Returns.Contents = []byte("some-yaml")
			err := manager.Initialize(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileIO.WriteFileCall.CallCount).To(Equal(1))
			Expect(fileIO.ReadFileCall.Receives.Filename).To(Equal(filepath.Join("some-bosh-deployment-dir", "runtime-configs", "dns.yml")))
			Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join("some-runtime-config-dir", "runtime-config.yml")))
			Expect(fileIO.WriteFileCall.Receives[0].Contents).To(MatchYAML([]byte("some-yaml")))
		})

		Context("runtime configs dir is missing", func() {
			It("should return a descriptive error", func() {
				dirProvider.GetRuntimeConfigDirCall.Returns.Error = errors.New("some-error")
				err := manager.Initialize(incomingState)
				Expect(err).To(MatchError("runtime config directory could not be found: some-error"))
			})
		})

		Context("director deployment dir is missing", func() {
			It("should return a descriptive error", func() {
				dirProvider.GetDirectorDeploymentDirCall.Returns.Error = errors.New("some-error")
				err := manager.Initialize(incomingState)
				Expect(err).To(MatchError("bosh-deployment directory could not be found: some-error"))
			})
		})

		Context("dns.yml file could not be read", func() {
			It("should return a descriptive error", func() {
				fileIO.ReadFileCall.Returns.Error = errors.New("some-error")
				err := manager.Initialize(incomingState)
				Expect(err).To(MatchError("failed to read runtime config dns.yml from bosh-deployment: some-error"))
			})
		})

		Context("runtime-config/runtime-config.yml could not be written", func() {
			It("should return a descriptive error", func() {
				fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{errors.New("some-error")}}
				err := manager.Initialize(incomingState)
				Expect(err).To(MatchError("failed to write runtime config: some-error"))
			})
		})
	})

	Describe("Update", func() {
		BeforeEach(func() {
			fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
				fakes.FileInfo{
					FileName: "shenanigans-ops.yml",
				},
				fakes.FileInfo{
					FileName: "cool-ops.yml",
				},
				fakes.FileInfo{
					FileName: "runtime-config.yml",
				},
			}

		})
		It("logs steps taken", func() {
			err := manager.Update(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.Messages).To(Equal([]string{
				"applying runtime config",
			}))
		})

		It("updates the bosh director with the correct runtime config and all user provided opsfiles", func() {
			authenticatedCLI := bosh.AuthenticatedCLI{
				BOSHExecutablePath: "some-bosh-executable-path",
			}
			configUpdater.InitializeAuthenticatedCLICall.Returns.AuthenticatedCLIRunner = authenticatedCLI
			err := manager.Update(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(configUpdater.InitializeAuthenticatedCLICall.CallCount).To(Equal(1))
			Expect(configUpdater.InitializeAuthenticatedCLICall.Receives.State).To(Equal(incomingState))

			Expect(configUpdater.UpdateRuntimeConfigCall.CallCount).To(Equal(1))
			Expect(configUpdater.UpdateRuntimeConfigCall.Receives.AuthenticatedCLIRunner).To(Equal(authenticatedCLI))
			Expect(configUpdater.UpdateRuntimeConfigCall.Receives.Filepath).To(Equal(filepath.Join("some-runtime-config-dir", "runtime-config.yml")))
			Expect(configUpdater.UpdateRuntimeConfigCall.Receives.OpsFilepaths).To(Equal([]string{
				filepath.Join("some-runtime-config-dir", "shenanigans-ops.yml"),
				filepath.Join("some-runtime-config-dir", "cool-ops.yml"),
			}))
			Expect(configUpdater.UpdateRuntimeConfigCall.Receives.Name).To(Equal("dns"))
		})

		Context("failure cases", func() {
			Context("when the config updater fails to initialize the authenticated bosh cli", func() {
				BeforeEach(func() {
					configUpdater.InitializeAuthenticatedCLICall.Returns.Error = errors.New("naval")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to initialize authenticated bosh cli: naval"))
				})
			})

			Context("when getting the runtime-config dir fails", func() {
				It("returns an error", func() {
					dirProvider.GetRuntimeConfigDirCall.Returns.Dir = ""
					dirProvider.GetRuntimeConfigDirCall.Returns.Error = errors.New("lime")

					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("could not find runtime-config directory: lime"))
				})
			})

			Context("when reading the runtime-config directory fails", func() {
				BeforeEach(func() {
					fileIO.ReadDirCall.Returns.Error = errors.New("aubergine")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to read the runtime-config directory: aubergine"))
				})
			})

			Context("when the config updater fails to update the runtime config", func() {
				BeforeEach(func() {
					configUpdater.UpdateRuntimeConfigCall.Returns.Error = errors.New("mandarin")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to update runtime-config: mandarin"))
				})
			})
		})
	})
})
