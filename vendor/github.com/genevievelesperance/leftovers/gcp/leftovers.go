package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/dns"
	"golang.org/x/oauth2/google"
	gcpcompute "google.golang.org/api/compute/v1"
	gcpdns "google.golang.org/api/dns/v1"
)

type resource interface {
	List(filter string) (map[string]string, error)
	Delete(items map[string]string)
}

type deleter struct {
	resource resource
	items    map[string]string
}

type Leftovers struct {
	logger    logger
	resources []resource
}

func (l Leftovers) Delete(filter string) error {
	var deleters []deleter

	for _, r := range l.resources {
		items, err := r.List(filter)
		if err != nil {
			return err
		}

		deleters = append(deleters, deleter{
			resource: r,
			items:    items,
		})
	}

	for _, d := range deleters {
		d.resource.Delete(d.items)
	}

	return nil
}

func NewLeftovers(logger logger, keyPath string) (Leftovers, error) {
	if keyPath == "" {
		return Leftovers{}, errors.New("Missing service account key path.")
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return Leftovers{}, fmt.Errorf("Reading service account key path %s: %s", keyPath, err)
	}

	p := struct {
		ProjectId string `json:"project_id"`
	}{}
	if err := json.Unmarshal(key, &p); err != nil {
		return Leftovers{}, fmt.Errorf("Unmarshalling account key for project id: %s", err)
	}

	logger.Println(fmt.Sprintf("Cleaning gcp project: %s.", p.ProjectId))

	config, err := google.JWTConfigFromJSON(key, gcpcompute.ComputeScope, gcpdns.NdevClouddnsReadwriteScope)
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating jwt config: %s", err)
	}

	service, err := gcpcompute.New(config.Client(context.Background()))
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating gcp client: %s", err)
	}

	client := compute.NewClient(p.ProjectId, service, logger)

	dnsService, err := gcpdns.New(config.Client(context.Background()))
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating gcp client: %s", err)
	}

	dnsClient := dns.NewClient(p.ProjectId, dnsService, logger)

	regions, err := client.ListRegions()
	if err != nil {
		return Leftovers{}, fmt.Errorf("Listing regions: %s", err)
	}

	zones, err := client.ListZones()
	if err != nil {
		return Leftovers{}, fmt.Errorf("Listing zones: %s", err)
	}

	return Leftovers{
		logger: logger,
		resources: []resource{
			compute.NewForwardingRules(client, logger, regions),
			compute.NewGlobalForwardingRules(client, logger),
			compute.NewFirewalls(client, logger),
			compute.NewTargetHttpProxies(client, logger),
			compute.NewTargetHttpsProxies(client, logger),
			compute.NewUrlMaps(client, logger),
			compute.NewTargetPools(client, logger, regions),
			compute.NewBackendServices(client, logger),
			compute.NewInstances(client, logger, zones),
			compute.NewInstanceGroups(client, logger, zones),
			compute.NewGlobalHealthChecks(client, logger),
			compute.NewHttpHealthChecks(client, logger),
			compute.NewHttpsHealthChecks(client, logger),
			compute.NewDisks(client, logger, zones),
			compute.NewSubnetworks(client, logger, regions),
			compute.NewNetworks(client, logger),
			compute.NewAddresses(client, logger, regions),
			dns.NewManagedZones(dnsClient, logger),
		},
	}, nil
}
