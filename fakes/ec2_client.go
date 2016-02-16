package fakes

import "github.com/aws/aws-sdk-go/service/ec2"

type EC2Client struct {
	ImportKeyPairCall struct {
		Receives struct {
			Input *ec2.ImportKeyPairInput
		}
		Returns struct {
			Error error
		}
	}
}

func (c *EC2Client) ImportKeyPair(input *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error) {
	c.ImportKeyPairCall.Receives.Input = input

	return nil, c.ImportKeyPairCall.Returns.Error
}
