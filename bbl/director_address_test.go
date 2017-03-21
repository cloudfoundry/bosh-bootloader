package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/rosenhouse/awsfaker"
)

var _ = Describe("director-address", func() {
	var (
		tempDirectory string
		args          []string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		args = []string{
			"--state-dir", tempDirectory,
			"director-address",
		}
	})

	Context("when bbl manages the director", func() {
		BeforeEach(func() {
			state := []byte(`{
				"version": 3,
				"bosh": {
					"directorAddress": "some-director-url"
				},
				"tfState": "some-tf-state"
			}`)
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the director address from the given state file", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("some-director-url"))
		})
	})

	Context("when bbl does not manage the director", func() {
		Context("gcp", func() {
			BeforeEach(func() {
				var err error
				fakeTerraformBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					switch request.URL.Path {
					case "/output/external_ip":
						responseWriter.Write([]byte("some-external-ip"))
					}
				}))

				state := []byte(`{
					"version":3,
					"iaas": "gcp",
					"noDirector": true,
					"tfState": "some-tf-state"
				}`)
				err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the external ip reserved for the director", func() {
				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("https://some-external-ip:25555"))
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

			It("returns the eip reserved for the director", func() {
				args = []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"director-address",
				}
				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("https://127.0.0.1:25555"))
			})

		})
	})

	Context("failure cases", func() {
		It("returns a non zero exit code when the bbl-state.json does not exist", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))

			expectedErrorMessage := fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)
			Expect(session.Err.Contents()).To(ContainSubstring(expectedErrorMessage))
		})

		It("returns a non zero exit code when the address does not exist", func() {
			state := []byte(`{"version":3}`)
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err.Contents()).To(ContainSubstring("Could not retrieve director address"))
		})
	})
})
