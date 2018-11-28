package actors

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"golang.org/x/oauth2/google"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	compute "google.golang.org/api/compute/v1"

	. "github.com/onsi/gomega"
)

type gcpLBHelper struct {
	service               *compute.Service
	projectID             string
	region                string
	serviceAccountKeyPath string
}

func NewGCPLBHelper(config acceptance.Config) gcpLBHelper {
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

	return gcpLBHelper{
		service:   service,
		projectID: p.ProjectID,
		region:    config.GCPRegion,
	}
}

func (g gcpLBHelper) GetLBArgs() []string {
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

func (g gcpLBHelper) VerifyCloudConfigExtensions(vmExtensions []string) {
	Expect(vmExtensions).To(ContainElement("cf-router-network-properties"))
	Expect(vmExtensions).To(ContainElement("diego-ssh-proxy-network-properties"))
	Expect(vmExtensions).To(ContainElement("cf-tcp-router-network-properties"))
}

func (g gcpLBHelper) ConfirmLBsExist(envID string) {
	targetPools := []string{envID + "-cf-ssh-proxy", envID + "-cf-tcp-router"}
	for _, p := range targetPools {
		targetPool, err := g.service.TargetPools.Get(g.projectID, g.region, p).Do()
		Expect(err).NotTo(HaveOccurred())
		Expect(targetPool.Name).NotTo(BeNil())
		Expect(targetPool.Name).To(Equal(p))
	}

	targetHTTPSProxy, err := g.service.TargetHttpsProxies.Get(g.projectID, envID+"-https-proxy").Do()
	Expect(err).NotTo(HaveOccurred())
	Expect(targetHTTPSProxy.SslCertificates).To(HaveLen(1))
}

func (g gcpLBHelper) ConfirmNoLBsExist(envID string) {
	targetPools := []string{envID + "-cf-ssh-proxy", envID + "-cf-tcp-router"}
	for _, p := range targetPools {
		_, err := g.service.TargetPools.Get(g.projectID, g.region, p).Do()
		Expect(err).To(MatchError(MatchRegexp(`The resource 'projects\/.+` + p + `' was not found`)))
	}
}

func (g gcpLBHelper) VerifyBblLBOutput(stdout string) {
	Expect(stdout).To(MatchRegexp("CF Router LB:.*"))
	Expect(stdout).To(MatchRegexp("CF SSH Proxy LB:.*"))
	Expect(stdout).To(MatchRegexp("CF TCP Router LB:.*"))
	Expect(stdout).To(MatchRegexp("CF WebSocket LB:.*"))
}

func (g gcpLBHelper) ConfirmNoStemcellsExist(stemcellIDs []string) {
	for _, stemcellID := range stemcellIDs {
		_, err := g.service.Images.Get(g.projectID, stemcellID).Do()
		Expect(err).To(MatchError(MatchRegexp("googleapi: Error 404.*")))
	}
}
