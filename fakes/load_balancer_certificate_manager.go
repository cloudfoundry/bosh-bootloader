package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"

type LoadBalancerCertificateManager struct {
	CreateCall struct {
		Returns struct {
			Output iam.CertificateCreateOutput
			Error  error
		}
		Receives struct {
			Input iam.CertificateCreateInput
		}
	}
}

func (l *LoadBalancerCertificateManager) Create(input iam.CertificateCreateInput, iamClient iam.Client) (iam.CertificateCreateOutput, error) {
	l.CreateCall.Receives.Input = input
	return l.CreateCall.Returns.Output, l.CreateCall.Returns.Error
}
