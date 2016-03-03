package ec2

type KeyPairManager struct {
	creator keypairCreator
	checker keypairChecker
	logger  logger
}

type keypairCreator interface {
	Create(Client) (KeyPair, error)
}

type keypairChecker interface {
	HasKeyPair(client Client, keypairName string) (bool, error)
}

type logger interface {
	Step(message string)
}

func NewKeyPairManager(creator keypairCreator, checker keypairChecker, logger logger) KeyPairManager {
	return KeyPairManager{
		creator: creator,
		checker: checker,
		logger:  logger,
	}
}

func (m KeyPairManager) Sync(ec2Client Client, keypair KeyPair) (KeyPair, error) {
	hasLocalKeyPair := !keypair.IsEmpty()
	hasRemoteKeyPair, err := m.checker.HasKeyPair(ec2Client, keypair.Name)
	if err != nil {
		return KeyPair{}, err
	}

	if !hasLocalKeyPair || !hasRemoteKeyPair {
		m.logger.Step("creating keypair")

		keypair, err = m.creator.Create(ec2Client)
		if err != nil {
			return KeyPair{}, err
		}
	} else {
		m.logger.Step("using existing keypair")
	}

	return keypair, nil
}
