package ec2

type KeyPair struct {
	Name       string
	PrivateKey string
	PublicKey  string
}

func (kp KeyPair) IsEmpty() bool {
	return kp.Name == "" && len(kp.PublicKey) == 0 && len(kp.PrivateKey) == 0
}
