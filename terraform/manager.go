package terraform

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Manager struct {
	executor executor
	logger   logger
}

type Outputs struct {
	ExternalIP             string   `json:"-"`
	NetworkName            string   `json:"-"`
	SubnetworkName         string   `json:"-"`
	BOSHTag                string   `json:"-"`
	InternalTag            string   `json:"-"`
	DirectorAddress        string   `json:"-"`
	RouterBackendService   string   `json:"-"`
	SSHProxyTargetPool     string   `json:"-"`
	TCPRouterTargetPool    string   `json:"-"`
	WSTargetPool           string   `json:"-"`
	ConcourseTargetPool    string   `json:"-"`
	RouterLBIP             string   `json:"cf_router_lb,omitempty"`
	SSHProxyLBIP           string   `json:"cf_ssh_proxy_lb,omitempty"`
	TCPRouterLBIP          string   `json:"cf_tcp_router_lb,omitempty"`
	WebSocketLBIP          string   `json:"cf_websocket_lb,omitempty"`
	ConcourseLBIP          string   `json:"-"`
	SystemDomainDNSServers []string `json:"cf_system_domain_dns_servers,omitempty"`
}

type executor interface {
	Destroy(serviceAccountKey, envID, projectID, zone, region, terraformTemplate, tfState string) (string, error)
	Output(string, string) (string, error)
}

type logger interface {
	Println(message string)
}

func NewManager(executor executor, logger logger) Manager {
	return Manager{
		executor: executor,
		logger:   logger,
	}
}

func (m Manager) Destroy(bblState storage.State) (storage.State, error) {
	if bblState.TFState == "" {
		return bblState, nil
	}

	tfState, err := m.executor.Destroy(bblState.GCP.ServiceAccountKey, bblState.EnvID, bblState.GCP.ProjectID, bblState.GCP.Zone, bblState.GCP.Region,
		VarsTemplate, bblState.TFState)
	switch err.(type) {
	case ExecutorDestroyError:
		executorDestroyError := err.(ExecutorDestroyError)
		bblState.TFState = executorDestroyError.tfState
		return storage.State{}, NewManagerDestroyError(bblState, executorDestroyError)
	case error:
		return storage.State{}, err
	}

	bblState.TFState = tfState
	return bblState, nil
}

func (m Manager) GetOutputs(tfState, lbType string, domainExists bool) (Outputs, error) {
	if tfState == "" {
		return Outputs{}, nil
	}

	externalIP, err := m.executor.Output(tfState, "external_ip")
	if err != nil {
		return Outputs{}, err
	}

	networkName, err := m.executor.Output(tfState, "network_name")
	if err != nil {
		return Outputs{}, err
	}

	subnetworkName, err := m.executor.Output(tfState, "subnetwork_name")
	if err != nil {
		return Outputs{}, err
	}

	boshTag, err := m.executor.Output(tfState, "bosh_open_tag_name")
	if err != nil {
		return Outputs{}, err
	}

	internalTag, err := m.executor.Output(tfState, "internal_tag_name")
	if err != nil {
		return Outputs{}, err
	}

	directorAddress, err := m.executor.Output(tfState, "director_address")
	if err != nil {
		return Outputs{}, err
	}

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
		systemDomainDNSServers    []string
	)

	if lbType == "cf" {
		routerBackendService, err = m.executor.Output(tfState, "router_backend_service")
		if err != nil {
			return Outputs{}, err
		}

		sshProxyTargetPool, err = m.executor.Output(tfState, "ssh_proxy_target_pool")
		if err != nil {
			return Outputs{}, err
		}

		tcpRouterTargetPool, err = m.executor.Output(tfState, "tcp_router_target_pool")
		if err != nil {
			return Outputs{}, err
		}

		wsTargetPool, err = m.executor.Output(tfState, "ws_target_pool")
		if err != nil {
			return Outputs{}, err
		}

		routerLBIP, err = m.executor.Output(tfState, "router_lb_ip")
		if err != nil {
			return Outputs{}, err
		}

		sshProxyLBIP, err = m.executor.Output(tfState, "ssh_proxy_lb_ip")
		if err != nil {
			return Outputs{}, err
		}

		tcpRouterLBIP, err = m.executor.Output(tfState, "tcp_router_lb_ip")
		if err != nil {
			return Outputs{}, err
		}

		webSocketLBIP, err = m.executor.Output(tfState, "ws_lb_ip")
		if err != nil {
			return Outputs{}, err
		}

		if domainExists {
			systemDomainDNSServersRaw, err = m.executor.Output(tfState, "system_domain_dns_servers")
			if err != nil {
				return Outputs{}, err
			}

			systemDomainDNSServers = strings.Split(systemDomainDNSServersRaw, ",\n")
		}
	}

	var concourseTargetPool, concourseLBIP string
	if lbType == "concourse" {
		concourseTargetPool, err = m.executor.Output(tfState, "concourse_target_pool")
		if err != nil {
			return Outputs{}, err
		}

		concourseLBIP, err = m.executor.Output(tfState, "concourse_lb_ip")
		if err != nil {
			return Outputs{}, err
		}
	}

	return Outputs{
		ExternalIP:             externalIP,
		NetworkName:            networkName,
		SubnetworkName:         subnetworkName,
		BOSHTag:                boshTag,
		InternalTag:            internalTag,
		DirectorAddress:        directorAddress,
		RouterBackendService:   routerBackendService,
		SSHProxyTargetPool:     sshProxyTargetPool,
		TCPRouterTargetPool:    tcpRouterTargetPool,
		WSTargetPool:           wsTargetPool,
		RouterLBIP:             routerLBIP,
		SSHProxyLBIP:           sshProxyLBIP,
		TCPRouterLBIP:          tcpRouterLBIP,
		WebSocketLBIP:          webSocketLBIP,
		SystemDomainDNSServers: systemDomainDNSServers,
		ConcourseTargetPool:    concourseTargetPool,
		ConcourseLBIP:          concourseLBIP,
	}, nil

}
