package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("Instances", func() {
	var (
		client *fakes.InstancesClient
		logger *fakes.Logger
		zones  map[string]string

		instances compute.Instances
	)

	BeforeEach(func() {
		client = &fakes.InstancesClient{}
		logger = &fakes.Logger{}
		zones = map[string]string{"https://zone-1": "zone-1"}

		instances = compute.NewInstances(client, logger, zones)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListInstancesCall.Returns.Output = []*gcpcompute.Instance{{
				Name: "banana-instance",
				Zone: "https://zone-1",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for instances to delete", func() {
			list, err := instances.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListInstancesCall.CallCount).To(Equal(1))
			Expect(client.ListInstancesCall.Receives.Zone).To(Equal("zone-1"))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Compute Instance"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-instance"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list instances", func() {
			BeforeEach(func() {
				client.ListInstancesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := instances.List(filter)
				Expect(err).To(MatchError("List Instances for zone zone-1: some error"))
			})
		})

		Context("when the clearer name for the instance group does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := instances.List("grape")
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
				list, err := instances.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
