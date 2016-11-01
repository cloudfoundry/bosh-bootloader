package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/boshinit"
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	UpCommand = "up"
)

type keyPairSynchronizer interface {
	Sync(keypair ec2.KeyPair) (ec2.KeyPair, error)
}

type infrastructureManager interface {
	Create(keyPairName string, numberOfAZs int, stackName, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error)
	Update(keyPairName string, numberOfAZs int, stackName, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error)
	Exists(stackName string) (bool, error)
	Delete(stackName string) error
	Describe(stackName string) (cloudformation.Stack, error)
}

type boshDeployer interface {
	Deploy(boshinit.DeployInput) (boshinit.DeployOutput, error)
}

type availabilityZoneRetriever interface {
	Retrieve(region string) ([]string, error)
}

type awsCredentialValidator interface {
	Validate() error
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

type envIDGenerator interface {
	Generate() (string, error)
}

type configProvider interface {
	SetConfig(config aws.Config)
}

type Up struct {
	awsCredentialValidator    awsCredentialValidator
	infrastructureManager     infrastructureManager
	keyPairSynchronizer       keyPairSynchronizer
	boshDeployer              boshDeployer
	stringGenerator           stringGenerator
	boshCloudConfigurator     boshCloudConfigurator
	availabilityZoneRetriever availabilityZoneRetriever
	certificateDescriber      certificateDescriber
	cloudConfigManager        cloudConfigManager
	boshClientProvider        boshClientProvider
	envIDGenerator            envIDGenerator
	stateStore                stateStore
	configProvider            configProvider
}

type upConfig struct {
	awsAccessKeyID     string
	awsSecretAccessKey string
	awsRegion          string
	iaas               string
}

func NewUp(
	awsCredentialValidator awsCredentialValidator, infrastructureManager infrastructureManager,
	keyPairSynchronizer keyPairSynchronizer, boshDeployer boshDeployer, stringGenerator stringGenerator,
	boshCloudConfigurator boshCloudConfigurator, availabilityZoneRetriever availabilityZoneRetriever,
	certificateDescriber certificateDescriber, cloudConfigManager cloudConfigManager,
	boshClientProvider boshClientProvider, envIDGenerator envIDGenerator, stateStore stateStore,
	configProvider configProvider) Up {

	return Up{
		awsCredentialValidator:    awsCredentialValidator,
		infrastructureManager:     infrastructureManager,
		keyPairSynchronizer:       keyPairSynchronizer,
		boshDeployer:              boshDeployer,
		stringGenerator:           stringGenerator,
		boshCloudConfigurator:     boshCloudConfigurator,
		availabilityZoneRetriever: availabilityZoneRetriever,
		certificateDescriber:      certificateDescriber,
		cloudConfigManager:        cloudConfigManager,
		boshClientProvider:        boshClientProvider,
		envIDGenerator:            envIDGenerator,
		stateStore:                stateStore,
		configProvider:            configProvider,
	}
}

func (u Up) Execute(subcommandFlags []string, state storage.State) error {
	config, err := u.parseFlags(subcommandFlags)
	if err != nil {
		return err
	}

	if config.iaas != "" && config.iaas != "gcp" && config.iaas != "aws" {
		return fmt.Errorf("%q is invalid; supported values: [gcp, aws]", config.iaas)
	}

	if state.IAAS == "" {
		if config.iaas == "" {
			return errors.New("--iaas [gcp, aws] must be provided")
		}

		state.IAAS = config.iaas
	}

	if config.iaas != "" && state.IAAS != config.iaas {
		return errors.New("the iaas provided must match the iaas in bbl-state.json")
	}

	if state.IAAS == "gcp" {
		if err := u.stateStore.Set(state); err != nil {
			return err
		}
		return nil
	}

	if u.awsCredentialsPresent(config) {
		state.AWS.AccessKeyID = config.awsAccessKeyID
		state.AWS.SecretAccessKey = config.awsSecretAccessKey
		state.AWS.Region = config.awsRegion
		if err := u.stateStore.Set(state); err != nil {
			return err
		}
		u.configProvider.SetConfig(aws.Config{
			AccessKeyID:     config.awsAccessKeyID,
			SecretAccessKey: config.awsSecretAccessKey,
			Region:          config.awsRegion,
		})
	} else if u.awsCredentialsNotPresent(config) {
		err := u.awsCredentialValidator.Validate()
		if err != nil {
			return err
		}
	} else {
		return u.awsMissingCredentials(config)
	}

	err = u.checkForFastFails(state)
	if err != nil {
		return err
	}

	if state.EnvID == "" {
		state.EnvID, err = u.envIDGenerator.Generate()
		if err != nil {
			return err
		}
	}

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
		stackEnvID := strings.Replace(state.EnvID, ":", "-", -1)
		state.Stack.Name = fmt.Sprintf("stack-%s", stackEnvID)

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

	stack, err := u.infrastructureManager.Create(state.KeyPair.Name, len(availabilityZones), state.Stack.Name, state.Stack.LBType, certificateARN, state.EnvID)
	if err != nil {
		return err
	}

	infrastructureConfiguration := boshinit.InfrastructureConfiguration{
		AWSRegion:        state.AWS.Region,
		SubnetID:         stack.Outputs["BOSHSubnet"],
		AvailabilityZone: stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        stack.Outputs["BOSHEIP"],
		AccessKeyID:      stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    stack.Outputs["BOSHSecurityGroup"],
	}

	deployInput, err := boshinit.NewDeployInput(state, infrastructureConfiguration, u.stringGenerator, state.EnvID)
	if err != nil {
		return err
	}

	deployOutput, err := u.boshDeployer.Deploy(deployInput)
	if err != nil {
		return err
	}

	if state.BOSH.IsEmpty() {
		state.BOSH = storage.BOSH{
			DirectorName:           deployInput.DirectorName,
			DirectorAddress:        stack.Outputs["BOSHURL"],
			DirectorUsername:       deployInput.DirectorUsername,
			DirectorPassword:       deployInput.DirectorPassword,
			DirectorSSLCA:          string(deployOutput.DirectorSSLKeyPair.CA),
			DirectorSSLCertificate: string(deployOutput.DirectorSSLKeyPair.Certificate),
			DirectorSSLPrivateKey:  string(deployOutput.DirectorSSLKeyPair.PrivateKey),
			Credentials:            deployOutput.Credentials,
		}
	}

	state.BOSH.State = deployOutput.BOSHInitState
	state.BOSH.Manifest = deployOutput.BOSHInitManifest

	err = u.stateStore.Set(state)
	if err != nil {
		return err
	}

	boshClient := u.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword)

	cloudConfigInput := u.boshCloudConfigurator.Configure(stack, availabilityZones)

	err = u.cloudConfigManager.Update(cloudConfigInput, boshClient)
	if err != nil {
		return err
	}

	err = u.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}

