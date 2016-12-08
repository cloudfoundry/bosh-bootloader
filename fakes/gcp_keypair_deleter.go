package fakes

type GCPKeyPairDeleter struct {
	DeleteCall struct {
		CallCount int
		Receives  struct {
			ProjectID string
			PublicKey string
		}
		Returns struct {
			Error error
		}
	}
}

func (g *GCPKeyPairDeleter) Delete(projectID, publicKey string) error {
	g.DeleteCall.CallCount++
	g.DeleteCall.Receives.ProjectID = projectID
	g.DeleteCall.Receives.PublicKey = publicKey
	return g.DeleteCall.Returns.Error
}
