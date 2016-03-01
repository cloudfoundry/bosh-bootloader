package ec2

type KeyPairManager struct {
	creator   keypairCreator
	retriever keypairRetriever
}

type keypairCreator interface {
	Create(Session) (KeyPair, error)
}

type keypairRetriever interface {
	Retrieve(session Session, keypairName string) (KeyPairInfo, bool, error)
}

func NewKeyPairManager(creator keypairCreator, retriever keypairRetriever) KeyPairManager {
	return KeyPairManager{
		creator:   creator,
		retriever: retriever,
	}
}

func (m KeyPairManager) Sync(ec2Session Session, keypair KeyPair) (KeyPair, error) {
	hasLocalKeyPair := !keypair.IsEmpty()
	_, hasRemoteKeyPair, err := m.retriever.Retrieve(ec2Session, keypair.Name)
	if err != nil {
		return KeyPair{}, err
	}

	if !hasLocalKeyPair || !hasRemoteKeyPair {
		keypair, err = m.creator.Create(ec2Session)
		if err != nil {
			return KeyPair{}, err
		}
	}

	return keypair, nil
}
