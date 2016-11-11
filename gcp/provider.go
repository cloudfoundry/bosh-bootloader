package gcp

import (
	"context"
	"net/http"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	compute "google.golang.org/api/compute/v1"
)

const (
	GoogleComputeAuth = "https://www.googleapis.com/auth/compute"
)

func clientFunc(config *jwt.Config) *http.Client {
	return config.Client(context.Background())
}

var client = clientFunc

type serviceWrapperStruct struct {
	service *compute.Service
}

type projectsService struct {
	projectsService *compute.ProjectsService
}

type Provider struct {
	serviceWrapperStruct serviceWrapperStruct
	basePath             string
}

func NewProvider(gcpBasePath string) *Provider {
	return &Provider{
		basePath: gcpBasePath,
	}
}

func (p *Provider) SetConfig(serviceAccountKey string) error {
	authURL := GoogleComputeAuth
	if p.basePath != "" {
		authURL = p.basePath
	}
	config, err := google.JWTConfigFromJSON([]byte(serviceAccountKey), authURL)
	if err != nil {
		return err
	}

	service, err := compute.New(client(config))
	if err != nil {
		return err
	}

	if p.basePath != "" {
		service.BasePath = p.basePath
	}

	p.serviceWrapperStruct = serviceWrapperStruct{
		service: service,
	}
	return nil
}

func (p *Provider) GetService() ServiceWrapper {
	return p.serviceWrapperStruct
}

func (s serviceWrapperStruct) GetProjectsService() ProjectsService {
	return projectsService{
		projectsService: s.service.Projects,
	}
}

func (p projectsService) Get(projectId string) (*compute.Project, error) {
	return p.projectsService.Get(projectId).Do()
}

func (p projectsService) SetCommonInstanceMetadata(projectId string, metadata *compute.Metadata) (*compute.Operation, error) {
	return p.projectsService.SetCommonInstanceMetadata(projectId, metadata).Do()
}
