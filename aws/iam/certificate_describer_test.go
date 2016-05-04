package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	awsiam "github.com/aws/aws-sdk-go/service/iam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateDescriber", func() {
	var (
		iamClient *fakes.IAMClient
		describer iam.CertificateDescriber
	)

	BeforeEach(func() {
		iamClient = &fakes.IAMClient{}
		describer = iam.NewCertificateDescriber()
	})

	Describe("Describe", func() {
		It("describes the certificate with the given name", func() {
			iamClient.GetServerCertificateCall.Returns.Output = &awsiam.GetServerCertificateOutput{
				ServerCertificate: &awsiam.ServerCertificate{
					CertificateBody: aws.String("some-certificate-body"),
					ServerCertificateMetadata: &awsiam.ServerCertificateMetadata{
						Path:                  aws.String("some-certificate-path"),
						Arn:                   aws.String("some-certificate-arn"),
						ServerCertificateId:   aws.String("some-server-certificate-id"),
						ServerCertificateName: aws.String("some-certificate"),
					},
				},
			}

			certificate, err := describer.Describe("some-certificate", iamClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(iamClient.GetServerCertificateCall.Receives.Input.ServerCertificateName).To(Equal(aws.String("some-certificate")))
			Expect(certificate.Name).To(Equal("some-certificate"))
			Expect(certificate.Body).To(Equal("some-certificate-body"))
			Expect(certificate.ARN).To(Equal("some-certificate-arn"))
		})

		Context("failure cases", func() {
			It("returns an error when the ServerCertificate is nil", func() {
				iamClient.GetServerCertificateCall.Returns.Output = &awsiam.GetServerCertificateOutput{
					ServerCertificate: nil,
				}

				_, err := describer.Describe("some-certificate", iamClient)
				Expect(err).To(MatchError(iam.CertificateDescriptionFailure))
			})

			It("returns an error when the ServerCertificateMetadata is nil", func() {
				iamClient.GetServerCertificateCall.Returns.Output = &awsiam.GetServerCertificateOutput{
					ServerCertificate: &awsiam.ServerCertificate{
						ServerCertificateMetadata: nil,
					},
				}

				_, err := describer.Describe("some-certificate", iamClient)
				Expect(err).To(MatchError(iam.CertificateDescriptionFailure))
			})

			It("returns an error when the certificate cannot be described", func() {
				iamClient.GetServerCertificateCall.Returns.Error = awserr.NewRequestFailure(
					awserr.New("boom",
						"something bad happened",
						errors.New(""),
					), 404, "0",
				)

				_, err := describer.Describe("some-certificate", iamClient)
				Expect(err).To(MatchError(ContainSubstring("something bad happened")))
			})

			It("returns a CertificateNotFound error when the certificate does not exist", func() {
				iamClient.GetServerCertificateCall.Returns.Error = awserr.NewRequestFailure(
					awserr.New("NoSuchEntity",
						"The Server Certificate with name some-certificate cannot be found.",
						errors.New(""),
					), 404, "0",
				)

				_, err := describer.Describe("some-certificate", iamClient)
				Expect(err).To(MatchError(iam.CertificateNotFound))
			})
		})
	})

})
