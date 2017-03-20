package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/rosenhouse/awsfaker"

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

type awsBOSHDeploymentVars struct {
	InternalCIDR          string   `yaml:"internal_cidr"`
	InternalGateway       string   `yaml:"internal_gw"`
	InternalIP            string   `yaml:"internal_ip"`
	DirectorName          string   `yaml:"director_name"`
	ExternalIP            string   `yaml:"external_ip"`
	AZ                    string   `yaml:"az"`
	SubnetID              string   `yaml:"subnet_id"`
	AccessKeyID           string   `yaml:"access_key_id"`
	SecretAccessKey       string   `yaml:"secret_access_key"`
	DefaultKeyName        string   `yaml:"default_key_name"`
	DefaultSecurityGroups []string `yaml:"default_security_groups"`
	Region                string   `yaml:"region"`
	PrivateKey            string   `yaml:"private_key"`
}

var _ = Describe("bosh-deployment-vars", func() {
	var (
		tempDirectory              string
		serviceAccountKeyPath      string
		pathToFakeTerraform        string
		pathToTerraform            string
		pathToFakeBOSH             string
		pathToBOSH                 string
		fakeBOSH                   *fakeBOSHDirector
		fakeBOSHServer             *httptest.Server
		fakeBOSHCLIBackendServer   *httptest.Server
		fakeTerraformBackendServer *httptest.Server
	)

	BeforeEach(func() {
		var err error

		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeBOSHCLIBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/version":
				responseWriter.Write([]byte("2.0.0"))
			}
		}))

		fakeTerraformBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/output/external_ip":
				responseWriter.Write([]byte("127.0.0.1"))
			case "/output/director_address":
				responseWriter.Write([]byte(fakeBOSHServer.URL))
			case "/output/network_name":
				responseWriter.Write([]byte("some-network-name"))
			case "/output/subnetwork_name":
				responseWriter.Write([]byte("some-subnetwork-name"))
			case "/output/internal_tag_name":
				responseWriter.Write([]byte("some-internal-tag"))
			case "/output/bosh_open_tag_name":
				responseWriter.Write([]byte("some-bosh-tag"))
			case "/version":
				responseWriter.Write([]byte("0.8.6"))
			}
		}))

		pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
		err = os.Rename(pathToFakeTerraform, pathToTerraform)
		Expect(err).NotTo(HaveOccurred())

		pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHCLIBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
		err = os.Rename(pathToFakeBOSH, pathToBOSH)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), filepath.Dir(pathToBOSH), originalPath}, ":"))
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)
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

	Context("AWS", func() {
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
		})

		It("prints a bosh create-env compatible vars-file", func() {
			args := []string{
				fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
				"--state-dir", tempDirectory,
				"bosh-deployment-vars",
			}
			session := executeCommand(args, 0)

			var vars awsBOSHDeploymentVars
			yaml.Unmarshal(session.Out.Contents(), &vars)
			Expect(vars.InternalCIDR).To(Equal("10.0.0.0/24"))
			Expect(vars.InternalGateway).To(Equal("10.0.0.1"))
			Expect(vars.InternalIP).To(Equal("10.0.0.6"))
			Expect(vars.DirectorName).To(Equal("bosh-some-env-id"))
			Expect(vars.ExternalIP).To(Equal("127.0.0.1"))
			Expect(vars.AZ).To(Equal("some-az"))
			Expect(vars.SubnetID).To(Equal("some-subnet-id"))
			Expect(vars.AccessKeyID).To(Equal("some-user-access-key"))
			Expect(vars.SecretAccessKey).To(Equal("some-user-secret-access-key"))
			Expect(vars.DefaultKeyName).To(Equal("keypair-some-env-id"))
			Expect(vars.DefaultSecurityGroups).To(Equal([]string{"some-default-security-group"}))
			Expect(vars.Region).To(Equal("some-region"))
			Expect(vars.PrivateKey).To(Equal(testhelpers.PRIVATE_KEY))
		})
	})

	Context("when the bosh cli version is <2.0", func() {
		BeforeEach(func() {
			fakeBOSHCLIBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/version":
					responseWriter.Write([]byte("1.9.0"))
				}
			}))

			var err error
			pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
				"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHCLIBackendServer.URL))
			Expect(err).NotTo(HaveOccurred())

			pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
			err = os.Rename(pathToFakeBOSH, pathToBOSH)
			Expect(err).NotTo(HaveOccurred())

			os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSH), originalPath}, ":"))
		})

		AfterEach(func() {
			os.Setenv("PATH", originalPath)
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
