package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("InstanceGroups", func() {
	var (
		client *fakes.InstanceGroupsClient
		logger *fakes.Logger
		zones  map[string]string

		instanceGroups compute.InstanceGroups
	)

	BeforeEach(func() {
		client = &fakes.InstanceGroupsClient{}
		logger = &fakes.Logger{}
		zones = map[string]string{"https://zone-1": "zone-1"}

		instanceGroups = compute.NewInstanceGroups(client, logger, zones)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListInstanceGroupsCall.Returns.Output = []*gcpcompute.InstanceGroup{{
				Name: "banana-group",
				Zone: "https://zone-1",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for instance groups to delete", func() {
			list, err := instanceGroups.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListInstanceGroupsCall.CallCount).To(Equal(1))
			Expect(client.ListInstanceGroupsCall.Receives.Zone).To(Equal("zone-1"))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Instance Group"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-group"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list instance groups", func() {
			BeforeEach(func() {
				client.ListInstanceGroupsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := instanceGroups.List(filter)
				Expect(err).To(MatchError("List Instance Groups for zone zone-1: some error"))
			})
		})

		Context("when the instance group name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := instanceGroups.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := instanceGroups.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
