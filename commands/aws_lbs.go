package commands

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type AWSLBs struct {
	terraformManager terraformManager
	logger           logger
}

func NewAWSLBs(terraformManager terraformManager, logger logger) AWSLBs {
	return AWSLBs{
		terraformManager: terraformManager,
		logger:           logger,
	}
}

func (l AWSLBs) Execute(subcommandFlags []string, state storage.State) error {
	terraformOutputs, err := l.terraformManager.GetOutputs()
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
				RouterLBName:           terraformOutputs.GetString("cf_router_lb_name"),
				RouterLBURL:            terraformOutputs.GetString("cf_router_lb_url"),
				SSHProxyLBName:         terraformOutputs.GetString("cf_ssh_lb_name"),
				SSHProxyLBURL:          terraformOutputs.GetString("cf_ssh_lb_url"),
				TCPRouterLBName:        terraformOutputs.GetString("cf_tcp_lb_name"),
				TCPRouterLBURL:         terraformOutputs.GetString("cf_tcp_lb_url"),
				SystemDomainDNSServers: terraformOutputs.GetStringSlice("env_dns_zone_name_servers"),
			})
			if err != nil {
				// not tested
				return err
			}

			l.logger.Println(string(lbOutput))
		} else {
			l.logger.Printf("CF Router LB: %s [%s]\n", terraformOutputs.GetString("cf_router_lb_name"), terraformOutputs.GetString("cf_router_lb_url"))
			l.logger.Printf("CF SSH Proxy LB: %s [%s]\n", terraformOutputs.GetString("cf_ssh_lb_name"), terraformOutputs.GetString("cf_ssh_lb_url"))
			l.logger.Printf("CF TCP Router LB: %s [%s]\n", terraformOutputs.GetString("cf_tcp_lb_name"), terraformOutputs.GetString("cf_tcp_lb_url"))

			dnsServers := terraformOutputs.GetStringSlice("env_dns_zone_name_servers")

			if len(dnsServers) > 0 {
				l.logger.Printf("CF System Domain DNS servers: %s\n", strings.Join(dnsServers, " "))
			}
		}
	case "concourse":
		l.logger.Printf("Concourse LB: %s [%s]\n", terraformOutputs.GetString("concourse_lb_name"), terraformOutputs.GetString("concourse_lb_url"))
	default:
		return errors.New("no lbs found")
	}

	return nil
}
