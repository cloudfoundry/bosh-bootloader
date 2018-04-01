package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ForwardingRule", func() {
	var (
		client *fakes.ForwardingRulesClient
		name   string
		region string

		forwardingRule compute.ForwardingRule
	)

	BeforeEach(func() {
		client = &fakes.ForwardingRulesClient{}
		name = "banana"
		region = "region"

		forwardingRule = compute.NewForwardingRule(client, name, region)
	})

	Describe("Delete", func() {
		It("deletes the forwarding rule", func() {
			err := forwardingRule.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteForwardingRuleCall.CallCount).To(Equal(1))
			Expect(client.DeleteForwardingRuleCall.Receives.ForwardingRule).To(Equal(name))
			Expect(client.DeleteForwardingRuleCall.Receives.Region).To(Equal(region))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteForwardingRuleCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := forwardingRule.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(forwardingRule.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns \"forwarding rule\"", func() {
			Expect(forwardingRule.Type()).To(Equal("Forwarding Rule"))
		})
	})
})
