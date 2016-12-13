package fakes

type GCPKeyPairDeleter struct {
	DeleteCall struct {
		CallCount int
		Receives  struct {
			PublicKey string
		}
		Returns struct {
			Error error
		}
	}
}

func (g *GCPKeyPairDeleter) Delete(publicKey string) error {
	g.DeleteCall.CallCount++
	g.DeleteCall.Receives.PublicKey = publicKey
	return g.DeleteCall.Returns.Error
}
