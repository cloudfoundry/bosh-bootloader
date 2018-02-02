package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
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
			logger.PromptCall.Returns.Proceed = true
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

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete target http proxy banana-target-http-proxy?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-target-http-proxy", ""))
		})

		Context("when the client fails to list target http proxies", func() {
			BeforeEach(func() {
				client.ListTargetHttpProxiesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := targetHttpProxies.List(filter)
				Expect(err).To(MatchError("Listing target http proxies: some error"))
			})
		})

		Context("when the target http proxy name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := targetHttpProxies.List("grape")
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
				list, err := targetHttpProxies.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-target-http-proxy": ""}
		})

		It("deletes target http proxies", func() {
			targetHttpProxies.Delete(list)

			Expect(client.DeleteTargetHttpProxyCall.CallCount).To(Equal(1))
			Expect(client.DeleteTargetHttpProxyCall.Receives.TargetHttpProxy).To(Equal("banana-target-http-proxy"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting target http proxy banana-target-http-proxy\n"}))
		})

		Context("when the client fails to delete a target http proxy", func() {
			BeforeEach(func() {
				client.DeleteTargetHttpProxyCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				targetHttpProxies.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting target http proxy banana-target-http-proxy: some error\n"}))
			})
		})
	})
})
