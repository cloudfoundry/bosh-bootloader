package iam_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateUploader", func() {
	var (
		iamClient       *fakes.IAMClient
		uuidGenerator   *fakes.UUIDGenerator
		uploader        iam.CertificateUploader
		certificateFile *os.File
		privateKeyFile  *os.File
		chainFile       *os.File
	)

	BeforeEach(func() {
		var err error
		iamClient = &fakes.IAMClient{}
		uuidGenerator = &fakes.UUIDGenerator{}
		uploader = iam.NewCertificateUploader(iamClient, uuidGenerator)

		certificateFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		privateKeyFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		chainFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{
			{
				String: "abcd",
			},
		}

	})

	Describe("Upload", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(certificateFile.Name(), []byte("some-certificate-body"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(privateKeyFile.Name(), []byte("some-private-key-body"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("uploads a certificate and private key with the given name", func() {
			iamClient.UploadServerCertificateCall.Returns.Output = &awsiam.UploadServerCertificateOutput{
				ServerCertificateMetadata: &awsiam.ServerCertificateMetadata{
					Arn:                   aws.String("arn:aws:iam::some-arn:server-certificate/some-certificate"),
					Path:                  aws.String("/"),
					ServerCertificateId:   aws.String("some-certificate-id"),
					ServerCertificateName: aws.String("test-certificate"),
				},
			}

			certificateName, err := uploader.Upload(certificateFile.Name(), privateKeyFile.Name(), "")
			Expect(err).NotTo(HaveOccurred())
			Expect(certificateName).To(Equal("bbl-cert-abcd"))

			Expect(iamClient.UploadServerCertificateCall.Receives.Input).To(Equal(
				&awsiam.UploadServerCertificateInput{
					CertificateBody:       aws.String("some-certificate-body"),
					PrivateKey:            aws.String("some-private-key-body"),
					ServerCertificateName: aws.String("bbl-cert-abcd"),
				},
			))
		})

		It("uploads a certificate with an optional chain", func() {
			err := ioutil.WriteFile(chainFile.Name(), []byte("some-chain-body"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			iamClient.UploadServerCertificateCall.Returns.Output = &awsiam.UploadServerCertificateOutput{
				ServerCertificateMetadata: &awsiam.ServerCertificateMetadata{
					Arn:                   aws.String("arn:aws:iam::some-arn:server-certificate/some-certificate"),
					Path:                  aws.String("/"),
					ServerCertificateId:   aws.String("some-certificate-id"),
					ServerCertificateName: aws.String("test-certificate"),
				},
			}

			_, err = uploader.Upload(certificateFile.Name(), privateKeyFile.Name(), chainFile.Name())
			Expect(err).NotTo(HaveOccurred())

			Expect(iamClient.UploadServerCertificateCall.Receives.Input).To(Equal(
				&awsiam.UploadServerCertificateInput{
					CertificateBody:       aws.String("some-certificate-body"),
					PrivateKey:            aws.String("some-private-key-body"),
					CertificateChain:      aws.String("some-chain-body"),
					ServerCertificateName: aws.String("bbl-cert-abcd"),
				},
			))
		})

		It("logs uploading certificate", func() {
			_, err := uploader.Upload(certificateFile.Name(), privateKeyFile.Name(), chainFile.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the certificate path does not exist", func() {
				_, err := uploader.Upload("/some/fake/path", privateKeyFile.Name(), "")
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when the private key path does not exist", func() {
				_, err := uploader.Upload(certificateFile.Name(), "/some/fake/path", "")
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when the chain path does not exist and was specified", func() {
				_, err := uploader.Upload(certificateFile.Name(), privateKeyFile.Name(), "/some/fake/chain/path")
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when the certificate fails to upload", func() {
				iamClient.UploadServerCertificateCall.Returns.Error = errors.New("failed to upload certificate")
				_, err := uploader.Upload(certificateFile.Name(), privateKeyFile.Name(), "")
				Expect(err).To(MatchError("failed to upload certificate"))
			})

			It("returns an error when the uuid generator fails", func() {
				uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{
					{
						Error: errors.New("failed to generate uuid"),
					},
				}
				_, err := uploader.Upload(certificateFile.Name(), privateKeyFile.Name(), "")
				Expect(err).To(MatchError("failed to generate uuid"))
			})
		})
	})
})
