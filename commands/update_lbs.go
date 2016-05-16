package commands

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type updateLBConfig struct {
	certPath string
	keyPath  string
}

type UpdateLBs struct {
	certificateManager        certificateManager
	clientProvider            awsClientProvider
	availabilityZoneRetriever availabilityZoneRetriever
	infrastructureManager     infrastructureManager
}

func NewUpdateLBs(certificateManager certificateManager, clientProvider awsClientProvider,
	availabilityZoneRetriever availabilityZoneRetriever, infrastructureManager infrastructureManager) UpdateLBs {

	return UpdateLBs{
		certificateManager:        certificateManager,
		clientProvider:            clientProvider,
		availabilityZoneRetriever: availabilityZoneRetriever,
		infrastructureManager:     infrastructureManager,
	}
}

func (c UpdateLBs) Execute(globalFlags GlobalFlags, subcommandFlags []string, state storage.State) (storage.State, error) {
	config, err := c.parseFlags(subcommandFlags)
	if err != nil {
		return state, err
	}

	awsConfig := aws.Config{
		AccessKeyID:      state.AWS.AccessKeyID,
		SecretAccessKey:  state.AWS.SecretAccessKey,
		Region:           state.AWS.Region,
		EndpointOverride: globalFlags.EndpointOverride,
	}

	cloudFormationClient, err := c.clientProvider.CloudFormationClient(awsConfig)
	if err != nil {
		return state, err
	}

	stackExists, err := c.infrastructureManager.Exists(state.Stack.Name, cloudFormationClient)
	if err != nil {
		return state, err
	}

	if !stackExists {
		return state, errors.New("a bbl environment could not be found, please create a new environment before running this command again")
	}

	iamClient, err := c.clientProvider.IAMClient(awsConfig)
	if err != nil {
		return state, err
	}

	certificateName, err := c.certificateManager.Create(config.certPath, config.keyPath, iamClient)
	if err != nil {
		return state, err
	}

	ec2Client, err := c.clientProvider.EC2Client(awsConfig)
	if err != nil {
		return state, err
	}

	availabilityZones, err := c.availabilityZoneRetriever.Retrieve(state.AWS.Region, ec2Client)
	if err != nil {
		return state, err
	}

	certificate, err := c.certificateManager.Describe(certificateName, iamClient)
	if err != nil {
		return state, err
	}

	_, err = c.infrastructureManager.Update(state.KeyPair.Name, len(availabilityZones), state.Stack.Name, state.Stack.LBType, certificate.ARN, cloudFormationClient)
	if err != nil {
		return state, err
	}

	err = c.certificateManager.Delete(state.CertificateName, iamClient)
	if err != nil {
		return state, err
	}

	state.CertificateName = certificateName

	return state, nil
}

func (UpdateLBs) parseFlags(subcommandFlags []string) (updateLBConfig, error) {
	lbFlags := flags.New("update-lbs")

	config := updateLBConfig{}
	lbFlags.String(&config.certPath, "cert", "")
	lbFlags.String(&config.keyPath, "key", "")

	err := lbFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}
