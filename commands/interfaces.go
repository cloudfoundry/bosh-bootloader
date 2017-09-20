package commands

import (
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

//go:generate counterfeiter -o ./fakes/terraform_applier.go --fake-name TerraformApplier . terraformApplier
type terraformApplier interface {
	ValidateVersion() error
	GetOutputs(storage.State) (map[string]interface{}, error)
	Apply(storage.State) (storage.State, error)
}

type terraformDestroyer interface {
	ValidateVersion() error
	GetOutputs(storage.State) (map[string]interface{}, error)
	Destroy(storage.State) (storage.State, error)
}

type terraformOutputter interface {
	GetOutputs(storage.State) (map[string]interface{}, error)
}

type boshManager interface {
	CreateDirector(bblState storage.State, terraformOutputs map[string]interface{}) (storage.State, error)
	CreateJumpbox(bblState storage.State, terraformOutputs map[string]interface{}) (storage.State, error)
	Delete(bblState storage.State, terraformOutputs map[string]interface{}) error
	DeleteJumpbox(bblState storage.State, terraformOutputs map[string]interface{}) error
	GetDirectorDeploymentVars(bblState storage.State, terraformOutputs map[string]interface{}) string
	GetJumpboxDeploymentVars(bblState storage.State, terraformOutputs map[string]interface{}) string
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

type stackManager interface {
	Describe(string) (cloudformation.Stack, error)
}

type certificateDeleter interface {
	Delete(certificateName string) error
}

type stateValidator interface {
	Validate() error
}

type networkInstancesChecker interface {
	ValidateSafeToDelete(networkName string) error
}

type certificateValidator interface {
	Validate(command, certPath, keyPath, chainPath string) error
}

type infrastructureManager interface {
	Exists(stackName string) (bool, error)
	Delete(stackName string) error
	Describe(stackName string) (cloudformation.Stack, error)
}

type logger interface {
	Step(string, ...interface{})
	Printf(string, ...interface{})
	Println(string)
	Prompt(string)
}

type stateStore interface {
	Set(state storage.State) error
}

type cloudConfigManager interface {
	Update(state storage.State) error
	Generate(state storage.State) (string, error)
}
