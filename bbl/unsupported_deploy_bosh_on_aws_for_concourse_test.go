package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bbl/awsbackend"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

type fakeBOSHDirector struct {
	mutex       sync.Mutex
	cloudConfig []byte
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
		privateKey     string
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

		contents, err := ioutil.ReadFile("fixtures/key.pem")
		Expect(err).NotTo(HaveOccurred())
		privateKey = string(contents)
	})

	Describe("unsupported-deploy-bosh-on-aws-for-concourse", func() {
		Context("when the cloudformation stack does not exist", func() {
			var stack awsbackend.Stack

			It("creates a stack and a keypair", func() {
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				state := readStateJson(tempDirectory)

				var ok bool
				stack, ok = fakeAWS.Stacks.Get(state.Stack.Name)
				Expect(ok).To(BeTrue())

				keyPairs := fakeAWS.KeyPairs.All()
				Expect(keyPairs).To(HaveLen(1))
				Expect(keyPairs[0].Name).To(MatchRegexp(`keypair-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))
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
						PrivateKey: privateKey,
					},
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
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

		Context("when an load balancer is attached", func() {
			It("attaches certificate to the load balancer", func() {
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				createLBs(fakeAWSServer.URL, tempDirectory, "concourse", 0, false)

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
				createLBs(fakeAWSServer.URL, tempDirectory, lbType, 0, false)
			}
			stdout := session.Out.Contents()

			Expect(stdout).To(ContainSubstring("step: generating cloud config"))
			Expect(stdout).To(ContainSubstring("step: applying cloud config"))
			Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))

			By("executing idempotently", func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"unsupported-deploy-bosh-on-aws-for-concourse",
				}

				executeCommand(args, 0)
				Expect(fakeBOSH.GetCloudConfig()).To(MatchYAML(string(contents)))
			})
		},
			Entry("generates a cloud config with no lb type", "", "fixtures/cloud-config-no-elb.yml"),
			Entry("generates a cloud config with cf lb type", "cf", "fixtures/cloud-config-cf-elb.yml"),
			Entry("generates a cloud config with concourse lb type", "concourse", "fixtures/cloud-config-concourse-elb.yml"),
		)
	})
})

func deployBOSHOnAWSForConcourse(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
		"--state-dir", tempDirectory,
		"unsupported-deploy-bosh-on-aws-for-concourse",
	}

	return executeCommand(args, exitCode)
}

func createLB(serverURL string, tempDirectory string, lbType string, exitCode int) *gexec.Session {
	dir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
		"--state-dir", tempDirectory,
		"unsupported-create-lbs",
		"--type", lbType,
		"--cert", filepath.Join(dir, "fixtures", "bbl.crt"),
		"--key", filepath.Join(dir, "fixtures", "bbl.key"),
	}

	return executeCommand(args, exitCode)
}
