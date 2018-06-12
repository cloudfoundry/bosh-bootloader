package container_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/container"
	"github.com/genevieve/leftovers/gcp/container/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster", func() {
	var (
		client *fakes.ClustersClient
		name   string

		cluster container.Cluster
	)

	BeforeEach(func() {
		client = &fakes.ClustersClient{}
		name = "banana"

		cluster = container.NewCluster(client, "zone", name)
	})

	Describe("Delete", func() {
		It("deletes the resource", func() {
			err := cluster.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteClusterCall.Receives.Zone).To(Equal("zone"))
			Expect(client.DeleteClusterCall.Receives.Cluster).To(Equal(name))
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				client.DeleteClusterCall.Returns.Error = errors.New("kiwi")
			})

			It("returns a helpful error message", func() {
				err := cluster.Delete()
				Expect(err).To(MatchError("Delete: kiwi"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(cluster.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(cluster.Type()).To(Equal("Container Cluster"))
		})
	})
})
