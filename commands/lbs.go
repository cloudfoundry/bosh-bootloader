package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	LBsCommand = "lbs"
)

type LBs struct {
	gcpLBs                gcpLBs
	credentialValidator   credentialValidator
	infrastructureManager infrastructureManager
	stateValidator        stateValidator
	logger                logger
}

type gcpLBs interface {
	Execute([]string, storage.State) error
}

func NewLBs(gcpLBs gcpLBs, credentialValidator credentialValidator, stateValidator stateValidator, infrastructureManager infrastructureManager, logger logger) LBs {
	return LBs{
		gcpLBs:                gcpLBs,
		credentialValidator:   credentialValidator,
		infrastructureManager: infrastructureManager,
		stateValidator:        stateValidator,
		logger:                logger,
	}
}

func (c LBs) Execute(subcommandFlags []string, state storage.State) error {
	err := c.stateValidator.Validate()
	if err != nil {
		return err
	}

	switch state.IAAS {
	case "aws":
		err = c.credentialValidator.Validate()
		if err != nil {
			return err
		}

		stack, err := c.infrastructureManager.Describe(state.Stack.Name)
		if err != nil {
			return err
		}

		switch state.Stack.LBType {
		case "cf":
			c.logger.Printf("CF Router LB: %s [%s]\n", stack.Outputs["CFRouterLoadBalancer"], stack.Outputs["CFRouterLoadBalancerURL"])
			c.logger.Printf("CF SSH Proxy LB: %s [%s]\n", stack.Outputs["CFSSHProxyLoadBalancer"], stack.Outputs["CFSSHProxyLoadBalancerURL"])
		case "concourse":
			c.logger.Printf("Concourse LB: %s [%s]\n", stack.Outputs["ConcourseLoadBalancer"], stack.Outputs["ConcourseLoadBalancerURL"])
		default:
			return errors.New("no lbs found")
		}
	case "gcp":
		if err := c.gcpLBs.Execute(subcommandFlags, state); err != nil {
			return err
		}
	}

	return nil
}
