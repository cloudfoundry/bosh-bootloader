package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type keyPairSynchronizer interface {
	Sync(keypair ec2.KeyPair, envID string) (ec2.KeyPair, error)
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
	Step(string)
	Println(string)
	Prompt(string)
}

type certificateDescriber interface {
	Describe(certificateName string) (iam.Certificate, error)
}

type envIDGenerator interface {
	Generate() (string, error)
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
}

func NewUp(
	awsCredentialValidator awsCredentialValidator, infrastructureManager infrastructureManager,
	keyPairSynchronizer keyPairSynchronizer, boshDeployer boshDeployer, stringGenerator stringGenerator,
	boshCloudConfigurator boshCloudConfigurator, availabilityZoneRetriever availabilityZoneRetriever,
	certificateDescriber certificateDescriber, cloudConfigManager cloudConfigManager,
	boshClientProvider boshClientProvider, envIDGenerator envIDGenerator) Up {

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
	}
}

func (u Up) Execute(subcommandFlags []string, state storage.State) (storage.State, error) {
	err := u.awsCredentialValidator.Validate()
	if err != nil {
		return state, err
	}

	err = u.checkForFastFails(state)
	if err != nil {
		return state, err
	}

	if state.EnvID == "" {
		state.EnvID, err = u.envIDGenerator.Generate()
		if err != nil {
			return state, err
		}
	}

	keyPair, err := u.keyPairSynchronizer.Sync(ec2.KeyPair{
		Name:       state.KeyPair.Name,
		PublicKey:  state.KeyPair.PublicKey,
		PrivateKey: state.KeyPair.PrivateKey,
	}, state.EnvID)
	if err != nil {
		return state, err
	}

	state.KeyPair.Name = keyPair.Name
	state.KeyPair.PublicKey = keyPair.PublicKey
	state.KeyPair.PrivateKey = keyPair.PrivateKey

	availabilityZones, err := u.availabilityZoneRetriever.Retrieve(state.AWS.Region)
	if err != nil {
		return state, err
	}

	if state.Stack.Name == "" {
		stackEnvID := strings.Replace(state.EnvID, ":", "-", -1)
		state.Stack.Name = fmt.Sprintf("stack-%s", stackEnvID)
	}

	var certificateARN string
	if lbExists(state.Stack.LBType) {
		certificate, err := u.certificateDescriber.Describe(state.Stack.CertificateName)
		if err != nil {
			return state, err
		}
		certificateARN = certificate.ARN
	}

	stack, err := u.infrastructureManager.Create(state.KeyPair.Name, len(availabilityZones), state.Stack.Name, state.Stack.LBType, certificateARN, state.EnvID)
	if err != nil {
		return state, err
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
		return state, err
	}

	deployOutput, err := u.boshDeployer.Deploy(deployInput)
	if err != nil {
		return state, err
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

	boshClient := u.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword)

	cloudConfigInput := u.boshCloudConfigurator.Configure(stack, availabilityZones)

	err = u.cloudConfigManager.Update(cloudConfigInput, boshClient)
	if err != nil {
		return state, err
	}

	return state, nil
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
				"https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you need assistance.",
			state.Stack.Name, state.AWS.Region)
	}

	return nil
}
