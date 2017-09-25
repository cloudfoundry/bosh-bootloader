package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

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

var _ = Describe("GCPUp", func() {
	var (
		gcpUp                 commands.GCPUp
		stateStore            *fakes.StateStore
		terraformManager      *fakes.TerraformManager
		boshManager           *fakes.BOSHManager
		cloudConfigManager    *fakes.CloudConfigManager
		envIDManager          *fakes.EnvIDManager
		terraformManagerError *fakes.TerraformManagerError
		gcpZones              *fakes.GCPClient

		expectedZonesState storage.State
	)

	BeforeEach(func() {
		boshManager = &fakes.BOSHManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		envIDManager = &fakes.EnvIDManager{}
		gcpZones = &fakes.GCPClient{}
		stateStore = &fakes.StateStore{}
		terraformManager = &fakes.TerraformManager{}
		terraformManagerError = &fakes.TerraformManagerError{}

		expectedAvailabilityZones := []string{"some-zone", "some-other-zone"}
		expectedZonesState = storage.State{
			GCP: storage.GCP{
				Region: "some-region",
				Zones:  expectedAvailabilityZones,
			},
		}
		expectedZonesState.GCP.Zones = expectedAvailabilityZones
		gcpZones.GetZonesCall.Returns.Zones = expectedAvailabilityZones

		envIDManager.SyncCall.Returns.State = storage.State{GCP: storage.GCP{Region: "some-region"}}

		gcpUp = commands.NewGCPUp(
			stateStore,
			terraformManager,
			boshManager,
			cloudConfigManager,
			envIDManager,
			gcpZones,
		)
	})

	Describe("Execute", func() {
		It("retrieves zones for a region", func() {
			err := gcpUp.Execute(commands.UpConfig{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			By("getting gcp availability zones", func() {
				Expect(gcpZones.GetZonesCall.CallCount).To(Equal(1))
				Expect(gcpZones.GetZonesCall.Receives.Region).To(Equal("some-region"))
			})

			By("saving gcp zones to the state", func() {
				Expect(stateStore.SetCall.Receives[1].State).To(Equal(expectedZonesState))
			})
		})

		Context("failure cases", func() {
			It("returns an error when GCP AZs cannot be retrieved", func() {
				gcpZones.GetZonesCall.Returns.Error = errors.New("can't get gcp availability zones")
				err := gcpUp.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("can't get gcp availability zones"))
			})

			It("returns an error when the state fails to be set after retrieving GCP zones", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("state failed to be set")}}
				err := gcpUp.Execute(commands.UpConfig{}, storage.State{})
				Expect(err).To(MatchError("state failed to be set"))
			})
		})
	})
})
