package runtimeconfig_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/runtimeconfig"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	var (
		logger             *fakes.Logger
		boshCLI            *fakes.BOSHCLI
		boshClientProvider *fakes.BOSHClientProvider
		manager            runtimeconfig.Manager
		incomingState      storage.State

		fileIO      *fakes.FileIO
		dirProvider *fakes.DirProvider
	)
	BeforeEach(func() {
		logger = &fakes.Logger{}
		dirProvider = &fakes.DirProvider{}
		fileIO = &fakes.FileIO{}
		boshCLI = &fakes.BOSHCLI{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		incomingState = storage.State{
			IAAS: "gcp",
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		}
		boshCLI.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error { return nil }
		boshClientProvider.BoshCLICall.Returns.BoshCLI = boshCLI
		manager = runtimeconfig.NewManager(logger, dirProvider, boshClientProvider, boshCLI, fileIO)
		dirProvider.GetDirectorDeploymentDirCall.Returns.Dir = "some-dir"
		dirProvider.GetRuntimeConfigDirCall.Returns.Dir = "some-runtime-config-dir"
	})

	Describe("Initialize", func() {
		It("returns a dns runtime config", func() {
			fileIO.ReadFileCall.Returns.Contents = []byte("some-yaml")
			err := manager.Initialize(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileIO.WriteFileCall.CallCount).To(Equal(1))
			Expect(fileIO.ReadFileCall.Receives.Filename).To(Equal(filepath.Join("some-dir", "runtime-configs", "dns.yml")))
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

	FDescribe("Interpolate", func() {
		BeforeEach(func() {
			fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
				fakes.FileInfo{
					FileName: "shenanigans-ops.yml", // should have path to bbl-state dir
				},
				fakes.FileInfo{
					FileName: "runtime-config.yml",
				},
			}

			boshCLI.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				stdout.Write([]byte("some-runtime-config"))
				return nil
			}
		})

		It("interpolates all other files in runtime-configs into runtime-config.yml", func() {
			runtimeConfigYAML, err := manager.Interpolate()
			Expect(err).NotTo(HaveOccurred())

			Expect(dirProvider.GetRuntimeConfigDirCall.CallCount).To(Equal(1))

			Expect(fileIO.ReadDirCall.Receives.Dirname).To(Equal("some-runtime-config-dir"))
			Expect(boshCLI.RunCallCount()).To(Equal(1))
			_, workingDirectory, args := boshCLI.RunArgsForCall(0)

			// not sure if the working directory actually matters.
			// please communicate to the team if you think it does.
			Expect(workingDirectory).To(Equal("some-runtime-config-dir"))

			Expect(args).To(Equal([]string{
				"interpolate", filepath.Join("some-runtime-config-dir", "runtime-config.yml"),
				"-o", filepath.Join("some-runtime-config-dir", "shenanigans-ops.yml"),
			}))

			Expect(runtimeConfigYAML).To(Equal("some-runtime-config"))
		})

		/* Context("failure cases", func() {
			Context("when getting the cloud config dir fails", func() {
				BeforeEach(func() {
					stateStore.GetCloudConfigDirCall.Returns.Error = errors.New("carrot")
				})

				It("returns an error", func() {
					_, err := manager.Interpolate()
					Expect(err).To(MatchError("carrot"))
				})
			})

			Context("when getting the vars dir fails", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("eggplant")
				})

				It("returns an error", func() {
					_, err := manager.Interpolate()
					Expect(err).To(MatchError("eggplant"))
				})
			})

			Context("when reading the cloud config dir fails", func() {
				BeforeEach(func() {
					fileIO.ReadDirCall.Returns.Error = errors.New("aubergine")
				})

				It("returns an error", func() {
					_, err := manager.Interpolate()
					Expect(err).To(MatchError("Read cloud config dir: aubergine"))
				})
			})

			Context("when command fails to run", func() {
				BeforeEach(func() {
					boshCLI.RunReturns(errors.New("Interpolate cloud config: failed to run"))
				})

				It("returns an error", func() {
					_, err := manager.Interpolate()
					Expect(err).To(MatchError("Interpolate cloud config: failed to run"))
				})
			})
		}) */
	})

	PDescribe("Update", func() {
		It("logs steps taken", func() {
			err := manager.Update(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.Messages).To(Equal([]string{
				"applying runtime config",
			}))
		})

		It("updates the bosh director with a runtime config provided a valid bbl state", func() {

			err := manager.Update(incomingState) // TODO: name config - after filename?
			Expect(err).NotTo(HaveOccurred())

			Expect(boshCLI.UpdateRuntimeConfigCall.Receives.Filepath).To(Equal("some-dir/runtime-configs/dns.yml"))
			Expect(boshCLI.UpdateRuntimeConfigCall.Receives.Name).To(Equal("dns"))
		})

		Context("failure cases", func() {
			Context("config director deployment dir does not exist", func() {
				It("returns an error", func() {
					dirProvider.GetDirectorDeploymentDirCall.Returns.Dir = ""
					dirProvider.GetDirectorDeploymentDirCall.Returns.Error = errors.New("lime")

					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("could not find bosh-deployment directory: lime"))
				})
			})

			Context("when bosh cli fails to update the runtime config", func() {
				BeforeEach(func() {
					boshCLI.UpdateRuntimeConfigCall.Returns.Error = errors.New("mandarin")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to update runtime-config: mandarin"))
				})
			})
		})
	})
})
