package main_test

import (
	"crypto/rsa"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/ssl"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
	"github.com/square/certstrap/pkix"
)

var _ = Describe("bbl cloud-config", func() {
	var (
		fakeBOSHServer *httptest.Server
		fakeBOSH       *fakeBOSHDirector

		tempDirectory         string
		serviceAccountKeyPath string
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

		fakeBOSHCLIBackendServer.SetCallRealInterpolate(true)
	})

	AfterEach(func() {
		fakeBOSHCLIBackendServer.ResetAll()
		fakeTerraformBackendServer.ResetAll()
	})

	Context("when there is no lb", func() {
		BeforeEach(func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
				"--no-director",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-west1",
			}

			executeCommand(args, 0)
		})

		It("returns the cloud config of the bbl environment", func() {
			contents, err := ioutil.ReadFile("../cloudconfig/fixtures/gcp-cloud-config-no-lb.yml")
			Expect(err).NotTo(HaveOccurred())
			args := []string{
				"--state-dir", tempDirectory,
				"cloud-config",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(MatchYAML(string(contents)))
		})
	})

	Context("when there is a concourse lb", func() {
		BeforeEach(func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
				"--no-director",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-east1",
			}

			executeCommand(args, 0)
		})

		It("returns the cloud config of a bbl environment with concourse lb", func() {
			contents, err := ioutil.ReadFile("../cloudconfig/fixtures/gcp-cloud-config-concourse-lb.yml")
			Expect(err).NotTo(HaveOccurred())
			args := []string{
				"--state-dir", tempDirectory,
				"create-lbs",
				"--type", "concourse",
			}

			executeCommand(args, 0)

			args = []string{
				"--state-dir", tempDirectory,
				"cloud-config",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(MatchYAML(string(contents)))
		})
	})

	Context("when there is a cf lb", func() {
		BeforeEach(func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
				"--no-director",
				"--iaas", "gcp",
				"--gcp-service-account-key", serviceAccountKeyPath,
				"--gcp-project-id", "some-project-id",
				"--gcp-zone", "some-zone",
				"--gcp-region", "us-east1",
			}

			executeCommand(args, 0)
		})

		It("returns the cloud config of a bbl environment with cf lb", func() {
			contents, err := ioutil.ReadFile("../cloudconfig/fixtures/gcp-cloud-config-cf-lb.yml")
			Expect(err).NotTo(HaveOccurred())

			keyPairGenerator := ssl.NewKeyPairGenerator(rsa.GenerateKey, pkix.CreateCertificateAuthority, pkix.CreateCertificateSigningRequest, pkix.CreateCertificateHost)
			keyPair, err := keyPairGenerator.Generate("127.0.0.1", "127.0.0.1")
			Expect(err).NotTo(HaveOccurred())
			cert := keyPair.Certificate
			key := keyPair.PrivateKey

			certPath := filepath.Join(tempDirectory, "some-cert")
			err = ioutil.WriteFile(certPath, cert, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			keyPath := filepath.Join(tempDirectory, "some-key")
			err = ioutil.WriteFile(filepath.Join(tempDirectory, "some-key"), key, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"create-lbs",
				"--type", "cf",
				"--cert", certPath,
				"--key", keyPath,
			}

			executeCommand(args, 0)

			args = []string{
				"--state-dir", tempDirectory,
				"cloud-config",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(MatchYAML(string(contents)))
		})
	})
})
