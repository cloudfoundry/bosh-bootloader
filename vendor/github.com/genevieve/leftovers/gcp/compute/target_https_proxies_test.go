package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("TargetHttpsProxies", func() {
	var (
		client *fakes.TargetHttpsProxiesClient
		logger *fakes.Logger

		targetHttpsProxies compute.TargetHttpsProxies
	)

	BeforeEach(func() {
		client = &fakes.TargetHttpsProxiesClient{}
		logger = &fakes.Logger{}

		targetHttpsProxies = compute.NewTargetHttpsProxies(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListTargetHttpsProxiesCall.Returns.Output = &gcpcompute.TargetHttpsProxyList{
				Items: []*gcpcompute.TargetHttpsProxy{{
					Name: "banana-target-https-proxy",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for target https proxies to delete", func() {
			list, err := targetHttpsProxies.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListTargetHttpsProxiesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Target Https Proxy"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-target-https-proxy"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list target https proxies", func() {
			BeforeEach(func() {
				client.ListTargetHttpsProxiesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := targetHttpsProxies.List(filter)
				Expect(err).To(MatchError("List Target Https Proxies: some error"))
			})
		})

		Context("when the target https proxy name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := targetHttpsProxies.List("grape")
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
				list, err := targetHttpsProxies.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
