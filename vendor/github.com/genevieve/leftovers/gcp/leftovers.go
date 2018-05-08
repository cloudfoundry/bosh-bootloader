package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	"github.com/genevieve/leftovers/gcp/common"
	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/container"
	"github.com/genevieve/leftovers/gcp/dns"
	"github.com/genevieve/leftovers/gcp/iam"
	"github.com/genevieve/leftovers/gcp/sql"
	"github.com/genevieve/leftovers/gcp/storage"
	"golang.org/x/oauth2/google"
	gcpcompute "google.golang.org/api/compute/v1"
	gcpcontainer "google.golang.org/api/container/v1"
	gcpdns "google.golang.org/api/dns/v1"
	gcpiam "google.golang.org/api/iam/v1"
	gcpsql "google.golang.org/api/sqladmin/v1beta4"
	gcpstorage "google.golang.org/api/storage/v1"
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

	absKeyPath, _ := filepath.Abs(keyPath)

	key, err := ioutil.ReadFile(absKeyPath)
	if err != nil {
		key = []byte(keyPath)
	}

	p := struct {
		ProjectId string `json:"project_id"`
	}{}
	if err := json.Unmarshal(key, &p); err != nil {
		return Leftovers{}, fmt.Errorf("Unmarshalling account key for project id: %s", err)
	}

	config, err := google.JWTConfigFromJSON(key, gcpcompute.CloudPlatformScope)
	if err != nil {
		return Leftovers{}, fmt.Errorf("Creating jwt config: %s", err)
	}

	httpClient := config.Client(context.Background())

	service, err := gcpcompute.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	client := compute.NewClient(p.ProjectId, service, logger)

	dnsService, err := gcpdns.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	dnsClient := dns.NewClient(p.ProjectId, dnsService, logger)

	sqlService, err := gcpsql.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	sqlClient := sql.NewClient(p.ProjectId, sqlService, logger)

	storageService, err := gcpstorage.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	storageClient := storage.NewClient(p.ProjectId, storageService, logger)

	iamService, err := gcpiam.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	iamClient := iam.NewClient(p.ProjectId, iamService, logger)

	containerService, err := gcpcontainer.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	containerClient := container.NewClient(p.ProjectId, containerService, logger)

	regions, err := client.ListRegions()
	if err != nil {
		return Leftovers{}, err
	}

	zones, err := client.ListZones()
	if err != nil {
		return Leftovers{}, err
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
			compute.NewSslCertificates(client, logger),
			iam.NewServiceAccounts(iamClient, logger),
			dns.NewManagedZones(dnsClient, dns.NewRecordSets(dnsClient), logger),
			sql.NewInstances(sqlClient, logger),
			storage.NewBuckets(storageClient, logger),
			container.NewClusters(containerClient, zones, logger),
		},
	}, nil
}
