package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TargetHttpProxy", func() {
	var (
		client *fakes.TargetHttpProxiesClient
		name   string

		targetHttpProxy compute.TargetHttpProxy
	)

	BeforeEach(func() {
		client = &fakes.TargetHttpProxiesClient{}
		name = "banana"

		targetHttpProxy = compute.NewTargetHttpProxy(client, name)
	})

	Describe("Delete", func() {
		It("deletes the target http proxy", func() {
			err := targetHttpProxy.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteTargetHttpProxyCall.CallCount).To(Equal(1))
			Expect(client.DeleteTargetHttpProxyCall.Receives.TargetHttpProxy).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteTargetHttpProxyCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := targetHttpProxy.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(targetHttpProxy.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(targetHttpProxy.Type()).To(Equal("Target Http Proxy"))
		})
	})
})
