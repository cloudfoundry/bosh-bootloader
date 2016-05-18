package iam

import (
	"io/ioutil"
	"strings"
)

type CertificateManager struct {
	certificateUploader  certificateUploader
	certificateDescriber certificateDescriber
	certificateDeleter   certificateDeleter
}

type Certificate struct {
	Name string
	Body string
	ARN  string
}

type certificateUploader interface {
	Upload(certificatePath, privateKeyPath string) (string, error)
}

type certificateDescriber interface {
	Describe(certificateName string) (Certificate, error)
}

type certificateDeleter interface {
	Delete(certificateName string) error
}

func NewCertificateManager(certificateUploader certificateUploader, certificateDescriber certificateDescriber, certificateDeleter certificateDeleter) CertificateManager {
	return CertificateManager{
		certificateUploader:  certificateUploader,
		certificateDescriber: certificateDescriber,
		certificateDeleter:   certificateDeleter,
	}
}

func (c CertificateManager) CreateOrUpdate(name, certificatePath, privateKeyPath string) (string, error) {
	if name == "" {
		return c.certificateUploader.Upload(certificatePath, privateKeyPath)
	}

	remoteCertificate, err := c.certificateDescriber.Describe(name)

	if err == CertificateNotFound {
		return c.certificateUploader.Upload(certificatePath, privateKeyPath)
	}

	if err != nil {
		return "", err
	}

	localCertificateBody, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return "", err
	}

	trimmedLocalCertificateBody := strings.TrimSpace(string(localCertificateBody))

	if remoteCertificate.Body != trimmedLocalCertificateBody {
		return c.overwriteCertificate(name, certificatePath, privateKeyPath)
	}

	return name, nil
}

func (c CertificateManager) Create(certificatePath, privateKeyPath string) (string, error) {
	return c.certificateUploader.Upload(certificatePath, privateKeyPath)
}

func (c CertificateManager) Delete(certificateName string) error {
	return c.certificateDeleter.Delete(certificateName)
}

func (c CertificateManager) Describe(certificateName string) (Certificate, error) {
	return c.certificateDescriber.Describe(certificateName)
}

func (c CertificateManager) overwriteCertificate(name, certificatePath, privateKeyPath string) (string, error) {
	err := c.certificateDeleter.Delete(name)
	if err != nil {
		return "", err
	}

	certificateName, err := c.certificateUploader.Upload(certificatePath, privateKeyPath)
	if err != nil {
		return "", err
	}
	return certificateName, err
}
