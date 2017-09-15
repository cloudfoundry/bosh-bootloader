package gcp

import (
	"net/http"

	"golang.org/x/oauth2/jwt"
)

func SetGCPHTTPClient(f func(*jwt.Config) *http.Client) {
	gcpHTTPClient = f
}

func ResetGCPHTTPClient() {
	gcpHTTPClient = gcpHTTPClientFunc
}

func NewClientWithInjectedComputeClient(computeClient ComputeClient, projectID, zone string) Client {
	return Client{
		computeClient: computeClient,
		projectID:     projectID,
		zone:          zone,
	}
}
