package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairRetriever struct {
	RetrieveCall struct {
		CallCount int
		Stub      func(ec2.Client, string) (ec2.KeyPairInfo, bool, error)
		Recieves  struct {
			Name   string
			Client ec2.Client
		}
		Returns struct {
			KeyPairInfo ec2.KeyPairInfo
			Present     bool
			Error       error
		}
	}
}

func (k *KeyPairRetriever) Retrieve(client ec2.Client, name string) (ec2.KeyPairInfo, bool, error) {
	k.RetrieveCall.CallCount++
	k.RetrieveCall.Recieves.Client = client
	k.RetrieveCall.Recieves.Name = name

	if k.RetrieveCall.Stub != nil {
		return k.RetrieveCall.Stub(client, name)
	}

	return k.RetrieveCall.Returns.KeyPairInfo,
		k.RetrieveCall.Returns.Present,
		k.RetrieveCall.Returns.Error
}
