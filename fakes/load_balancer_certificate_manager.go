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
	IsValidLBTypeCall struct {
		Returns struct {
			Result bool
		}
		Receives struct {
			LBType string
		}
	}
}

func (l *LoadBalancerCertificateManager) Create(input iam.CertificateCreateInput) (iam.CertificateCreateOutput, error) {
	l.CreateCall.Receives.Input = input
	return l.CreateCall.Returns.Output, l.CreateCall.Returns.Error
}

func (l *LoadBalancerCertificateManager) IsValidLBType(lbType string) bool {
	l.IsValidLBTypeCall.Receives.LBType = lbType
	return l.IsValidLBTypeCall.Returns.Result
}
