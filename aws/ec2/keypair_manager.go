package ec2

type KeyPairManager struct {
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

func NewKeyPairManager(creator keypairCreator, checker keypairChecker, logger logger) KeyPairManager {
	return KeyPairManager{
		creator: creator,
		checker: checker,
		logger:  logger,
	}
}

func (m KeyPairManager) Sync(keypair KeyPair) (KeyPair, error) {
	hasLocalKeyPair := len(keypair.PublicKey) != 0 || len(keypair.PrivateKey) != 0
	hasRemoteKeyPair, err := m.checker.HasKeyPair(keypair.Name)
	if err != nil {
		return KeyPair{}, err
	}

	if !hasLocalKeyPair || !hasRemoteKeyPair {
		keyPairName := keypair.Name
		m.logger.Step("creating keypair: %q", keyPairName)

		keypair, err = m.creator.Create(keyPairName)
		if err != nil {
			return KeyPair{}, err
		}
	} else {
		m.logger.Step("using existing keypair")
	}

	return keypair, nil
}
