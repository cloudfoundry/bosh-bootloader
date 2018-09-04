package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("ForwardingRules", func() {
	var (
		client  *fakes.ForwardingRulesClient
		logger  *fakes.Logger
		regions map[string]string

		forwardingRules compute.ForwardingRules
	)

	BeforeEach(func() {
		client = &fakes.ForwardingRulesClient{}
		logger = &fakes.Logger{}
		regions = map[string]string{"https://region-1": "region-1"}

		forwardingRules = compute.NewForwardingRules(client, logger, regions)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListForwardingRulesCall.Returns.Output = []*gcpcompute.ForwardingRule{{
				Name:   "banana-rule",
				Region: "https://region-1",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for forwarding rules to delete", func() {
			list, err := forwardingRules.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListForwardingRulesCall.CallCount).To(Equal(1))
			Expect(client.ListForwardingRulesCall.Receives.Region).To(Equal("region-1"))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Forwarding Rule"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-rule"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list forwarding rules", func() {
			BeforeEach(func() {
				client.ListForwardingRulesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := forwardingRules.List(filter)
				Expect(err).To(MatchError("List Forwarding Rules for region region-1: some error"))
			})
		})

		Context("when the forwarding rule name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := forwardingRules.List("grape")
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
				list, err := forwardingRules.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
