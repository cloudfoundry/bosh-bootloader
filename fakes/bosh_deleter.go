package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type BOSHDeleter struct {
	DeleteCall struct {
		CallCount int
		Receives  struct {
			BOSHInitManifest string
			BOSHInitState    boshinit.State
			EC2PrivateKey    string
		}
		Returns struct {
			Error error
		}
	}
}

func (d *BOSHDeleter) Delete(boshInitManifest string, boshInitState boshinit.State, ec2PrivateKey string) error {
	d.DeleteCall.CallCount++
	d.DeleteCall.Receives.BOSHInitManifest = boshInitManifest
	d.DeleteCall.Receives.BOSHInitState = boshInitState
	d.DeleteCall.Receives.EC2PrivateKey = ec2PrivateKey

	return d.DeleteCall.Returns.Error
}
