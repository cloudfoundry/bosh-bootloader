package container_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/container"
	"github.com/genevieve/leftovers/gcp/container/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcontainer "google.golang.org/api/container/v1"
)

var _ = Describe("Clusters", func() {
	var (
		client *fakes.ClustersClient
		logger *fakes.Logger
		filter string

		clusters container.Clusters
	)

	BeforeEach(func() {
		client = &fakes.ClustersClient{}
		logger = &fakes.Logger{}
		filter = "banana"
		zones := map[string]string{"url": "zone-1"}

		logger.PromptWithDetailsCall.Returns.Proceed = true

		clusters = container.NewClusters(client, zones, logger)
	})

	Describe("List", func() {
		BeforeEach(func() {
			client.ListClustersCall.Returns.Output = &gcpcontainer.ListClustersResponse{
				Clusters: []*gcpcontainer.Cluster{{
					Name: "banana-cluster",
				}},
			}
		})

		It("returns a list of clusters to delete", func() {
			list, err := clusters.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Container Cluster"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-cluster"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the user does not want to delete that resource", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return the resource in the list", func() {
				list, err := clusters.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the resource name does not contain the filter", func() {
			BeforeEach(func() {
				client.ListClustersCall.Returns.Output = &gcpcontainer.ListClustersResponse{
					Clusters: []*gcpcontainer.Cluster{{
						Name: "kiwi-cluster",
					}},
				}
			})

			It("does not return the resource in the list", func() {
				list, err := clusters.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				client.ListClustersCall.Returns.Error = errors.New("panic time")
			})

			It("wraps it in a helpful error message", func() {
				_, err := clusters.List(filter)
				Expect(err).To(MatchError("List Clusters for Zone zone-1: panic time"))
			})
		})
	})
})
