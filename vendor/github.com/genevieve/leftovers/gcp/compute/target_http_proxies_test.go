package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("TargetHttpProxies", func() {
	var (
		client *fakes.TargetHttpProxiesClient
		logger *fakes.Logger

		targetHttpProxies compute.TargetHttpProxies
	)

	BeforeEach(func() {
		client = &fakes.TargetHttpProxiesClient{}
		logger = &fakes.Logger{}

		targetHttpProxies = compute.NewTargetHttpProxies(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListTargetHttpProxiesCall.Returns.Output = &gcpcompute.TargetHttpProxyList{
				Items: []*gcpcompute.TargetHttpProxy{{
					Name: "banana-target-http-proxy",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for target http proxies to delete", func() {
			list, err := targetHttpProxies.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListTargetHttpProxiesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Target Http Proxy"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-target-http-proxy"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list target http proxies", func() {
			BeforeEach(func() {
				client.ListTargetHttpProxiesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := targetHttpProxies.List(filter)
				Expect(err).To(MatchError("List Target Http Proxies: some error"))
			})
		})

		Context("when the target http proxy name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := targetHttpProxies.List("grape")
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
				list, err := targetHttpProxies.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
