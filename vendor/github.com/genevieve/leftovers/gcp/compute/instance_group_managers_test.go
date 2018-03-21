package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("InstanceGroupManagers", func() {
	var (
		client *fakes.InstanceGroupManagersClient
		logger *fakes.Logger
		zones  map[string]string

		instanceGroupManagers compute.InstanceGroupManagers
	)

	BeforeEach(func() {
		client = &fakes.InstanceGroupManagersClient{}
		logger = &fakes.Logger{}
		zones = map[string]string{"https://zone-1": "zone-1"}

		instanceGroupManagers = compute.NewInstanceGroupManagers(client, logger, zones)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.ListInstanceGroupManagersCall.Returns.Output = &gcpcompute.InstanceGroupManagerList{
				Items: []*gcpcompute.InstanceGroupManager{{
					Name: "banana-group",
					Zone: "https://zone-1",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for instance group managers to delete", func() {
			list, err := instanceGroupManagers.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListInstanceGroupManagersCall.CallCount).To(Equal(1))
			Expect(client.ListInstanceGroupManagersCall.Receives.Zone).To(Equal("zone-1"))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete instance group manager banana-group?"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list instance group managers", func() {
			BeforeEach(func() {
				client.ListInstanceGroupManagersCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := instanceGroupManagers.List(filter)
				Expect(err).To(MatchError("Listing instance group managers for zone zone-1: some error"))
			})
		})

		Context("when the instance group manager name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := instanceGroupManagers.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := instanceGroupManagers.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
