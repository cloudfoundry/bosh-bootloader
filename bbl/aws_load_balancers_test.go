package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"sync"
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

		callRealInterpolate      bool
		callRealInterpolateMutex sync.Mutex
	)

	BeforeEach(func() {
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeBOSHCLIBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/version":
				responseWriter.Write([]byte("2.0.0"))
			case "/path":
				responseWriter.Write([]byte(noFakesPath))
			case "/call-real-interpolate":
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				if callRealInterpolate {
					responseWriter.Write([]byte("true"))
				} else {
					responseWriter.Write([]byte("false"))
				}
			}
		}))

		fakeAWS = awsbackend.New(fakeBOSHServer.URL)
		fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

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
		var createLBsTests = func(terraform bool) {
			DescribeTable("creates lbs with the specified cert, key, and chain attached",
				func(lbType, fixtureLocation string) {
					contents, err := ioutil.ReadFile(fixtureLocation)
					Expect(err).NotTo(HaveOccurred())

					createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, lbType, 0, false)

					if !terraform {
						checkCertificatesForCloudFormation(fakeAWS, lbType)
					}

					Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))
				},
				Entry("attaches a cf lb type", "cf", "fixtures/cloud-config-cf-elb.yml"),
				Entry("attaches a concourse lb type", "concourse", "fixtures/cloud-config-concourse-elb.yml"),
			)

			It("logs all the steps", func() {
				session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "concourse", 0, false)
				stdout := session.Out.Contents()
				if !terraform {
					Expect(stdout).To(ContainSubstring("step: uploading certificate"))
					Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
					Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
				}
				Expect(stdout).To(ContainSubstring("step: generating cloud config"))
				Expect(stdout).To(ContainSubstring("step: applying cloud config"))
			})

			if !terraform {
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
			}

			Context("failure cases", func() {
				Context("when the bosh cli version is <2.0", func() {
					BeforeEach(func() {
						fakeBOSHCLIBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
							switch request.URL.Path {
							case "/version":
								responseWriter.Write([]byte("1.9.0"))
							}
						}))
					})

					It("fast fails with a helpful error message", func() {
						session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 1, false)
						Expect(session.Err.Contents()).To(ContainSubstring("BOSH version must be at least v2.0.0"))
					})
				})

				if !terraform {
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
				}

				It("exits 1 when an unknown lb-type is supplied", func() {
					session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "some-fake-lb-type", 1, false)
					stderr := session.Err.Contents()

					Expect(stderr).To(ContainSubstring("\"some-fake-lb-type\" is not a valid lb type, valid lb types are: concourse and cf"))
				})

				Context("when the environment has not been provisioned", func() {
					if terraform {
						It("exits 1 when the terraform state does not exist", func() {
							state := readStateJson(tempDirectory)
							state.TFState = ""
							writeStateJson(state, tempDirectory)

							session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 1, false)
							stderr := session.Err.Contents()

							Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
						})
					} else {
						It("exits 1 when the cloudformation stack does not exist", func() {
							state := readStateJson(tempDirectory)

							fakeAWS.Stacks.Delete(state.Stack.Name)
							session := createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 1, false)
							stderr := session.Err.Contents()

							Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
						})
					}

					It("exits 1 when the BOSH director does not exist", func() {
						writeStateJson(storage.State{
							Version: 3,
							IAAS:    "aws",
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
		}

		Context("when bbl'd up with cloudformation", func() {
			BeforeEach(func() {
				upAWS(fakeAWSServer.URL, tempDirectory, 0)

				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = true
			})

			AfterEach(func() {
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = false
			})

			createLBsTests(false)
		})

		Context("when bbl'd up with terraform", func() {
			BeforeEach(func() {
				fakeTerraformBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					switch request.URL.Path {
					case "/output/--json":
						responseWriter.Write([]byte(fmt.Sprintf(`{
							"bosh_eip": {
								"value": "some-bosh-eip"
							},
							"bosh_url": {
								"value": %q
							},
							"bosh_user_access_key": {
								"value": "some-bosh-user-access-key"
							},
							"bosh_user_secret_access_key": {
								"value": "some-bosh-user-secret-access_key"
							},
							"nat_eip": {
								"value": "some-nat-eip"
							},
							"bosh_subnet_id": {
								"value": "some-bosh-subnet-id"
							},
							"bosh_subnet_availability_zone": {
								"value": "some-bosh-subnet-availability-zone"
							},
							"bosh_security_group": {
								"value": "some-bosh-security-group"
							},
							"env_dns_zone_name_servers": {
								"value": [
									"name-server-1.",
									"name-server-2."
								]
							},
							"internal_security_group": {
								"value": "some-internal-security-group"
							},
							"internal_subnet_ids": {
								"value": [
									"some-internal-subnet-ids-1",
									"some-internal-subnet-ids-2",
									"some-internal-subnet-ids-3"
								]
							},
							"internal_subnet_cidrs": {
								"value": [
									"10.0.16.0/20",
									"10.0.32.0/20",
									"10.0.48.0/20"
								]
							},
							"vpc_id": {
								"value": "some-vpc-id"
							},
							"cf_router_lb_name": {
								"value": "some-cf-router-lb"
							},
							"cf_router_lb_internal_security_group": {
								"value": "some-cf-router-internal-security-group"
							},
							"cf_ssh_lb_name":  {
								"value": "some-cf-ssh-proxy-lb"
							},
							"cf_ssh_lb_internal_security_group":  {
								"value": "some-cf-ssh-proxy-internal-security-group"
							},
							"concourse_lb_name":  {
								"value": "some-concourse-lb"
							},
							"concourse_lb_internal_security_group":  {
								"value": "some-concourse-internal-security-group"
							}
						}`, fakeBOSHServer.URL)))
					case "/version":
						responseWriter.Write([]byte("0.8.6"))
					}
				}))

				upAWSWithAdditionalFlags(fakeAWSServer.URL, tempDirectory, []string{"--terraform"}, 0)

				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = true
			})

			AfterEach(func() {
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = false
			})

			createLBsTests(true)

			Context("when domain is provided", func() {
				var (
					expectedCloudConfig []byte
				)

				BeforeEach(func() {
					var err error
					expectedCloudConfig, err = ioutil.ReadFile("fixtures/cloud-config-cf-elb.yml")
					Expect(err).NotTo(HaveOccurred())
				})

				It("creates and attaches a cf lb type and ns when domain is provided", func() {
					args := []string{
						fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
						"--state-dir", tempDirectory,
						"create-lbs",
						"--type", "cf",
						"--cert", lbCertPath,
						"--key", lbKeyPath,
						"--domain", "cf.example.com",
					}

					executeCommand(args, 0)

					Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(expectedCloudConfig)))

					state := readStateJson(tempDirectory)
					Expect(state.LB).NotTo(BeNil())
					Expect(state.LB.Type).To(Equal("cf"))
					Expect(state.LB.Cert).To(Equal(testhelpers.BBL_CERT))
					Expect(state.LB.Key).To(Equal(testhelpers.BBL_KEY))
					Expect(state.LB.Domain).To(Equal("cf.example.com"))
				})
			})

		})
	})

	Describe("update-lbs", func() {
		It("updates the load balancer with the given cert, key and chain", func() {
			upAWS(fakeAWSServer.URL, tempDirectory, 0)

			writeStateJson(storage.State{
				Version: 3,
				IAAS:    "aws",
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
			upAWS(fakeAWSServer.URL, tempDirectory, 0)

			writeStateJson(storage.State{
				Version: 3,
				IAAS:    "aws",
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
			upAWS(fakeAWSServer.URL, tempDirectory, 0)

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
			upAWS(fakeAWSServer.URL, tempDirectory, 0)

			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(0))

			session := updateLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, "", 0, true)

			certificates = fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(0))

			stdout := session.Out.Contents()
			Expect(stdout).To(ContainSubstring(`no lb type exists, skipping...`))
		})

		Context("failure cases", func() {
			BeforeEach(func() {
				upAWS(fakeAWSServer.URL, tempDirectory, 0)
			})

			Context("when the bosh cli version is < 2.0.0", func() {
				BeforeEach(func() {
					fakeBOSHCLIBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
						switch request.URL.Path {
						case "/version":
							responseWriter.Write([]byte("1.9.0"))
						}
					}))
				})

				It("fast fails with a helpful error message", func() {
					session := updateLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, "", 1, false)
					Expect(session.Err.Contents()).To(ContainSubstring("BOSH version must be at least v2.0.0"))
				})
			})

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
						Version: 3,
						IAAS:    "aws",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key",
							SecretAccessKey: "some-access-secret",
							Region:          "some-region",
						},
						Stack: storage.Stack{
							LBType: "concourse",
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
						Version: 3,
						IAAS:    "aws",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key",
							SecretAccessKey: "some-access-secret",
							Region:          "some-region",
						},
						Stack: storage.Stack{
							Name:   "some-stack-name",
							LBType: "concourse",
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
		var deleteLBsTests = func(terraform bool) {
			It("deletes the load balancer", func() {
				cloudformationNoELB, err := ioutil.ReadFile("fixtures/cloudformation-no-elb.json")
				Expect(err).NotTo(HaveOccurred())

				cloudConfigFixture, err := ioutil.ReadFile("fixtures/cloud-config-no-elb.yml")
				Expect(err).NotTo(HaveOccurred())

				state := readStateJson(tempDirectory)

				if !terraform {
					state.KeyPair.Name = "some-keypair-name"
					state.EnvID = "bbl-env-lake-timestamp"
					writeStateJson(state, tempDirectory)
					fakeAWS.Stacks.Set(awsbackend.Stack{
						Name: state.Stack.Name,
					})

					fakeAWS.Certificates.Set(awsbackend.Certificate{
						Name: state.Stack.CertificateName,
					})
				}

				deleteLBs(fakeAWSServer.URL, tempDirectory, 0, false)

				if !terraform {
					certificates := fakeAWS.Certificates.All()
					Expect(certificates).To(HaveLen(0))

					stack, ok := fakeAWS.Stacks.Get(state.Stack.Name)
					Expect(ok).To(BeTrue())
					Expect(stack.WasUpdated).To(BeTrue())
					Expect(stack.Template).To(MatchJSON(string(cloudformationNoELB)))
				}

				Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(cloudConfigFixture)))
			})

			It("logs all the steps", func() {
				session := deleteLBs(fakeAWSServer.URL, tempDirectory, 0, false)
				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: generating cloud config"))
				Expect(stdout).To(ContainSubstring("step: applying cloud config"))
				if terraform {
					Expect(stdout).To(ContainSubstring("step: generating terraform template"))
					Expect(stdout).To(ContainSubstring("step: applied terraform template"))
				} else {
					Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
					Expect(stdout).To(ContainSubstring("step: updating cloudformation stack"))
					Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
					Expect(stdout).To(ContainSubstring("step: deleting certificate"))
				}
			})

			It("no-ops if --skip-if-missing is provided and an lb does not exist", func() {
				if !terraform {
					certificates := fakeAWS.Certificates.All()
					Expect(certificates).To(HaveLen(1))
				}

				session := deleteLBs(fakeAWSServer.URL, tempDirectory, 0, true)
				stdout := session.Out.Contents()

				if terraform {
					Expect(stdout).To(ContainSubstring("step: generating terraform template"))
					Expect(stdout).To(ContainSubstring("step: applied terraform template"))
				} else {
					Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
					Expect(stdout).To(ContainSubstring("step: updating cloudformation stack"))
					Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
					Expect(stdout).To(ContainSubstring("step: deleting certificate"))
				}

				if !terraform {
					certificates := fakeAWS.Certificates.All()
					Expect(certificates).To(HaveLen(0))
				}

				session = deleteLBs(fakeAWSServer.URL, tempDirectory, 0, true)

				stdout = session.Out.Contents()
				Expect(stdout).To(ContainSubstring(`no lb type exists, skipping...`))

				if !terraform {
					certificates := fakeAWS.Certificates.All()
					Expect(certificates).To(HaveLen(0))
				}
			})

			Context("failure cases", func() {
				Context("when the environment has not been provisioned", func() {
					if terraform {
						It("exits 1 when the terraform state does not exist", func() {
							state := readStateJson(tempDirectory)
							state.TFState = ""
							writeStateJson(state, tempDirectory)

							session := deleteLBs(fakeAWSServer.URL, tempDirectory, 1, false)
							stderr := session.Err.Contents()

							Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
						})
					} else {
						It("exits 1 when the cloudformation stack does not exist", func() {
							state := readStateJson(tempDirectory)

							fakeAWS.Stacks.Delete(state.Stack.Name)
							session := deleteLBs(fakeAWSServer.URL, tempDirectory, 1, false)
							stderr := session.Err.Contents()

							Expect(stderr).To(ContainSubstring(commands.BBLNotFound.Error()))
						})
					}

					It("exits 1 when the BOSH director does not exist", func() {
						fakeAWS.Stacks.Set(awsbackend.Stack{
							Name: "some-stack-name",
						})

						writeStateJson(storage.State{
							Version: 3,
							IAAS:    "aws",
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
		}

		Context("when bbl'd up with cloudformation", func() {
			BeforeEach(func() {
				upAWS(fakeAWSServer.URL, tempDirectory, 0)
				createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 0, false)

				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = true
			})

			AfterEach(func() {
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = false
			})

			deleteLBsTests(false)
		})

		Context("when bbl'd up with terraform", func() {
			BeforeEach(func() {
				fakeTerraformBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					switch request.URL.Path {
					case "/output/--json":
						responseWriter.Write([]byte(fmt.Sprintf(`{
							"bosh_eip": {
								"value": "some-bosh-eip"
							},
							"bosh_url": {
								"value": %q
							},
							"bosh_user_access_key": {
								"value": "some-bosh-user-access-key"
							},
							"bosh_user_secret_access_key": {
								"value": "some-bosh-user-secret-access_key"
							},
							"nat_eip": {
								"value": "some-nat-eip"
							},
							"bosh_subnet_id": {
								"value": "some-bosh-subnet-id"
							},
							"bosh_subnet_availability_zone": {
								"value": "some-bosh-subnet-availability-zone"
							},
							"bosh_security_group": {
								"value": "some-bosh-security-group"
							},
							"internal_security_group": {
								"value": "some-internal-security-group"
							},
							"internal_subnet_ids": {
								"value": [
									"some-internal-subnet-ids-1",
									"some-internal-subnet-ids-2",
									"some-internal-subnet-ids-3"
								]
							},
							"internal_subnet_cidrs": {
								"value": [
									"10.0.16.0/20",
									"10.0.32.0/20",
									"10.0.48.0/20"
								]
							},
							"vpc_id": {
								"value": "some-vpc-id"
							},
							"cf_router_lb_name": {
								"value": "some-cf-router-lb"
							},
							"cf_router_lb_internal_security_group": {
								"value": "some-cf-router-internal-security-group"
							},
							"cf_ssh_lb_name":  {
								"value": "some-cf-ssh-proxy-lb"
							},
							"cf_ssh_lb_internal_security_group":  {
								"value": "some-cf-ssh-proxy-internal-security-group"
							},
							"concourse_lb_name":  {
								"value": "some-concourse-lb"
							},
							"concourse_lb_internal_security_group":  {
								"value": "some-concourse-internal-security-group"
							}
						}`, fakeBOSHServer.URL)))
					case "/version":
						responseWriter.Write([]byte("0.8.6"))
					}
				}))

				upAWSWithAdditionalFlags(fakeAWSServer.URL, tempDirectory, []string{"--terraform"}, 0)
				createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 0, false)

				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = true
			})

			AfterEach(func() {
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = false
			})

			deleteLBsTests(true)
		})
	})

	Describe("lbs", func() {
		Context("when bbl'd up with terraform", func() {
			BeforeEach(func() {
				fakeTerraformBackendServer.SetHandler(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
					switch request.URL.Path {
					case "/output/--json":
						responseWriter.Write([]byte(fmt.Sprintf(`{
							"bosh_eip": {
								"value": "some-bosh-eip"
							},
							"bosh_url": {
								"value": %q
							},
							"bosh_user_access_key": {
								"value": "some-bosh-user-access-key"
							},
							"bosh_user_secret_access_key": {
								"value": "some-bosh-user-secret-access_key"
							},
							"nat_eip": {
								"value": "some-nat-eip"
							},
							"bosh_subnet_id": {
								"value": "some-bosh-subnet-id"
							},
							"bosh_subnet_availability_zone": {
								"value": "some-bosh-subnet-availability-zone"
							},
							"bosh_security_group": {
								"value": "some-bosh-security-group"
							},
							"env_dns_zone_name_servers": {
								"value": [
									"name-server-1.",
									"name-server-2."
								]
							},
							"internal_security_group": {
								"value": "some-internal-security-group"
							},
							"internal_subnet_ids": {
								"value": [
									"some-internal-subnet-ids-1",
									"some-internal-subnet-ids-2",
									"some-internal-subnet-ids-3"
								]
							},
							"internal_subnet_cidrs": {
								"value": [
									"10.0.16.0/20",
									"10.0.32.0/20",
									"10.0.48.0/20"
								]
							},
							"vpc_id": {
								"value": "some-vpc-id"
							},
							"cf_router_lb_name": {
								"value": "some-router-lb-name"
							},
							"cf_router_lb_url": {
								"value": "some-router-lb-url"
							},
							"cf_router_lb_internal_security_group": {
								"value": "some-cf-router-internal-security-group"
							},
							"cf_ssh_lb_name":  {
								"value": "some-ssh-proxy-lb-name"
							},
							"cf_ssh_lb_url":  {
								"value": "some-ssh-proxy-lb-url"
							},
							"cf_ssh_lb_internal_security_group":  {
								"value": "some-cf-ssh-proxy-internal-security-group"
							},
							"concourse_lb_name":  {
								"value": "some-concourse-lb"
							},
							"concourse_lb_internal_security_group":  {
								"value": "some-concourse-internal-security-group"
							}
						}`, fakeBOSHServer.URL)))
					}
				}))
				upAWSWithAdditionalFlags(fakeAWSServer.URL, tempDirectory, []string{"--terraform"}, 0)

				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"--debug",
					"create-lbs",
					"--type", "cf",
					"--cert", lbCertPath,
					"--key", lbKeyPath,
					"--domain", "cf.example.com",
				}

				executeCommand(args, 0)
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = true
			})

			AfterEach(func() {
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = false
			})

			It("prints out the currently attached lb names and urls", func() {
				session := lbs("", []string{}, tempDirectory, 0)
				stdout := session.Out.Contents()

				Expect(stdout).To(ContainSubstring("CF Router LB: some-router-lb-name [some-router-lb-url]\n"))
				Expect(stdout).To(ContainSubstring("CF SSH Proxy LB: some-ssh-proxy-lb-name [some-ssh-proxy-lb-url]\n"))
				Expect(stdout).To(ContainSubstring("CF System Domain DNS servers: name-server-1. name-server-2.\n"))
			})

			It("prints out the currently attached lb names and urls in JSON", func() {
				session := lbs("", []string{"--json"}, tempDirectory, 0)
				stdout := session.Out.Contents()

				Expect(stdout).To(MatchJSON(`{
					"cf_router_lb": "some-router-lb-name",
					"cf_router_lb_url": "some-router-lb-url",
					"cf_ssh_proxy_lb": "some-ssh-proxy-lb-name",
					"cf_ssh_proxy_lb_url": "some-ssh-proxy-lb-url",
					"env_dns_zone_name_servers": [
						"name-server-1.",
						"name-server-2."
					]
				}`))
			})
		})

		Context("when bbl'd up with cloudformation", func() {
			BeforeEach(func() {
				upAWS(fakeAWSServer.URL, tempDirectory, 0)
			})

			It("prints out the currently attached lb names and urls", func() {
				createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "cf", 0, false)

				session := lbs(fakeAWSServer.URL, []string{}, tempDirectory, 0)
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

	Describe("when no bosh director exists", func() {
		BeforeEach(func() {
			args := []string{
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

			executeCommand(args, 0)

		})

		It("creates a concourse lb", func() {
			createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "concourse", 0, false)

			certificates := fakeAWS.Certificates.All()
			Expect(certificates).To(HaveLen(1))
			Expect(certificates[0].CertificateBody).To(Equal(testhelpers.BBL_CERT))
			Expect(certificates[0].PrivateKey).To(Equal(testhelpers.BBL_KEY))
			Expect(certificates[0].Chain).To(Equal(testhelpers.BBL_CHAIN))
			Expect(certificates[0].Name).To(MatchRegexp(`concourse-elb-cert-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))
		})

		It("updates a cf lb", func() {
			writeStateJson(storage.State{
				Version: 3,
				IAAS:    "aws",
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

		It("deletes lbs", func() {
			writeStateJson(storage.State{
				Version: 3,
				IAAS:    "aws",
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
		})
	})

})

func lbs(endpointOverrideURL string, subcommandFlags []string, stateDir string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", endpointOverrideURL),
		"--state-dir", stateDir,
		"lbs",
	}

	return executeCommand(append(args, subcommandFlags...), exitCode)
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
		"--debug",
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

func checkCertificatesForCloudFormation(fakeAWS *awsbackend.Backend, lbType string) {
	certificates := fakeAWS.Certificates.All()
	Expect(certificates).To(HaveLen(1))
	Expect(certificates[0].CertificateBody).To(Equal(testhelpers.BBL_CERT))
	Expect(certificates[0].PrivateKey).To(Equal(testhelpers.BBL_KEY))
	Expect(certificates[0].Chain).To(Equal(testhelpers.BBL_CHAIN))
	Expect(certificates[0].Name).To(MatchRegexp(fmt.Sprintf(`%s-elb-cert-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`, lbType)))
}
