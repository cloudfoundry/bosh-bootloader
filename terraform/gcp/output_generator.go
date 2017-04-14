package gcp

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type executor interface {
	Output(string, string) (string, error)
}

type OutputGenerator struct {
	executor executor
}

func NewOutputGenerator(executor executor) OutputGenerator {
	return OutputGenerator{
		executor: executor,
	}
}

func (g OutputGenerator) Generate(bblState storage.State) (map[string]interface{}, error) {
	outputs := map[string]interface{}{}
	if bblState.TFState == "" {
		return outputs, nil
	}

	externalIP, err := g.executor.Output(bblState.TFState, "external_ip")
	if err != nil {
		return map[string]interface{}{}, err
	}
	outputs["external_ip"] = externalIP

	networkName, err := g.executor.Output(bblState.TFState, "network_name")
	if err != nil {
		return map[string]interface{}{}, err
	}
	outputs["network_name"] = networkName

	subnetworkName, err := g.executor.Output(bblState.TFState, "subnetwork_name")
	if err != nil {
		return map[string]interface{}{}, err
	}
	outputs["subnetwork_name"] = subnetworkName

	boshTag, err := g.executor.Output(bblState.TFState, "bosh_open_tag_name")
	if err != nil {
		return map[string]interface{}{}, err
	}
	outputs["bosh_open_tag_name"] = boshTag

	internalTag, err := g.executor.Output(bblState.TFState, "internal_tag_name")
	if err != nil {
		return map[string]interface{}{}, err
	}
	outputs["internal_tag_name"] = internalTag

	directorAddress, err := g.executor.Output(bblState.TFState, "director_address")
	if err != nil {
		return map[string]interface{}{}, err
	}
	outputs["director_address"] = directorAddress

	var (
		routerBackendService      string
		sshProxyTargetPool        string
		tcpRouterTargetPool       string
		wsTargetPool              string
		routerLBIP                string
		sshProxyLBIP              string
		tcpRouterLBIP             string
		webSocketLBIP             string
		systemDomainDNSServersRaw string
	)

	if bblState.LB.Type == "cf" {
		routerBackendService, err = g.executor.Output(bblState.TFState, "router_backend_service")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["router_backend_service"] = routerBackendService

		sshProxyTargetPool, err = g.executor.Output(bblState.TFState, "ssh_proxy_target_pool")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["ssh_proxy_target_pool"] = sshProxyTargetPool

		tcpRouterTargetPool, err = g.executor.Output(bblState.TFState, "tcp_router_target_pool")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["tcp_router_target_pool"] = tcpRouterTargetPool

		wsTargetPool, err = g.executor.Output(bblState.TFState, "ws_target_pool")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["ws_target_pool"] = wsTargetPool

		routerLBIP, err = g.executor.Output(bblState.TFState, "router_lb_ip")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["router_lb_ip"] = routerLBIP

		sshProxyLBIP, err = g.executor.Output(bblState.TFState, "ssh_proxy_lb_ip")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["ssh_proxy_lb_ip"] = sshProxyLBIP

		tcpRouterLBIP, err = g.executor.Output(bblState.TFState, "tcp_router_lb_ip")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["tcp_router_lb_ip"] = tcpRouterLBIP

		webSocketLBIP, err = g.executor.Output(bblState.TFState, "ws_lb_ip")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["ws_lb_ip"] = webSocketLBIP

		if bblState.LB.Domain != "" {
			systemDomainDNSServersRaw, err = g.executor.Output(bblState.TFState, "system_domain_dns_servers")
			if err != nil {
				return map[string]interface{}{}, err
			}

			outputs["system_domain_dns_servers"] = strings.Split(systemDomainDNSServersRaw, ",\n")
		}
	}

	if bblState.LB.Type == "concourse" {
		concourseTargetPool, err := g.executor.Output(bblState.TFState, "concourse_target_pool")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["concourse_target_pool"] = concourseTargetPool

		concourseLBIP, err := g.executor.Output(bblState.TFState, "concourse_lb_ip")
		if err != nil {
			return map[string]interface{}{}, err
		}
		outputs["concourse_lb_ip"] = concourseLBIP
	}

	return outputs, nil
}
