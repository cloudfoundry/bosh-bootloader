package commands

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSLBs struct {
	terraformManager terraformOutputter
	logger           logger
}

func NewAWSLBs(terraformManager terraformOutputter, logger logger) AWSLBs {
	return AWSLBs{
		terraformManager: terraformManager,
		logger:           logger,
	}
}

func (l AWSLBs) Execute(subcommandFlags []string, state storage.State) error {
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
					TCPRouterLBName        string   `json:"cf_tcp_lb,omitempty"`
					TCPRouterLBURL         string   `json:"cf_tcp_lb_url,omitempty"`
					SystemDomainDNSServers []string `json:"env_dns_zone_name_servers,omitempty"`
				}{
					RouterLBName:           terraformOutputs["cf_router_lb_name"].(string),
					RouterLBURL:            terraformOutputs["cf_router_lb_url"].(string),
					SSHProxyLBName:         terraformOutputs["cf_ssh_lb_name"].(string),
					SSHProxyLBURL:          terraformOutputs["cf_ssh_lb_url"].(string),
					TCPRouterLBName:        terraformOutputs["cf_tcp_lb_name"].(string),
					TCPRouterLBURL:         terraformOutputs["cf_tcp_lb_url"].(string),
					SystemDomainDNSServers: terraformOutputs["env_dns_zone_name_servers"].([]string),
				})
				if err != nil {
					// not tested
					return err
				}

				l.logger.Println(string(lbOutput))
			} else {
				l.logger.Printf("CF Router LB: %s [%s]\n", terraformOutputs["cf_router_lb_name"], terraformOutputs["cf_router_lb_url"])
				l.logger.Printf("CF SSH Proxy LB: %s [%s]\n", terraformOutputs["cf_ssh_lb_name"], terraformOutputs["cf_ssh_lb_url"])
				l.logger.Printf("CF TCP Router LB: %s [%s]\n", terraformOutputs["cf_tcp_lb_name"], terraformOutputs["cf_tcp_lb_url"])

				if dnsServers, ok := terraformOutputs["env_dns_zone_name_servers"]; ok {
					l.logger.Printf("CF System Domain DNS servers: %s\n", strings.Join(dnsServers.([]string), " "))
				}
			}
		case "concourse":
			l.logger.Printf("Concourse LB: %s [%s]\n", terraformOutputs["concourse_lb_name"], terraformOutputs["concourse_lb_url"])
		default:
			return errors.New("no lbs found")
		}
	}

	return nil
}
