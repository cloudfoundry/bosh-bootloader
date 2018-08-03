package fakes

type DirProvider struct {
	GetDirectorDeploymentDirCall struct {
		CallCount int
		Returns   struct {
			Dir   string
			Error error
		}
	}
}

func (d *DirProvider) GetDirectorDeploymentDir() (string, error) {
	d.GetDirectorDeploymentDirCall.CallCount++

	return d.GetDirectorDeploymentDirCall.Returns.Dir, d.GetDirectorDeploymentDirCall.Returns.Error
}
