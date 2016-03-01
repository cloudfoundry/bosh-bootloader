package ec2

type KeyPairManager struct {
	creator   keypairCreator
	retriever keypairRetriever
}

type keypairCreator interface {
	Create(Client) (KeyPair, error)
}

type keypairRetriever interface {
	Retrieve(client Client, keypairName string) (KeyPairInfo, bool, error)
}

func NewKeyPairManager(creator keypairCreator, retriever keypairRetriever) KeyPairManager {
	return KeyPairManager{
		creator:   creator,
		retriever: retriever,
	}
}

func (m KeyPairManager) Sync(ec2Client Client, keypair KeyPair) (KeyPair, error) {
	hasLocalKeyPair := !keypair.IsEmpty()
	_, hasRemoteKeyPair, err := m.retriever.Retrieve(ec2Client, keypair.Name)
	if err != nil {
		return KeyPair{}, err
	}

	if !hasLocalKeyPair || !hasRemoteKeyPair {
		keypair, err = m.creator.Create(ec2Client)
		if err != nil {
			return KeyPair{}, err
		}
	}

	return keypair, nil
}
