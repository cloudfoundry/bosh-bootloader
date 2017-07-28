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

func gcpHTTPClientFunc(config *jwt.Config) *http.Client {
	return config.Client(context.Background())
}

var gcpHTTPClient = gcpHTTPClientFunc

type ClientProvider struct {
	basePath string
	client   Client
}

func NewClientProvider(gcpBasePath string) *ClientProvider {
	return &ClientProvider{
		basePath: gcpBasePath,
	}
}

func (p *ClientProvider) SetConfig(serviceAccountKey, projectID, region, zone string) error {
	config, err := google.JWTConfigFromJSON([]byte(serviceAccountKey), compute.ComputeScope)
	if err != nil {
		return err
	}

	if p.basePath != "" {
		config.TokenURL = p.basePath
	}

	service, err := compute.New(gcpHTTPClient(config))
	if err != nil {
		return err
	}

	if p.basePath != "" {
		service.BasePath = p.basePath
	}

	p.client = GCPClient{
		service:   service,
		projectID: projectID,
		zone:      zone,
	}

	_, err = p.client.GetRegion(region)
	if err != nil {
		return err
	}

	_, err = p.client.GetZone(zone)
	if err != nil {
		return err
	}

	return nil
}

func (p *ClientProvider) Client() Client {
	return p.client
}
