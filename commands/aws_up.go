package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	UpCommand = "up"
)

type keyPairSynchronizer interface {
	Sync(keypair ec2.KeyPair) (ec2.KeyPair, error)
}

type infrastructureManager interface {
	Create(keyPairName string, azs []string, stackName, boshAZ, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error)
	Update(keyPairName string, azs []string, stackName, boshAZ, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error)
	Exists(stackName string) (bool, error)
	Delete(stackName string) error
	Describe(stackName string) (cloudformation.Stack, error)
}

type availabilityZoneRetriever interface {
	Retrieve(region string) ([]string, error)
}

type credentialValidator interface {
	ValidateAWS() error
	ValidateGCP() error
}

type logger interface {
	Step(string, ...interface{})
	Println(string)
	Prompt(string)
}

type stateStore interface {
	Set(state storage.State) error
}

type certificateDescriber interface {
	Describe(certificateName string) (iam.Certificate, error)
}

type configProvider interface {
	SetConfig(config aws.Config)
}

type cloudConfigManager interface {
	Update(state storage.State) error
	Generate(state storage.State) (string, error)
}

type AWSUp struct {
	credentialValidator       credentialValidator
	infrastructureManager     infrastructureManager
	keyPairSynchronizer       keyPairSynchronizer
	boshManager               boshManager
	availabilityZoneRetriever availabilityZoneRetriever
	certificateDescriber      certificateDescriber
	cloudConfigManager        cloudConfigManager
	stateStore                stateStore
	configProvider            configProvider
	envIDManager              envIDManager
}

type AWSUpConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	OpsFilePath     string
	BOSHAZ          string
	Name            string
	NoDirector      bool
}

func NewAWSUp(
	credentialValidator credentialValidator, infrastructureManager infrastructureManager,
	keyPairSynchronizer keyPairSynchronizer, boshManager boshManager,
	availabilityZoneRetriever availabilityZoneRetriever,
	certificateDescriber certificateDescriber, cloudConfigManager cloudConfigManager,
	stateStore stateStore,
	configProvider configProvider, envIDManager envIDManager) AWSUp {

	return AWSUp{
		credentialValidator:       credentialValidator,
		infrastructureManager:     infrastructureManager,
		keyPairSynchronizer:       keyPairSynchronizer,
		boshManager:               boshManager,
		availabilityZoneRetriever: availabilityZoneRetriever,
		certificateDescriber:      certificateDescriber,
		cloudConfigManager:        cloudConfigManager,
		stateStore:                stateStore,
		configProvider:            configProvider,
		envIDManager:              envIDManager,
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
		err := u.credentialValidator.ValidateAWS()
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

	envID, err := u.envIDManager.Sync(state, config.Name)
	if err != nil {
		return err
	}

	state.EnvID = envID

	if state.KeyPair.Name == "" {
		state.KeyPair.Name = fmt.Sprintf("keypair-%s", state.EnvID)
	}

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	keyPair, err := u.keyPairSynchronizer.Sync(ec2.KeyPair{
		Name:       state.KeyPair.Name,
		PublicKey:  state.KeyPair.PublicKey,
		PrivateKey: state.KeyPair.PrivateKey,
	})
	if err != nil {
		return err
	}

	state.KeyPair.PublicKey = keyPair.PublicKey
	state.KeyPair.PrivateKey = keyPair.PrivateKey

	if err := u.stateStore.Set(state); err != nil {
		return err
	}

	availabilityZones, err := u.availabilityZoneRetriever.Retrieve(state.AWS.Region)
	if err != nil {
		return err
	}

	if state.Stack.Name == "" {
		state.Stack.Name = fmt.Sprintf("stack-%s", strings.Replace(state.EnvID, ":", "-", -1))
		state.Stack.BOSHAZ = config.BOSHAZ

		if err := u.stateStore.Set(state); err != nil {
			return err
		}
	}

	var certificateARN string
	if lbExists(state.Stack.LBType) {
		certificate, err := u.certificateDescriber.Describe(state.Stack.CertificateName)
		if err != nil {
			return err
		}
		certificateARN = certificate.ARN
	}

	_, err = u.infrastructureManager.Create(state.KeyPair.Name, availabilityZones, state.Stack.Name, state.Stack.BOSHAZ, state.Stack.LBType, certificateARN, state.EnvID)
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

		state, err = u.boshManager.Create(state, opsFile)
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
	stackExists, err := u.infrastructureManager.Exists(state.Stack.Name)
	if err != nil {
		return err
	}

	if !state.BOSH.IsEmpty() && !stackExists {
		return fmt.Errorf(
			"Found BOSH data in state directory, but Cloud Formation stack %q cannot be found "+
				"for region %q and given AWS credentials. bbl cannot safely proceed. Open an issue on GitHub at "+
				"https://github.com/cloudfoundry/bosh-bootloader/issues/new if you need assistance.",
			state.Stack.Name, state.AWS.Region)
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
