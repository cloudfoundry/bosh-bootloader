package fakes

import "github.com/aws/aws-sdk-go/service/iam"

type IAMClient struct {
	UploadServerCertificateCall struct {
		Receives struct {
			Input *iam.UploadServerCertificateInput
		}
		Returns struct {
			Output *iam.UploadServerCertificateOutput
			Error  error
		}
	}
}

func (c *IAMClient) UploadServerCertificate(input *iam.UploadServerCertificateInput) (*iam.UploadServerCertificateOutput, error) {
	c.UploadServerCertificateCall.Receives.Input = input
	return c.UploadServerCertificateCall.Returns.Output, c.UploadServerCertificateCall.Returns.Error

}
