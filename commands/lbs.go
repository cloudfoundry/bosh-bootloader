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
	credentialValidator   credentialValidator
	infrastructureManager infrastructureManager
	stateValidator        stateValidator
	terraformOutputter    terraformOutputter
	stdout                io.Writer
}

func NewLBs(credentialValidator credentialValidator, stateValidator stateValidator, infrastructureManager infrastructureManager, terraformOutputter terraformOutputter, stdout io.Writer) LBs {
	return LBs{
		credentialValidator:   credentialValidator,
		infrastructureManager: infrastructureManager,
		stateValidator:        stateValidator,
		terraformOutputter:    terraformOutputter,
		stdout:                stdout,
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
		switch state.LB.Type {
		case "cf":
			routerLB, err := c.terraformOutputter.Get(state.TFState, "router_lb_ip")
			if err != nil {
				return err
			}

			sshProxyLB, err := c.terraformOutputter.Get(state.TFState, "ssh_proxy_lb_ip")
			if err != nil {
				return err
			}

			tcpRouterLB, err := c.terraformOutputter.Get(state.TFState, "tcp_router_lb_ip")
			if err != nil {
				return err
			}

			fmt.Fprintf(c.stdout, "CF Router LB: %s\n", routerLB)
			fmt.Fprintf(c.stdout, "CF SSH Proxy LB: %s\n", sshProxyLB)
			fmt.Fprintf(c.stdout, "CF TCP Router LB: %s\n", tcpRouterLB)
		case "concourse":
			concourseLB, err := c.terraformOutputter.Get(state.TFState, "concourse_lb_ip")
			if err != nil {
				return err
			}
			fmt.Fprintf(c.stdout, "Concourse LB: %s\n", concourseLB)
		default:
			return errors.New("no lbs found")
		}
	}

	return nil
}
