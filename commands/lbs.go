package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	LBsCommand = "lbs"
)

type LBs struct {
	credentialValidator     credentialValidator
	infrastructureManager   infrastructureManager
	stateValidator          stateValidator
	terraformOutputProvider terraformOutputProvider
	stdout                  io.Writer
}

func NewLBs(credentialValidator credentialValidator, stateValidator stateValidator, infrastructureManager infrastructureManager, terraformOutputProvider terraformOutputProvider, stdout io.Writer) LBs {
	return LBs{
		credentialValidator:     credentialValidator,
		infrastructureManager:   infrastructureManager,
		stateValidator:          stateValidator,
		terraformOutputProvider: terraformOutputProvider,
		stdout:                  stdout,
	}
}

func (c LBs) Execute(subcommandFlags []string, state storage.State) error {
	err := c.stateValidator.Validate()
	if err != nil {
		return err
	}

	switch state.IAAS {
	case "aws":
		err = c.credentialValidator.ValidateAWS()
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
	case "gcp":
		terraformOutputs, err := c.terraformOutputProvider.Get(state.TFState, state.LB.Type)
		if err != nil {
			return err
		}

		switch state.LB.Type {
		case "cf":
			fmt.Fprintf(c.stdout, "CF Router LB: %s\n", terraformOutputs.RouterLBIP)
			fmt.Fprintf(c.stdout, "CF SSH Proxy LB: %s\n", terraformOutputs.SSHProxyLBIP)
			fmt.Fprintf(c.stdout, "CF TCP Router LB: %s\n", terraformOutputs.TCPRouterLBIP)
			fmt.Fprintf(c.stdout, "CF WebSocket LB: %s\n", terraformOutputs.WebSocketLBIP)
		case "concourse":
			fmt.Fprintf(c.stdout, "Concourse LB: %s\n", terraformOutputs.ConcourseLBIP)
		default:
			return errors.New("no lbs found")
		}
	}

	return nil
}
