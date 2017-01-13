package main_test

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/ssl"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/onsi/gomega/gexec"
	"github.com/square/certstrap/pkix"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("load balancers", func() {
	var (
		tempDirectory              string
		serviceAccountKeyPath      string
		pathToFakeTerraform        string
		pathToTerraform            string
		fakeTerraformBackendServer *httptest.Server
		fakeBOSHServer             *httptest.Server
		fakeBOSH                   *fakeBOSHDirector
		fastFailTerraform          bool
		fastFailTerraformMutex     sync.Mutex
	)

	var setFastFailTerraform = func(on bool) {
		fastFailTerraformMutex.Lock()
		defer fastFailTerraformMutex.Unlock()
		fastFailTerraform = on
	}

	var getFastFailTerraform = func() bool {
		fastFailTerraformMutex.Lock()
		defer fastFailTerraformMutex.Unlock()
		return fastFailTerraform
	}

	BeforeEach(func() {
		var err error
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
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
				responseWriter.Write([]byte("some-tag"))
			case "/output/bosh_open_tag_name":
				responseWriter.Write([]byte("some-bosh-open-tag"))
			case "/output/concourse_target_pool":
				responseWriter.Write([]byte("concourse-target-pool"))
			case "/fastfail":
				if getFastFailTerraform() {
					responseWriter.WriteHeader(http.StatusInternalServerError)
				}
			case "/output/router_backend_service":
				responseWriter.Write([]byte("router-backend-service"))
			case "/output/ssh_proxy_target_pool":
				responseWriter.Write([]byte("ssh-proxy-target-pool"))
			case "/output/tcp_router_target_pool":
				responseWriter.Write([]byte("tcp-router-target-pool"))
			case "/output/router_lb_ip":
				responseWriter.Write([]byte("some-router-lb-ip"))
			case "/output/ssh_proxy_lb_ip":
				responseWriter.Write([]byte("some-ssh-proxy-lb-ip"))
			case "/output/tcp_router_lb_ip":
				responseWriter.Write([]byte("some-tcp-router-lb-ip"))
			case "/output/concourse_lb_ip":
				responseWriter.Write([]byte("some-concourse-lb-ip"))
			}
		}))

		pathToFakeTerraform, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/faketerraform",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeTerraformBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToTerraform = filepath.Join(filepath.Dir(pathToFakeTerraform), "terraform")
		err = os.Rename(pathToFakeTerraform, pathToTerraform)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToTerraform), os.Getenv("PATH")}, ":"))

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		executeCommand([]string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "gcp",
			"--gcp-service-account-key", serviceAccountKeyPath,
			"--gcp-project-id", "some-project-id",
			"--gcp-zone", "us-east1-a",
			"--gcp-region", "us-east1",
		}, 0)
	})

	AfterEach(func() {
		setFastFailTerraform(false)
	})

	Describe("create-lbs", func() {
		It("creates and attaches a concourse lb type", func() {
			contents, err := ioutil.ReadFile("../cloudconfig/gcp/fixtures/cloud-config-concourse-lb.yml")
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"create-lbs",
				"--type", "concourse",
			}

			executeCommand(args, 0)

			Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))

			state := readStateJson(tempDirectory)
			Expect(state.LB).NotTo(BeNil())
			Expect(state.LB.Type).To(Equal("concourse"))
			Expect(state.LB.Cert).To(Equal(""))
			Expect(state.LB.Key).To(Equal(""))
		})

		Context("cf lb", func() {
			var certPath, keyPath string
			var contents []byte

			BeforeEach(func() {
				certPath = filepath.Join(tempDirectory, "some-cert")
				err := ioutil.WriteFile(certPath, []byte("cert-contents"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				keyPath = filepath.Join(tempDirectory, "some-key")
				err = ioutil.WriteFile(filepath.Join(tempDirectory, "some-key"), []byte("key-contents"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				contents, err = ioutil.ReadFile("../cloudconfig/gcp/fixtures/cloud-config-cf-lb.yml")
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates and attaches a cf lb type and ns when domain is provided", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"create-lbs",
					"--type", "cf",
					"--cert", certPath,
					"--key", keyPath,
					"--domain", "cf.example.com",
				}

				executeCommand(args, 0)

				Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))

				state := readStateJson(tempDirectory)
				Expect(state.LB).NotTo(BeNil())
				Expect(state.LB.Type).To(Equal("cf"))
				Expect(state.LB.Cert).To(Equal("cert-contents"))
				Expect(state.LB.Key).To(Equal("key-contents"))
				Expect(state.LB.Domain).To(Equal("cf.example.com"))
			})

			It("creates and attaches only a cf lb type when domain is not provided", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"create-lbs",
					"--type", "cf",
					"--cert", certPath,
					"--key", keyPath,
				}

				executeCommand(args, 0)

				Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))

				state := readStateJson(tempDirectory)
				Expect(state.LB).NotTo(BeNil())
				Expect(state.LB.Type).To(Equal("cf"))
				Expect(state.LB.Cert).To(Equal("cert-contents"))
				Expect(state.LB.Key).To(Equal("key-contents"))
				Expect(state.LB.Domain).To(Equal(""))
			})
		})

		It("logs all the steps", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"create-lbs",
				"--type", "concourse",
			}

			session := executeCommand(args, 0)
			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring("step: generating terraform template"))
			Expect(stdout).To(ContainSubstring("step: finished applying terraform template"))
			Expect(stdout).To(ContainSubstring("step: generating cloud config"))
			Expect(stdout).To(ContainSubstring("step: applying cloud config"))
		})

		It("no-ops if --skip-if-exists is provided and an lb exists", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"create-lbs",
				"--type", "concourse",
			}
			executeCommand(args, 0)

			args = []string{
				"--state-dir", tempDirectory,
				"create-lbs",
				"--type", "concourse",
				"--skip-if-exists",
			}
			session := executeCommand(args, 0)
			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(`lb type "concourse" exists, skipping...`))
		})
	})

	Describe("update-lbs", func() {
		var (
			certPath    string
			keyPath     string
			newCertPath string
			newKeyPath  string
			cert        []byte
			key         []byte
			newCert     []byte
			newKey      []byte
		)

		BeforeEach(func() {
			keyPairGenerator := ssl.NewKeyPairGenerator(rsa.GenerateKey, pkix.CreateCertificateAuthority, pkix.CreateCertificateSigningRequest, pkix.CreateCertificateHost)
			keyPair, err := keyPairGenerator.Generate("127.0.0.1", "127.0.0.1")
			Expect(err).NotTo(HaveOccurred())
			cert = keyPair.Certificate
			key = keyPair.PrivateKey

			newKeyPair, err := keyPairGenerator.Generate("127.0.0.2", "127.0.0.2")
			Expect(err).NotTo(HaveOccurred())
			newCert = newKeyPair.Certificate
			newKey = newKeyPair.PrivateKey

			certPath = filepath.Join(tempDirectory, "some-cert")
			err = ioutil.WriteFile(certPath, cert, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			keyPath = filepath.Join(tempDirectory, "some-key")
			err = ioutil.WriteFile(filepath.Join(tempDirectory, "some-key"), key, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			newCertPath = filepath.Join(tempDirectory, "some-new-cert")
			err = ioutil.WriteFile(newCertPath, newCert, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			newKeyPath = filepath.Join(tempDirectory, "some-new-key")
			err = ioutil.WriteFile(newKeyPath, newKey, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a cf lb exists", func() {
			BeforeEach(func() {
				args := []string{
					"--state-dir", tempDirectory,
					"create-lbs",
					"--type", "cf",
					"--cert", certPath,
					"--key", keyPath,
				}

				executeCommand(args, 0)
			})

			It("updates the load balancer with the given cert and key", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"update-lbs",
					"--cert", newCertPath,
					"--key", newKeyPath,
				}

				executeCommand(args, 0)

				state := readStateJson(tempDirectory)
				Expect(state.LB.Cert).To(Equal(string(newCert)))
				Expect(state.LB.Key).To(Equal(string(newKey)))
			})

			It("does nothing if the certificate is unchanged", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"update-lbs",
					"--cert", certPath,
					"--key", keyPath,
				}

				executeCommand(args, 0)

				state := readStateJson(tempDirectory)
				Expect(state.LB.Cert).To(Equal(string(cert)))
				Expect(state.LB.Key).To(Equal(string(key)))
			})
		})

		It("no-ops if --skip-if-missing is provided and an lb does not exist", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"update-lbs",
				"--cert", certPath,
				"--key", keyPath,
				"--skip-if-missing",
			}

			session := executeCommand(args, 0)
			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(`no lb type exists, skipping...`))
		})

		Context("failure cases", func() {
			Context("when an lb type does not exist", func() {
				It("exits 1", func() {
					args := []string{
						"--state-dir", tempDirectory,
						"update-lbs",
						"--cert", certPath,
						"--key", keyPath,
					}

					session := executeCommand(args, 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring("no load balancer has been found for this bbl environment"))
				})
			})

			Context("when bbl environment is not up", func() {
				It("exits 1 when the BOSH director does not exist", func() {
					writeStateJson(storage.State{
						IAAS: "gcp",
						GCP: storage.GCP{
							ProjectID:         "some-project-id",
							ServiceAccountKey: "some-service-account-key",
							Region:            "some-region",
							Zone:              "some-zone",
						},
						BOSH: storage.BOSH{
							DirectorAddress: "127.2.5.4",
						},
						LB: storage.LB{
							Type: "cf",
						},
					}, tempDirectory)

					args := []string{
						"--state-dir", tempDirectory,
						"update-lbs",
						"--cert", certPath,
						"--key", keyPath,
					}

					session := executeCommand(args, 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})
			})

			Context("when bbl-state.json does not exist", func() {
				It("exits with status 1 and outputs helpful error message", func() {
					tempDirectory, err := ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())

					args := []string{
						"--state-dir", tempDirectory,
						"update-lbs",
					}

					session := executeCommand(args, 1)
					Expect(session.Err.Contents()).To(ContainSubstring(fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)))
				})
			})
		})
	})

	Describe("delete-lbs", func() {
		It("deletes lbs", func() {
			var session *gexec.Session
			var stdout []byte

			By("running create-lbs", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"create-lbs",
					"--type", "concourse",
				}

				session = executeCommand(args, 0)
			})

			By("running delete-lbs", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"delete-lbs",
				}

				session := executeCommand(args, 0)
				stdout = session.Out.Contents()
			})

			By("logging the steps", func() {
				Expect(stdout).To(ContainSubstring("step: generating terraform template"))
				Expect(stdout).To(ContainSubstring("step: finished applying terraform template"))
				Expect(stdout).To(ContainSubstring("step: generating cloud config"))
				Expect(stdout).To(ContainSubstring("step: applying cloud config"))
			})

			By("removing the lb vm_extention from cloud config", func() {
				contents, err := ioutil.ReadFile("../cloudconfig/gcp/fixtures/cloud-config-no-lb.yml")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))
			})
		})

		It("the command is re-entrant when terraform apply fails", func() {
			var (
				session *gexec.Session
				stdout  []byte
			)
			By("running create-lbs", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"create-lbs",
					"--type", "concourse",
				}

				session = executeCommand(args, 0)
			})

			By("running delete-lbs and it fails", func() {
				setFastFailTerraform(true)
				args := []string{
					"--state-dir", tempDirectory,
					"delete-lbs",
				}

				session := executeCommand(args, 1)
				stderr := session.Err.Contents()

				Expect(stderr).To(ContainSubstring("failed to terraform"))
			})

			By("running delete-lbs and it succeeds", func() {
				args := []string{
					"--state-dir", tempDirectory,
					"delete-lbs",
				}

				session := executeCommand(args, 1)
				stdout = session.Out.Contents()
			})
		})
	})

	Describe("lbs", func() {
		Context("when cf lb was created", func() {
			BeforeEach(func() {
				certPath := filepath.Join(tempDirectory, "some-cert")
				err := ioutil.WriteFile(certPath, []byte("cert-contents"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				keyPath := filepath.Join(tempDirectory, "some-key")
				err = ioutil.WriteFile(filepath.Join(tempDirectory, "some-key"), []byte("key-contents"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				args := []string{
					"--state-dir", tempDirectory,
					"create-lbs",
					"--type", "cf",
					"--cert", certPath,
					"--key", keyPath,
				}

				executeCommand(args, 0)
			})

			It("prints out the currently attached lb names and urls", func() {
				session := lbs("", tempDirectory, 0)
				stdout := session.Out.Contents()

				Expect(stdout).To(ContainSubstring("CF Router LB: some-router-lb-ip"))
				Expect(stdout).To(ContainSubstring("CF SSH Proxy LB: some-ssh-proxy-lb-ip"))
				Expect(stdout).To(ContainSubstring("CF TCP Router LB: some-tcp-router-lb-ip"))
			})
		})

		Context("when concourse lb was created", func() {
			BeforeEach(func() {
				args := []string{
					"--state-dir", tempDirectory,
					"create-lbs",
					"--type", "concourse",
				}

				executeCommand(args, 0)
			})

			It("prints out the currently attached lb names and urls", func() {
				session := lbs("", tempDirectory, 0)
				stdout := session.Out.Contents()

				Expect(stdout).To(ContainSubstring("Concourse LB: some-concourse-lb-ip"))
			})
		})
	})
})
