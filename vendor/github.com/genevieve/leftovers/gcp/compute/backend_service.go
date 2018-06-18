package compute

import "fmt"

type BackendService struct {
	client backendServicesClient
	name   string
}

func NewBackendService(client backendServicesClient, name string) BackendService {
	return BackendService{
		client: client,
		name:   name,
	}
}

func (b BackendService) Delete() error {
	err := b.client.DeleteBackendService(b.name)
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (b BackendService) Name() string {
	return b.name
}

func (b BackendService) Type() string {
	return "Backend Service"
}
