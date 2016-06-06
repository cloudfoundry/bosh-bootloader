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
	iamClient     Client
	uuidGenerator uuidGenerator
	logger        logger
}

func NewCertificateUploader(iamClient Client, uuidGenerator uuidGenerator, logger logger) CertificateUploader {
	return CertificateUploader{
		iamClient:     iamClient,
		uuidGenerator: uuidGenerator,
		logger:        logger,
	}
}

func (c CertificateUploader) Upload(certificatePath, privateKeyPath, chainPath string) (string, error) {
	c.logger.Step("uploading certificate")

	certificate, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return "", err
	}

	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return "", err
	}

	var chain *string
	if chainPath != "" {
		chainContents, err := ioutil.ReadFile(chainPath)
		chain = aws.String(string(chainContents))
		if err != nil {
			return "", err
		}
	}

	uuid, err := c.uuidGenerator.Generate()
	if err != nil {
		return "", err
	}

	newName := fmt.Sprintf("bbl-cert-%s", uuid)

	_, err = c.iamClient.UploadServerCertificate(&awsiam.UploadServerCertificateInput{
		CertificateBody:       aws.String(string(certificate)),
		PrivateKey:            aws.String(string(privateKey)),
		CertificateChain:      chain,
		ServerCertificateName: aws.String(newName),
	})
	if err != nil {
		return "", err
	}
	return newName, nil
}
