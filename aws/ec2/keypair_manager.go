package ec2

type KeyPairManager struct {
	creator   keypairCreator
	retriever keypairRetriever
	logger    logger
}

type keypairCreator interface {
	Create(Client) (KeyPair, error)
}

type keypairRetriever interface {
	Retrieve(client Client, keypairName string) (KeyPairInfo, bool, error)
}

type logger interface {
	Step(message string)
}

func NewKeyPairManager(creator keypairCreator, retriever keypairRetriever, logger logger) KeyPairManager {
	return KeyPairManager{
		creator:   creator,
		retriever: retriever,
		logger:    logger,
	}
}

func (m KeyPairManager) Sync(ec2Client Client, keypair KeyPair) (KeyPair, error) {
	hasLocalKeyPair := !keypair.IsEmpty()
	_, hasRemoteKeyPair, err := m.retriever.Retrieve(ec2Client, keypair.Name)
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
