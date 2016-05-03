package iam

import "io/ioutil"

type CertificateManager struct {
	certificateUploader  certificateUploader
	certificateDescriber certificateDescriber
	certificateDeleter   certificateDeleter
}

type Certificate struct {
	Name string
	Body string
}

type certificateUploader interface {
	Upload(certificatePath, privateKeyPath string, iamClient Client) (string, error)
}

type certificateDescriber interface {
	Describe(certificateName string, iamClient Client) (Certificate, error)
}

type certificateDeleter interface {
	Delete(certificateName string, iamClient Client) error
}

func NewCertificateManager(certificateUploader certificateUploader, certificateDescriber certificateDescriber, certificateDeleter certificateDeleter) CertificateManager {
	return CertificateManager{
		certificateUploader:  certificateUploader,
		certificateDescriber: certificateDescriber,
		certificateDeleter:   certificateDeleter,
	}
}

func (c CertificateManager) CreateOrUpdate(name, certificatePath, privateKeyPath string, iamClient Client) (string, error) {
	if name == "" {
		return c.certificateUploader.Upload(certificatePath, privateKeyPath, iamClient)
	}

	remoteCertificate, err := c.certificateDescriber.Describe(name, iamClient)

	if err == CertificateNotFound {
		return c.certificateUploader.Upload(certificatePath, privateKeyPath, iamClient)
	}

	if err != nil {
		return "", err
	}

	localCertificate, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return "", err
	}

	if remoteCertificate.Body != string(localCertificate) {
		return c.overwriteCertificate(name, certificatePath, privateKeyPath, iamClient)
	}

	return name, nil
}

func (c CertificateManager) overwriteCertificate(name, certificatePath, privateKeyPath string, iamClient Client) (string, error) {
	err := c.certificateDeleter.Delete(name, iamClient)
	if err != nil {
		return "", err
	}

	certificateName, err := c.certificateUploader.Upload(certificatePath, privateKeyPath, iamClient)
	if err != nil {
		return "", err
	}
	return certificateName, err
}
