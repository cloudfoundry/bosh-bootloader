package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TargetHttpsProxy", func() {
	var (
		client *fakes.TargetHttpsProxiesClient
		name   string

		targetHttpsProxy compute.TargetHttpsProxy
	)

	BeforeEach(func() {
		client = &fakes.TargetHttpsProxiesClient{}
		name = "banana"

		targetHttpsProxy = compute.NewTargetHttpsProxy(client, name)
	})

	Describe("Delete", func() {
		It("deletes the target https proxy", func() {
			err := targetHttpsProxy.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteTargetHttpsProxyCall.CallCount).To(Equal(1))
			Expect(client.DeleteTargetHttpsProxyCall.Receives.TargetHttpsProxy).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteTargetHttpsProxyCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := targetHttpsProxy.Delete()
				Expect(err).To(MatchError("ERROR deleting target https proxy banana: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(targetHttpsProxy.Name()).To(Equal(name))
		})
	})
})
