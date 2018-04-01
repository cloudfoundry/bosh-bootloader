package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/fatih/color"
	"github.com/genevieve/leftovers/gcp/common"
	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/dns"
	"golang.org/x/oauth2/google"
	gcpcompute "google.golang.org/api/compute/v1"
	gcpdns "google.golang.org/api/dns/v1"
)

type resource interface {
	List(filter string) ([]common.Deletable, error)
}

type Leftovers struct {
	logger    logger
	resources []resource
}

func (l Leftovers) List(filter string) {
	l.logger.NoConfirm()

	var deletables []common.Deletable

	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			l.logger.Println(color.YellowString(err.Error()))
		}

		deletables = append(deletables, list...)
	}

	for _, d := range deletables {
		l.logger.Println(fmt.Sprintf("[%s: %s]", d.Type(), d.Name()))
	}
}

func (l Leftovers) Delete(filter string) error {
	deletables := [][]common.Deletable{}

	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			l.logger.Println(color.YellowString(err.Error()))
		}

		deletables = append(deletables, list)
	}

	var wg sync.WaitGroup

	for _, list := range deletables {
		for _, d := range list {
			wg.Add(1)

			go func(d common.Deletable) {
				defer wg.Done()

				l.logger.Println(fmt.Sprintf("[%s: %s] Deleting...", d.Type(), d.Name()))

				if err := d.Delete(); err != nil {
					l.logger.Println(fmt.Sprintf("[%s: %s] %s", d.Type(), d.Name(), color.YellowString(err.Error())))
				} else {
					l.logger.Println(fmt.Sprintf("[%s: %s] %s", d.Type(), d.Name(), color.GreenString("Deleted!")))
				}
			}(d)
		}

		wg.Wait()
	}

	return nil
}

func NewLeftovers(logger logger, keyPath string) (Leftovers, error) {
	if keyPath == "" {
		return Leftovers{}, errors.New("Missing service account key path.")
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		key = []byte(keyPath)
	}

	p := struct {
		ProjectId string `json:"project_id"`
	}{}
	if err := json.Unmarshal(key, &p); err != nil {
		return Leftovers{}, fmt.Errorf("Unmarshalling account key for project id: %s", err)
	}

	config, err := google.JWTConfigFromJSON(key, gcpcompute.ComputeScope, gcpdns.NdevClouddnsReadwriteScope)
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating jwt config: %s", err)
	}

	service, err := gcpcompute.New(config.Client(context.Background()))
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating gcp client: %s", err)
	}
	client := compute.NewClient(p.ProjectId, service, logger)

	regions, err := client.ListRegions()
	if err != nil {
		return Leftovers{}, fmt.Errorf("Listing regions: %s", err)
	}

	zones, err := client.ListZones()
	if err != nil {
		return Leftovers{}, fmt.Errorf("Listing zones: %s", err)
	}

	dnsService, err := gcpdns.New(config.Client(context.Background()))
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating gcp client: %s", err)
	}
	dnsClient := dns.NewClient(p.ProjectId, dnsService, logger)

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
			compute.NewInstanceTemplates(client, logger),
			compute.NewInstanceGroupManagers(client, logger, zones),
			compute.NewInstances(client, logger, zones),
			compute.NewInstanceGroups(client, logger, zones),
			compute.NewGlobalHealthChecks(client, logger),
			compute.NewHttpHealthChecks(client, logger),
			compute.NewHttpsHealthChecks(client, logger),
			compute.NewImages(client, logger),
			compute.NewDisks(client, logger, zones),
			compute.NewSubnetworks(client, logger, regions),
			compute.NewNetworks(client, logger),
			compute.NewAddresses(client, logger, regions),
			compute.NewGlobalAddresses(client, logger),
			dns.NewManagedZones(dnsClient, dns.NewRecordSets(dnsClient), logger),
		},
	}, nil
}
