package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type KeypairUploader struct {
	UploadCall struct {
		Receives struct {
			Session ec2.Session
			Keypair ec2.Keypair
		}
		Returns struct {
			Error error
		}
	}
}

func (u *KeypairUploader) Upload(session ec2.Session, keypair ec2.Keypair) error {
	u.UploadCall.Receives.Session = session
	u.UploadCall.Receives.Keypair = keypair

	return u.UploadCall.Returns.Error
}
