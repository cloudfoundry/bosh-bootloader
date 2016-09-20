package main_test

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/onsi/gomega/gexec"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

type fakeBOSHDirector struct {
	mutex           sync.Mutex
	cloudConfig     []byte
	cloudConfigFail bool
}

func (b *fakeBOSHDirector) SetCloudConfig(cloudConfig []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.cloudConfig = cloudConfig
}

func (b *fakeBOSHDirector) GetCloudConfig() []byte {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.cloudConfig
}

func (b *fakeBOSHDirector) SetCloudConfigEndpointFail(fail bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.cloudConfigFail = fail
}

func (b *fakeBOSHDirector) GetCloudConfigEndpointFail() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.cloudConfigFail
}

func (b *fakeBOSHDirector) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/info":
		responseWriter.Write([]byte(`{
			"name": "some-bosh-director",
			"uuid": "some-uuid",
			"version": "some-version"
		}`))

		return
	case "/cloud_configs":
		if b.GetCloudConfigEndpointFail() {
			responseWriter.WriteHeader(0)
			return
		}
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		b.SetCloudConfig(buf)
		responseWriter.WriteHeader(http.StatusCreated)

		return
	default:
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
}

var _ = Describe("bbl", func() {
	var (
		fakeAWS        *awsbackend.Backend
		fakeAWSServer  *httptest.Server
		fakeBOSHServer *httptest.Server
		fakeBOSH       *fakeBOSHDirector
		tempDirectory  string
		lbCertPath     string
		lbChainPath    string
		lbKeyPath      string
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

		lbCertPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		lbChainPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		lbKeyPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("up", func() {
		Context("when the cloudformation stack does not exist", func() {
			var stack awsbackend.Stack

			It("creates a stack and a keypair", func() {
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				state := readStateJson(tempDirectory)

				var ok bool
				stack, ok = fakeAWS.Stacks.Get(state.Stack.Name)
				Expect(ok).To(BeTrue())
				Expect(state.Stack.Name).To(MatchRegexp(`stack-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}T\d{2}-\d{2}Z`))

				keyPairs := fakeAWS.KeyPairs.All()
				Expect(keyPairs).To(HaveLen(1))
				Expect(keyPairs[0].Name).To(MatchRegexp(`keypair-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}T\d{2}:\d{2}Z`))
			})

			It("creates an IAM user", func() {
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				state := readStateJson(tempDirectory)

				var ok bool
				stack, ok = fakeAWS.Stacks.Get(state.Stack.Name)
				Expect(ok).To(BeTrue())

				var template struct {
					Resources struct {
						BOSHUser struct {
							Properties templates.IAMUser
							Type       string
						}
					}
				}

				err := json.Unmarshal([]byte(stack.Template), &template)
				Expect(err).NotTo(HaveOccurred())

				Expect(template.Resources.BOSHUser.Properties.Policies).To(HaveLen(1))
				Expect(template.Resources.BOSHUser.Properties.UserName).To(MatchRegexp(`bosh-iam-user-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}T\d{2}-\d{2}Z`))
			})

			It("does not change the iam user name when state exists", func() {
				fakeAWS.Stacks.Set(awsbackend.Stack{
					Name: "some-stack-name",
				})
				fakeAWS.KeyPairs.Set(awsbackend.KeyPair{
					Name: "some-keypair-name",
				})

				writeStateJson(storage.State{
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
					KeyPair: storage.KeyPair{
						Name: "some-keypair-name",
					},
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
					BOSH: storage.BOSH{
						DirectorAddress: fakeBOSHServer.URL,
					},
				}, tempDirectory)
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				state := readStateJson(tempDirectory)

				var ok bool
				stack, ok = fakeAWS.Stacks.Get(state.Stack.Name)
				Expect(ok).To(BeTrue())

				var template struct {
					Resources struct {
						BOSHUser struct {
							Properties templates.IAMUser
							Type       string
						}
					}
				}

				err := json.Unmarshal([]byte(stack.Template), &template)
				Expect(err).NotTo(HaveOccurred())

				Expect(template.Resources.BOSHUser.Properties.Policies).To(HaveLen(1))
				Expect(template.Resources.BOSHUser.Properties.UserName).To(BeEmpty())
			})

			It("logs the steps and bosh-init manifest", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: creating keypair"))
				Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: creating cloudformation stack"))
				Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: generating bosh-init manifest"))
				Expect(stdout).To(ContainSubstring("step: deploying bosh director"))
			})

			It("invokes bosh-init", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init deploy bosh.yml]"))
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-state.json: {}"))
			})

			It("names the bosh director with env id", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh director name: bosh-bbl-"))
			})

			It("does not change the bosh director name when state exists", func() {
				fakeAWS.Stacks.Set(awsbackend.Stack{
					Name: "some-stack-name",
				})
				fakeAWS.KeyPairs.Set(awsbackend.KeyPair{
					Name: "some-keypair-name",
				})

				writeStateJson(storage.State{
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					},
					KeyPair: storage.KeyPair{
						Name: "some-keypair-name",
					},
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
					BOSH: storage.BOSH{
						DirectorAddress: fakeBOSHServer.URL,
					},
				}, tempDirectory)
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh director name: my-bosh"))
			})

			It("signs bosh-init cert and key with the generated CA cert", func() {
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				state := readStateJson(tempDirectory)

				caCert := state.BOSH.DirectorSSLCA
				cert := state.BOSH.DirectorSSLCertificate

				rawCACertBlock, rest := pem.Decode([]byte(caCert))
				Expect(rest).To(HaveLen(0))

				rawCertBlock, rest := pem.Decode([]byte(cert))
				Expect(rest).To(HaveLen(0))

				rawCACert, err := x509.ParseCertificate(rawCACertBlock.Bytes)
				Expect(err).NotTo(HaveOccurred())

				rawCert, err := x509.ParseCertificate(rawCertBlock.Bytes)
				Expect(err).NotTo(HaveOccurred())

				err = rawCert.CheckSignatureFrom(rawCACert)
				Expect(err).NotTo(HaveOccurred())
			})

			It("can invoke bosh-init idempotently", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init deploy bosh.yml]"))
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-state.json: {}"))

				session = deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init deploy bosh.yml]"))
				Expect(session.Out.Contents()).To(ContainSubstring(`bosh-state.json: {"key":"value","md5checksum":`))
				Expect(session.Out.Contents()).To(ContainSubstring("No new changes, skipping deployment..."))
			})

			It("fast fails if the bosh state exists", func() {
				writeStateJson(storage.State{BOSH: storage.BOSH{DirectorAddress: "some-director-address"}}, tempDirectory)
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 1)
				Expect(session.Err.Contents()).To(ContainSubstring("Found BOSH data in state directory"))
			})
		})

		Context("when the keypair and cloudformation stack already exist", func() {
			BeforeEach(func() {
				fakeAWS.Stacks.Set(awsbackend.Stack{
					Name: "some-stack-name",
				})
				fakeAWS.KeyPairs.Set(awsbackend.KeyPair{
					Name: "some-keypair-name",
				})
			})

			It("updates the stack with the cloudformation template", func() {
				buf, err := json.Marshal(storage.State{
					KeyPair: storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: testhelpers.BBL_KEY,
					},
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
					EnvID: "bbl-env-lake-timestamp",
				})
				Expect(err).NotTo(HaveOccurred())

				ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), buf, os.ModePerm)

				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				template, err := ioutil.ReadFile("fixtures/cloudformation-no-elb.json")
				Expect(err).NotTo(HaveOccurred())

				stack, ok := fakeAWS.Stacks.Get("some-stack-name")
				Expect(ok).To(BeTrue())
				Expect(stack.Name).To(Equal("some-stack-name"))
				Expect(stack.WasUpdated).To(Equal(true))
				Expect(stack.Template).To(MatchJSON(string(template)))

				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: using existing keypair"))
				Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: updating cloudformation stack"))
				Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
			})
		})

		Context("when a load balancer is attached", func() {
			It("attaches certificate to the load balancer", func() {
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, "concourse", 0, false)

				state := readStateJson(tempDirectory)

				stack, ok := fakeAWS.Stacks.Get(state.Stack.Name)
				Expect(ok).To(BeTrue())

				type listener struct {
					SSLCertificateId string
				}

				var template struct {
					Resources struct {
						ConcourseLoadBalancer struct {
							Properties struct {
								Listeners []listener
							}
						}
					}
				}

				err := json.Unmarshal([]byte(stack.Template), &template)
				Expect(err).NotTo(HaveOccurred())

				Expect(template.Resources.ConcourseLoadBalancer.Properties.Listeners).To(ContainElement(listener{
					SSLCertificateId: "some-certificate-arn",
				}))
			})
		})

		DescribeTable("cloud config", func(lbType, fixtureLocation string) {
			contents, err := ioutil.ReadFile(fixtureLocation)
			Expect(err).NotTo(HaveOccurred())

			session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
			if lbType != "" {
				createLBs(fakeAWSServer.URL, tempDirectory, lbCertPath, lbKeyPath, lbChainPath, lbType, 0, false)
			}
			stdout := session.Out.Contents()

			Expect(stdout).To(ContainSubstring("step: generating cloud config"))
			Expect(stdout).To(ContainSubstring("step: applying cloud config"))
			Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))

			By("executing idempotently", func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"up",
				}

				executeCommand(args, 0)
				Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))
			})
		},
			Entry("generates a cloud config with no lb type", "", "fixtures/cloud-config-no-elb.yml"),
			Entry("generates a cloud config with cf lb type", "cf", "fixtures/cloud-config-cf-elb.yml"),
			Entry("generates a cloud config with concourse lb type", "concourse", "fixtures/cloud-config-concourse-elb.yml"),
		)

		Describe("reentrant", func() {
			Context("when the keypair fails to create", func() {
				It("saves the keypair name to the state", func() {
					fakeAWS.KeyPairs.SetCreateKeyPairReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to create keypair",
					})
					session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 1)
					stdout := session.Out.Contents()
					stderr := session.Err.Contents()

					Expect(stdout).To(MatchRegexp(`step: checking if keypair "keypair-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}T\d{2}:\d{2}Z" exists`))
					Expect(stdout).To(ContainSubstring("step: creating keypair"))
					Expect(stderr).To(ContainSubstring("failed to create keypair"))

					state := readStateJson(tempDirectory)

					Expect(state.KeyPair.Name).To(MatchRegexp(`keypair-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}T\d{2}:\d{2}Z`))
				})
			})

			Context("when the stack fails to create", func() {
				It("saves the stack name to the state", func() {
					fakeAWS.Stacks.SetCreateStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to create stack",
					})
					session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 1)
					stdout := session.Out.Contents()
					stderr := session.Err.Contents()

					Expect(stdout).To(MatchRegexp(`step: checking if cloudformation stack "stack-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}T\d{2}-\d{2}Z" exists`))
					Expect(stdout).To(ContainSubstring("step: creating cloudformation stack"))
					Expect(stderr).To(ContainSubstring("failed to create stack"))

					state := readStateJson(tempDirectory)

					Expect(state.Stack.Name).To(MatchRegexp(`stack-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}T\d{2}-\d{2}Z`))
				})

				It("saves the private key to the state", func() {
					fakeAWS.Stacks.SetCreateStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to create stack",
					})
					deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 1)
					state := readStateJson(tempDirectory)

					Expect(state.KeyPair.PrivateKey).To(ContainSubstring(testhelpers.PRIVATE_KEY))
				})

				It("does not create a new key pair on second call", func() {
					fakeAWS.Stacks.SetCreateStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to create stack",
					})
					deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 1)

					fakeAWS.Stacks.SetCreateStackReturnError(nil)
					deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

					Expect(fakeAWS.CreateKeyPairCallCount).To(Equal(int64(1)))
				})
			})

			Context("when bosh init fails to create", func() {
				It("does not re-provision stack", func() {
					originalPath := os.Getenv("PATH")

					By("rebuilding bosh-init with fail fast flag", func() {
						pathToFakeBOSHInit, err := gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakeboshinit",
							"-ldflags",
							"-X main.FailFast=true")
						Expect(err).NotTo(HaveOccurred())

						pathToBOSHInit = filepath.Join(filepath.Dir(pathToFakeBOSHInit), "bosh-init")
						err = os.Rename(pathToFakeBOSHInit, pathToBOSHInit)
						Expect(err).NotTo(HaveOccurred())

						os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSHInit), os.Getenv("PATH")}, ":"))
					})

					By("running up twice and checking if it created one stack", func() {
						deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 1)

						os.Setenv("PATH", originalPath)
						deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

						Expect(fakeAWS.CreateStackCallCount).To(Equal(int64(1)))
					})
				})
			})

			Context("when bosh cloud config fails to update", func() {
				It("saves the bosh properties to the state", func() {
					fakeBOSH.SetCloudConfigEndpointFail(true)
					deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 1)
					state := readStateJson(tempDirectory)

					Expect(state.BOSH.DirectorName).To(MatchRegexp(`bosh-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}T\d{2}:\d{2}Z`))

					originalBOSHState := state.BOSH

					fakeBOSH.SetCloudConfigEndpointFail(false)
					deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
					state = readStateJson(tempDirectory)

					Expect(state.BOSH).To(Equal(originalBOSHState))
				})
			})
		})
	})

})

func deployBOSHOnAWSForConcourse(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
		"--state-dir", tempDirectory,
		"up",
	}

	return executeCommand(args, exitCode)
}

func createLB(serverURL string, tempDirectory string, lbType string, certPath string, keyPath string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
		"--state-dir", tempDirectory,
		"create-lbs",
		"--type", lbType,
		"--cert", certPath,
		"--key", keyPath,
	}

	return executeCommand(args, exitCode)
}
