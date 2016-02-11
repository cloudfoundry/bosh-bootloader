package fakes

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Ec2 struct {
	ImportKeyPairCall struct {
		Receives struct {
			Name      string
			PublicKey []byte
		}
		Returns struct {
			Error error
		}
	}

	Ec2ClientCall struct {
		Receives struct {
			Key    string
			Secret string
			Region string
		}
	}
}

func NewEc2() *Ec2 {
	return &Ec2{}

}

func (e *Ec2) ImportKeyPair(input *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error) {
	e.ImportKeyPairCall.Receives.Name = *input.KeyName
	e.ImportKeyPairCall.Receives.PublicKey = input.PublicKeyMaterial
	return nil, e.ImportKeyPairCall.Returns.Error
}
