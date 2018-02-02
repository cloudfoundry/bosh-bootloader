package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
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
			logger.PromptCall.Returns.Proceed = true
			client.ListForwardingRulesCall.Returns.Output = &gcpcompute.ForwardingRuleList{
				Items: []*gcpcompute.ForwardingRule{{
					Name:   "banana-rule",
					Region: "https://region-1",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for forwarding rules to delete", func() {
			list, err := forwardingRules.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListForwardingRulesCall.CallCount).To(Equal(1))
			Expect(client.ListForwardingRulesCall.Receives.Region).To(Equal("region-1"))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete forwarding rule banana-rule?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-rule", "region-1"))
		})

		Context("when the client fails to list forwarding rules", func() {
			BeforeEach(func() {
				client.ListForwardingRulesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := forwardingRules.List(filter)
				Expect(err).To(MatchError("Listing forwarding rules for region region-1: some error"))
			})
		})

		Context("when the forwarding rule name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := forwardingRules.List("grape")
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
				list, err := forwardingRules.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-rule": "region-1"}
		})

		It("deletes forwarding rules", func() {
			forwardingRules.Delete(list)

			Expect(client.DeleteForwardingRuleCall.CallCount).To(Equal(1))
			Expect(client.DeleteForwardingRuleCall.Receives.Region).To(Equal("region-1"))
			Expect(client.DeleteForwardingRuleCall.Receives.ForwardingRule).To(Equal("banana-rule"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting forwarding rule banana-rule\n"}))
		})

		Context("when the client fails to delete a forwarding rule", func() {
			BeforeEach(func() {
				client.DeleteForwardingRuleCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				forwardingRules.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting forwarding rule banana-rule: some error\n"}))
			})
		})
	})
})
