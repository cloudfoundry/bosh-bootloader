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
		boshClientProvider.BoshCLICall.Returns.BoshCLI = boshCLI
		manager = runtimeconfig.NewManager(logger, dirProvider, boshClientProvider)
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
			dirProvider.GetDirectorDeploymentDirCall.Returns.Dir = "some-dir"

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
