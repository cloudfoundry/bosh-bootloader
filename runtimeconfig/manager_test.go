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

		fileIO           *fakes.FileIO
		dirProvider      *fakes.DirProvider
		dnsRuntimeConfig []byte
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
		manager = runtimeconfig.NewManager(logger, dirProvider, boshClientProvider, boshCLI)
		dirProvider.GetDirectorDeploymentDirCall.Returns.Dir = "some-dir"
	})

	FDescribe("Initialize", func() {
		It("returns a dns runtime config", func() {
			err := manager.Initialize(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileIO.WriteFileCall.CallCount).To(Equal(1))
			Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join("some-runtime-config-dir", "default-runtime-config.yml")))
			Expect(fileIO.WriteFileCall.Receives[0].Contents).To(MatchYAML(dnsRuntimeConfig))
		})
	})

	Describe("Update", func() {
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

	Describe("Interpolate", func() {
		BeforeEach(func() {
			fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
				fakes.FileInfo{
					FileName: "shenanigans-ops.yml",
				},
				fakes.FileInfo{
					FileName: "dns.yml",
				},
			}
		})

		It("returns a cloud config yaml provided a valid bbl state", func() {
			runtimeConfigYAML, err := manager.Interpolate()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshCLI.RunCallCount()).To(Equal(1))
			_, workingDirectory, args := boshCLI.RunArgsForCall(0)

			Expect(args).To(Equal([]string{
				"interpolate", filepath.Join(workingDirectory, "runtime-configs", "dns.yml"),
				"-o", filepath.Join(workingDirectory, "runtime-configs", "shenanigans-ops.yml"),
			}))

			Expect(runtimeConfigYAML).To(Equal("some-cloud-config"))
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
})
