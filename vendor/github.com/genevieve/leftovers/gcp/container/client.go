package container

import (
	"fmt"

	gcpcontainer "google.golang.org/api/container/v1"
	"google.golang.org/api/googleapi"
)

type client struct {
	project string
	logger  logger

	service    *gcpcontainer.Service
	containers *gcpcontainer.ProjectsZonesClustersService
}

func NewClient(project string, service *gcpcontainer.Service, logger logger) client {
	return client{
		project: project,
		logger:  logger,

		service:    service,
		containers: service.Projects.Zones.Clusters,
	}
}

func (c client) ListClusters(zone string) (*gcpcontainer.ListClustersResponse, error) {
	return c.containers.List(c.project, zone).Do()
}

func (c client) DeleteCluster(zone string, cluster string) error {
	return c.wait(c.containers.Delete(c.project, zone, cluster))
}

type request interface {
	Do(...googleapi.CallOption) (*gcpcontainer.Operation, error)
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
