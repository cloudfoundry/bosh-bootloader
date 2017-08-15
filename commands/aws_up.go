package commands

import (
	"errors"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type infrastructureManager interface {
	Exists(stackName string) (bool, error)
	Delete(stackName string) error
	Describe(stackName string) (cloudformation.Stack, error)
}

type credentialValidator interface {
	Validate() error
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

type configProvider interface {
	SetConfig(config aws.Config)
}

type cloudConfigManager interface {
	Update(state storage.State) error
	Generate(state storage.State) (string, error)
}

type brokenEnvironmentValidator interface {
	Validate(state storage.State) error
}

type AWSUp struct {
	credentialValidator        credentialValidator
	boshManager                boshManager
	cloudConfigManager         cloudConfigManager
	stateStore                 stateStore
	configProvider             configProvider
	envIDManager               envIDManager
	terraformManager           terraformApplier
	brokenEnvironmentValidator brokenEnvironmentValidator
}

type AWSUpConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	OpsFilePath     string
	BOSHAZ          string
	Name            string
	NoDirector      bool
	Terraform       bool
}

func NewAWSUp(
	credentialValidator credentialValidator,
	boshManager boshManager,
	cloudConfigManager cloudConfigManager,
	stateStore stateStore, configProvider configProvider, envIDManager envIDManager,
	terraformManager terraformApplier, brokenEnvironmentValidator brokenEnvironmentValidator) AWSUp {

	return AWSUp{
		credentialValidator:        credentialValidator,
		boshManager:                boshManager,
		cloudConfigManager:         cloudConfigManager,
		stateStore:                 stateStore,
		configProvider:             configProvider,
		envIDManager:               envIDManager,
		terraformManager:           terraformManager,
		brokenEnvironmentValidator: brokenEnvironmentValidator,
	}
}

func (u AWSUp) Execute(config AWSUpConfig, state storage.State) error {
	state.IAAS = "aws"

	if u.awsCredentialsPresent(config) {
		state.AWS.AccessKeyID = config.AccessKeyID
		state.AWS.SecretAccessKey = config.SecretAccessKey
		state.AWS.Region = config.Region
		if err := u.stateStore.Set(state); err != nil {
			return err
		}
		u.configProvider.SetConfig(aws.Config{
			AccessKeyID:     config.AccessKeyID,
			SecretAccessKey: config.SecretAccessKey,
			Region:          config.Region,
		})
	} else if u.awsCredentialsNotPresent(config) {
		err := u.credentialValidator.Validate()
		if err != nil {
			return err
		}
	} else {
		return u.awsMissingCredentials(config)
	}

	if config.NoDirector {
		if !state.BOSH.IsEmpty() {
			return errors.New(`Director already exists, you must re-create your environment to use "--no-director"`)
		}

		state.NoDirector = true
	}

	err := u.checkForFastFails(state, config)
	if err != nil {
		return err
	}

	state, err = u.envIDManager.Sync(state, config.Name)
	if err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	state.Stack.BOSHAZ = config.BOSHAZ
	state, err = u.terraformManager.Apply(state)
	if err != nil {
		return handleTerraformError(err, u.stateStore)
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return err
	}

	terraformOutputs, err := u.terraformManager.GetOutputs(state)
	if err != nil {
		return err
	}

	if !state.NoDirector {
		opsFile := []byte{}
		if config.OpsFilePath != "" {
			opsFile, err = ioutil.ReadFile(config.OpsFilePath)
			if err != nil {
				return err
			}
		}
		state.BOSH.UserOpsFile = string(opsFile)

		state, err = u.boshManager.CreateDirector(state, terraformOutputs)
		switch err.(type) {
		case bosh.ManagerCreateError:
			bcErr := err.(bosh.ManagerCreateError)
			if setErr := u.stateStore.Set(bcErr.State()); setErr != nil {
				errorList := helpers.Errors{}
				errorList.Add(err)
				errorList.Add(setErr)
				return errorList
			}
			return err
		case error:
			return err
		}

		err = u.stateStore.Set(state)
		if err != nil {
			return err
		}

		err = u.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u AWSUp) checkForFastFails(state storage.State, config AWSUpConfig) error {
	err := u.brokenEnvironmentValidator.Validate(state)
	if err != nil {
		return err
	}

	if state.Stack.Name != "" && state.Stack.BOSHAZ != config.BOSHAZ {
		return errors.New("The --aws-bosh-az cannot be changed for existing environments.")
	}

	return nil
}

func (AWSUp) awsCredentialsPresent(config AWSUpConfig) bool {
	return config.AccessKeyID != "" && config.SecretAccessKey != "" && config.Region != ""
}

func (AWSUp) awsCredentialsNotPresent(config AWSUpConfig) bool {
	return config.AccessKeyID == "" && config.SecretAccessKey == "" && config.Region == ""
}

func (AWSUp) awsMissingCredentials(config AWSUpConfig) error {
	switch {
	case config.AccessKeyID == "":
		return errors.New("AWS access key ID must be provided")
	case config.SecretAccessKey == "":
		return errors.New("AWS secret access key must be provided")
	case config.Region == "":
		return errors.New("AWS region must be provided")
	}

	return nil
}
