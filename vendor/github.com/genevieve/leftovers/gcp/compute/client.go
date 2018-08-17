package compute

import (
	"fmt"

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
	subnetworks           *gcpcompute.SubnetworksService
	sslCertificates       *gcpcompute.SslCertificatesService
	networks              *gcpcompute.NetworksService
	targetHttpProxies     *gcpcompute.TargetHttpProxiesService
	targetHttpsProxies    *gcpcompute.TargetHttpsProxiesService
	targetPools           *gcpcompute.TargetPoolsService
	urlMaps               *gcpcompute.UrlMapsService
	regions               *gcpcompute.RegionsService
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
		sslCertificates:       service.SslCertificates,
		subnetworks:           service.Subnetworks,
		networks:              service.Networks,
		targetHttpProxies:     service.TargetHttpProxies,
		targetHttpsProxies:    service.TargetHttpsProxies,
		targetPools:           service.TargetPools,
		urlMaps:               service.UrlMaps,
		regions:               service.Regions,
		zones:                 service.Zones,
	}
}

func (c client) ListAddresses(region string) (*gcpcompute.AddressList, error) {
	return c.addresses.List(c.project, region).Do()
}

func (c client) DeleteAddress(region, address string) error {
	return c.wait(c.addresses.Delete(c.project, region, address))
}

func (c client) ListGlobalAddresses() (*gcpcompute.AddressList, error) {
	return c.globalAddresses.List(c.project).Do()
}

func (c client) DeleteGlobalAddress(address string) error {
	return c.wait(c.globalAddresses.Delete(c.project, address))
}

func (c client) ListBackendServices() (*gcpcompute.BackendServiceList, error) {
	return c.backendServices.List(c.project).Do()
}

func (c client) DeleteBackendService(backendService string) error {
	return c.wait(c.backendServices.Delete(c.project, backendService))
}

func (c client) ListDisks(zone string) (*gcpcompute.DiskList, error) {
	return c.disks.List(c.project, zone).Do()
}

func (c client) DeleteDisk(zone, disk string) error {
	return c.wait(c.disks.Delete(c.project, zone, disk))
}

func (c client) ListImages() (*gcpcompute.ImageList, error) {
	return c.images.List(c.project).Do()
}

func (c client) DeleteImage(image string) error {
	return c.wait(c.images.Delete(c.project, image))
}

func (c client) ListInstances(zone string) (*gcpcompute.InstanceList, error) {
	return c.instances.List(c.project, zone).Do()
}

func (c client) DeleteInstance(zone, instance string) error {
	return c.wait(c.instances.Delete(c.project, zone, instance))
}

func (c client) ListInstanceTemplates() (*gcpcompute.InstanceTemplateList, error) {
	return c.instanceTemplates.List(c.project).Do()
}

func (c client) DeleteInstanceTemplate(instanceTemplate string) error {
	return c.wait(c.instanceTemplates.Delete(c.project, instanceTemplate))
}

func (c client) ListInstanceGroups(zone string) (*gcpcompute.InstanceGroupList, error) {
	return c.instanceGroups.List(c.project, zone).Do()
}

func (c client) DeleteInstanceGroup(zone, instanceGroup string) error {
	return c.wait(c.instanceGroups.Delete(c.project, zone, instanceGroup))
}

func (c client) ListInstanceGroupManagers(zone string) (*gcpcompute.InstanceGroupManagerList, error) {
	return c.instanceGroupManagers.List(c.project, zone).Do()
}

func (c client) DeleteInstanceGroupManager(zone, instanceGroupManager string) error {
	return c.wait(c.instanceGroupManagers.Delete(c.project, zone, instanceGroupManager))
}

func (c client) ListGlobalHealthChecks() (*gcpcompute.HealthCheckList, error) {
	return c.globalHealthChecks.List(c.project).Do()
}

func (c client) DeleteGlobalHealthCheck(globalHealthCheck string) error {
	return c.wait(c.globalHealthChecks.Delete(c.project, globalHealthCheck))
}

func (c client) ListHttpHealthChecks() (*gcpcompute.HttpHealthCheckList, error) {
	return c.httpHealthChecks.List(c.project).Do()
}

func (c client) DeleteHttpHealthCheck(httpHealthCheck string) error {
	return c.wait(c.httpHealthChecks.Delete(c.project, httpHealthCheck))
}

func (c client) ListHttpsHealthChecks() (*gcpcompute.HttpsHealthCheckList, error) {
	return c.httpsHealthChecks.List(c.project).Do()
}

func (c client) DeleteHttpsHealthCheck(httpsHealthCheck string) error {
	return c.wait(c.httpsHealthChecks.Delete(c.project, httpsHealthCheck))
}

func (c client) ListFirewalls() (*gcpcompute.FirewallList, error) {
	return c.firewalls.List(c.project).Do()
}

func (c client) DeleteFirewall(firewall string) error {
	return c.wait(c.firewalls.Delete(c.project, firewall))
}

func (c client) ListGlobalForwardingRules() (*gcpcompute.ForwardingRuleList, error) {
	return c.globalForwardingRules.List(c.project).Do()
}

func (c client) DeleteGlobalForwardingRule(globalForwardingRule string) error {
	return c.wait(c.globalForwardingRules.Delete(c.project, globalForwardingRule))
}

func (c client) ListForwardingRules(region string) (*gcpcompute.ForwardingRuleList, error) {
	return c.forwardingRules.List(c.project, region).Do()
}

func (c client) DeleteForwardingRule(region, forwardingRule string) error {
	return c.wait(c.forwardingRules.Delete(c.project, region, forwardingRule))
}

func (c client) ListNetworks() (*gcpcompute.NetworkList, error) {
	return c.networks.List(c.project).Do()
}

func (c client) DeleteNetwork(network string) error {
	return c.wait(c.networks.Delete(c.project, network))
}

func (c client) ListSubnetworks(region string) (*gcpcompute.SubnetworkList, error) {
	return c.subnetworks.List(c.project, region).Do()
}

func (c client) DeleteSubnetwork(region, subnetwork string) error {
	return c.wait(c.subnetworks.Delete(c.project, region, subnetwork))
}

func (c client) ListSslCertificates() (*gcpcompute.SslCertificateList, error) {
	return c.sslCertificates.List(c.project).Do()
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

func (c client) ListUrlMaps() (*gcpcompute.UrlMapList, error) {
	return c.urlMaps.List(c.project).Do()
}

func (c client) DeleteUrlMap(urlMap string) error {
	return c.wait(c.urlMaps.Delete(c.project, urlMap))
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
