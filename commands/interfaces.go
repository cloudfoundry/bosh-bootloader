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
	Setup(storage.State) error
	Init(storage.State) error
	Apply(storage.State) (storage.State, error)
	Validate(storage.State) (storage.State, error)
	Destroy(storage.State) (storage.State, error)
	IsPaved() (bool, error)
}

type boshManager interface {
	InitializeDirector(bblState storage.State) error
	CreateDirector(bblState storage.State, terraformOutputs terraform.Outputs) (storage.State, error)
	InitializeJumpbox(bblState storage.State) error
	CleanUpDirector(state storage.State) error
	CreateJumpbox(bblState storage.State, terraformOutputs terraform.Outputs) (storage.State, error)
	DeleteDirector(bblState storage.State, terraformOutputs terraform.Outputs) error
	DeleteJumpbox(bblState storage.State, terraformOutputs terraform.Outputs) error
	GetDirectorDeploymentVars(bblState storage.State, terraformOutputs terraform.Outputs) string
	GetJumpboxDeploymentVars(bblState storage.State, terraformOutputs terraform.Outputs) string
	Path() string
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
	ReadAndValidate(certPath, keyPath string) (certs.CertData, error)
	Read(certPath, keyPath string) (certs.CertData, error)
	ReadPKCS12(certPath, passwordPath string) (certs.CertData, error)
	ReadAndValidatePKCS12(certPath, passwordPath string) (certs.CertData, error)
}

type lbArgsHandler interface {
	GetLBState(string, LBArgs) (storage.LB, error)
	Merge(storage.LB, storage.LB) storage.LB
}

type createLBsCmd interface {
	Execute(state storage.State) error
}

type logger interface {
	Step(string, ...interface{})
	Printf(string, ...interface{})
	Println(string)
	Prompt(string) bool
}

type stateStore interface {
	Set(state storage.State) error
	GetOldBblDir() string
	GetVarsDir() (string, error)
	GetCloudConfigDir() (string, error)
}

type cloudConfigManager interface {
	Update(state storage.State) error
	Initialize(state storage.State) error
	IsPresentCloudConfig() bool
	IsPresentCloudConfigVars() bool
}

type runtimeConfigManager interface {
	Initialize(state storage.State) error
	Update(state storage.State) error
}
