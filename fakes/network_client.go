package fakes

type NetworkClient struct {
	CheckExistsCall struct {
		CallCount int
		Receives  struct {
			Name string
		}
		Returns struct {
			Exists bool
			Error  error
		}
	}
}

func (n *NetworkClient) CheckExists(name string) (bool, error) {
	n.CheckExistsCall.CallCount++
	n.CheckExistsCall.Receives.Name = name
	return n.CheckExistsCall.Returns.Exists, n.CheckExistsCall.Returns.Error
}
