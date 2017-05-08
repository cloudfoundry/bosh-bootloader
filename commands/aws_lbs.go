package commands

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSLBs struct {
	credentialValidator   credentialValidator
	infrastructureManager infrastructureManager
	terraformManager      terraformManager
	logger                logger
}

func NewAWSLBs(credentialValidator credentialValidator, infrastructureManager infrastructureManager, terraformManager terraformManager, logger logger) AWSLBs {
	return AWSLBs{
		credentialValidator:   credentialValidator,
		infrastructureManager: infrastructureManager,
		terraformManager:      terraformManager,
		logger:                logger,
	}
}

func (l AWSLBs) Execute(subcommandFlags []string, state storage.State) error {
	err := l.credentialValidator.Validate()
	if err != nil {
		return err
	}

	if state.TFState != "" {
		terraformOutputs, err := l.terraformManager.GetOutputs(state)
		if err != nil {
			return err
		}

		switch state.LB.Type {
		case "cf":
			if len(subcommandFlags) > 0 && subcommandFlags[0] == "--json" {
				lbOutput, err := json.Marshal(struct {
					RouterLBName           string   `json:"cf_router_lb,omitempty"`
					RouterLBURL            string   `json:"cf_router_lb_url,omitempty"`
					SSHProxyLBName         string   `json:"cf_ssh_proxy_lb,omitempty"`
					SSHProxyLBURL          string   `json:"cf_ssh_proxy_lb_url,omitempty"`
					SystemDomainDNSServers []string `json:"env_dns_zone_name_servers,omitempty"`
				}{
					RouterLBName:           terraformOutputs["cf_router_load_balancer"].(string),
					RouterLBURL:            terraformOutputs["cf_router_load_balancer_url"].(string),
					SSHProxyLBName:         terraformOutputs["cf_ssh_proxy_load_balancer"].(string),
					SSHProxyLBURL:          terraformOutputs["cf_ssh_proxy_load_balancer_url"].(string),
					SystemDomainDNSServers: terraformOutputs["cf_system_domain_dns_servers"].([]string),
				})
				if err != nil {
					// not tested
					return err
				}

				l.logger.Println(string(lbOutput))
			} else {
				l.logger.Printf("CF Router LB: %s [%s]\n", terraformOutputs["cf_router_load_balancer"], terraformOutputs["cf_router_load_balancer_url"])
				l.logger.Printf("CF SSH Proxy LB: %s [%s]\n", terraformOutputs["cf_ssh_proxy_load_balancer"], terraformOutputs["cf_ssh_proxy_load_balancer_url"])

				if dnsServers, ok := terraformOutputs["cf_system_domain_dns_servers"]; ok {
					l.logger.Printf("CF System Domain DNS servers: %s\n", strings.Join(dnsServers.([]string), " "))
				}
			}
		case "concourse":
			l.logger.Printf("Concourse LB: %s [%s]\n", terraformOutputs["concourse_load_balancer"], terraformOutputs["concourse_load_balancer_url"])
		default:
			return errors.New("no lbs found")
		}

	} else {
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
	}
	return nil
}
