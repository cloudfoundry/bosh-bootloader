package gcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	compute "google.golang.org/api/compute/v1"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

func gcpHTTPClientFunc(config *jwt.Config) *http.Client {
	return config.Client(context.Background())
}

var gcpHTTPClient = gcpHTTPClientFunc

func NewClient(gcpConfig storage.GCP, basePath string, fileIO fileio.FileIO) (Client, error) {
	key, err := fileIO.ReadFile(gcpConfig.ServiceAccountKey)
	if err != nil {
		panic(err)
	}
	config, err := google.JWTConfigFromJSON(key, compute.ComputeScope)
	if err != nil {
		return Client{}, fmt.Errorf("parse service account key: %s", err)
	}

	if basePath != "" {
		config.TokenURL = basePath
	}

	service, err := compute.New(gcpHTTPClient(config))
	if err != nil {
		return Client{}, fmt.Errorf("create gcp client: %s", err)
	}

	if basePath != "" {
		service.BasePath = basePath
	}

	client := Client{
		computeClient: gcpComputeClient{service: service},
		projectID:     gcpConfig.ProjectID,
		zone:          gcpConfig.Zone,
	}

	_, err = client.GetRegion(gcpConfig.Region)
	if err != nil {
		return Client{}, fmt.Errorf("get region: %s", err)
	}

	return client, nil
}
