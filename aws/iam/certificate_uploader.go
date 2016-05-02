package iam

import (
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type CertificateUploader struct{}

func NewCertificateUploader() CertificateUploader {
	return CertificateUploader{}
}

func (CertificateUploader) Upload(name, certificatePath, privateKeyPath string, iamClient Client) error {
	certificate, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return err
	}

	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}

	_, err = iamClient.UploadServerCertificate(&awsiam.UploadServerCertificateInput{
		CertificateBody:       aws.String(string(certificate)),
		PrivateKey:            aws.String(string(privateKey)),
		ServerCertificateName: aws.String(name),
	})
	if err != nil {
		return err
	}
	return nil
}
