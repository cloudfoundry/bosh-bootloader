package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type up interface {
	CheckFastFails([]string, storage.State) error
	ParseArgs([]string, storage.State) (UpConfig, error)
	Execute([]string, storage.State) error
}

type terraformManager interface {
	ValidateVersion() error
	GetOutputs(storage.State) (terraform.Outputs, error)
	Init(storage.State) error
	Apply(storage.State) (storage.State, error)
	Destroy(storage.State) (storage.State, error)
}

type terraformOutputter interface {
	GetOutputs(storage.State) (terraform.Outputs, error)
}

type boshManager interface {
	InitializeDirector(bblState storage.State, terraformOutputs terraform.Outputs) error
	CreateDirector(bblState storage.State) (storage.State, error)
	InitializeJumpbox(bblState storage.State, terraformOutputs terraform.Outputs) error
	CreateJumpbox(bblState storage.State, jumpboxURL string) (storage.State, error)
	DeleteDirector(bblState storage.State, terraformOutputs terraform.Outputs) error
	DeleteJumpbox(bblState storage.State, terraformOutputs terraform.Outputs) error
	GetDirectorDeploymentVars(bblState storage.State, terraformOutputs terraform.Outputs) string
	GetJumpboxDeploymentVars(bblState storage.State, terraformOutputs terraform.Outputs) string
	Version() (string, error)
}

type envIDManager interface {
	Sync(storage.State, string) (storage.State, error)
}

type environmentValidator interface {
	Validate(state storage.State) error
}

type terraformManagerError interface {
	Error() string
	BBLState() (storage.State, error)
}

type vpcStatusChecker interface {
	ValidateSafeToDelete(vpcID string, envID string) error
}

type certificateDeleter interface {
	Delete(certificateName string) error
}

type stateValidator interface {
	Validate() error
}

type certificateValidator interface {
	Validate(command, certPath, keyPath, chainPath string) error
}

type logger interface {
	Step(string, ...interface{})
	Printf(string, ...interface{})
	Println(string)
	Prompt(string)
}

type stateStore interface {
	Set(state storage.State) error
	GetBblDir() (string, error)
}

type cloudConfigManager interface {
	Update(state storage.State) error
	Generate(state storage.State) (string, error)
}
