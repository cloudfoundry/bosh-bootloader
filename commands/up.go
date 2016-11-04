package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Up struct {
	awsUp awsUp
	gcpUp gcpUp
}

type awsUp interface {
	Execute(awsUpConfig AWSUpConfig, state storage.State) error
}

type gcpUp interface {
	Execute(state storage.State) error
}

type upConfig struct {
	awsAccessKeyID     string
	awsSecretAccessKey string
	awsRegion          string
	iaas               string
}

func NewUp(awsUp awsUp, gcpUp gcpUp) Up {
	return Up{
		awsUp: awsUp,
		gcpUp: gcpUp,
	}
}

func (u Up) Execute(args []string, state storage.State) error {
	var desiredIAAS string

	config, err := u.parseArgs(args)
	if err != nil {
		return err
	}

	switch {
	case state.IAAS == "" && config.iaas == "":
		return errors.New("--iaas [gcp, aws] must be provided")
	case state.IAAS == "" && config.iaas != "":
		desiredIAAS = config.iaas
	case state.IAAS != "" && config.iaas == "":
		desiredIAAS = state.IAAS
	case state.IAAS != "" && config.iaas != "":
		if state.IAAS != config.iaas {
			return errors.New("the iaas provided must match the iaas in bbl-state.json")
		} else {
			desiredIAAS = state.IAAS
		}
	}

	switch desiredIAAS {
	case "aws":
		awsConfig := AWSUpConfig{
			AccessKeyID:     config.awsAccessKeyID,
			SecretAccessKey: config.awsSecretAccessKey,
			Region:          config.awsRegion,
		}
		err = u.awsUp.Execute(awsConfig, state)
	case "gcp":
		err = u.gcpUp.Execute(state)
	default:
		return fmt.Errorf("%q is invalid; supported values: [gcp, aws]", desiredIAAS)
	}

	if err != nil {
		return err
	}

	return nil
}

func (u Up) parseArgs(args []string) (upConfig, error) {
	var config upConfig

	upFlags := flags.New("up")
	upFlags.String(&config.awsAccessKeyID, "aws-access-key-id", os.Getenv("BBL_AWS_ACCESS_KEY_ID"))
	upFlags.String(&config.awsSecretAccessKey, "aws-secret-access-key", os.Getenv("BBL_AWS_SECRET_ACCESS_KEY"))
	upFlags.String(&config.awsRegion, "aws-region", os.Getenv("BBL_AWS_REGION"))
	upFlags.String(&config.iaas, "iaas", "")

	err := upFlags.Parse(args)
	if err != nil {
		return upConfig{}, err
	}

	return config, nil
}
