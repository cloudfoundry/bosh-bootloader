package iam

type LoadBalancerCertificateManager struct {
	certificateManager certificateManager
}

type CertificateCreateInput struct {
	CurrentCertificateName string
	CurrentLBType          string
	DesiredLBType          string
	CertPath               string
	KeyPath                string
}

type CertificateCreateOutput struct {
	CertificateName string
	CertificateARN  string
	LBType          string
}

type certificateManager interface {
	CreateOrUpdate(name string, certificate string, privateKey string, client Client) (string, error)
	Delete(certificateName string, iamClient Client) error
	Describe(certificateName string, iamClient Client) (Certificate, error)
}

func NewLoadBalancerCertificateManager(certificateManager certificateManager) LoadBalancerCertificateManager {
	return LoadBalancerCertificateManager{
		certificateManager: certificateManager,
	}
}

func (l LoadBalancerCertificateManager) Create(input CertificateCreateInput, iamClient Client) (CertificateCreateOutput, error) {
	var err error
	certOutput := CertificateCreateOutput{
		LBType:          determineLBType(input.CurrentLBType, input.DesiredLBType),
		CertificateName: input.CurrentCertificateName,
	}

	if certOutput.LBType != "none" && input.CertPath != "" && input.KeyPath != "" {
		certOutput.CertificateName, err = l.certificateManager.CreateOrUpdate(certOutput.CertificateName, input.CertPath, input.KeyPath, iamClient)
		if err != nil {
			return certOutput, err
		}
	}

	if certOutput.LBType == "none" && certOutput.CertificateName != "" {
		err = l.certificateManager.Delete(certOutput.CertificateName, iamClient)
		if err != nil {
			return certOutput, err
		}

		certOutput.CertificateName = ""
	}

	if certOutput.CertificateName != "" {
		certificate, err := l.certificateManager.Describe(certOutput.CertificateName, iamClient)
		certOutput.CertificateARN = certificate.ARN
		if err != nil {
			return certOutput, err
		}
	}

	return certOutput, nil
}

func determineLBType(currentLBType, desiredLBType string) string {
	switch {
	case desiredLBType == "" && currentLBType == "":
		return "none"
	case desiredLBType != "":
		return desiredLBType
	default:
		return currentLBType
	}
}
