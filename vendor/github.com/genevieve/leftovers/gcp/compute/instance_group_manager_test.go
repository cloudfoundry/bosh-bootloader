package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceGroupManager", func() {
	var (
		client *fakes.InstanceGroupManagersClient
		name   string
		zone   string

		instanceGroupManager compute.InstanceGroupManager
	)

	BeforeEach(func() {
		client = &fakes.InstanceGroupManagersClient{}
		name = "banana"
		zone = "zone"

		instanceGroupManager = compute.NewInstanceGroupManager(client, name, zone)
	})

	Describe("Delete", func() {
		It("deletes the instance group manager", func() {
			err := instanceGroupManager.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteInstanceGroupManagerCall.CallCount).To(Equal(1))
			Expect(client.DeleteInstanceGroupManagerCall.Receives.InstanceGroupManager).To(Equal(name))
			Expect(client.DeleteInstanceGroupManagerCall.Receives.Zone).To(Equal(zone))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteInstanceGroupManagerCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := instanceGroupManager.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(instanceGroupManager.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(instanceGroupManager.Type()).To(Equal("Instance Group Manager"))
		})
	})
})
