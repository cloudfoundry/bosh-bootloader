package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type LBs struct {
	awsCredentialValidator awsCredentialValidator
	infrastructureManager  infrastructureManager
	stdout                 io.Writer
}

func NewLBs(awsCredentialValidator awsCredentialValidator, infrastructureManager infrastructureManager, stdout io.Writer) LBs {
	return LBs{
		awsCredentialValidator: awsCredentialValidator,
		infrastructureManager:  infrastructureManager,
		stdout:                 stdout,
	}
}

func (c LBs) Execute(subcommandFlags []string, state storage.State) (storage.State, error) {
	err := c.awsCredentialValidator.Validate()
	if err != nil {
		return state, err
	}

	stack, err := c.infrastructureManager.Describe(state.Stack.Name)
	if err != nil {
		return state, err
	}

	if state.Stack.LBType == "cf" {
		fmt.Fprintf(c.stdout, "%s: %s\n", stack.Outputs["CFRouterLoadBalancer"], stack.Outputs["CFRouterLoadBalancerURL"])
		fmt.Fprintf(c.stdout, "%s: %s\n", stack.Outputs["CFSSHProxyLoadBalancer"], stack.Outputs["CFSSHProxyLoadBalancerURL"])
	} else if state.Stack.LBType == "concourse" {
		fmt.Fprintf(c.stdout, "%s: %s\n", stack.Outputs["ConcourseLoadBalancer"], stack.Outputs["ConcourseLoadBalancerURL"])
	} else {
		return state, errors.New("no lbs found")
	}

	return state, nil
}
