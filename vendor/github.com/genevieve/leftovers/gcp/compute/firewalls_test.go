package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("Firewalls", func() {
	var (
		client *fakes.FirewallsClient
		logger *fakes.Logger

		firewalls compute.Firewalls
	)

	BeforeEach(func() {
		client = &fakes.FirewallsClient{}
		logger = &fakes.Logger{}

		firewalls = compute.NewFirewalls(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListFirewallsCall.Returns.Output = []*gcpcompute.Firewall{{
				Name: "banana-firewall",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for firewalls to delete", func() {
			list, err := firewalls.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListFirewallsCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Firewall"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-firewall"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list firewalls", func() {
			BeforeEach(func() {
				client.ListFirewallsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := firewalls.List(filter)
				Expect(err).To(MatchError("Listing firewalls: some error"))
			})
		})

		Context("when the firewall name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := firewalls.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the firewall name contains default", func() {
			BeforeEach(func() {
				client.ListFirewallsCall.Returns.Output = []*gcpcompute.Firewall{{
					Name: "default-allow-banana",
				}}
			})

			It("does not add it to the list", func() {
				list, err := firewalls.List("banana")
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
				list, err := firewalls.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
