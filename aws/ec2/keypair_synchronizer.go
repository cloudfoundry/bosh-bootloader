package ec2

type keyPairManager interface {
	Sync(keypair KeyPair) (KeyPair, error)
}

type KeyPairSynchronizer struct {
	keyPairManager keyPairManager
}

func NewKeyPairSynchronizer(keyPairManager keyPairManager) KeyPairSynchronizer {
	return KeyPairSynchronizer{
		keyPairManager: keyPairManager,
	}
}

func (s KeyPairSynchronizer) Sync(keyPair KeyPair) (KeyPair, error) {
	ec2KeyPair, err := s.keyPairManager.Sync(KeyPair{
		Name:       keyPair.Name,
		PrivateKey: keyPair.PrivateKey,
		PublicKey:  keyPair.PublicKey,
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
