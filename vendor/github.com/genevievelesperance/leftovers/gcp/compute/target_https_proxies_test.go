package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
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
			logger.PromptCall.Returns.Proceed = true
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

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete target https proxy banana-target-https-proxy?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-target-https-proxy", ""))
		})

		Context("when the client fails to list target https proxies", func() {
			BeforeEach(func() {
				client.ListTargetHttpsProxiesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := targetHttpsProxies.List(filter)
				Expect(err).To(MatchError("Listing target https proxies: some error"))
			})
		})

		Context("when the target https proxy name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := targetHttpsProxies.List("grape")
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
				list, err := targetHttpsProxies.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-target-https-proxy": ""}
		})

		It("deletes target https proxies", func() {
			targetHttpsProxies.Delete(list)

			Expect(client.DeleteTargetHttpsProxyCall.CallCount).To(Equal(1))
			Expect(client.DeleteTargetHttpsProxyCall.Receives.TargetHttpsProxy).To(Equal("banana-target-https-proxy"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting target https proxy banana-target-https-proxy\n"}))
		})

		Context("when the client fails to delete a target https proxy", func() {
			BeforeEach(func() {
				client.DeleteTargetHttpsProxyCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				targetHttpsProxies.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting target https proxy banana-target-https-proxy: some error\n"}))
			})
		})
	})
})
