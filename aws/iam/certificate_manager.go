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
	Upload(certificateName, certificatePath, privateKeyPath string, iamClient Client) error
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

func (c CertificateManager) CreateOrUpdate(name, certificatePath, privateKeyPath string, iamClient Client) error {
	remoteCertificate, err := c.certificateDescriber.Describe(name, iamClient)

	if err == CertificateNotFound {
		return c.certificateUploader.Upload(name, certificatePath, privateKeyPath, iamClient)
	}

	if err != nil {
		return err
	}

	localCertificate, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return err
	}

	if remoteCertificate.Body != string(localCertificate) {
		return c.overwriteCertificate(name, certificatePath, privateKeyPath, iamClient)
	}

	return nil
}

func (c CertificateManager) overwriteCertificate(name, certificatePath, privateKeyPath string, iamClient Client) error {
	err := c.certificateDeleter.Delete(name, iamClient)
	if err != nil {
		return err
	}

	err = c.certificateUploader.Upload(name, certificatePath, privateKeyPath, iamClient)
	if err != nil {
		return err
	}
	return err
}
