package compute

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
	return b.client.DeleteBackendService(b.name)
}
