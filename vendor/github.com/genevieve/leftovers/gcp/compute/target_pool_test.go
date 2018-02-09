package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TargetPool", func() {
	var (
		client *fakes.TargetPoolsClient
		name   string
		region string

		targetPool compute.TargetPool
	)

	BeforeEach(func() {
		client = &fakes.TargetPoolsClient{}
		name = "banana"
		region = "region"

		targetPool = compute.NewTargetPool(client, name, region)
	})

	Describe("Delete", func() {
		It("deletes the target pool", func() {
			err := targetPool.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteTargetPoolCall.CallCount).To(Equal(1))
			Expect(client.DeleteTargetPoolCall.Receives.TargetPool).To(Equal(name))
			Expect(client.DeleteTargetPoolCall.Receives.Region).To(Equal(region))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteTargetPoolCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := targetPool.Delete()
				Expect(err).To(MatchError("ERROR deleting target pool banana: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(targetPool.Name()).To(Equal(name))
		})
	})
})
