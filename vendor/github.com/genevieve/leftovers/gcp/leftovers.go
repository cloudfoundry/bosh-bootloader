package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/fatih/color"
	"github.com/genevieve/leftovers/app"
	"github.com/genevieve/leftovers/common"
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
	Type() string
}

type Leftovers struct {
	logger       logger
	asyncDeleter app.AsyncDeleter
	resources    []resource
}

// NewLeftovers returns a new Leftovers for GCP that can be used to list resources,
// list types, or delete resources for the provided account. It returns an error
// if the credentials provided are invalid or if a client fails to be created.
func NewLeftovers(logger logger, keyPath string) (Leftovers, error) {
	if keyPath == "" {
		return Leftovers{}, errors.New("Missing service account key path.")
	}

	if keyPath[0] == '~' {
		var err error
		keyPath, err = homedir.Expand(keyPath)
		if err != nil {
			return Leftovers{}, fmt.Errorf("Invalid service account key path: %s", keyPath)
		}
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
	dnsClient := dns.NewClient(p.ProjectId, dnsService)

	sqlService, err := gcpsql.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	sqlClient := sql.NewClient(p.ProjectId, sqlService, logger)

	storageService, err := gcpstorage.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	storageClient := storage.NewClient(p.ProjectId, storageService)

	iamService, err := gcpiam.New(httpClient)
	if err != nil {
		return Leftovers{}, err
	}
	iamClient := iam.NewClient(p.ProjectId, iamService)

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

	asyncDeleter := app.NewAsyncDeleter(logger)

	return Leftovers{
		logger:       logger,
		asyncDeleter: asyncDeleter,
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

// List will print all of the resources that match the provided filter.
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

// Types will print all the resource types that can
// be deleted on this IaaS.
func (l Leftovers) Types() {
	l.logger.NoConfirm()

	for _, r := range l.resources {
		l.logger.Println(r.Type())
	}
}

// Delete will collect all resources that contain
// the provided filter in the resource's identifier, prompt
// you to confirm deletion (if enabled), and delete those
// that are selected.
func (l Leftovers) Delete(filter string) error {
	deletables := [][]common.Deletable{}

	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			l.logger.Println(color.YellowString(err.Error()))
		}

		deletables = append(deletables, list)
	}

	return l.asyncDeleter.Run(deletables)
}

// DeleteType will collect all resources of the provied type that contain
// the provided filter in the resource's identifier, prompt
// you to confirm deletion (if enabled), and delete those
// that are selected.
func (l Leftovers) DeleteType(filter, rType string) error {
	deletables := [][]common.Deletable{}

	for _, r := range l.resources {
		if r.Type() == rType {
			list, err := r.List(filter)
			if err != nil {
				l.logger.Println(color.YellowString(err.Error()))
			}

			deletables = append(deletables, list)
		}
	}

	return l.asyncDeleter.Run(deletables)
}
