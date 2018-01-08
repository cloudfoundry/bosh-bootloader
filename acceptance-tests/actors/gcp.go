package actors

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	compute "google.golang.org/api/compute/v1"

	. "github.com/onsi/gomega"
)

type GCP struct {
	service       *compute.Service
	projectID     string
	region        string
	stateFilePath string
}

func NewGCP(config acceptance.Config) GCP {
	rawServiceAccountKey, err := ioutil.ReadFile(config.GCPServiceAccountKey)
	if err != nil {
		rawServiceAccountKey = []byte(config.GCPServiceAccountKey)
	}

	googleConfig, err := google.JWTConfigFromJSON(rawServiceAccountKey, "https://www.googleapis.com/auth/compute")
	Expect(err).NotTo(HaveOccurred())

	p := struct {
		ProjectID string `json:"project_id"`
	}{}
	err = json.Unmarshal(rawServiceAccountKey, &p)
	Expect(err).NotTo(HaveOccurred())

	service, err := compute.New(googleConfig.Client(context.Background()))
	Expect(err).NotTo(HaveOccurred())

	stateFilePath := filepath.Join(config.StateFileDir, "bbl-state.json")

	return GCP{
		service:       service,
		projectID:     p.ProjectID,
		region:        config.GCPRegion,
		stateFilePath: stateFilePath,
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
	zone := getZoneFromStateFile(g.stateFilePath)
	list, err := g.service.Instances.List(g.projectID, zone).
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

func getZoneFromStateFile(path string) string {
	p := struct {
		GCP struct {
			Zone string `json:"zone"`
		} `json:"gcp"`
	}{}

	contents, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	err = json.Unmarshal(contents, &p)
	Expect(err).NotTo(HaveOccurred())
	return p.GCP.Zone
}

type gcpIaasLbHelper struct {
	gcp GCP
}

func (g gcpIaasLbHelper) GetLBArgs() []string {
	certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
	Expect(err).NotTo(HaveOccurred())
	keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
	Expect(err).NotTo(HaveOccurred())

	return []string{
		"--lb-type", "cf",
		"--lb-cert", certPath,
		"--lb-key", keyPath,
	}
}

func (g gcpIaasLbHelper) ConfirmLBsExist(envID string) {
	targetPools := []string{envID + "-cf-ssh-proxy", envID + "-cf-tcp-router"}
	for _, p := range targetPools {
		targetPool, err := g.gcp.GetTargetPool(p)
		Expect(err).NotTo(HaveOccurred())
		Expect(targetPool.Name).NotTo(BeNil())
		Expect(targetPool.Name).To(Equal(p))
	}

	targetHTTPSProxy, err := g.gcp.GetTargetHTTPSProxy(envID + "-https-proxy")
	Expect(err).NotTo(HaveOccurred())
	Expect(targetHTTPSProxy.SslCertificates).To(HaveLen(1))
}

func (g gcpIaasLbHelper) ConfirmNoLBsExist(envID string) {
	targetPools := []string{envID + "-cf-ssh-proxy", envID + "-cf-tcp-router"}
	for _, p := range targetPools {
		_, err := g.gcp.GetTargetPool(p)
		Expect(err).To(MatchError(MatchRegexp(`The resource 'projects\/.+` + p + `' was not found`)))
	}
}
