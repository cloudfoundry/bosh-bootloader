package fakes

type DirProvider struct {
	GetDirectorDeploymentDirCall struct {
		CallCount int
		Returns   struct {
			Dir   string
			Error error
		}
	}

	GetCloudConfigDirCall struct {
		CallCount int
		Returns   struct {
			Directory string
			Error     error
		}
	}

	GetVarsDirCall struct {
		CallCount int
		Returns   struct {
			Directory string
			Error     error
		}
	}

	GetRuntimeConfigDirCall struct {
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

func (d *DirProvider) GetCloudConfigDir() (string, error) {
	d.GetCloudConfigDirCall.CallCount++

	return d.GetCloudConfigDirCall.Returns.Directory, d.GetCloudConfigDirCall.Returns.Error
}

func (d *DirProvider) GetVarsDir() (string, error) {
	d.GetVarsDirCall.CallCount++

	return d.GetVarsDirCall.Returns.Directory, d.GetVarsDirCall.Returns.Error
}

func (d *DirProvider) GetRuntimeConfigDir() (string, error) {
	d.GetRuntimeConfigDirCall.CallCount++

	return d.GetRuntimeConfigDirCall.Returns.Dir, d.GetRuntimeConfigDirCall.Returns.Error
}
