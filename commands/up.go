package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type keyPairSynchronizer interface {
	Sync(keypair ec2.KeyPair) (ec2.KeyPair, error)
}

type infrastructureManager interface {
	Create(keyPairName string, numberOfAZs int, stackName string, lbType string, lbCertificateARN string) (cloudformation.Stack, error)
	Update(keyPairName string, numberOfAZs int, stackName string, lbType string, lbCertificateARN string) (cloudformation.Stack, error)
	Exists(stackName string) (bool, error)
	Delete(stackName string) error
	Describe(stackName string) (cloudformation.Stack, error)
}

type boshDeployer interface {
	Deploy(boshinit.DeployInput) (boshinit.DeployOutput, error)
}

type cloudConfigurator interface {
	Configure(stack cloudformation.Stack, azs []string, boshClient bosh.Client) error
}

type availabilityZoneRetriever interface {
	Retrieve(region string) ([]string, error)
}

type elbDescriber interface {
	Describe(elbName string) ([]string, error)
}

type loadBalancerCertificateManager interface {
	Create(input iam.CertificateCreateInput) (iam.CertificateCreateOutput, error)
	IsValidLBType(lbType string) bool
}

type awsCredentialValidator interface {
	Validate() error
}

type logger interface {
	Step(string)
	Println(string)
	Prompt(string)
}

type upConfig struct {
	lbType   string
	certPath string
	keyPath  string
}

type Up struct {
	awsCredentialValidator         awsCredentialValidator
	infrastructureManager          infrastructureManager
	keyPairSynchronizer            keyPairSynchronizer
	boshDeployer                   boshDeployer
	stringGenerator                stringGenerator
	cloudConfigurator              cloudConfigurator
	availabilityZoneRetriever      availabilityZoneRetriever
	elbDescriber                   elbDescriber
	loadBalancerCertificateManager loadBalancerCertificateManager
}

func NewUp(
	awsCredentialValidator awsCredentialValidator, infrastructureManager infrastructureManager,
	keyPairSynchronizer keyPairSynchronizer, boshDeployer boshDeployer, stringGenerator stringGenerator,
	cloudConfigurator cloudConfigurator, availabilityZoneRetriever availabilityZoneRetriever,
	elbDescriber elbDescriber, loadBalancerCertificateManager loadBalancerCertificateManager) Up {

	return Up{
		awsCredentialValidator:         awsCredentialValidator,
		infrastructureManager:          infrastructureManager,
		keyPairSynchronizer:            keyPairSynchronizer,
		boshDeployer:                   boshDeployer,
		stringGenerator:                stringGenerator,
		cloudConfigurator:              cloudConfigurator,
		availabilityZoneRetriever:      availabilityZoneRetriever,
		elbDescriber:                   elbDescriber,
		loadBalancerCertificateManager: loadBalancerCertificateManager,
	}
}

func (u Up) Execute(subcommandFlags []string, state storage.State) (storage.State, error) {
	err := u.awsCredentialValidator.Validate()
	if err != nil {
		return state, err
	}

	config, err := u.parseFlags(subcommandFlags)
	if err != nil {
		return state, err
	}

	err = u.checkForFastFails(config, state)
	if err != nil {
		return state, err
	}

	certOutput, err := u.loadBalancerCertificateManager.Create(iam.CertificateCreateInput{
		CurrentLBType:          state.Stack.LBType,
		DesiredLBType:          config.lbType,
		CurrentCertificateName: state.Stack.CertificateName,
		CertPath:               config.certPath,
		KeyPath:                config.keyPath,
	})
	if err != nil {
		return state, err
	}
	state.Stack.CertificateName = certOutput.CertificateName

	keyPair, err := u.keyPairSynchronizer.Sync(ec2.KeyPair{
		Name:       state.KeyPair.Name,
		PublicKey:  state.KeyPair.PublicKey,
		PrivateKey: state.KeyPair.PrivateKey,
	})
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
		state.Stack.Name, err = u.stringGenerator.Generate("bbl-aws-", 5)
		if err != nil {
			return state, err
		}
	}

	stack, err := u.infrastructureManager.Create(state.KeyPair.Name, len(availabilityZones), state.Stack.Name, certOutput.LBType, certOutput.CertificateARN)
	if err != nil {
		return state, err
	}
	state.Stack.LBType = certOutput.LBType

	infrastructureConfiguration := boshinit.InfrastructureConfiguration{
		AWSRegion:        state.AWS.Region,
		SubnetID:         stack.Outputs["BOSHSubnet"],
		AvailabilityZone: stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        stack.Outputs["BOSHEIP"],
		AccessKeyID:      stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    stack.Outputs["BOSHSecurityGroup"],
	}

	deployInput, err := boshinit.NewDeployInput(state, infrastructureConfiguration, u.stringGenerator)
	if err != nil {
		return state, err
	}

	deployOutput, err := u.boshDeployer.Deploy(deployInput)
	if err != nil {
		return state, err
	}

	if state.BOSH.IsEmpty() {
		state.BOSH = storage.BOSH{
			DirectorAddress:        stack.Outputs["BOSHURL"],
			DirectorUsername:       deployInput.DirectorUsername,
			DirectorPassword:       deployInput.DirectorPassword,
			DirectorSSLCertificate: string(deployOutput.DirectorSSLKeyPair.Certificate),
			DirectorSSLPrivateKey:  string(deployOutput.DirectorSSLKeyPair.PrivateKey),
			Credentials:            deployOutput.Credentials,
		}
	}

	state.BOSH.State = deployOutput.BOSHInitState
	state.BOSH.Manifest = deployOutput.BOSHInitManifest

	boshClient := bosh.NewClient(
		stack.Outputs["BOSHURL"],
		deployInput.DirectorUsername,
		deployInput.DirectorPassword,
	)
	err = u.cloudConfigurator.Configure(stack, availabilityZones, boshClient)
	if err != nil {
		return state, err
	}

	return state, nil
}

func (Up) parseFlags(subcommandFlags []string) (upConfig, error) {
	upFlags := flags.New("unsupported-deploy-bosh-on-aws-for-concourse")

	config := upConfig{}
	upFlags.String(&config.lbType, "lb-type", "")
	upFlags.String(&config.certPath, "cert", "")
	upFlags.String(&config.keyPath, "key", "")

	err := upFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (u Up) checkForFastFails(config upConfig, state storage.State) error {
	if !u.loadBalancerCertificateManager.IsValidLBType(config.lbType) {
		return fmt.Errorf("Unknown lb-type %q, supported lb-types are: concourse, cf or none", config.lbType)
	}

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

	if stackExists && state.Stack.LBType != config.lbType {
		stack, err := u.infrastructureManager.Describe(state.Stack.Name)
		if err != nil {
			return err
		}

		for _, lbName := range []string{"ConcourseLoadBalancer", "CFSSHProxyLoadBalancer", "CFRouterLoadBalancer"} {
			if id, ok := stack.Outputs[lbName]; ok {
				instances, err := u.elbDescriber.Describe(id)
				if err != nil {
					return err
				}

				if len(instances) > 0 {
					return fmt.Errorf("Load balancer %q cannot be deleted since it has attached instances: %s", id, strings.Join(instances, ", "))
				}
			}
		}
	}

	return nil
}
