package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeypairRetriever struct {
	RetrieveCall struct {
		CallCount int

		Returns struct {
			KeyPairInfo ec2.KeyPairInfo
			Error       error
		}

		Recieves struct {
			Name    string
			Session ec2.Session
		}
	}
}

func (k *KeypairRetriever) Retrieve(session ec2.Session, name string) (ec2.KeyPairInfo, error) {
	k.RetrieveCall.Recieves.Session = session
	k.RetrieveCall.Recieves.Name = name
	k.RetrieveCall.CallCount++
	return k.RetrieveCall.Returns.KeyPairInfo, k.RetrieveCall.Returns.Error
}
