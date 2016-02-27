package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairRetriever struct {
	RetrieveCall struct {
		CallCount int
		Stub      func(ec2.Session, string) (ec2.KeyPairInfo, bool, error)
		Recieves  struct {
			Name    string
			Session ec2.Session
		}
		Returns struct {
			KeyPairInfo ec2.KeyPairInfo
			Present     bool
			Error       error
		}
	}
}

func (k *KeyPairRetriever) Retrieve(session ec2.Session, name string) (ec2.KeyPairInfo, bool, error) {
	k.RetrieveCall.CallCount++
	k.RetrieveCall.Recieves.Session = session
	k.RetrieveCall.Recieves.Name = name

	if k.RetrieveCall.Stub != nil {
		return k.RetrieveCall.Stub(session, name)
	}

	return k.RetrieveCall.Returns.KeyPairInfo,
		k.RetrieveCall.Returns.Present,
		k.RetrieveCall.Returns.Error
}
