package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPUp", func() {
	var (
		gcpUp    commands.GCPUp
		gcpZones *fakes.GCPClient

		incomingState      storage.State
		expectedZonesState storage.State
	)

	BeforeEach(func() {
		gcpZones = &fakes.GCPClient{}

		incomingState = storage.State{GCP: storage.GCP{Region: "some-region"}}
		expectedZonesState = storage.State{GCP: storage.GCP{Region: "some-region", Zones: []string{"zone-1"}}}
		gcpZones.GetZonesCall.Returns.Zones = []string{"zone-1"}

		gcpUp = commands.NewGCPUp(gcpZones)
	})

	Describe("Execute", func() {
		It("retrieves zones for a region", func() {
			returnedState, err := gcpUp.Execute(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(gcpZones.GetZonesCall.CallCount).To(Equal(1))
			Expect(gcpZones.GetZonesCall.Receives.Region).To(Equal("some-region"))

			Expect(returnedState).To(Equal(expectedZonesState))
		})

		Context("failure cases", func() {
			It("returns an error when GCP AZs cannot be retrieved", func() {
				gcpZones.GetZonesCall.Returns.Error = errors.New("canteloupe")
				_, err := gcpUp.Execute(storage.State{})
				Expect(err).To(MatchError("Retrieving availability zones: canteloupe"))
			})
		})
	})
})
