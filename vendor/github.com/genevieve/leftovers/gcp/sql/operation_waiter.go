package sql

import (
	"fmt"

	"github.com/genevieve/leftovers/gcp/common"

	gcpsql "google.golang.org/api/sqladmin/v1beta4"
)

type operationWaiter struct {
	op      *gcpsql.Operation
	service *gcpsql.Service
	project string
	logger  logger
}

func NewOperationWaiter(op *gcpsql.Operation, service *gcpsql.Service, project string, logger logger) operationWaiter {
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

	result, ok := raw.(*gcpsql.Operation)
	if ok && result.Error != nil && len(result.Error.Errors) > 0 {
		return fmt.Errorf("Operation error: %s", result.Error.Errors[0].Message)
	}

	return nil
}

func (c *operationWaiter) refreshFunc() common.StateRefreshFunc {
	return func() (interface{}, string, error) {
		op, err := c.service.Operations.Get(c.project, c.op.Name).Do()

		if err != nil {
			return nil, "", fmt.Errorf("Refreshing operation request: %s", err)
		}

		return op, op.Status, nil
	}
}
