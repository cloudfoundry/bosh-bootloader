package actors

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

type GCP struct {
	service   *compute.Service
	projectID string
	region    string
}

func NewGCP(config integration.Config) GCP {
	rawServiceAccountKey, err := ioutil.ReadFile(config.GCPServiceAccountKeyPath)
	if err != nil {
		rawServiceAccountKey = []byte(config.GCPServiceAccountKeyPath)
	}

	googleConfig, err := google.JWTConfigFromJSON(rawServiceAccountKey, "https://www.googleapis.com/auth/compute")
	if err != nil {
		panic(err)
	}

	service, err := compute.New(googleConfig.Client(context.Background()))
	if err != nil {
		panic(err)
	}

	return GCP{
		service:   service,
		projectID: config.GCPProjectID,
		region:    config.GCPRegion,
	}
}

func (g GCP) SSHKey() (string, error) {
	project, err := g.service.Projects.Get(g.projectID).Do()
	if err != nil {
		return "", err
	}

	for _, item := range project.CommonInstanceMetadata.Items {
		if item.Key == "sshKeys" {
			return *item.Value, nil
		}
	}

	return "", nil
}

func (g GCP) RemoveSSHKey(sshKey string) error {
	project, err := g.service.Projects.Get(g.projectID).Do()
	if err != nil {
		return err
	}

	for i, item := range project.CommonInstanceMetadata.Items {
		if item.Key == "sshKeys" {
			newSSHKeys := []string{}

			for _, keyFromGCP := range strings.Split(*item.Value, "\n") {
				if keyFromGCP != sshKey {
					newSSHKeys = append(newSSHKeys, keyFromGCP)
				}
			}

			newValue := strings.Join(newSSHKeys, "\n")
			project.CommonInstanceMetadata.Items[i].Value = &newValue
			break
		}
	}

	_, err = g.service.Projects.SetCommonInstanceMetadata(g.projectID, project.CommonInstanceMetadata).Do()
	if err != nil {
		return err
	}

	return nil
}

func (g GCP) GetNetwork(networkName string) (*compute.Network, error) {
	return g.service.Networks.Get(g.projectID, networkName).Do()
}

func (g GCP) GetSubnet(subnetName string) (*compute.Subnetwork, error) {
	return g.service.Subnetworks.Get(g.projectID, g.region, subnetName).Do()
}

func (g GCP) GetAddress(addressName string) (*compute.Address, error) {
	aggregatedList, err := g.service.Addresses.AggregatedList(g.projectID).Filter(fmt.Sprintf("name eq %s", addressName)).Do()
	if err != nil {
		return nil, err
	}

	items, ok := aggregatedList.Items["regions/"+g.region]
	if !ok {
		return nil, nil
	}

	if len(items.Addresses) == 0 {
		return nil, nil
	}

	return items.Addresses[0], nil
}

func (g GCP) GetFirewallRule(firewallRuleName string) (*compute.Firewall, error) {
	return g.service.Firewalls.Get(g.projectID, firewallRuleName).Do()
}

func (g GCP) GetTargetPool(targetPoolName string) (*compute.TargetPool, error) {
	return g.service.TargetPools.Get(g.projectID, g.region, targetPoolName).Do()
}

func (g GCP) GetTargetHTTPSProxy(name string) (*compute.TargetHttpsProxy, error) {
	return g.service.TargetHttpsProxies.Get(g.projectID, name).Do()
}

func (g GCP) GetHealthCheck(healthCheckName string) (*compute.HttpHealthCheck, error) {
	return g.service.HttpHealthChecks.Get(g.projectID, healthCheckName).Do()
}
