package fakes

type AllProxyGetter struct {
	GeneratePrivateKeyCall struct {
		CallCount int
		Returns   struct {
			PrivateKey string
			Error      error
		}
	}

	BoshAllProxyCall struct {
		CallCount int
		Receives  struct {
			JumpboxURL string
			PrivateKey string
		}
		Returns struct {
			URL string
		}
	}
}

func (a *AllProxyGetter) GeneratePrivateKey() (string, error) {
	a.GeneratePrivateKeyCall.CallCount++
	return a.GeneratePrivateKeyCall.Returns.PrivateKey, a.GeneratePrivateKeyCall.Returns.Error
}

func (a *AllProxyGetter) BoshAllProxy(jumpboxURL, privateKey string) string {
	a.BoshAllProxyCall.CallCount++
	a.BoshAllProxyCall.Receives.JumpboxURL = jumpboxURL
	a.BoshAllProxyCall.Receives.PrivateKey = privateKey

	return a.BoshAllProxyCall.Returns.URL
}
