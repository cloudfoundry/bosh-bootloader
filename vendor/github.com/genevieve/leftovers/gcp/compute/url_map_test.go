package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UrlMap", func() {
	var (
		client *fakes.UrlMapsClient
		name   string

		urlMap compute.UrlMap
	)

	BeforeEach(func() {
		client = &fakes.UrlMapsClient{}
		name = "banana"

		urlMap = compute.NewUrlMap(client, name)
	})

	Describe("Delete", func() {
		It("deletes the url map", func() {
			err := urlMap.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteUrlMapCall.CallCount).To(Equal(1))
			Expect(client.DeleteUrlMapCall.Receives.UrlMap).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteUrlMapCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := urlMap.Delete()
				Expect(err).To(MatchError("ERROR deleting url map banana: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(urlMap.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns \"url map\"", func() {
			Expect(urlMap.Type()).To(Equal("url map"))
		})
	})
})
