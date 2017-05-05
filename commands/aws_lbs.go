package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSLBs struct {
	credentialValidator   credentialValidator
	infrastructureManager infrastructureManager
	logger                logger
}

func NewAWSLBs(credentialValidator credentialValidator, infrastructureManager infrastructureManager, logger logger) AWSLBs {
	return AWSLBs{
		credentialValidator:   credentialValidator,
		infrastructureManager: infrastructureManager,
		logger:                logger,
	}
}

func (l AWSLBs) Execute(state storage.State) error {
	err := l.credentialValidator.Validate()
	if err != nil {
		return err
	}

	stack, err := l.infrastructureManager.Describe(state.Stack.Name)
	if err != nil {
		return err
	}

	switch state.Stack.LBType {
	case "cf":
		l.logger.Printf("CF Router LB: %s [%s]\n", stack.Outputs["CFRouterLoadBalancer"], stack.Outputs["CFRouterLoadBalancerURL"])
		l.logger.Printf("CF SSH Proxy LB: %s [%s]\n", stack.Outputs["CFSSHProxyLoadBalancer"], stack.Outputs["CFSSHProxyLoadBalancerURL"])
	case "concourse":
		l.logger.Printf("Concourse LB: %s [%s]\n", stack.Outputs["ConcourseLoadBalancer"], stack.Outputs["ConcourseLoadBalancerURL"])
	default:
		return errors.New("no lbs found")
	}
	return nil
}
