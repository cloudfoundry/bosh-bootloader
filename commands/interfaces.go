package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type plan interface {
	CheckFastFails([]string, storage.State) error
	ParseArgs([]string, storage.State) (PlanConfig, error)
	Execute([]string, storage.State) error
	InitializePlan(PlanConfig, storage.State) (storage.State, error)
	IsInitialized(storage.State) bool
}

type up interface {
	CheckFastFails([]string, storage.State) error
	ParseArgs([]string, storage.State) (PlanConfig, error)
	Execute([]string, storage.State) error
}

type terraformManager interface {
	ValidateVersion() error
	GetOutputs() (terraform.Outputs, error)
	Init(storage.State) error
	Apply(storage.State) (storage.State, error)
	Destroy(storage.State) (storage.State, error)
}

type boshManager interface {
	InitializeDirector(bblState storage.State) error
	CreateDirector(bblState storage.State, terraformOutputs terraform.Outputs) (storage.State, error)
	InitializeJumpbox(bblState storage.State) error
	CreateJumpbox(bblState storage.State, terraformOutputs terraform.Outputs) (storage.State, error)
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
	ReadAndValidate(certPath, keyPath, chainPath string) (certs.CertData, error)
	Read(certPath, keyPath, chainPath string) (certs.CertData, error)
}

type lbArgsHandler interface {
	GetLBState(string, CreateLBsConfig) (storage.LB, error)
	Merge(storage.LB, storage.LB) storage.LB
}

type createLBsCmd interface {
	Execute(state storage.State) error
}

type logger interface {
	Step(string, ...interface{})
	Printf(string, ...interface{})
	Println(string)
	Prompt(string)
}

type stateStore interface {
	Set(state storage.State) error
	GetVarsDir() (string, error)
}

type cloudConfigManager interface {
	Update(state storage.State) error
	Initialize(state storage.State) error
	GenerateVars(state storage.State) error
	Interpolate() (string, error)
	IsPresentCloudConfig() bool
	IsPresentCloudConfigVars() bool
}
