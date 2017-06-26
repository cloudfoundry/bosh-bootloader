package commands

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPLBs struct {
	terraformManager terraformOutputter
	logger           logger
}

func NewGCPLBs(terraformManager terraformOutputter, logger logger) GCPLBs {
	return GCPLBs{
		terraformManager: terraformManager,
		logger:           logger,
	}
}

func (l GCPLBs) Execute(subcommandFlags []string, state storage.State) error {
	terraformOutputs, err := l.terraformManager.GetOutputs(state)
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

			l.logger.Println(string(lbOutput))
		} else {
			l.logger.Printf("CF Router LB: %s\n", terraformOutputs["router_lb_ip"])
			l.logger.Printf("CF SSH Proxy LB: %s\n", terraformOutputs["ssh_proxy_lb_ip"])
			l.logger.Printf("CF TCP Router LB: %s\n", terraformOutputs["tcp_router_lb_ip"])
			l.logger.Printf("CF WebSocket LB: %s\n", terraformOutputs["ws_lb_ip"])

			if dnsServers, ok := terraformOutputs["system_domain_dns_servers"]; ok {
				l.logger.Printf("CF System Domain DNS servers: %s\n", strings.Join(dnsServers.([]string), " "))
			}
		}
	case "concourse":
		l.logger.Printf("Concourse LB: %s\n", terraformOutputs["concourse_lb_ip"])
	default:
		return errors.New("no lbs found")
	}

	return nil
}
