package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairDeleter struct {
	DeleteCall struct {
		Receives struct {
			Client ec2.Client
			Name   string
		}
		Returns struct {
			Error error
		}
	}
}

func (d *KeyPairDeleter) Delete(client ec2.Client, name string) error {
	d.DeleteCall.Receives.Client = client
	d.DeleteCall.Receives.Name = name

	return d.DeleteCall.Returns.Error
}
