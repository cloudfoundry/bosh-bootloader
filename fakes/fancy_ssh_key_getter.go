package fakes

type FancySSHKeyGetter struct {
	JumpboxGetCall struct {
		CallCount int
		Returns   struct {
			PrivateKey string
			Error      error
		}
	}
	DirectorGetCall struct {
		CallCount int
		Returns   struct {
			PrivateKey string
			Error      error
		}
	}
}

func (s *FancySSHKeyGetter) Get(deployment string) (string, error) {
	if deployment == "jumpbox" {
		s.JumpboxGetCall.CallCount++

		return s.JumpboxGetCall.Returns.PrivateKey, s.JumpboxGetCall.Returns.Error
	} else if deployment == "director" {
		s.DirectorGetCall.CallCount++

		return s.DirectorGetCall.Returns.PrivateKey, s.DirectorGetCall.Returns.Error
	} else {
		panic("expected deployment to be either jumpbox or director")
	}
}
