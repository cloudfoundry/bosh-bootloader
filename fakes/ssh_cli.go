package fakes

type SSHCLI struct {
	RunCall struct {
		CallCount int
		Receives  []SSHRunReceive
		Returns   []SSHRunReturn
	}
}
type SSHRunReceive struct {
	Args []string
}
type SSHRunReturn struct {
	Error error
}

func (s *SSHCLI) Run(args []string) error {
	s.RunCall.CallCount++

	s.RunCall.Receives = append(s.RunCall.Receives, SSHRunReceive{
		Args: args,
	})

	if len(s.RunCall.Returns) < s.RunCall.CallCount {
		return nil
	}

	return s.RunCall.Returns[s.RunCall.CallCount-1].Error
}
