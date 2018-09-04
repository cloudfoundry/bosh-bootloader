package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("SslCertificates", func() {
	var (
		client *fakes.SslCertificatesClient
		logger *fakes.Logger

		sslCertificates compute.SslCertificates
	)

	BeforeEach(func() {
		client = &fakes.SslCertificatesClient{}
		logger = &fakes.Logger{}

		sslCertificates = compute.NewSslCertificates(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListSslCertificatesCall.Returns.Output = []*gcpcompute.SslCertificate{{
				Name: "banana-certificate",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for instance templates to delete", func() {
			list, err := sslCertificates.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListSslCertificatesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Compute Ssl Certificate"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-certificate"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list instance templates", func() {
			BeforeEach(func() {
				client.ListSslCertificatesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := sslCertificates.List(filter)
				Expect(err).To(MatchError("List Ssl Certificates: some error"))
			})
		})

		Context("when the ssl certificate name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := sslCertificates.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := sslCertificates.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
