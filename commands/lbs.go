package commands

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const (
	LBsCommand = "lbs"
)

type LBs struct {
	credentialValidator   credentialValidator
	infrastructureManager infrastructureManager
	stateValidator        stateValidator
	terraformManager      terraformManager
	logger                logger
}

func NewLBs(credentialValidator credentialValidator, stateValidator stateValidator, infrastructureManager infrastructureManager, terraformManager terraformManager, logger logger) LBs {
	return LBs{
		credentialValidator:   credentialValidator,
		infrastructureManager: infrastructureManager,
		stateValidator:        stateValidator,
		terraformManager:      terraformManager,
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
		terraformOutputs, err := c.terraformManager.GetOutputs(state)
		if err != nil {
			return err
		}

		switch state.LB.Type {
		case "cf":
			if len(subcommandFlags) > 0 && subcommandFlags[0] == "--json" {
				lbOutput, err := json.Marshal(struct {
					RouterLBIP             string   `json:"cf_router_lb,omitempty"`
					SSHProxyLBIP           string   `json:"cf_ssh_proxy_lb,omitempty"`
					TCPRouterLBIP          string   `json:"cf_tcp_router_lb,omitempty"`
					WebSocketLBIP          string   `json:"cf_websocket_lb,omitempty"`
					SystemDomainDNSServers []string `json:"cf_system_domain_dns_servers,omitempty"`
				}{
					RouterLBIP:             terraformOutputs["router_lb_ip"].(string),
					SSHProxyLBIP:           terraformOutputs["ssh_proxy_lb_ip"].(string),
					TCPRouterLBIP:          terraformOutputs["tcp_router_lb_ip"].(string),
					WebSocketLBIP:          terraformOutputs["ws_lb_ip"].(string),
					SystemDomainDNSServers: terraformOutputs["system_domain_dns_servers"].([]string),
				})
				if err != nil {
					// not tested
					return err
				}

				c.logger.Println(string(lbOutput))
			} else {
				c.logger.Printf("CF Router LB: %s\n", terraformOutputs["router_lb_ip"])
				c.logger.Printf("CF SSH Proxy LB: %s\n", terraformOutputs["ssh_proxy_lb_ip"])
				c.logger.Printf("CF TCP Router LB: %s\n", terraformOutputs["tcp_router_lb_ip"])
				c.logger.Printf("CF WebSocket LB: %s\n", terraformOutputs["ws_lb_ip"])

				if dnsServers, ok := terraformOutputs["system_domain_dns_servers"]; ok {
					c.logger.Printf("CF System Domain DNS servers: %s\n", strings.Join(dnsServers.([]string), " "))
				}
			}
		case "concourse":
			c.logger.Printf("Concourse LB: %s\n", terraformOutputs["concourse_lb_ip"])
		default:
			return errors.New("no lbs found")
		}
	}

	return nil
}
