package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	yaml "gopkg.in/yaml.v2"
)

type gcpBOSHDeploymentVars struct {
	InternalCIDR       string   `yaml:"internal_cidr"`
	InternalGateway    string   `yaml:"internal_gw"`
	InternalIP         string   `yaml:"internal_ip"`
	DirectorName       string   `yaml:"director_name"`
	ExternalIP         string   `yaml:"external_ip"`
	Zone               string   `yaml:"zone"`
	Network            string   `yaml:"network"`
	Subnetwork         string   `yaml:"subnetwork"`
	Tags               []string `yaml:"tags"`
	ProjectID          string   `yaml:"project_id"`
	GCPCredentialsJSON string   `yaml:"gcp_credentials_json"`
}

var _ = Describe("bosh-deployment-vars", func() {
	var (
		tempDirectory         string
		serviceAccountKeyPath string
		fakeBOSH              *fakeBOSHDirector
		fakeBOSHServer        *httptest.Server
	)

	BeforeEach(func() {
		var err error

		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeTerraformBackendServer.SetFakeBOSHServer(fakeBOSHServer.URL)

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		fakeBOSHCLIBackendServer.ResetAll()
		fakeTerraformBackendServer.ResetAll()
	})

	Context("GCP", func() {
		BeforeEach(func() {
			args := []string{
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--name", "some-env-id",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "some-region",
			}
			executeCommand(args, 0)
		})

		It("prints a bosh create-env compatible vars-file", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"bosh-deployment-vars",
			}
			session := executeCommand(args, 0)

			var vars gcpBOSHDeploymentVars
			yaml.Unmarshal(session.Out.Contents(), &vars)

			var realAccountKey map[interface{}]interface{}
			var returnedAccountKey map[interface{}]interface{}
			yaml.Unmarshal([]byte(serviceAccountKey), &realAccountKey)
			yaml.Unmarshal([]byte(vars.GCPCredentialsJSON), &returnedAccountKey)

			Expect(vars.InternalCIDR).To(Equal("10.0.0.0/24"))
			Expect(vars.InternalGateway).To(Equal("10.0.0.1"))
			Expect(vars.InternalIP).To(Equal("10.0.0.6"))
			Expect(vars.DirectorName).To(Equal("bosh-some-env-id"))
			Expect(vars.ExternalIP).To(Equal("127.0.0.1"))
			Expect(vars.Zone).To(Equal("some-zone"))
			Expect(vars.Network).To(Equal("some-network-name"))
			Expect(vars.Subnetwork).To(Equal("some-subnetwork-name"))
			Expect(vars.Tags).To(Equal([]string{"some-bosh-tag", "some-internal-tag"}))
			Expect(vars.ProjectID).To(Equal("some-project-id"))
			Expect(returnedAccountKey).To(Equal(realAccountKey))
		})
	})

	Context("when the bosh cli version is <2.0", func() {
		var (
			fakeAWS       *awsbackend.Backend
			fakeAWSServer *httptest.Server
		)

		BeforeEach(func() {
			fakeAWS = awsbackend.New(fakeBOSHServer.URL)
			fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

			args := []string{
				fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
				"--state-dir", tempDirectory,
				"--debug",
				"up",
				"--name", "some-env-id",
				"--iaas", "aws",
				"--aws-access-key-id", "some-access-key",
				"--aws-secret-access-key", "some-access-secret",
				"--aws-region", "some-region",
			}

			executeCommand(args, 0)

			fakeBOSHCLIBackendServer.SetVersion("1.9.0")
		})

		It("fast fails with a helpful error message", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"bosh-deployment-vars",
			}
			session := executeCommand(args, 1)

			Expect(session.Err.Contents()).To(ContainSubstring("BOSH version must be at least v2.0.0"))
		})
	})
})
