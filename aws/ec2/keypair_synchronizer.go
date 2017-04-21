package ec2

type KeyPairSynchronizer struct {
	creator keypairCreator
	checker keypairChecker
	logger  logger
}

type keypairCreator interface {
	Create(keyPairName string) (KeyPair, error)
}

type keypairChecker interface {
	HasKeyPair(keypairName string) (bool, error)
}

type logger interface {
	Step(message string, a ...interface{})
}

func NewKeyPairSynchronizer(creator keypairCreator, checker keypairChecker, logger logger) KeyPairSynchronizer {
	return KeyPairSynchronizer{
		creator: creator,
		checker: checker,
		logger:  logger,
	}
}

func (k KeyPairSynchronizer) Sync(keyPair KeyPair) (KeyPair, error) {
	hasLocalKeyPair := len(keyPair.PublicKey) != 0 || len(keyPair.PrivateKey) != 0

	k.logger.Step("checking if keypair %q exists", keyPair.Name)
	hasRemoteKeyPair, err := k.checker.HasKeyPair(keyPair.Name)
	if err != nil {
		return KeyPair{}, err
	}

	if !hasLocalKeyPair || !hasRemoteKeyPair {
		keyPairName := keyPair.Name
		k.logger.Step("creating keypair")

		keyPair, err = k.creator.Create(keyPairName)
		if err != nil {
			return KeyPair{}, err
		}
	} else {
		k.logger.Step("using existing keypair")
	}

	return keyPair, nil
}
