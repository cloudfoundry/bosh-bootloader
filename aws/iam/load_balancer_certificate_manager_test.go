package iam_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadBalancerCertificateManager", func() {
	var (
		certificateManager   *fakes.CertificateManager
		lbCertificateManager iam.LoadBalancerCertificateManager
	)

	BeforeEach(func() {
		certificateManager = &fakes.CertificateManager{}
		lbCertificateManager = iam.NewLoadBalancerCertificateManager(certificateManager)
	})
	Describe("Create", func() {
		Context("when desired lb type is specified", func() {
			Context("when cert and key are provided", func() {
				It("creates a new certificate", func() {
					certificateManager.CreateOrUpdateCall.Returns.CertificateName = "some-certificate-name"
					certificateManager.DescribeCall.Returns.Certificate = iam.Certificate{
						Name: "some-certificate-name",
						ARN:  "some-certificate-arn",
					}

					output, err := lbCertificateManager.Create(iam.CertificateCreateInput{
						DesiredLBType: "some-lb-type",
						CertPath:      "some-cert-path",
						KeyPath:       "some-key-path",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(certificateManager.CreateOrUpdateCall.Receives.Certificate).To(Equal("some-cert-path"))
					Expect(certificateManager.CreateOrUpdateCall.Receives.PrivateKey).To(Equal("some-key-path"))
					Expect(output).To(Equal(iam.CertificateCreateOutput{
						CertificateName: "some-certificate-name",
						CertificateARN:  "some-certificate-arn",
						LBType:          "some-lb-type",
					}))
				})
			})

			Context("when cert and key are not provided", func() {
				It("does not upload a cert and key", func() {
					_, err := lbCertificateManager.Create(iam.CertificateCreateInput{
						DesiredLBType: "some-lb-type",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(certificateManager.CreateOrUpdateCall.CallCount).To(Equal(0))
				})
			})
		})

		Context("when desired lb type is none", func() {
			It("doesn't upload a cert and key", func() {
				output, err := lbCertificateManager.Create(iam.CertificateCreateInput{
					CurrentCertificateName: "some-certificate",
					DesiredLBType:          "none",
					CertPath:               "some-cert-path",
					KeyPath:                "some-key-path",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(output).To(Equal(iam.CertificateCreateOutput{
					LBType: "none",
				}))

				Expect(certificateManager.CreateOrUpdateCall.CallCount).To(Equal(0))
			})

			Context("when current lb type is specified", func() {
				It("deletes cert and key", func() {
					output, err := lbCertificateManager.Create(iam.CertificateCreateInput{
						CurrentCertificateName: "some-certificate",
						DesiredLBType:          "none",
						CertPath:               "some-cert-path",
						KeyPath:                "some-key-path",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(output).To(Equal(iam.CertificateCreateOutput{
						LBType: "none",
					}))

					Expect(certificateManager.DeleteCall.CallCount).To(Equal(1))
					Expect(certificateManager.DescribeCall.CallCount).To(Equal(0))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when it fails to create certificate", func() {
				certificateManager.CreateOrUpdateCall.Returns.Error = errors.New("failed to create certificate")

				_, err := lbCertificateManager.Create(iam.CertificateCreateInput{
					DesiredLBType: "some-lb-type",
					CertPath:      "some-cert-path",
					KeyPath:       "some-key-path",
				})
				Expect(err).To(MatchError("failed to create certificate"))
			})

			It("returns an error when it fails to delete certificate", func() {
				certificateManager.DeleteCall.Returns.Error = errors.New("failed to delete certificate")

				_, err := lbCertificateManager.Create(iam.CertificateCreateInput{
					CurrentCertificateName: "some-non-existant-certificate",
					DesiredLBType:          "none",
					CertPath:               "some-cert-path",
					KeyPath:                "some-key-path",
				})
				Expect(err).To(MatchError("failed to delete certificate"))
			})

			It("returns an error when it fails to describe certificate", func() {
				certificateManager.DescribeCall.Returns.Error = errors.New("failed to describe certificate")

				_, err := lbCertificateManager.Create(iam.CertificateCreateInput{
					CurrentCertificateName: "some-non-existant-certificate",
					CurrentLBType:          "some-lb-type",
				})
				Expect(err).To(MatchError("failed to describe certificate"))
			})
		})
	})

	Describe("IsValidLBType", func() {
		DescribeTable("check for valid lb type", func(lbType string) {
			isValid := lbCertificateManager.IsValidLBType(lbType)
			Expect(isValid).To(BeTrue())
		},
			Entry("concourse is supported", "concourse"),
			Entry("cf is supported", "cf"),
			Entry("none is supported", "none"),
			Entry("empty is supported", "none"),
		)

		It("returns false if something other than the supported lb types are passed in", func() {
			isValid := lbCertificateManager.IsValidLBType("some-unsupported-type")
			Expect(isValid).To(BeFalse())
		})
	})
})
