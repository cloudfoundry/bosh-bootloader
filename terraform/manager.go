package terraform

import "github.com/cloudfoundry/bosh-bootloader/storage"

type Manager struct {
	executor executor
	logger   logger
}

type executor interface {
	Destroy(serviceAccountKey, envID, projectID, zone, region, terraformTemplate, tfState string) (string, error)
}

type logger interface {
	Println(message string)
}

func NewManager(executor executor, logger logger) Manager {
	return Manager{
		executor: executor,
		logger:   logger,
	}
}

func (m Manager) Destroy(bblState storage.State) (storage.State, error) {
	tfState, err := m.executor.Destroy(bblState.GCP.ServiceAccountKey, bblState.EnvID, bblState.GCP.ProjectID, bblState.GCP.Zone, bblState.GCP.Region,
		VarsTemplate, bblState.TFState)
	switch err.(type) {
	case ExecutorDestroyError:
		executorDestroyError := err.(ExecutorDestroyError)
		bblState.TFState = executorDestroyError.tfState
		return storage.State{}, NewManagerDestroyError(bblState, executorDestroyError)
	case error:
		return storage.State{}, err
	}

	bblState.TFState = tfState
	return bblState, nil
}
