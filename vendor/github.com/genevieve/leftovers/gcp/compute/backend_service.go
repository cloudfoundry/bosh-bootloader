package compute

import "fmt"

type BackendService struct {
	client backendServicesClient
	name   string
	kind   string
}

func NewBackendService(client backendServicesClient, name string) BackendService {
	return BackendService{
		client: client,
		name:   name,
		kind:   "backend-service",
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

func (b BackendService) Kind() string {
	return b.kind
}
