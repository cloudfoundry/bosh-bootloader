package fakes

type RuntimeStateStore struct {
	GetDirectorDeploymentDirCall struct {
		CallCount int
		Returns   struct {
			Dir   string
			Error error
		}
	}
}

func (ss *RuntimeStateStore) GetDirectorDeploymentDir() (string, error) {
	ss.GetDirectorDeploymentDirCall.CallCount++

	return ss.GetDirectorDeploymentDirCall.Returns.Dir, ss.GetDirectorDeploymentDirCall.Returns.Error
}