func (u Up) checkForFastFails(state storage.State) error {
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

	return nil
}

func (Up) parseFlags(subcommandFlags []string) (upConfig, error) {
	upFlags := flags.New("up")

	config := upConfig{}
	upFlags.String(&config.awsAccessKeyID, "aws-access-key-id", os.Getenv("BBL_AWS_ACCESS_KEY_ID"))
	upFlags.String(&config.awsSecretAccessKey, "aws-secret-access-key", os.Getenv("BBL_AWS_SECRET_ACCESS_KEY"))
	upFlags.String(&config.awsRegion, "aws-region", os.Getenv("BBL_AWS_REGION"))
	upFlags.String(&config.iaas, "iaas", "")

	err := upFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (Up) awsCredentialsPresent(config upConfig) bool {
	return config.awsAccessKeyID != "" && config.awsSecretAccessKey != "" && config.awsRegion != ""
}

func (Up) awsCredentialsNotPresent(config upConfig) bool {
	return config.awsAccessKeyID == "" && config.awsSecretAccessKey == "" && config.awsRegion == ""
}

func (Up) awsMissingCredentials(config upConfig) error {
	switch {
	case config.awsAccessKeyID == "":
		return errors.New("AWS access key ID must be provided")
	case config.awsSecretAccessKey == "":
		return errors.New("AWS secret access key must be provided")
	case config.awsRegion == "":
		return errors.New("AWS region must be provided")
	}

	return nil
}
