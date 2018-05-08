package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/gcp/common"

	gcpcompute "google.golang.org/api/compute/v1"
)

type operationWaiter struct {
	op      *gcpcompute.Operation
	service *gcpcompute.Service
	project string
	logger  logger
}

func NewOperationWaiter(op *gcpcompute.Operation, service *gcpcompute.Service, project string, logger logger) operationWaiter {
	return operationWaiter{
		op:      op,
		service: service,
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

	result, ok := raw.(*gcpcompute.Operation)
	if ok && result.Error != nil && len(result.Error.Errors) > 0 {
		return fmt.Errorf("Operation error: %s", result.Error.Errors[0].Message)
	}

	return nil
}

func (c *operationWaiter) refreshFunc() common.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var op *gcpcompute.Operation
		var err error

		if c.op.Zone != "" {
			zoneURLParts := strings.Split(c.op.Zone, "/")
			zone := zoneURLParts[len(zoneURLParts)-1]
			op, err = c.service.ZoneOperations.Get(c.project, zone, c.op.Name).Do()
		} else if c.op.Region != "" {
			regionURLParts := strings.Split(c.op.Region, "/")
			region := regionURLParts[len(regionURLParts)-1]
			op, err = c.service.RegionOperations.Get(c.project, region, c.op.Name).Do()
		} else {
			op, err = c.service.GlobalOperations.Get(c.project, c.op.Name).Do()
		}

		if err != nil {
			return nil, "", fmt.Errorf("Refreshing operation request: %s", err)
		}

		return op, op.Status, nil
	}
}
