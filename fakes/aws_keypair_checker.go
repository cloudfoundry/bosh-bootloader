package fakes

type KeyPairChecker struct {
	HasKeyPairCall struct {
		CallCount int
		Stub      func(string) (bool, error)
		Recieves  struct {
			Name string
		}
		Returns struct {
			Present bool
			Error   error
		}
	}
}

func (k *KeyPairChecker) HasKeyPair(name string) (bool, error) {
	k.HasKeyPairCall.CallCount++
	k.HasKeyPairCall.Recieves.Name = name

	if k.HasKeyPairCall.Stub != nil {
		return k.HasKeyPairCall.Stub(name)
	}

	return k.HasKeyPairCall.Returns.Present,
		k.HasKeyPairCall.Returns.Error
}
