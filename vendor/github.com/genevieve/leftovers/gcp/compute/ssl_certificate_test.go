package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SslCertificate", func() {
	var (
		client *fakes.SslCertificatesClient
		name   string

		sslCertificate compute.SslCertificate
	)

	BeforeEach(func() {
		client = &fakes.SslCertificatesClient{}
		name = "banana"

		sslCertificate = compute.NewSslCertificate(client, name)
	})

	Describe("Delete", func() {
		It("deletes the ssl certificate", func() {
			err := sslCertificate.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteSslCertificateCall.CallCount).To(Equal(1))
			Expect(client.DeleteSslCertificateCall.Receives.SslCertificate).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteSslCertificateCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := sslCertificate.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(sslCertificate.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(sslCertificate.Type()).To(Equal("Compute Ssl Certificate"))
		})
	})
})
