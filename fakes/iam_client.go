package fakes

import "github.com/aws/aws-sdk-go/service/iam"

type IAMClient struct {
	UploadServerCertificateCall struct {
		CallCount int
		Receives  struct {
			Input *iam.UploadServerCertificateInput
		}
		Returns struct {
			Output *iam.UploadServerCertificateOutput
			Error  error
		}
	}

	GetServerCertificateCall struct {
		CallCount int
		Receives  struct {
			Input *iam.GetServerCertificateInput
		}
		Returns struct {
			Output *iam.GetServerCertificateOutput
			Error  error
		}
	}

	DeleteServerCertificateCall struct {
		CallCount int
		Receives  struct {
			Input *iam.DeleteServerCertificateInput
		}
		Returns struct {
			Output *iam.DeleteServerCertificateOutput
			Error  error
		}
	}
}

func (c *IAMClient) UploadServerCertificate(input *iam.UploadServerCertificateInput) (*iam.UploadServerCertificateOutput, error) {
	c.UploadServerCertificateCall.CallCount++
	c.UploadServerCertificateCall.Receives.Input = input
	return c.UploadServerCertificateCall.Returns.Output, c.UploadServerCertificateCall.Returns.Error
}

func (c *IAMClient) GetServerCertificate(input *iam.GetServerCertificateInput) (*iam.GetServerCertificateOutput, error) {
	c.GetServerCertificateCall.CallCount++
	c.GetServerCertificateCall.Receives.Input = input
	return c.GetServerCertificateCall.Returns.Output, c.GetServerCertificateCall.Returns.Error
}

func (c *IAMClient) DeleteServerCertificate(input *iam.DeleteServerCertificateInput) (*iam.DeleteServerCertificateOutput, error) {
	c.DeleteServerCertificateCall.CallCount++
	c.DeleteServerCertificateCall.Receives.Input = input
	return c.DeleteServerCertificateCall.Returns.Output, c.DeleteServerCertificateCall.Returns.Error
}
