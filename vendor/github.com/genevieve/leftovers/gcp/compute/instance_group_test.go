package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceGroup", func() {
	var (
		client *fakes.InstanceGroupsClient
		name   string
		zone   string

		instanceGroup compute.InstanceGroup
	)

	BeforeEach(func() {
		client = &fakes.InstanceGroupsClient{}
		name = "banana"
		zone = "zone"

		instanceGroup = compute.NewInstanceGroup(client, name, zone)
	})

	Describe("Delete", func() {
		It("deletes the instance group", func() {
			err := instanceGroup.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteInstanceGroupCall.CallCount).To(Equal(1))
			Expect(client.DeleteInstanceGroupCall.Receives.InstanceGroup).To(Equal(name))
			Expect(client.DeleteInstanceGroupCall.Receives.Zone).To(Equal(zone))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteInstanceGroupCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := instanceGroup.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(instanceGroup.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(instanceGroup.Type()).To(Equal("Instance Group"))
		})
	})
})
