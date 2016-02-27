package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeyPairUploader struct {
	UploadCall struct {
		CallCount int
		Receives  struct {
			Session ec2.Session
			KeyPair ec2.KeyPair
		}
		Returns struct {
			Error error
		}
	}
}

func (u *KeyPairUploader) Upload(session ec2.Session, keypair ec2.KeyPair) error {
	u.UploadCall.Receives.Session = session
	u.UploadCall.Receives.KeyPair = keypair
	u.UploadCall.CallCount++
	return u.UploadCall.Returns.Error
}
