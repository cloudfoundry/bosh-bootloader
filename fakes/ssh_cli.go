package fakes

import "os/exec"

type SSHCLI struct {
	RunCall struct {
		CallCount int
		Receives  [][]string
		Returns   []error
	}

	StartCall struct {
		CallCount int
		Receives  [][]string
		Returns   []SSHStartReturn
	}
}
type SSHRunReceive struct {
	Args []string
}
type SSHStartReturn struct {
	Cmd   *exec.Cmd
	Error error
}

func (s *SSHCLI) Run(args []string) error {
	s.RunCall.CallCount++

	s.RunCall.Receives = append(s.RunCall.Receives, args)

	if len(s.RunCall.Returns) < s.RunCall.CallCount {
		return nil
	}

	return s.RunCall.Returns[s.RunCall.CallCount-1]
}

func (s *SSHCLI) Start(args []string) (*exec.Cmd, error) {
	s.StartCall.CallCount++

	s.StartCall.Receives = append(s.StartCall.Receives, args)

	if len(s.StartCall.Returns) < s.StartCall.CallCount {
		return nil, nil
	}

	return s.StartCall.Returns[s.StartCall.CallCount-1].Cmd, s.StartCall.Returns[s.StartCall.CallCount-1].Error
}
