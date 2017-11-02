package commands

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPLBs struct {
	terraformManager terraformManager
	logger           logger
}

func NewGCPLBs(terraformManager terraformManager, logger logger) GCPLBs {
	return GCPLBs{
		terraformManager: terraformManager,
		logger:           logger,
	}
}

func (l GCPLBs) Execute(subcommandFlags []string, state storage.State) error {
	terraformOutputs, err := l.terraformManager.GetOutputs()
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
				CredhubLBIP            string   `json:"cf_credhub_lb,omitempty"`
				SystemDomainDNSServers []string `json:"cf_system_domain_dns_servers,omitempty"`
			}{
				RouterLBIP:             terraformOutputs.GetString("router_lb_ip"),
				SSHProxyLBIP:           terraformOutputs.GetString("ssh_proxy_lb_ip"),
				TCPRouterLBIP:          terraformOutputs.GetString("tcp_router_lb_ip"),
				WebSocketLBIP:          terraformOutputs.GetString("ws_lb_ip"),
				CredhubLBIP:            terraformOutputs.GetString("credhub_lb_ip"),
				SystemDomainDNSServers: terraformOutputs.GetStringSlice("system_domain_dns_servers"),
			})
			if err != nil {
				// not tested
				return err
			}

			l.logger.Println(string(lbOutput))
		} else {
			l.logger.Printf("CF Router LB: %s\n", terraformOutputs.GetString("router_lb_ip"))
			l.logger.Printf("CF SSH Proxy LB: %s\n", terraformOutputs.GetString("ssh_proxy_lb_ip"))
			l.logger.Printf("CF TCP Router LB: %s\n", terraformOutputs.GetString("tcp_router_lb_ip"))
			l.logger.Printf("CF WebSocket LB: %s\n", terraformOutputs.GetString("ws_lb_ip"))
			l.logger.Printf("CF Credhub LB: %s\n", terraformOutputs.GetString("credhub_lb_ip"))
			dnsServers := terraformOutputs.GetStringSlice("system_domain_dns_servers")
			if len(dnsServers) > 0 {
				l.logger.Printf("CF System Domain DNS servers: %s\n", strings.Join(dnsServers, " "))
			}
		}
	case "concourse":
		l.logger.Printf("Concourse LB: %s\n", terraformOutputs.GetString("concourse_lb_ip"))
	default:
		return errors.New("no lbs found")
	}

	return nil
}
