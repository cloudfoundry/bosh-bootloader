package terraform

type Outputs struct {
	ExternalIP           string
	NetworkName          string
	SubnetworkName       string
	BOSHTag              string
	InternalTag          string
	DirectorAddress      string
	RouterBackendService string
	SSHProxyTargetPool   string
	TCPRouterTargetPool  string
	WSTargetPool         string
	ConcourseTargetPool  string
	RouterLBIP           string
	SSHProxyLBIP         string
	TCPRouterLBIP        string
	WebSocketLBIP        string
	ConcourseLBIP        string
}

type outputter interface {
	Get(string, string) (string, error)
}

type OutputProvider struct {
	outputter outputter
}

func NewOutputProvider(outputter outputter) OutputProvider {
	return OutputProvider{
		outputter: outputter,
	}
}

func (o OutputProvider) Get(tfState, lbType string) (Outputs, error) {
	externalIP, err := o.outputter.Get(tfState, "external_ip")
	if err != nil {
		return Outputs{}, err
	}

	networkName, err := o.outputter.Get(tfState, "network_name")
	if err != nil {
		return Outputs{}, err
	}

	subnetworkName, err := o.outputter.Get(tfState, "subnetwork_name")
	if err != nil {
		return Outputs{}, err
	}

	boshTag, err := o.outputter.Get(tfState, "bosh_open_tag_name")
	if err != nil {
		return Outputs{}, err
	}

	internalTag, err := o.outputter.Get(tfState, "internal_tag_name")
	if err != nil {
		return Outputs{}, err
	}

	directorAddress, err := o.outputter.Get(tfState, "director_address")
	if err != nil {
		return Outputs{}, err
	}

	var routerBackendService, sshProxyTargetPool, tcpRouterTargetPool, wsTargetPool, routerLBIP, sshProxyLBIP, tcpRouterLBIP, webSocketLBIP string
	if lbType == "cf" {
		routerBackendService, err = o.outputter.Get(tfState, "router_backend_service")
		if err != nil {
			return Outputs{}, err
		}

		sshProxyTargetPool, err = o.outputter.Get(tfState, "ssh_proxy_target_pool")
		if err != nil {
			return Outputs{}, err
		}

		tcpRouterTargetPool, err = o.outputter.Get(tfState, "tcp_router_target_pool")
		if err != nil {
			return Outputs{}, err
		}

		wsTargetPool, err = o.outputter.Get(tfState, "ws_target_pool")
		if err != nil {
			return Outputs{}, err
		}

		routerLBIP, err = o.outputter.Get(tfState, "router_lb_ip")
		if err != nil {
			return Outputs{}, err
		}

		sshProxyLBIP, err = o.outputter.Get(tfState, "ssh_proxy_lb_ip")
		if err != nil {
			return Outputs{}, err
		}

		tcpRouterLBIP, err = o.outputter.Get(tfState, "tcp_router_lb_ip")
		if err != nil {
			return Outputs{}, err
		}

		webSocketLBIP, err = o.outputter.Get(tfState, "ws_lb_ip")
		if err != nil {
			return Outputs{}, err
		}
	}

	var concourseTargetPool, concourseLBIP string
	if lbType == "concourse" {
		concourseTargetPool, err = o.outputter.Get(tfState, "concourse_target_pool")
		if err != nil {
			return Outputs{}, err
		}

		concourseLBIP, err = o.outputter.Get(tfState, "concourse_lb_ip")
		if err != nil {
			return Outputs{}, err
		}
	}

	return Outputs{
		ExternalIP:           externalIP,
		NetworkName:          networkName,
		SubnetworkName:       subnetworkName,
		BOSHTag:              boshTag,
		InternalTag:          internalTag,
		DirectorAddress:      directorAddress,
		RouterBackendService: routerBackendService,
		SSHProxyTargetPool:   sshProxyTargetPool,
		TCPRouterTargetPool:  tcpRouterTargetPool,
		WSTargetPool:         wsTargetPool,
		RouterLBIP:           routerLBIP,
		SSHProxyLBIP:         sshProxyLBIP,
		TCPRouterLBIP:        tcpRouterLBIP,
		WebSocketLBIP:        webSocketLBIP,
		ConcourseTargetPool:  concourseTargetPool,
		ConcourseLBIP:        concourseLBIP,
	}, nil
}
