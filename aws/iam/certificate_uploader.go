package iam

import (
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type CertificateUploader struct {
	iamClient Client
}

func NewCertificateUploader(iamClient Client) CertificateUploader {
	return CertificateUploader{
		iamClient: iamClient,
	}
}

func (c CertificateUploader) Upload(certificatePath, privateKeyPath, chainPath, certificateName string) error {
	certificate, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return err
	}

	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}

	var chain *string
	if chainPath != "" {
		chainContents, err := ioutil.ReadFile(chainPath)
		chain = aws.String(string(chainContents))
		if err != nil {
			return err
		}
	}

	_, err = c.iamClient.UploadServerCertificate(&awsiam.UploadServerCertificateInput{
		CertificateBody:       aws.String(string(certificate)),
		PrivateKey:            aws.String(string(privateKey)),
		CertificateChain:      chain,
		ServerCertificateName: aws.String(certificateName),
	})
	if err != nil {
		return err
	}
	return nil
}
