package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bbl/awsbackend"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("load balancers", func() {
	var (
		fakeAWS        *awsbackend.Backend
		fakeAWSServer  *httptest.Server
		fakeBOSHServer *httptest.Server
		fakeBOSH       *fakeBOSHDirector
		lbCert         []byte
		lbKey          []byte
		lbChain        []byte
		tempDirectory  string
	)

	BeforeEach(func() {
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeAWS = awsbackend.New(fakeBOSHServer.URL)
		fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

		lbCert, err = ioutil.ReadFile("fixtures/lb-cert.pem")
		Expect(err).NotTo(HaveOccurred())

		lbKey, err = ioutil.ReadFile("fixtures/lb-key.pem")
		Expect(err).NotTo(HaveOccurred())

		lbChain, err = ioutil.ReadFile("fixtures/lb-chain.pem")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("create-lbs", func() {
		DescribeTable("creates lbs with the specified cert, key, and chain attached",
			func(lbType, fixtureLocation string) {
				contents, err := ioutil.ReadFile(fixtureLocation)
				Expect(err).NotTo(HaveOccurred())

				session := createLBs(fakeAWSServer.URL, tempDirectory, lbType, 0)

				certificates := fakeAWS.Certificates.All()
				Expect(certificates).To(HaveLen(1))
				Expect(certificates[0].CertificateBody).To(Equal(string(lbCert)))
				Expect(certificates[0].PrivateKey).To(Equal(string(lbKey)))
				Expect(certificates[0].Chain).To(Equal(string(lbChain)))
				Expect(certificates[0].Name).To(MatchRegexp(`bbl-cert-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))

				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: generating cloud config"))
				Expect(stdout).To(ContainSubstring("step: applying cloud config"))

				Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))
			},
			Entry("it attaches a cf lb type", "cf", "fixtures/cloud-config-cf-elb.yml"),
			Entry("it attaches a concourse lb type", "concourse", "fixtures/cloud-config-concourse-elb.yml"),
		)

		It("no-ops if --skip-if-exists is provided and an lb exists", func() {
			createLBs(fakeAWSServer.URL, tempDirectory, "cf", 0)

			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(1))

			originalCertificate := certificates[0]

			session := createLBsSkipIfExists(fakeAWSServer.URL, tempDirectory, "cf", 0)

			certificates = fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(1))

			Expect(certificates[0].Name).To(Equal(originalCertificate.Name))

			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(`lb type "cf" exists, skipping...`))
		})

		Context("failure cases", func() {
			Context("when an lb already exists", func() {
				BeforeEach(func() {
					createLBs(fakeAWSServer.URL, tempDirectory, "concourse", 0)
				})

				It("exits 1", func() {
					session := createLBs(fakeAWSServer.URL, tempDirectory, "cf", 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring("bbl already has a concourse load balancer attached, please remove the previous load balancer before attaching a new one"))
				})
			})

			It("exits 1 when an unknown lb-type is supplied", func() {
				session := createLBs(fakeAWSServer.URL, tempDirectory, "some-fake-lb-type", 1)
				stderr := session.Err.Contents()

				Expect(stderr).To(ContainSubstring("\"some-fake-lb-type\" is not a valid lb type, valid lb types are: concourse and cf"))
			})

			Context("when the environment has not been provisioned", func() {
				It("exits 1 when the cloudformation stack does not exist", func() {
					state := readStateJson(tempDirectory)

					fakeAWS.Stacks.Delete(state.Stack.Name)
					session := createLBs(fakeAWSServer.URL, tempDirectory, "cf", 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})

				It("exits 1 when the BOSH director does not exist", func() {
					writeStateJson(storage.State{
						Stack: storage.Stack{
							Name: "some-stack-name",
						},
						BOSH: storage.BOSH{
							DirectorUsername: "admin",
							DirectorPassword: "admin",
							DirectorAddress:  "",
						},
					}, tempDirectory)

					session := createLBs(fakeAWSServer.URL, tempDirectory, "cf", 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})
			})
		})
	})

	Describe("update-lbs", func() {
		It("updates the load balancer with the given key and cert", func() {
			writeStateJson(storage.State{
				Stack: storage.Stack{
					Name:            "some-stack-name",
					LBType:          "cf",
					CertificateName: "bbl-cert-old-certificate",
				},
				BOSH: storage.BOSH{
					DirectorUsername: "admin",
					DirectorPassword: "admin",
					DirectorAddress:  fakeBOSHServer.URL,
				},
			}, tempDirectory)

			fakeAWS.Stacks.Set(awsbackend.Stack{
				Name: "some-stack-name",
			})

			fakeAWS.Certificates.Set(awsbackend.Certificate{
				Name:            "bbl-cert-old-certificate",
				CertificateBody: "some-old-certificate-body",
				PrivateKey:      "some-old-private-key",
			})

			updateLBs(fakeAWSServer.URL, tempDirectory, temporaryFileContaining("some-new-certificate-body"), temporaryFileContaining("some-new-private-key"), 0)

			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(1))
			Expect(certificates[0].CertificateBody).To(Equal("some-new-certificate-body"))
			Expect(certificates[0].PrivateKey).To(Equal("some-new-private-key"))
			Expect(certificates[0].Name).To(MatchRegexp(`bbl-cert-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))

			stack, ok := fakeAWS.Stacks.Get("some-stack-name")
			Expect(ok).To(BeTrue())
			Expect(stack.WasUpdated).To(BeTrue())
		})

		Context("failure cases", func() {
			Context("when an lb type does not exist", func() {
				It("exits 1", func() {
					session := updateLBs(fakeAWSServer.URL, tempDirectory, temporaryFileContaining("some-new-certificate-body"), temporaryFileContaining("some-new-private-key"), 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring("no load balancer has been found for this bbl environment"))
				})
			})

			Context("when bbl environment is not up", func() {
				It("exits 1 when the cloudformation stack does not exist", func() {
					writeStateJson(storage.State{}, tempDirectory)
					session := updateLBs(fakeAWSServer.URL, tempDirectory, temporaryFileContaining("some-new-certificate-body"), temporaryFileContaining("some-new-private-key"), 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})

				It("exits 1 when the BOSH director does not exist", func() {
					fakeAWS.Stacks.Set(awsbackend.Stack{
						Name: "some-stack-name",
					})

					writeStateJson(storage.State{
						Stack: storage.Stack{
							Name: "some-stack-name",
						},
					}, tempDirectory)

					session := updateLBs(fakeAWSServer.URL, tempDirectory, temporaryFileContaining("some-new-certificate-body"), temporaryFileContaining("some-new-private-key"), 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})
			})
		})
	})

	Describe("delete-lbs", func() {
		It("deletes the load balancer", func() {
			cloudformationNoELB, err := ioutil.ReadFile("fixtures/cloudformation-no-elb.json")
			Expect(err).NotTo(HaveOccurred())

			cloudConfigFixture, err := ioutil.ReadFile("fixtures/cloud-config-no-elb.yml")
			Expect(err).NotTo(HaveOccurred())

			writeStateJson(storage.State{
				Stack: storage.Stack{
					Name:            "some-stack-name",
					LBType:          "cf",
					CertificateName: "bbl-cert-old-certificate",
				},
				BOSH: storage.BOSH{
					DirectorUsername: "admin",
					DirectorPassword: "admin",
					DirectorAddress:  fakeBOSHServer.URL,
				},
				KeyPair: storage.KeyPair{
					Name: "some-keypair-name",
				},
			}, tempDirectory)

			fakeAWS.Stacks.Set(awsbackend.Stack{
				Name: "some-stack-name",
			})

			fakeAWS.Certificates.Set(awsbackend.Certificate{
				Name: "bbl-cert-old-certificate",
			})

			deleteLBs(fakeAWSServer.URL, tempDirectory, 0)

			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(0))

			stack, ok := fakeAWS.Stacks.Get("some-stack-name")
			Expect(ok).To(BeTrue())
			Expect(stack.WasUpdated).To(BeTrue())
			Expect(stack.Template).To(MatchJSON(string(cloudformationNoELB)))

			Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(cloudConfigFixture)))
		})

		Context("failure cases", func() {
			Context("when the environment has not been provisioned", func() {
				It("exits 1 when the cloudformation stack does not exist", func() {
					state := readStateJson(tempDirectory)

					fakeAWS.Stacks.Delete(state.Stack.Name)
					session := deleteLBs(fakeAWSServer.URL, tempDirectory, 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})

				It("exits 1 when the BOSH director does not exist", func() {
					fakeAWS.Stacks.Set(awsbackend.Stack{
						Name: "some-stack-name",
					})

					writeStateJson(storage.State{
						Stack: storage.Stack{
							Name: "some-stack-name",
						},
					}, tempDirectory)

					session := deleteLBs(fakeAWSServer.URL, tempDirectory, 1)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})
			})
		})
	})

	Describe("lbs", func() {
		It("prints out the currently attached lb names and urls", func() {
			createLBs(fakeAWSServer.URL, tempDirectory, "cf", 0)

			session := lbs(fakeAWSServer.URL, tempDirectory, 0)
			stdout := session.Out.Contents()

			Expect(stdout).To(ContainSubstring("CF Router LB: some-cf-router-lb [some-cf-router-lb-url]"))
			Expect(stdout).To(ContainSubstring("CF SSH Proxy LB: some-cf-ssh-proxy-lb [some-cf-ssh-proxy-lb-url]"))
		})
	})
})

