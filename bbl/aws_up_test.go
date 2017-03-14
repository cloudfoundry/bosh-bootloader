package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	"github.com/onsi/gomega/gexec"
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

var _ = Describe("bbl up aws", func() {
	var (
		fakeAWS                  *awsbackend.Backend
		fakeAWSServer            *httptest.Server
		fakeBOSHServer           *httptest.Server
		fakeBOSHCLIBackendServer *httptest.Server
		fakeBOSH                 *fakeBOSHDirector
		pathToFakeBOSH           string
		pathToBOSH               string
		tempDirectory            string
		lbCertPath               string
		lbChainPath              string
		lbKeyPath                string

		fastFail                 bool
		fastFailMutex            sync.Mutex
		callRealInterpolate      bool
		callRealInterpolateMutex sync.Mutex

		interpolateArgs []string
	)

	BeforeEach(func() {
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeBOSHCLIBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/version":
				responseWriter.Write([]byte("2.0.0"))
			case "/path":
				responseWriter.Write([]byte(originalPath))
			case "/interpolate/args":
				body, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				interpolateArgs = append(interpolateArgs, string(body))
			case "/create-env/fastfail":
				fastFailMutex.Lock()
				defer fastFailMutex.Unlock()
				if fastFail {
					responseWriter.WriteHeader(http.StatusInternalServerError)
				} else {
					responseWriter.WriteHeader(http.StatusOK)
				}
				return
			case "/call-real-interpolate":
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				if callRealInterpolate {
					responseWriter.Write([]byte("true"))
				} else {
					responseWriter.Write([]byte("false"))
				}
			default:
				responseWriter.WriteHeader(http.StatusOK)
				return
			}
		}))

		fakeAWS = awsbackend.New(fakeBOSHServer.URL)
		fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

		var err error
		pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHCLIBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
		err = os.Rename(pathToFakeBOSH, pathToBOSH)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSH), originalPath}, ":"))

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		lbCertPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		lbChainPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		lbKeyPath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		fastFailMutex.Lock()
		defer fastFailMutex.Unlock()
		fastFail = false
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)
	})

	Describe("up", func() {
		Context("when AWS creds are provided through environment variables", func() {
			It("honors the environment variables", func() {
				os.Setenv("BBL_AWS_ACCESS_KEY_ID", "some-access-key")
				os.Setenv("BBL_AWS_SECRET_ACCESS_KEY", "some-access-secret")
				os.Setenv("BBL_AWS_REGION", "some-region")
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"up",
					"--iaas", "aws",
				}

				executeCommand(args, 0)

				state := readStateJson(tempDirectory)
				Expect(state.AWS).To(Equal(storage.AWS{
					AccessKeyID:     "some-access-key",
					SecretAccessKey: "some-access-secret",
					Region:          "some-region",
				}))
			})
		})

		Context("when bbl-state.json contains aws details", func() {
			BeforeEach(func() {
				buf, err := json.Marshal(storage.State{
					Version: 3,
					IAAS:    "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-access-key",
						Region:          "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), buf, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not require --iaas flag and exits 0", func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"up",
				}

				executeCommand(args, 0)
			})

			Context("when called with --iaas gcp", func() {
				It("exits 1 and prints error message", func() {
					session := executeCommand([]string{
						"--state-dir", tempDirectory,
						"up",
						"--iaas", "gcp",
					}, 1)

					Expect(session.Err.Contents()).To(ContainSubstring("The iaas type cannot be changed for an existing environment. The current iaas type is aws."))
				})
			})
		})

		Context("when bbl-state.json contains aws details", func() {
			BeforeEach(func() {
				buf, err := json.Marshal(storage.State{
					Version: 3,
					IAAS:    "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-access-key",
						Region:          "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), buf, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not require --iaas flag and exits 0", func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"up",
				}

				executeCommand(args, 0)
			})

			Context("when called with --iaas gcp", func() {
				It("exits 1 and prints error message", func() {
					session := executeCommand([]string{
						"--state-dir", tempDirectory,
						"up",
						"--iaas", "gcp",
					}, 1)

					Expect(session.Err.Contents()).To(ContainSubstring("The iaas type cannot be changed for an existing environment. The current iaas type is aws."))
				})
			})
		})

		Context("when ops files are provides via --ops-file flag", func() {
			It("passes those ops files to bosh create env", func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"--debug",
					"up",
					"--iaas", "aws",
					"--aws-access-key-id", "some-access-key",
					"--aws-secret-access-key", "some-access-secret",
					"--aws-region", "some-region",
					"--ops-file", "fixtures/ops-file.yml",
				}

				executeCommand(args, 0)

				Expect(interpolateArgs[0]).To(MatchRegexp(`\"-o\",\".*user-ops-file.yml\"`))
			})
		})

		Context("when an az is provided for bosh via --aws-bosh-az flag", func() {
			It("creates a bosh subnet in the stack with provided az", func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"--debug",
					"up",
					"--iaas", "aws",
					"--aws-access-key-id", "some-access-key",
					"--aws-secret-access-key", "some-access-secret",
					"--aws-region", "some-region",
					"--aws-bosh-az", "some-bosh-az",
				}

				executeCommand(args, 0)

				state := readStateJson(tempDirectory)

				var stack awsbackend.Stack
				var ok bool
				stack, ok = fakeAWS.Stacks.Get(state.Stack.Name)
				Expect(ok).To(BeTrue())

				var template struct {
					Resources struct {
						BOSHSubnet struct {
							Properties templates.Subnet
							Type       string
						}
					}
				}

				err := json.Unmarshal([]byte(stack.Template), &template)
				Expect(err).NotTo(HaveOccurred())
				Expect(template.Resources.BOSHSubnet.Properties.AvailabilityZone).To(Equal("some-bosh-az"))
			})
		})

		Context("when the cloudformation stack does not exist", func() {
			var stack awsbackend.Stack

			It("creates a stack and a keypair", func() {
				upAWS(fakeAWSServer.URL, tempDirectory, 0)

				state := readStateJson(tempDirectory)

				var ok bool
				stack, ok = fakeAWS.Stacks.Get(state.Stack.Name)
				Expect(ok).To(BeTrue())
				Expect(state.Stack.Name).To(MatchRegexp(`stack-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z`))

				keyPairs := fakeAWS.KeyPairs.All()
				Expect(keyPairs).To(HaveLen(1))
				Expect(keyPairs[0].Name).To(MatchRegexp(`keypair-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z`))
			})

			It("creates an IAM user", func() {
				upAWS(fakeAWSServer.URL, tempDirectory, 0)

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
				Expect(template.Resources.BOSHUser.Properties.UserName).To(MatchRegexp(`bosh-iam-user-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z`))
			})

			It("does not change the iam user name when state exists", func() {
				fakeAWS.Stacks.Set(awsbackend.Stack{
					Name: "some-stack-name",
				})
				fakeAWS.KeyPairs.Set(awsbackend.KeyPair{
					Name: "some-keypair-name",
				})

				writeStateJson(storage.State{
					Version: 3,
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
				upAWS(fakeAWSServer.URL, tempDirectory, 0)

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

			It("logs the steps", func() {
				session := upAWS(fakeAWSServer.URL, tempDirectory, 0)

				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: creating keypair"))
				Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: creating cloudformation stack"))
				Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
			})

			It("invokes the bosh cli", func() {
				session := upAWS(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh create-env"))
			})

			It("names the bosh director with env id", func() {
				session := upAWS(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Err.Contents()).To(ContainSubstring("bosh director name: bosh-bbl-"))
			})

			It("does not change the bosh director name when state exists", func() {
				fakeAWS.Stacks.Set(awsbackend.Stack{
					Name: "some-stack-name",
				})
				fakeAWS.KeyPairs.Set(awsbackend.KeyPair{
					Name: "some-keypair-name",
				})

				writeStateJson(storage.State{
					Version: 3,
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
					EnvID: "lakename",
				}, tempDirectory)
				session := upAWS(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Err.Contents()).To(ContainSubstring("bosh director name: bosh-lakename"))
			})

			It("can invoke the bosh cli idempotently", func() {
				session := upAWS(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh create-env"))

				session = upAWS(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh create-env"))
				Expect(session.Out.Contents()).To(ContainSubstring("No new changes, skipping deployment..."))
			})

			It("fast fails if the bosh state exists", func() {
				writeStateJson(storage.State{Version: 3, BOSH: storage.BOSH{DirectorAddress: "some-director-address"}}, tempDirectory)
				session := upAWS(fakeAWSServer.URL, tempDirectory, 1)
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
					Version: 3,
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

				ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), buf, os.ModePerm)

				session := upAWS(fakeAWSServer.URL, tempDirectory, 0)

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
				upAWS(fakeAWSServer.URL, tempDirectory, 0)
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
			By("allowing the call of real interpolate in fake bosh", func() {
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = true
			})

			contents, err := ioutil.ReadFile(fixtureLocation)
			Expect(err).NotTo(HaveOccurred())

			session := upAWS(fakeAWSServer.URL, tempDirectory, 0)
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

			By("disabling the call of real interpolate in fake bosh", func() {
				callRealInterpolateMutex.Lock()
				defer callRealInterpolateMutex.Unlock()
				callRealInterpolate = false
			})
		},
			Entry("generates a cloud config with no lb type", "", "fixtures/cloud-config-no-elb.yml"),
			Entry("generates a cloud config with cf lb type", "cf", "fixtures/cloud-config-cf-elb.yml"),
			Entry("generates a cloud config with concourse lb type", "concourse", "fixtures/cloud-config-concourse-elb.yml"),
		)

		Context("when the --no-director flag is provided", func() {
			It("creates the infrastructure for a bosh director", func() {
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

				session := executeCommand(args, 0)

				stdout := session.Out.Contents()

				Expect(stdout).To(MatchRegexp(`step: checking if keypair "keypair-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z" exists`))
				Expect(stdout).To(ContainSubstring("step: creating keypair"))
				Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: creating cloudformation stack"))
				Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
			})

			It("does not invoke the bosh cli or create a cloud config", func() {
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

				session := executeCommand(args, 0)

				Expect(session.Out.Contents()).NotTo(ContainSubstring("bosh create-env"))
				Expect(session.Err.Contents()).NotTo(ContainSubstring("bosh director name: bosh-bbl-"))
				Expect(session.Err.Contents()).NotTo(ContainSubstring("step: generating cloud config"))
				Expect(session.Err.Contents()).NotTo(ContainSubstring("step: applying cloud config"))
			})
		})

		Describe("reentrant", func() {
			Context("when the keypair fails to create", func() {
				It("saves the keypair name to the state", func() {
					fakeAWS.KeyPairs.SetCreateKeyPairReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to create keypair",
					})
					session := upAWS(fakeAWSServer.URL, tempDirectory, 1)
					stdout := session.Out.Contents()
					stderr := session.Err.Contents()

					Expect(stdout).To(MatchRegexp(`step: checking if keypair "keypair-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z" exists`))
					Expect(stdout).To(ContainSubstring("step: creating keypair"))
					Expect(stderr).To(ContainSubstring("failed to create keypair"))

					state := readStateJson(tempDirectory)

					Expect(state.KeyPair.Name).To(MatchRegexp(`keypair-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z`))
				})
			})

			Context("when the stack fails to create", func() {
				It("saves the stack name to the state", func() {
					fakeAWS.Stacks.SetCreateStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to create stack",
					})
					session := upAWS(fakeAWSServer.URL, tempDirectory, 1)
					stdout := session.Out.Contents()
					stderr := session.Err.Contents()

					Expect(stdout).To(MatchRegexp(`step: checking if cloudformation stack "stack-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z" exists`))
					Expect(stdout).To(ContainSubstring("step: creating cloudformation stack"))
					Expect(stderr).To(ContainSubstring("failed to create stack"))

					state := readStateJson(tempDirectory)

					Expect(state.Stack.Name).To(MatchRegexp(`stack-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z`))
				})

				It("saves the private key to the state", func() {
					fakeAWS.Stacks.SetCreateStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to create stack",
					})
					upAWS(fakeAWSServer.URL, tempDirectory, 1)
					state := readStateJson(tempDirectory)

					Expect(state.KeyPair.PrivateKey).To(ContainSubstring(testhelpers.PRIVATE_KEY))
				})

				It("does not create a new key pair on second call", func() {
					fakeAWS.Stacks.SetCreateStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to create stack",
					})
					upAWS(fakeAWSServer.URL, tempDirectory, 1)

					fakeAWS.Stacks.SetCreateStackReturnError(nil)
					upAWS(fakeAWSServer.URL, tempDirectory, 0)

					Expect(fakeAWS.CreateKeyPairCallCount).To(Equal(int64(1)))
				})
			})

			Context("when the bosh cli fails to create", func() {
				It("does not re-provision stack", func() {
					fastFailMutex.Lock()
					fastFail = true
					fastFailMutex.Unlock()
					upAWS(fakeAWSServer.URL, tempDirectory, 1)

					fastFailMutex.Lock()
					fastFail = false
					fastFailMutex.Unlock()
					upAWS(fakeAWSServer.URL, tempDirectory, 0)

					Expect(fakeAWS.CreateStackCallCount).To(Equal(int64(1)))
				})

				It("stores a partial bosh state", func() {
					fastFailMutex.Lock()
					fastFail = true
					fastFailMutex.Unlock()
					upAWS(fakeAWSServer.URL, tempDirectory, 1)

					state := readStateJson(tempDirectory)
					Expect(state.BOSH.State).To(Equal(map[string]interface{}{
						"partial": "bosh-state",
					}))
				})
			})

			Context("when bosh cloud config fails to update", func() {
				It("saves the bosh properties to the state", func() {
					fakeBOSH.SetCloudConfigEndpointFail(true)
					upAWS(fakeAWSServer.URL, tempDirectory, 1)
					state := readStateJson(tempDirectory)

					Expect(state.BOSH.DirectorName).To(MatchRegexp(`bosh-bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z`))

					originalBOSHState := state.BOSH

					fakeBOSH.SetCloudConfigEndpointFail(false)
					upAWS(fakeAWSServer.URL, tempDirectory, 0)
					state = readStateJson(tempDirectory)

					Expect(state.BOSH).To(Equal(originalBOSHState))
				})
			})

			It("prints the up usage when --help is provided", func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"up",
					"--help",
				}

				session := executeCommand(args, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("Deploys BOSH director on an IAAS"))
				Expect(session.Out.Contents()).To(ContainSubstring("--iaas"))
			})

			Context("--iaas", func() {
				Context("when bbl-state.json does not exist", func() {
					It("writes iaas: aws to state and creates resources", func() {
						upAWS(fakeAWSServer.URL, tempDirectory, 0)

						state := readStateJson(tempDirectory)
						Expect(state.IAAS).To(Equal("aws"))

						var ok bool
						_, ok = fakeAWS.Stacks.Get(state.Stack.Name)
						Expect(ok).To(BeTrue())
					})
				})
			})
		})
	})
})

func upAWS(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--state-dir", tempDirectory,
		"--debug",
		"up",
		"--iaas", "aws",
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
	}

	return executeCommand(args, exitCode)
}

func createLB(serverURL string, tempDirectory string, lbType string, certPath string, keyPath string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--state-dir", tempDirectory,
		"create-lbs",
		"--type", lbType,
		"--cert", certPath,
		"--key", keyPath,
	}

	return executeCommand(args, exitCode)
}
