package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("TargetPools", func() {
	var (
		client  *fakes.TargetPoolsClient
		logger  *fakes.Logger
		regions map[string]string

		targetPools compute.TargetPools
	)

	BeforeEach(func() {
		client = &fakes.TargetPoolsClient{}
		logger = &fakes.Logger{}
		regions = map[string]string{"https://region-1": "region-1"}

		targetPools = compute.NewTargetPools(client, logger, regions)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListTargetPoolsCall.Returns.Output = &gcpcompute.TargetPoolList{
				Items: []*gcpcompute.TargetPool{{
					Name:   "banana-pool",
					Region: "https://region-1",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for target pools to delete", func() {
			list, err := targetPools.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListTargetPoolsCall.CallCount).To(Equal(1))
			Expect(client.ListTargetPoolsCall.Receives.Region).To(Equal("region-1"))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Target Pool"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-pool"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list target pools", func() {
			BeforeEach(func() {
				client.ListTargetPoolsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := targetPools.List(filter)
				Expect(err).To(MatchError("List Target Pools for region region-1: some error"))
			})
		})

		Context("when the target pool name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := targetPools.List("grape")
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
				list, err := targetPools.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(list).To(HaveLen(0))
			})
		})
	})
})
