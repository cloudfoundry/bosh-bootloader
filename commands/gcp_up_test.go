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

		incomingState storage.State
	)

	BeforeEach(func() {
		gcpZones = &fakes.GCPClient{}

		incomingState = storage.State{GCP: storage.GCP{Region: "some-region"}}
		gcpZones.GetZonesCall.Returns.Zones = []string{"zone-1", "zone-2"}

		gcpUp = commands.NewGCPUp(gcpZones)
	})

	Describe("Execute", func() {
		It("retrieves zones for a region", func() {
			returnedState, err := gcpUp.Execute(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(gcpZones.GetZonesCall.CallCount).To(Equal(1))
			Expect(gcpZones.GetZonesCall.Receives.Region).To(Equal("some-region"))

			Expect(returnedState.GCP.Zones).To(Equal([]string{"zone-1", "zone-2"}))
		})

		It("picks a zone and sets it on the state", func() {
			returnedState, err := gcpUp.Execute(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedState.GCP.Zone).To(Equal("zone-1"))
		})

		Context("when zone is already set on the state", func() {
			BeforeEach(func() {
				incomingState.GCP.Zone = "zone-2"
			})

			It("uses existing zone", func() {
				returnedState, err := gcpUp.Execute(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(returnedState.GCP.Zone).To(Equal("zone-2"))
			})
		})

		Context("failure cases", func() {
			Context("when GCP AZs cannot be retrieved", func() {
				BeforeEach(func() {
					gcpZones.GetZonesCall.Returns.Error = errors.New("canteloupe")
				})

				It("returns an error", func() {
					_, err := gcpUp.Execute(storage.State{})
					Expect(err).To(MatchError("Retrieving availability zones: canteloupe"))
				})
			})

			Context("when no zones are retrieved", func() {
				BeforeEach(func() {
					gcpZones.GetZonesCall.Returns.Zones = []string{}
				})

				It("returns an error", func() {
					_, err := gcpUp.Execute(storage.State{})
					Expect(err).To(MatchError("Zone list is empty"))
				})
			})
		})
	})
})
