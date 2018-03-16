package fakes

type SSHCmd struct {
	RunCall struct {
		Receives struct {
			Args []string
		}
		Returns struct {
			Error error
		}
	}
}

func (s *SSHCmd) Run(args []string) error {
	s.RunCall.Receives.Args = args
	return s.RunCall.Returns.Error
}
