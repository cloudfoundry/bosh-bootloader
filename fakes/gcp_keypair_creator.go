package fakes

type GCPKeyPairCreator struct {
	CreateCall struct {
		CallCount int
		Returns   struct {
			PrivateKey string
			PublicKey  string
			Error      error
		}
	}
}

func (k *GCPKeyPairCreator) Create() (string, string, error) {
	k.CreateCall.CallCount++
	return k.CreateCall.Returns.PrivateKey, k.CreateCall.Returns.PublicKey, k.CreateCall.Returns.Error
}
