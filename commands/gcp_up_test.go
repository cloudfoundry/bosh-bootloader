package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("gcp up", func() {
	var (
		stateStore *fakes.StateStore
		gcpUp      commands.GCPUp
	)

	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		gcpUp = commands.NewGCPUp(stateStore)
	})

	Context("Execute", func() {
		It("saves gcp details to the state", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKey: "some-service-account-key",
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "some-region",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: "some-service-account-key",
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
			}))
		})

		Context("failure cases", func() {
			It("returns an error when state store fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("set call failed")}}
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: "sak",
					ProjectID:         "p",
					Zone:              "z",
					Region:            "r",
				}, storage.State{})
				Expect(err).To(MatchError("set call failed"))
			})

			DescribeTable("up config validation", func(upConfig commands.GCPUpConfig, expectedErr string) {
				err := gcpUp.Execute(upConfig, storage.State{})
				Expect(err).To(MatchError(expectedErr))
			},
				Entry("returns an error when service account key is missing", commands.GCPUpConfig{
					ProjectID: "p",
					Zone:      "z",
					Region:    "r",
				}, "GCP service account key must be provided"),
				Entry("returns an error when project ID is missing", commands.GCPUpConfig{
					ServiceAccountKey: "sak",
					Zone:              "z",
					Region:            "r",
				}, "GCP project ID must be provided"),
				Entry("returns an error when zone is missing", commands.GCPUpConfig{
					ServiceAccountKey: "sak",
					ProjectID:         "p",
					Region:            "r",
				}, "GCP zone must be provided"),
				Entry("returns an error when region is missing", commands.GCPUpConfig{
					ServiceAccountKey: "sak",
					ProjectID:         "p",
					Zone:              "z",
				}, "GCP region must be provided"),
			)
		})

		Context("when state contains gcp details", func() {
			It("overwrites them with the up config details", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKey: "new-service-account-key",
					ProjectID:         "new-project-id",
					Zone:              "new-zone",
					Region:            "new-region",
				}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: "new-service-account-key",
						ProjectID:         "new-project-id",
						Zone:              "new-zone",
						Region:            "new-region",
					},
				}))
			})

			It("does not require details from up config", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				}))
			})

			DescribeTable("up config contains subset of the details", func(upConfig commands.GCPUpConfig, expectedErr string) {
				err := gcpUp.Execute(upConfig, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).To(MatchError(expectedErr))
			},
				Entry("returns an error when the service account key is not provided", commands.GCPUpConfig{
					ProjectID: "new-project-id",
					Zone:      "new-zone",
					Region:    "new-region",
				}, "GCP service account key must be provided"),
				Entry("returns an error when the project ID is not provided", commands.GCPUpConfig{
					ServiceAccountKey: "new-service-account-key",
					Zone:              "new-zone",
					Region:            "new-region",
				}, "GCP project ID must be provided"),
				Entry("returns an error when the zone is not provided", commands.GCPUpConfig{
					ServiceAccountKey: "new-service-account-key",
					ProjectID:         "new-project-id",
					Region:            "new-region",
				}, "GCP zone must be provided"),
				Entry("returns an error when the region is not provided", commands.GCPUpConfig{
					ServiceAccountKey: "new-service-account-key",
					ProjectID:         "new-project-id",
					Zone:              "new-zone",
				}, "GCP region must be provided"),
			)
		})
	})
})
