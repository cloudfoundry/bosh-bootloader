package iam

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type CertificateUploader struct {
	iamClientProvider iamClientProvider
}

func NewCertificateUploader(iamClientProvider iamClientProvider) CertificateUploader {
	return CertificateUploader{
		iamClientProvider: iamClientProvider,
	}
}

func (c CertificateUploader) Upload(certificatePath, privateKeyPath, chainPath, certificateName string) error {
	if strings.ContainsAny(certificateName, ":") {
		return fmt.Errorf("%q is an invalid certificate name, it must not contain %q", certificateName, ":")
	}

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

	_, err = c.iamClientProvider.GetIAMClient().UploadServerCertificate(&awsiam.UploadServerCertificateInput{
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
