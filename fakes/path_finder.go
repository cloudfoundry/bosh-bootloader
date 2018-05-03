package fakes

type PathFinder struct {
	CommandExistsCall struct {
		CallCount int
		Receives  struct {
			Command string
		}
		Returns struct {
			Exists bool
		}
	}
}

func (p *PathFinder) CommandExists(command string) bool {
	p.CommandExistsCall.CallCount++
	p.CommandExistsCall.Receives.Command = command
	return p.CommandExistsCall.Returns.Exists
}
