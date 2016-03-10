package ssl

type KeyPair struct {
	Certificate []byte
	PrivateKey  []byte
}

func (k KeyPair) IsEmpty() bool {
	return len(k.Certificate) == 0 || len(k.PrivateKey) == 0
}
