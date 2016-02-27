package ec2

type KeyPair struct {
	Name       string
	PublicKey  []byte
	PrivateKey []byte
}

func (kp KeyPair) IsEmpty() bool {
	return kp.Name == "" && len(kp.PublicKey) == 0 && len(kp.PrivateKey) == 0
}
