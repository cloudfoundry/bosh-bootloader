package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairChecker struct {
	HasKeyPairCall struct {
		CallCount int
		Stub      func(ec2.Client, string) (bool, error)
		Recieves  struct {
			Name   string
			Client ec2.Client
		}
		Returns struct {
			Present bool
			Error   error
		}
	}
}

func (k *KeyPairChecker) HasKeyPair(client ec2.Client, name string) (bool, error) {
	k.HasKeyPairCall.CallCount++
	k.HasKeyPairCall.Recieves.Client = client
	k.HasKeyPairCall.Recieves.Name = name

	if k.HasKeyPairCall.Stub != nil {
		return k.HasKeyPairCall.Stub(client, name)
	}

	return k.HasKeyPairCall.Returns.Present,
		k.HasKeyPairCall.Returns.Error
}
