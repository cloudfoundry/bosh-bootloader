package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/rosenhouse/awsfaker"
)

var _ = Describe("bbl", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("prints the bosh environment variables when print-env is called to be sourced", func() {
		state := []byte(`{
			"version": 3,
			"bosh": {
				"directorAddress": "some-director-address",
				"directorUsername": "some-director-username",
				"directorPassword": "some-director-password",
				"directorSSLCA": "some-director-ca-cert"
			},
			"envID": "some-env-id"
		}`)
		err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		args := []string{
			"--state-dir", tempDirectory,
			"print-env",
		}

		session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_CLIENT=some-director-username"))
		Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_CLIENT_SECRET=some-director-password"))
		Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_CA_CERT='some-director-ca-cert'"))
		Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_ENVIRONMENT=some-director-address"))
	})

	Context("when the bosh environment has no director", func() {

		Context("gcp", func() {
			var (
				pathToFakeTerraform        string
				pathToTerraform            string
				fakeTerraformBackendServer *httptest.Server
			)

			BeforeEach(func() {
				var err error
				fakeTerraformBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					switch request.URL.Path {
					case "/output/external_ip":
						responseWriter.Write([]byte("some-external-ip"))
					}
				}))

				pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
					"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
				Expect(err).NotTo(HaveOccurred())

				pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
				err = os.Rename(pathToFakeTerraform, pathToTerraform)
				Expect(err).NotTo(HaveOccurred())

				os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), originalPath}, ":"))

				state := []byte(`{"version":3,"iaas": "gcp", "noDirector": true}`)
				err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.Setenv("PATH", originalPath)
			})

			It("prints only the external ip", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"print-env",
				}

				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_ENVIRONMENT=https://some-external-ip:25555"))
				Expect(session.Out.Contents()).NotTo(ContainSubstring("export BOSH_CLIENT="))
				Expect(session.Out.Contents()).NotTo(ContainSubstring("export BOSH_CLIENT_SECRET="))
				Expect(session.Out.Contents()).NotTo(ContainSubstring("export BOSH_CA_CERT="))
			})
		})

		Context("aws", func() {
			var (
				fakeAWS       *awsbackend.Backend
				fakeAWSServer *httptest.Server

				fakeBOSHServer *httptest.Server
				fakeBOSH       *fakeBOSHDirector
			)

			BeforeEach(func() {
				fakeBOSH = &fakeBOSHDirector{}
				fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					fakeBOSH.ServeHTTP(responseWriter, request)
				}))

				fakeAWS = awsbackend.New(fakeBOSHServer.URL)
				fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

				upArgs := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"--debug",
					"up",
					"--no-director",
					"--iaas", "aws",
					"--aws-access-key-id", "some-access-key",
					"--aws-secret-access-key", "some-access-secret",
					"--aws-region", "some-region",
				}

				executeCommand(upArgs, 0)
			})

			It("prints only the external ip", func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"print-env",
				}

				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("export BOSH_ENVIRONMENT=https://127.0.0.1:25555"))
				Expect(session.Out.Contents()).NotTo(ContainSubstring("export BOSH_CLIENT="))
				Expect(session.Out.Contents()).NotTo(ContainSubstring("export BOSH_CLIENT_SECRET="))
				Expect(session.Out.Contents()).NotTo(ContainSubstring("export BOSH_CA_CERT="))
			})

		})

	})
})