func lbs(endpointOverrideURL string, stateDir string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--aws-access-key-id", "some-access-key-id",
		"--aws-secret-access-key", "some-secret-access-key",
		"--aws-region", "some-region",
		"--state-dir", stateDir,
		"unsupported-lbs",
	}

	return executeCommand(args, exitCode)
}

func deleteLBs(endpointOverrideURL string, stateDir string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--aws-access-key-id", "some-access-key-id",
		"--aws-secret-access-key", "some-secret-access-key",
		"--aws-region", "some-region",
		"--state-dir", stateDir,
		"unsupported-delete-lbs",
	}

	return executeCommand(args, exitCode)
}

func updateLBs(endpointOverrideURL string, stateDir string, certName string, keyName string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--aws-access-key-id", "some-access-key-id",
		"--aws-secret-access-key", "some-secret-access-key",
		"--aws-region", "some-region",
		"--state-dir", stateDir,
		"unsupported-update-lbs",
		"--cert", certName,
		"--key", keyName,
	}

	return executeCommand(args, exitCode)
}

func createLBsSkipIfExists(endpointOverrideURL string, stateDir string, lbType string, exitCode int) *gexec.Session {
	dir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--aws-access-key-id", "some-access-key-id",
		"--aws-secret-access-key", "some-secret-access-key",
		"--aws-region", "some-region",
		"--state-dir", stateDir,
		"unsupported-create-lbs",
		"--skip-if-exists",
		"--type", lbType,
		"--cert", filepath.Join(dir, "fixtures", "lb-cert.pem"),
		"--key", filepath.Join(dir, "fixtures", "lb-key.pem"),
		"--chain", filepath.Join(dir, "fixtures", "lb-chain.pem"),
	}

	return executeCommand(args, exitCode)
}

func createLBs(endpointOverrideURL string, stateDir string, lbType string, exitCode int) *gexec.Session {
	dir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--aws-access-key-id", "some-access-key-id",
		"--aws-secret-access-key", "some-secret-access-key",
		"--aws-region", "some-region",
		"--state-dir", stateDir,
		"unsupported-create-lbs",
		"--type", lbType,
		"--cert", filepath.Join(dir, "fixtures", "lb-cert.pem"),
		"--key", filepath.Join(dir, "fixtures", "lb-key.pem"),
		"--chain", filepath.Join(dir, "fixtures", "lb-chain.pem"),
	}

	return executeCommand(args, exitCode)
}

func temporaryFileContaining(fileContents string) string {
	temporaryFile, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())

	err = ioutil.WriteFile(temporaryFile.Name(), []byte(fileContents), os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	return temporaryFile.Name()
}
