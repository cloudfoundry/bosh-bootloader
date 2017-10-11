package actors

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	"golang.org/x/oauth2/google"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	compute "google.golang.org/api/compute/v1"

	. "github.com/onsi/gomega"
)

type GCP struct {
	service   *compute.Service
	projectID string
	region    string
	zone      string
}

func NewGCP(config acceptance.Config) GCP {
	rawServiceAccountKey, err := ioutil.ReadFile(config.GCPServiceAccountKey)
	if err != nil {
		rawServiceAccountKey = []byte(config.GCPServiceAccountKey)
	}

	googleConfig, err := google.JWTConfigFromJSON(rawServiceAccountKey, "https://www.googleapis.com/auth/compute")
	if err != nil {
		panic(err)
	}

	p := struct {
		ProjectID string `json:"project_id"`
	}{}
	err = json.Unmarshal(rawServiceAccountKey, &p)
	if err != nil {
		panic(err)
	}

	service, err := compute.New(googleConfig.Client(context.Background()))
	if err != nil {
		panic(err)
	}

	return GCP{
		service:   service,
		projectID: p.ProjectID,
		region:    config.GCPRegion,
		zone:      config.GCPZone,
	}
}

func (g GCP) GetNetwork(networkName string) (*compute.Network, error) {
	return g.service.Networks.Get(g.projectID, networkName).Do()
}

func (g GCP) GetTargetPool(targetPoolName string) (*compute.TargetPool, error) {
	return g.service.TargetPools.Get(g.projectID, g.region, targetPoolName).Do()
}

func (g GCP) GetTargetHTTPSProxy(name string) (*compute.TargetHttpsProxy, error) {
	return g.service.TargetHttpsProxies.Get(g.projectID, name).Do()
}

func (g GCP) NetworkHasBOSHDirector(envID string) bool {
	list, err := g.service.Instances.List(g.projectID, g.zone).
		Filter("labels.director:bosh-init").
		Do()
	Expect(err).NotTo(HaveOccurred())

	for _, item := range list.Items {
		for _, networkInterface := range item.NetworkInterfaces {
			if strings.Contains(networkInterface.Network, envID) {
				return true
			}
		}
	}

	return false
}
