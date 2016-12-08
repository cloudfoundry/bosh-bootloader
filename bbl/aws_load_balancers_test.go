package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	"github.com/onsi/gomega/gexec"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("load balancers", func() {
	var (
		fakeAWS          *awsbackend.Backend
		fakeAWSServer    *httptest.Server
		fakeBOSHServer   *httptest.Server
		fakeBOSH         *fakeBOSHDirector
		tempDirectory    string
		lbCertPath       string
		lbChainPath      string
		lbKeyPath        string
		otherLBCertPath  string
		otherLBChainPath string
		otherLBKeyPath   string
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

		upAWS(fakeAWSServer.URL, tempDirectory, 0)

		lbCertPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		lbChainPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		lbKeyPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		otherLBCertPath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		otherLBChainPath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		otherLBKeyPath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_KEY)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("create-lbs", func() {
		DescribeTable("creates lbs with the specified cert, key, and chain attached",
			func(lbType, fixtureLocation string) {
				contents, err := ioutil.ReadFile(fixtureLocation)
				Expect(err).NotTo(HaveOccurred())

				createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, lbType, 0, false)

				certificates := fakeAWS.Certificates.All()
				Expect(certificates).To(HaveLen(1))
				Expect(certificates[0].CertificateBody).To(Equal(testhelpers.BBL_CERT))
				Expect(certificates[0].PrivateKey).To(Equal(testhelpers.BBL_KEY))
				Expect(certificates[0].Chain).To(Equal(testhelpers.BBL_CHAIN))
				Expect(certificates[0].Name).To(MatchRegexp(fmt.Sprintf(`%s-elb-cert-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`, lbType)))

				Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))
			},
			Entry("attaches a cf lb type", "cf", "fixtures/cloud-config-cf-elb.yml"),
			Entry("attaches a concourse lb type", "concourse", "fixtures/cloud-config-concourse-elb.yml"),
		)

		It("logs all the steps", func() {
			session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "concourse", 0, false)
			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring("step: uploading certificate"))
			Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
			Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
			Expect(stdout).To(ContainSubstring("step: generating cloud config"))
			Expect(stdout).To(ContainSubstring("step: applying cloud config"))
		})

		It("no-ops if --skip-if-exists is provided and an lb exists", func() {
			createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 0, false)

			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(1))

			originalCertificate := certificates[0]

			session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 0, true)

			certificates = fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(1))

			Expect(certificates[0].Name).To(Equal(originalCertificate.Name))

			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(`lb type "cf" exists, skipping...`))
		})

		Context("failure cases", func() {
			Context("when an lb already exists", func() {
				BeforeEach(func() {
					createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "concourse", 0, false)
				})

				It("exits 1", func() {
					session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 1, false)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring("bbl already has a concourse load balancer attached, please remove the previous load balancer before attaching a new one"))
				})
			})

			It("exits 1 when an unknown lb-type is supplied", func() {
				session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "some-fake-lb-type", 1, false)
				stderr := session.Err.Contents()

				Expect(stderr).To(ContainSubstring("\"some-fake-lb-type\" is not a valid lb type, valid lb types are: concourse and cf"))
			})

			Context("when the environment has not been provisioned", func() {
				It("exits 1 when the cloudformation stack does not exist", func() {
					state := readStateJson(tempDirectory)

					fakeAWS.Stacks.Delete(state.Stack.Name)
					session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 1, false)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})

				It("exits 1 when the BOSH director does not exist", func() {
					writeStateJson(storage.State{
						IAAS: "aws",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key",
							SecretAccessKey: "some-access-secret",
							Region:          "some-region",
						},
						Stack: storage.Stack{
							Name: "some-stack-name",
						},
						BOSH: storage.BOSH{
							DirectorUsername: "admin",
							DirectorPassword: "admin",
							DirectorAddress:  "",
						},
					}, tempDirectory)

					session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 1, false)
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
						"create-lbs",
					}
					cmd := exec.Command(pathToBBL, args...)

					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session, 10*time.Second).Should(gexec.Exit(1))

					Expect(session.Err.Contents()).To(ContainSubstring(fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)))
				})
			})
		})
	})

	Describe("update-lbs", func() {
		It("updates the load balancer with the given cert, key and chain", func() {
			writeStateJson(storage.State{
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key",
					SecretAccessKey: "some-access-secret",
					Region:          "some-region",
				},
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

			updateLBs(fakeAWSServer.URL, tempDirectory, otherLBCertPath,
				otherLBKeyPath, otherLBChainPath, 0, false)

			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(1))
			Expect(certificates[0].Chain).To(Equal(testhelpers.OTHER_BBL_CHAIN))
			Expect(certificates[0].CertificateBody).To(Equal(testhelpers.OTHER_BBL_CERT))
			Expect(certificates[0].PrivateKey).To(Equal(testhelpers.OTHER_BBL_KEY))
			Expect(certificates[0].Name).To(MatchRegexp(`cf-elb-cert-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))

			stack, ok := fakeAWS.Stacks.Get("some-stack-name")
			Expect(ok).To(BeTrue())
			Expect(stack.WasUpdated).To(BeTrue())
		})

		It("does nothing if the certificate is unchanged", func() {
			writeStateJson(storage.State{
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key",
					SecretAccessKey: "some-access-secret",
					Region:          "some-region",
				},
				Stack: storage.Stack{
					Name:            "some-stack-name",
					LBType:          "cf",
					CertificateName: "bbl-cert-certificate",
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
				Name:            "bbl-cert-certificate",
				CertificateBody: testhelpers.BBL_CERT,
				PrivateKey:      testhelpers.BBL_KEY,
			})

			session := updateLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, "", 0, false)
			stdout := session.Out.Contents()

			Expect(stdout).To(ContainSubstring("no updates are to be performed"))

			stack, ok := fakeAWS.Stacks.Get("some-stack-name")
			Expect(ok).To(BeTrue())
			Expect(stack.WasUpdated).To(BeFalse())
		})

		It("logs all the steps", func() {
			createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "concourse", 0, false)
			session := updateLBs(fakeAWSServer.URL, tempDirectory, otherLBCertPath, otherLBKeyPath, "", 0, false)
			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring("step: uploading new certificate"))
			Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
			Expect(stdout).To(ContainSubstring("step: updating cloudformation stack"))
			Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
			Expect(stdout).To(ContainSubstring("step: deleting old certificate"))
		})

		It("no-ops if --skip-if-missing is provided and an lb does not exist", func() {
			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(0))

			session := updateLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, "", 0, true)

			certificates = fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(0))

			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(`no lb type exists, skipping...`))
		})

		Context("failure cases", func() {
			Context("when an lb type does not exist", func() {
				It("exits 1", func() {
					session := updateLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, "", 1, false)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring("no load balancer has been found for this bbl environment"))
				})
			})

			Context("when bbl environment is not up", func() {
				It("exits 1 when the cloudformation stack does not exist", func() {
					writeStateJson(storage.State{
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key",
							SecretAccessKey: "some-access-secret",
							Region:          "some-region",
						},
					}, tempDirectory)
					session := updateLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, "", 1, false)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})

				It("exits 1 when the BOSH director does not exist", func() {
					fakeAWS.Stacks.Set(awsbackend.Stack{
						Name: "some-stack-name",
					})

					writeStateJson(storage.State{
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key",
							SecretAccessKey: "some-access-secret",
							Region:          "some-region",
						},
						Stack: storage.Stack{
							Name: "some-stack-name",
						},
					}, tempDirectory)

					session := updateLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, "", 1, false)
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
					cmd := exec.Command(pathToBBL, args...)

					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session, 10*time.Second).Should(gexec.Exit(1))

					Expect(session.Err.Contents()).To(ContainSubstring(fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)))
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
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key",
					SecretAccessKey: "some-access-secret",
					Region:          "some-region",
				},
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
				EnvID: "bbl-env-lake-timestamp",
			}, tempDirectory)

			fakeAWS.Stacks.Set(awsbackend.Stack{
				Name: "some-stack-name",
			})

			fakeAWS.Certificates.Set(awsbackend.Certificate{
				Name: "bbl-cert-old-certificate",
			})

			deleteLBs(fakeAWSServer.URL, tempDirectory, 0, false)

			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(0))

			stack, ok := fakeAWS.Stacks.Get("some-stack-name")
			Expect(ok).To(BeTrue())
			Expect(stack.WasUpdated).To(BeTrue())
			Expect(stack.Template).To(MatchJSON(string(cloudformationNoELB)))

			Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(cloudConfigFixture)))
		})

		It("logs all the steps", func() {
			createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "concourse", 0, false)
			session := deleteLBs(fakeAWSServer.URL, tempDirectory, 0, false)
			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring("step: generating cloud config"))
			Expect(stdout).To(ContainSubstring("step: applying cloud config"))
			Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
			Expect(stdout).To(ContainSubstring("step: updating cloudformation stack"))
			Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
			Expect(stdout).To(ContainSubstring("step: deleting certificate"))
		})

		It("no-ops if --skip-if-missing is provided and an lb does not exist", func() {
			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(0))

			session := deleteLBs(fakeAWSServer.URL, tempDirectory, 0, true)

			certificates = fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(0))

			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(`no lb type exists, skipping...`))
		})

		Context("failure cases", func() {
			Context("when the environment has not been provisioned", func() {
				It("exits 1 when the cloudformation stack does not exist", func() {
					state := readStateJson(tempDirectory)

					fakeAWS.Stacks.Delete(state.Stack.Name)
					session := deleteLBs(fakeAWSServer.URL, tempDirectory, 1, false)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
				})

				It("exits 1 when the BOSH director does not exist", func() {
					fakeAWS.Stacks.Set(awsbackend.Stack{
						Name: "some-stack-name",
					})

					writeStateJson(storage.State{
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key",
							SecretAccessKey: "some-access-secret",
							Region:          "some-region",
						},
						Stack: storage.Stack{
							Name: "some-stack-name",
						},
					}, tempDirectory)

					session := deleteLBs(fakeAWSServer.URL, tempDirectory, 1, false)
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
						"delete-lbs",
					}
					cmd := exec.Command(pathToBBL, args...)

					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session, 10*time.Second).Should(gexec.Exit(1))

					Expect(session.Err.Contents()).To(ContainSubstring(fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)))
				})
			})
		})
	})

	Describe("lbs", func() {
		It("prints out the currently attached lb names and urls", func() {
			createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 0, false)

			session := lbs(fakeAWSServer.URL, tempDirectory, 0)
			stdout := session.Out.Contents()

			Expect(stdout).To(ContainSubstring("CF Router LB: some-cf-router-lb [some-cf-router-lb-url]"))
			Expect(stdout).To(ContainSubstring("CF SSH Proxy LB: some-cf-ssh-proxy-lb [some-cf-ssh-proxy-lb-url]"))
		})

		Context("when bbl-state.json does not exist", func() {
			It("exits with status 1 and outputs helpful error message", func() {
				tempDirectory, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				args := []string{
					"--state-dir", tempDirectory,
					"lbs",
				}
				cmd := exec.Command(pathToBBL, args...)

				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(1))

				Expect(session.Err.Contents()).To(ContainSubstring(fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)))
			})
		})
	})
})

func lbs(endpointOverrideURL string, stateDir string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--state-dir", stateDir,
		"lbs",
	}

	return executeCommand(args, exitCode)
}

func deleteLBs(endpointOverrideURL string, stateDir string, exitCode int, skipIfMissing bool) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--state-dir", stateDir,
		"delete-lbs",
	}

	if skipIfMissing {
		args = append(args, "--skip-if-missing")
	}

	return executeCommand(args, exitCode)
}

func updateLBs(endpointOverrideURL string, stateDir string, certName string, keyName string, chainName string, exitCode int, skipIfMissing bool) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--state-dir", stateDir,
		"update-lbs",
		"--cert", certName,
		"--key", keyName,
		"--chain", chainName,
	}

	if skipIfMissing {
		args = append(args, "--skip-if-missing")
	}

	return executeCommand(args, exitCode)
}

func createLBs(endpointOverrideURL string, stateDir string, certName string, keyName string, chainName string, lbType string, exitCode int, skipIfExists bool) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--state-dir", stateDir,
		"create-lbs",
		"--type", lbType,
		"--cert", certName,
		"--key", keyName,
		"--chain", chainName,
	}

	if skipIfExists {
		args = append(args, "--skip-if-exists")
	}

	return executeCommand(args, exitCode)
}
