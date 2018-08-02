package runtimeconfig_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/runtimeconfig"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	var (
		logger             *fakes.Logger
		boshClient         *fakes.BOSHClient
		boshClientProvider *fakes.BOSHClientProvider
		manager            runtimeconfig.Manager
		incomingState      storage.State

		fileIO     *fakes.FileIO
		stateStore *fakes.RuntimeStateStore
	)
	BeforeEach(func() {
		logger = &fakes.Logger{}
		stateStore = &fakes.RuntimeStateStore{}
		fileIO = &fakes.FileIO{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		boshClient = &fakes.BOSHClient{}
		boshClientProvider.ClientCall.Returns.Client = boshClient
		incomingState = storage.State{
			IAAS: "gcp",
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		}
		manager = runtimeconfig.NewManager(logger, boshClientProvider, fileIO, stateStore)
	})
	Describe("Update", func() {
		It("logs steps taken", func() {
			err := manager.Update(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.Messages).To(Equal([]string{
				"loading runtime config",
				"applying runtime config",
			}))
		})

		It("updates the bosh director with a runtime config provided a valid bbl state", func() {
			fileIO.ReadFileCall.Returns.Contents = []byte("some-runtime-config")
			fileIO.ReadFileCall.Returns.Error = nil

			err := manager.Update(incomingState) // TODO: name config - after filename?
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

			Expect(boshClient.UpdateRuntimeConfigCall.Receives.Yaml).To(Equal([]byte("some-runtime-config")))
		})

		Context("failure cases", func() {
			Context("bosh client provider encounteres some error", func() {
				It("returns an error", func() {
					boshClientProvider.ClientCall.Returns.Error = errors.New("orange")

					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("orange"))
				})
			})

			Context("config file does not exist", func() {
				It("returns an error", func() {
					fileIO.ReadFileCall.Returns.Contents = nil
					fileIO.ReadFileCall.Returns.Error = errors.New("lemon")

					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("could not open runtime config file \"dns.yml\": lemon"))
				})
			})

			Context("config director deployment dir does not exist", func() {
				It("returns an error", func() {
					stateStore.GetDirectorDeploymentDirCall.Returns.Dir = ""
					stateStore.GetDirectorDeploymentDirCall.Returns.Error = errors.New("lime")

					fileIO.ReadFileCall.Returns.Contents = nil
					fileIO.ReadFileCall.Returns.Error = nil

					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("could not open runtime config file \"dns.yml\": lime"))
				})
			})

			Context("when bosh client fails to update cloud config", func() {
				BeforeEach(func() {
					boshClient.UpdateRuntimeConfigCall.Returns.Error = errors.New("failed to update")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to update"))
				})
			})
		})
	})
})
