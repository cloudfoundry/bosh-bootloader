package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type StateStore struct {
	SetCall struct {
		CallCount int
		Receives  []SetCallReceive
		Returns   []SetCallReturn
	}

	GetCall struct {
		CallCount int
		Receives  struct {
			Dir string
		}
		Returns struct {
			State storage.State
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

	GetStateDirCall struct {
		CallCount int
		Returns   struct {
			Directory string
		}
	}

	GetBblDirCall struct {
		CallCount int
		Returns   struct {
			Directory string
			Error     error
		}
	}

	GetTerraformDirCall struct {
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

	GetDirectorDeploymentDirCall struct {
		CallCount int
		Returns   struct {
			Directory string
			Error     error
		}
	}

	GetJumpboxDeploymentDirCall struct {
		CallCount int
		Returns   struct {
			Directory string
			Error     error
		}
	}
}

type SetCallReceive struct {
	State storage.State
}

type SetCallReturn struct {
	Error error
}

func (s *StateStore) Set(state storage.State) error {
	s.SetCall.CallCount++

	s.SetCall.Receives = append(s.SetCall.Receives, SetCallReceive{State: state})

	if len(s.SetCall.Returns) < s.SetCall.CallCount {
		return nil
	}

	return s.SetCall.Returns[s.SetCall.CallCount-1].Error
}

func (s *StateStore) GetCloudConfigDir() (string, error) {
	s.GetCloudConfigDirCall.CallCount++

	return s.GetCloudConfigDirCall.Returns.Directory, s.GetCloudConfigDirCall.Returns.Error
}

func (s *StateStore) GetStateDir() string {
	s.GetStateDirCall.CallCount++

	return s.GetStateDirCall.Returns.Directory
}

func (s *StateStore) GetBblDir() (string, error) {
	s.GetBblDirCall.CallCount++

	return s.GetBblDirCall.Returns.Directory, s.GetBblDirCall.Returns.Error
}

func (s *StateStore) GetTerraformDir() (string, error) {
	s.GetTerraformDirCall.CallCount++

	return s.GetTerraformDirCall.Returns.Directory, s.GetTerraformDirCall.Returns.Error
}

func (s *StateStore) GetVarsDir() (string, error) {
	s.GetVarsDirCall.CallCount++

	return s.GetVarsDirCall.Returns.Directory, s.GetVarsDirCall.Returns.Error
}

func (s *StateStore) GetDirectorDeploymentDir() (string, error) {
	s.GetDirectorDeploymentDirCall.CallCount++

	return s.GetDirectorDeploymentDirCall.Returns.Directory, s.GetDirectorDeploymentDirCall.Returns.Error
}

func (s *StateStore) GetJumpboxDeploymentDir() (string, error) {
	s.GetJumpboxDeploymentDirCall.CallCount++

	return s.GetJumpboxDeploymentDirCall.Returns.Directory, s.GetJumpboxDeploymentDirCall.Returns.Error
}
