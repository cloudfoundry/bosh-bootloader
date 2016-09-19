package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const (
	LBsCommand = "lbs"
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

func (c LBs) Execute(subcommandFlags []string, state storage.State) error {
	err := c.awsCredentialValidator.Validate()
	if err != nil {
		return err
	}

	stack, err := c.infrastructureManager.Describe(state.Stack.Name)
	if err != nil {
		return err
	}

	switch state.Stack.LBType {
	case "cf":
		fmt.Fprintf(c.stdout, "CF Router LB: %s [%s]\n", stack.Outputs["CFRouterLoadBalancer"], stack.Outputs["CFRouterLoadBalancerURL"])
		fmt.Fprintf(c.stdout, "CF SSH Proxy LB: %s [%s]\n", stack.Outputs["CFSSHProxyLoadBalancer"], stack.Outputs["CFSSHProxyLoadBalancerURL"])
	case "concourse":
		fmt.Fprintf(c.stdout, "Concourse LB: %s [%s]\n", stack.Outputs["ConcourseLoadBalancer"], stack.Outputs["ConcourseLoadBalancerURL"])
	default:
		return errors.New("no lbs found")
	}

	return nil
}
