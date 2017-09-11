package helpers

import (
	"errors"
	"fmt"
	"regexp"

	compute "google.golang.org/api/compute/v1"

	awslib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var matchString = regexp.MatchString

type EnvIDManager struct {
	envIDGenerator        envIDGenerator
	gcpClient             gcpClient
	infrastructureManager infrastructureManager
	ec2Client             ec2Client
}

type envIDGenerator interface {
	Generate() (string, error)
}

type infrastructureManager interface {
	Exists(stackName string) (bool, error)
}

type ec2Client interface {
	DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error)
}

type gcpClient interface {
	GetNetworks(name string) (*compute.NetworkList, error)
}

func NewEnvIDManager(envIDGenerator envIDGenerator, gcpClient gcpClient, infrastructureManager infrastructureManager, ec2Client ec2Client) EnvIDManager {
	return EnvIDManager{
		envIDGenerator:        envIDGenerator,
		gcpClient:             gcpClient,
		infrastructureManager: infrastructureManager,
		ec2Client:             ec2Client,
	}
}

func (e EnvIDManager) Sync(state storage.State, envID string) (storage.State, error) {
	if state.EnvID != "" {
		return state, nil
	}

	err := e.checkFastFail(state.IAAS, envID)
	if err != nil {
		return storage.State{}, err
	}

	err = e.validateName(envID)
	if err != nil {
		return storage.State{}, err
	}

	if envID != "" {
		state.EnvID = envID
	} else {
		state.EnvID, err = e.envIDGenerator.Generate()
		if err != nil {
			return storage.State{}, err
		}
	}

	return state, nil
}

func (e EnvIDManager) checkFastFail(iaas, envID string) error {
	switch iaas {
	case "gcp":
		networkName := envID + "-network"
		networkList, err := e.gcpClient.GetNetworks(networkName)
		if err != nil {
			return err
		}
		if len(networkList.Items) > 0 {
			return errors.New(fmt.Sprintf("It looks like a bbl environment already exists with the name '%s'. Please provide a different name.", envID))
		}
	case "aws":
		stackName := "stack-" + envID
		stackExists, err := e.infrastructureManager.Exists(stackName)
		if err != nil {
			return err
		}
		if stackExists {
			return errors.New(fmt.Sprintf("It looks like a bbl environment already exists with the name '%s'. Please provide a different name.", envID))
		}

		vpcs, err := e.ec2Client.DescribeVpcs(&ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
				{
					Name: awslib.String("tag:Name"),
					Values: []*string{
						awslib.String(fmt.Sprintf("%s-vpc", envID)),
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("Failed to check vpc existence: %s", err)
		}

		if len(vpcs.Vpcs) > 0 {
			return errors.New(fmt.Sprintf("It looks like a bbl environment already exists with the name '%s'. Please provide a different name.", envID))
		}
	}
	return nil
}

func (e EnvIDManager) validateName(envID string) error {
	if envID == "" {
		return nil
	}

	matched, err := matchString("^(?:[a-z](?:[-a-z0-9]+[a-z0-9])?)$", envID)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("Names must start with a letter and be alphanumeric or hyphenated.")
	}

	return nil
}
