package route53_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/route53"
	"github.com/genevieve/leftovers/aws/route53/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HealthCheck", func() {
	var (
		healthCheck route53.HealthCheck
		client      *fakes.HealthChecksClient
		id          *string
	)

	BeforeEach(func() {
		client = &fakes.HealthChecksClient{}
		id = aws.String("the-id")

		healthCheck = route53.NewHealthCheck(client, id)
	})

	Describe("Delete", func() {
		It("deletes the health check", func() {
			err := healthCheck.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteHealthCheckCall.CallCount).To(Equal(1))
			Expect(client.DeleteHealthCheckCall.Receives.Input.HealthCheckId).To(Equal(id))
		})

		Context("when the client fails to delete the health check", func() {
			BeforeEach(func() {
				client.DeleteHealthCheckCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := healthCheck.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(healthCheck.Name()).To(Equal("the-id"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(healthCheck.Type()).To(Equal("Route53 Health Check"))
		})
	})
})
