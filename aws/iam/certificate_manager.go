package iam

type CertificateManager struct {
	certificateUploader  certificateUploader
	certificateDescriber certificateDescriber
	certificateDeleter   certificateDeleter
}

type Certificate struct {
	Name  string
	Body  string
	ARN   string
	Chain string
}

type certificateUploader interface {
	Upload(certificatePath, privateKeyPath, chainPath string) (string, error)
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

func (c CertificateManager) Create(certificatePath, privateKeyPath, chainPath string) (string, error) {
	return c.certificateUploader.Upload(certificatePath, privateKeyPath, chainPath)
}

func (c CertificateManager) Delete(certificateName string) error {
	return c.certificateDeleter.Delete(certificateName)
}

func (c CertificateManager) Describe(certificateName string) (Certificate, error) {
	return c.certificateDescriber.Describe(certificateName)
}
