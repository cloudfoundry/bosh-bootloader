package unsupported

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type keyPairManager interface {
	Sync(ec2Client ec2.Client, keypair ec2.KeyPair) (ec2.KeyPair, error)
}

type KeyPairSynchronizer struct {
	keyPairManager keyPairManager
}

type KeyPair struct {
	Name       string
	PrivateKey string
	PublicKey  string
}

func NewKeyPairSynchronizer(keyPairManager keyPairManager) KeyPairSynchronizer {
	return KeyPairSynchronizer{
		keyPairManager: keyPairManager,
	}
}

func (s KeyPairSynchronizer) Sync(keyPair KeyPair, ec2Client ec2.Client) (KeyPair, error) {
	ec2KeyPair, err := s.keyPairManager.Sync(ec2Client, ec2.KeyPair{
		Name:       keyPair.Name,
		PrivateKey: []byte(keyPair.PrivateKey),
		PublicKey:  []byte(keyPair.PublicKey),
	})
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		Name:       ec2KeyPair.Name,
		PrivateKey: string(ec2KeyPair.PrivateKey),
		PublicKey:  string(ec2KeyPair.PublicKey),
	}, nil
}
