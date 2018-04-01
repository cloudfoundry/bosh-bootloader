package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HttpsHealthCheck", func() {
	var (
		client *fakes.HttpsHealthChecksClient
		name   string

		httpsHealthCheck compute.HttpsHealthCheck
	)

	BeforeEach(func() {
		client = &fakes.HttpsHealthChecksClient{}
		name = "banana"

		httpsHealthCheck = compute.NewHttpsHealthCheck(client, name)
	})

	Describe("Delete", func() {
		It("deletes the https health check", func() {
			err := httpsHealthCheck.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteHttpsHealthCheckCall.CallCount).To(Equal(1))
			Expect(client.DeleteHttpsHealthCheckCall.Receives.HttpsHealthCheck).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteHttpsHealthCheckCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := httpsHealthCheck.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(httpsHealthCheck.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(httpsHealthCheck.Type()).To(Equal("Https Health Check"))
		})
	})
})
