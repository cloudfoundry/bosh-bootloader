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
		uploader        iam.CertificateUploader
		certificateFile *os.File
		privateKeyFile  *os.File
	)

	BeforeEach(func() {
		var err error
		iamClient = &fakes.IAMClient{}
		uploader = iam.NewCertificateUploader()

		certificateFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		privateKeyFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

	})

	Describe("Upload", func() {
		It("uploads a certificate and private key with the given name", func() {
			err := ioutil.WriteFile(certificateFile.Name(), []byte("some-certificate-body"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(privateKeyFile.Name(), []byte("some-private-key-body"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			iamClient.UploadServerCertificateCall.Returns.Output = &awsiam.UploadServerCertificateOutput{
				ServerCertificateMetadata: &awsiam.ServerCertificateMetadata{
					Arn:                   aws.String("arn:aws:iam::some-arn:server-certificate/some-certificate"),
					Path:                  aws.String("/"),
					ServerCertificateId:   aws.String("some-certificate-id"),
					ServerCertificateName: aws.String("test-certificate"),
				},
			}

			err = uploader.Upload("some-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(iamClient.UploadServerCertificateCall.Receives.Input).To(Equal(
				&awsiam.UploadServerCertificateInput{
					CertificateBody:       aws.String("some-certificate-body"),
					PrivateKey:            aws.String("some-private-key-body"),
					ServerCertificateName: aws.String("some-certificate"),
				},
			))

		})

		Context("failure cases", func() {
			It("returns an error when the certificate path does not exist", func() {
				err := uploader.Upload("some-certificate", "/some/fake/path", privateKeyFile.Name(), iamClient)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when the private key path does not exist", func() {
				err := uploader.Upload("some-certificate", certificateFile.Name(), "/some/fake/path", iamClient)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when the certificate fails to upload", func() {
				iamClient.UploadServerCertificateCall.Returns.Error = errors.New("failed to upload certificate")
				err := uploader.Upload("some-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
				Expect(err).To(MatchError("failed to upload certificate"))
			})
		})
	})
})
