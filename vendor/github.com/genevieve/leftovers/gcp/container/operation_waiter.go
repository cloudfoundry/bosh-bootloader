package container

import (
	"fmt"

	"github.com/genevieve/leftovers/gcp/common"

	gcpcontainer "google.golang.org/api/container/v1"
)

type operationWaiter struct {
	op      *gcpcontainer.Operation
	service *gcpcontainer.ProjectsZonesService
	project string
	logger  logger
}

func NewOperationWaiter(op *gcpcontainer.Operation, service *gcpcontainer.Service, project string, logger logger) operationWaiter {
	return operationWaiter{
		op:      op,
		service: service.Projects.Zones,
		project: project,
		logger:  logger,
	}
}

func (w *operationWaiter) Wait() error {
	state := common.NewState(w.logger, w.refreshFunc())

	raw, err := state.Wait()
	if err != nil {
		return fmt.Errorf("Waiting for operation to complete: %s", err)
	}

	result, ok := raw.(*gcpcontainer.Operation)
	if ok && result.Status != "DONE" {
		return fmt.Errorf("Operation error: %s", result.Status)
	}

	return nil
}

func (c *operationWaiter) refreshFunc() common.StateRefreshFunc {
	return func() (interface{}, string, error) {
		op, err := c.service.Operations.Get(c.project, c.op.Zone, c.op.Name).Do()

		if err != nil {
			return nil, "", fmt.Errorf("Refreshing operation request: %s", err)
		}

		return op, op.Status, nil
	}
}
