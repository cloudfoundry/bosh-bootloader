package iam

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type uuidGenerator interface {
	Generate() (string, error)
}

type CertificateUploader struct {
	uuidGenerator uuidGenerator
}

func NewCertificateUploader(uuidGenerator uuidGenerator) CertificateUploader {
	return CertificateUploader{uuidGenerator: uuidGenerator}
}

func (c CertificateUploader) Upload(certificatePath, privateKeyPath string, iamClient Client) (string, error) {
	certificate, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return "", err
	}

	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return "", err
	}

	uuid, err := c.uuidGenerator.Generate()
	if err != nil {
		return "", err
	}

	newName := fmt.Sprintf("bbl-cert-%s", uuid)

	_, err = iamClient.UploadServerCertificate(&awsiam.UploadServerCertificateInput{
		CertificateBody:       aws.String(string(certificate)),
		PrivateKey:            aws.String(string(privateKey)),
		ServerCertificateName: aws.String(newName),
	})
	if err != nil {
		return "", err
	}
	return newName, nil
}
