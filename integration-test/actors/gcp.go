package actors

import (
	"context"
	"io/ioutil"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

type GCP struct {
	service   *compute.Service
	projectID string
}

func NewGCP(config integration.Config) GCP {
	serviceAccountKey, err := ioutil.ReadFile(config.GCPServiceAccountKeyPath)

	googleConfig, err := google.JWTConfigFromJSON(serviceAccountKey, "https://www.googleapis.com/auth/compute")
	if err != nil {
		panic(err)
	}

	service, err := compute.New(googleConfig.Client(context.Background()))
	if err != nil {
		panic(err)
	}

	return GCP{
		service:   service,
		projectID: config.GCPProjectID,
	}
}

func (g GCP) SSHKey() (string, error) {
	project, err := g.service.Projects.Get(g.projectID).Do()
	if err != nil {
		return "", err
	}

	for _, item := range project.CommonInstanceMetadata.Items {
		if item.Key == "sshKeys" {
			return *item.Value, nil
		}
	}

	return "", nil
}

func (g GCP) RemoveSSHKey() error {
	project, err := g.service.Projects.Get(g.projectID).Do()
	if err != nil {
		return err
	}

	for i, item := range project.CommonInstanceMetadata.Items {
		if item.Key == "sshKeys" {
			newValue := ""
			project.CommonInstanceMetadata.Items[i].Value = &newValue
			break
		}
	}

	_, err = g.service.Projects.SetCommonInstanceMetadata(g.projectID, project.CommonInstanceMetadata).Do()
	if err != nil {
		return err
	}

	return nil
}
