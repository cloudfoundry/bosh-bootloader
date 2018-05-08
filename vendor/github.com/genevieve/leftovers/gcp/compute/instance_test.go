package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	gcpcompute "google.golang.org/api/compute/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instance", func() {
	var (
		client *fakes.InstancesClient
		name   string
		zone   string
		tags   *gcpcompute.Tags

		instance compute.Instance
	)

	BeforeEach(func() {
		client = &fakes.InstancesClient{}
		name = "banana"
		zone = "zone"
		tags = &gcpcompute.Tags{Items: []string{"tag-1"}}

		instance = compute.NewInstance(client, name, zone, tags)
	})

	Describe("Delete", func() {
		It("deletes the instance", func() {
			err := instance.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteInstanceCall.CallCount).To(Equal(1))
			Expect(client.DeleteInstanceCall.Receives.Instance).To(Equal(name))
			Expect(client.DeleteInstanceCall.Receives.Zone).To(Equal(zone))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteInstanceCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(instance.Name()).To(Equal("banana (tag-1)"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(instance.Type()).To(Equal("Compute Instance"))
		})
	})
})
