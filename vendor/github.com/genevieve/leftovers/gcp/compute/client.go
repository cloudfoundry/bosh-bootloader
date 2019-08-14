package compute

import (
	"fmt"
	"time"

	gcpcompute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

type client struct {
	project string
	logger  logger

	service               *gcpcompute.Service
	addresses             *gcpcompute.AddressesService
	globalAddresses       *gcpcompute.GlobalAddressesService
	backendServices       *gcpcompute.BackendServicesService
	disks                 *gcpcompute.DisksService
	globalHealthChecks    *gcpcompute.HealthChecksService
	httpHealthChecks      *gcpcompute.HttpHealthChecksService
	httpsHealthChecks     *gcpcompute.HttpsHealthChecksService
	images                *gcpcompute.ImagesService
	instanceTemplates     *gcpcompute.InstanceTemplatesService
	instances             *gcpcompute.InstancesService
	instanceGroups        *gcpcompute.InstanceGroupsService
	instanceGroupManagers *gcpcompute.InstanceGroupManagersService
	firewalls             *gcpcompute.FirewallsService
	forwardingRules       *gcpcompute.ForwardingRulesService
	globalForwardingRules *gcpcompute.GlobalForwardingRulesService
	routes                *gcpcompute.RoutesService
	routers               *gcpcompute.RoutersService
	subnetworks           *gcpcompute.SubnetworksService
	sslCertificates       *gcpcompute.SslCertificatesService
	networks              *gcpcompute.NetworksService
	targetHttpProxies     *gcpcompute.TargetHttpProxiesService
	targetHttpsProxies    *gcpcompute.TargetHttpsProxiesService
	targetPools           *gcpcompute.TargetPoolsService
	targetVpnGateways     *gcpcompute.TargetVpnGatewaysService
	urlMaps               *gcpcompute.UrlMapsService
	regions               *gcpcompute.RegionsService
	vpnTunnels            *gcpcompute.VpnTunnelsService
	zones                 *gcpcompute.ZonesService
}

func NewClient(project string, service *gcpcompute.Service, logger logger) client {
	return client{
		project:               project,
		logger:                logger,
		service:               service,
		addresses:             service.Addresses,
		globalAddresses:       service.GlobalAddresses,
		backendServices:       service.BackendServices,
		disks:                 service.Disks,
		globalHealthChecks:    service.HealthChecks,
		httpHealthChecks:      service.HttpHealthChecks,
		httpsHealthChecks:     service.HttpsHealthChecks,
		images:                service.Images,
		instanceTemplates:     service.InstanceTemplates,
		instances:             service.Instances,
		instanceGroups:        service.InstanceGroups,
		instanceGroupManagers: service.InstanceGroupManagers,
		firewalls:             service.Firewalls,
		forwardingRules:       service.ForwardingRules,
		globalForwardingRules: service.GlobalForwardingRules,
		routes:                service.Routes,
		routers:               service.Routers,
		sslCertificates:       service.SslCertificates,
		subnetworks:           service.Subnetworks,
		networks:              service.Networks,
		targetHttpProxies:     service.TargetHttpProxies,
		targetHttpsProxies:    service.TargetHttpsProxies,
		targetPools:           service.TargetPools,
		targetVpnGateways:     service.TargetVpnGateways,
		urlMaps:               service.UrlMaps,
		vpnTunnels:            service.VpnTunnels,
		regions:               service.Regions,
		zones:                 service.Zones,
	}
}

func (c client) ListAddresses(region string) ([]*gcpcompute.Address, error) {
	var token string
	list := []*gcpcompute.Address{}

	for {
		resp, err := c.addresses.List(c.project, region).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteAddress(region, address string) error {
	return c.wait(c.addresses.Delete(c.project, region, address))
}

func (c client) ListGlobalAddresses() ([]*gcpcompute.Address, error) {
	var token string
	list := []*gcpcompute.Address{}

	for {
		resp, err := c.globalAddresses.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteGlobalAddress(address string) error {
	return c.wait(c.globalAddresses.Delete(c.project, address))
}

func (c client) ListBackendServices() ([]*gcpcompute.BackendService, error) {
	var token string
	list := []*gcpcompute.BackendService{}

	for {
		resp, err := c.backendServices.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteBackendService(backendService string) error {
	return c.wait(c.backendServices.Delete(c.project, backendService))
}

// ListDisks returns the full list of disks.
func (c client) ListDisks(zone string) ([]*gcpcompute.Disk, error) {
	var token string
	list := []*gcpcompute.Disk{}

	for {
		resp, err := c.disks.List(c.project, zone).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteDisk(zone, disk string) error {
	return c.wait(c.disks.Delete(c.project, zone, disk))
}

// ListImages returns the full list of images.
func (c client) ListImages() ([]*gcpcompute.Image, error) {
	var token string
	list := []*gcpcompute.Image{}

	for {
		resp, err := c.images.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteImage(image string) error {
	return c.wait(c.images.Delete(c.project, image))
}

func (c client) ListInstances(zone string) ([]*gcpcompute.Instance, error) {
	var token string
	list := []*gcpcompute.Instance{}

	for {
		resp, err := c.instances.List(c.project, zone).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteInstance(zone, instance string) error {
	return c.wait(c.instances.Delete(c.project, zone, instance))
}

func (c client) ListInstanceTemplates() ([]*gcpcompute.InstanceTemplate, error) {
	var token string
	list := []*gcpcompute.InstanceTemplate{}

	for {
		resp, err := c.instanceTemplates.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteInstanceTemplate(instanceTemplate string) error {
	return c.wait(c.instanceTemplates.Delete(c.project, instanceTemplate))
}

func (c client) ListInstanceGroups(zone string) ([]*gcpcompute.InstanceGroup, error) {
	var token string
	list := []*gcpcompute.InstanceGroup{}

	for {
		resp, err := c.instanceGroups.List(c.project, zone).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteInstanceGroup(zone, instanceGroup string) error {
	return c.wait(c.instanceGroups.Delete(c.project, zone, instanceGroup))
}

func (c client) ListInstanceGroupManagers(zone string) ([]*gcpcompute.InstanceGroupManager, error) {
	var token string
	list := []*gcpcompute.InstanceGroupManager{}

	for {
		resp, err := c.instanceGroupManagers.List(c.project, zone).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteInstanceGroupManager(zone, instanceGroupManager string) error {
	return c.wait(c.instanceGroupManagers.Delete(c.project, zone, instanceGroupManager))
}

func (c client) ListGlobalHealthChecks() ([]*gcpcompute.HealthCheck, error) {
	var token string
	list := []*gcpcompute.HealthCheck{}

	for {
		resp, err := c.globalHealthChecks.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteGlobalHealthCheck(globalHealthCheck string) error {
	return c.wait(c.globalHealthChecks.Delete(c.project, globalHealthCheck))
}

func (c client) ListHttpHealthChecks() ([]*gcpcompute.HttpHealthCheck, error) {
	var token string
	list := []*gcpcompute.HttpHealthCheck{}

	for {
		resp, err := c.httpHealthChecks.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteHttpHealthCheck(httpHealthCheck string) error {
	return c.wait(c.httpHealthChecks.Delete(c.project, httpHealthCheck))
}

func (c client) ListHttpsHealthChecks() ([]*gcpcompute.HttpsHealthCheck, error) {
	var token string
	list := []*gcpcompute.HttpsHealthCheck{}

	for {
		resp, err := c.httpsHealthChecks.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteHttpsHealthCheck(httpsHealthCheck string) error {
	return c.wait(c.httpsHealthChecks.Delete(c.project, httpsHealthCheck))
}

func (c client) ListFirewalls() ([]*gcpcompute.Firewall, error) {
	var token string
	list := []*gcpcompute.Firewall{}

	for {
		resp, err := c.firewalls.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteFirewall(firewall string) error {
	return c.wait(c.firewalls.Delete(c.project, firewall))
}

func (c client) ListGlobalForwardingRules() ([]*gcpcompute.ForwardingRule, error) {
	var token string
	list := []*gcpcompute.ForwardingRule{}

	for {
		resp, err := c.globalForwardingRules.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteGlobalForwardingRule(globalForwardingRule string) error {
	return c.wait(c.globalForwardingRules.Delete(c.project, globalForwardingRule))
}

func (c client) ListForwardingRules(region string) ([]*gcpcompute.ForwardingRule, error) {
	var token string
	list := []*gcpcompute.ForwardingRule{}

	for {
		resp, err := c.forwardingRules.List(c.project, region).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteForwardingRule(region, forwardingRule string) error {
	return c.wait(c.forwardingRules.Delete(c.project, region, forwardingRule))
}

func (c client) ListRoutes() ([]*gcpcompute.Route, error) {
	var token string
	list := []*gcpcompute.Route{}

	for {
		resp, err := c.routes.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteRoute(route string) error {
	return c.wait(c.routes.Delete(c.project, route))
}

func (c client) ListRouters(region string) ([]*gcpcompute.Router, error) {
	var token string
	list := []*gcpcompute.Router{}

	for {
		resp, err := c.routers.List(c.project, region).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteRouter(region, router string) error {
	return c.wait(c.routers.Delete(c.project, region, router))
}

func (c client) ListNetworks() ([]*gcpcompute.Network, error) {
	var token string
	list := []*gcpcompute.Network{}

	for {
		resp, err := c.networks.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteNetwork(network string) error {
	return c.wait(c.networks.Delete(c.project, network))
}

func (c client) ListSubnetworks(region string) ([]*gcpcompute.Subnetwork, error) {
	var token string
	list := []*gcpcompute.Subnetwork{}

	for {
		resp, err := c.subnetworks.List(c.project, region).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteSubnetwork(region, subnetwork string) error {
	return c.wait(c.subnetworks.Delete(c.project, region, subnetwork))
}

func (c client) ListSslCertificates() ([]*gcpcompute.SslCertificate, error) {
	var token string
	list := []*gcpcompute.SslCertificate{}

	for {
		resp, err := c.sslCertificates.List(c.project).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteSslCertificate(certificate string) error {
	return c.wait(c.sslCertificates.Delete(c.project, certificate))
}

func (c client) ListTargetHttpProxies() (*gcpcompute.TargetHttpProxyList, error) {
	return c.targetHttpProxies.List(c.project).Do()
}

func (c client) DeleteTargetHttpProxy(targetHttpProxy string) error {
	return c.wait(c.targetHttpProxies.Delete(c.project, targetHttpProxy))
}

func (c client) ListTargetHttpsProxies() (*gcpcompute.TargetHttpsProxyList, error) {
	return c.targetHttpsProxies.List(c.project).Do()
}

func (c client) DeleteTargetHttpsProxy(targetHttpsProxy string) error {
	return c.wait(c.targetHttpsProxies.Delete(c.project, targetHttpsProxy))
}

func (c client) ListTargetPools(region string) (*gcpcompute.TargetPoolList, error) {
	return c.targetPools.List(c.project, region).Do()
}

func (c client) DeleteTargetPool(region string, targetPool string) error {
	return c.wait(c.targetPools.Delete(c.project, region, targetPool))
}

func (c client) ListTargetVpnGateways(region string) ([]*gcpcompute.TargetVpnGateway, error) {
	var token string
	list := []*gcpcompute.TargetVpnGateway{}

	for {
		resp, err := c.targetVpnGateways.List(c.project, region).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteTargetVpnGateway(region, targetVpnGateway string) error {
	return c.wait(c.targetVpnGateways.Delete(c.project, region, targetVpnGateway))
}

func (c client) ListUrlMaps() (*gcpcompute.UrlMapList, error) {
	return c.urlMaps.List(c.project).Do()
}

func (c client) DeleteUrlMap(urlMap string) error {
	return c.wait(c.urlMaps.Delete(c.project, urlMap))
}

func (c client) ListVpnTunnels(region string) ([]*gcpcompute.VpnTunnel, error) {
	var token string
	list := []*gcpcompute.VpnTunnel{}

	for {
		resp, err := c.vpnTunnels.List(c.project, region).PageToken(token).Do()
		if err != nil {
			return nil, err
		}

		list = append(list, resp.Items...)

		token = resp.NextPageToken
		if token == "" {
			break
		}

		time.Sleep(time.Second)
	}

	return list, nil
}

func (c client) DeleteVpnTunnel(region, vpnTunnel string) error {
	return c.wait(c.vpnTunnels.Delete(c.project, region, vpnTunnel))
}

func (c client) ListRegions() (map[string]string, error) {
	regions := map[string]string{}

	list, err := c.regions.List(c.project).Do()
	if err != nil {
		return regions, fmt.Errorf("List Regions: %s", err)
	}

	for _, r := range list.Items {
		regions[r.SelfLink] = r.Name
	}
	return regions, nil
}

func (c client) ListZones() (map[string]string, error) {
	zones := map[string]string{}

	list, err := c.zones.List(c.project).Do()
	if err != nil {
		return zones, fmt.Errorf("List Zones: %s", err)
	}

	for _, z := range list.Items {
		zones[z.SelfLink] = z.Name
	}
	return zones, nil
}

type request interface {
	Do(...googleapi.CallOption) (*gcpcompute.Operation, error)
}

func (c client) wait(request request) error {
	op, err := request.Do()
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok {
			if gerr.Code == 404 {
				return nil
			}
		}
		return fmt.Errorf("Do request: %s", err)
	}

	waiter := NewOperationWaiter(op, c.service, c.project, c.logger)

	return waiter.Wait()
}
