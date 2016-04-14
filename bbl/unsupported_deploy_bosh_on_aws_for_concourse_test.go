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
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

type fakeBOSHDirector struct {
	mutex       sync.Mutex
	CloudConfig []byte
}

func (b *fakeBOSHDirector) SetCloudConfig(cloudConfig []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.CloudConfig = cloudConfig
}

func (b *fakeBOSHDirector) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	buf, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	b.SetCloudConfig(buf)
	responseWriter.WriteHeader(http.StatusCreated)
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
				Expect(stdout).To(ContainSubstring("Director Address:  127.0.0.1"))
			})

			It("prints out randomized bosh director credentials", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				stdout := session.Out.Contents()
				Expect(stdout).To(MatchRegexp(`Director Username: user-\w{7}`))
				Expect(stdout).To(MatchRegexp(`Director Password: p-\w{15}`))
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

		Context("cloud config", func() {
			Context("when no elb is specified", func() {
				It("applies the cloud config", func() {
					contents, err := ioutil.ReadFile("fixtures/cloud-config-no-elb.yml")
					Expect(err).NotTo(HaveOccurred())

					args := []string{
						fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
						"--aws-access-key-id", "some-access-key",
						"--aws-secret-access-key", "some-access-secret",
						"--aws-region", "some-region",
						"--state-dir", tempDirectory,
						"unsupported-deploy-bosh-on-aws-for-concourse",
					}

					session := executeCommand(args, 0)
					stdout := session.Out.Contents()

					Expect(stdout).To(ContainSubstring("step: generating cloud config"))
					Expect(stdout).To(ContainSubstring("step: applying cloud config"))
					Expect(fakeBOSH.CloudConfig).To(MatchYAML(string(contents)))
				})
			})

			Context("when elb type is concourse", func() {
				It("applies the cloud config", func() {
					contents, err := ioutil.ReadFile("fixtures/cloud-config-concourse-elb.yml")
					Expect(err).NotTo(HaveOccurred())

					args := []string{
						fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
						"--aws-access-key-id", "some-access-key",
						"--aws-secret-access-key", "some-access-secret",
						"--aws-region", "some-region",
						"--state-dir", tempDirectory,
						"unsupported-deploy-bosh-on-aws-for-concourse",
						"--lb-type", "concourse",
					}

					session := executeCommand(args, 0)
					stdout := session.Out.Contents()

					Expect(stdout).To(ContainSubstring("step: generating cloud config"))
					Expect(stdout).To(ContainSubstring("step: applying cloud config"))
					Expect(fakeBOSH.CloudConfig).To(MatchYAML(string(contents)))
				})

				It("idempotently applies the cloud config", func() {
					contents, err := ioutil.ReadFile("fixtures/cloud-config-concourse-elb.yml")
					Expect(err).NotTo(HaveOccurred())

					By("invoking bbl with the lb-type flag", func() {
						args := []string{
							fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
							"--aws-access-key-id", "some-access-key",
							"--aws-secret-access-key", "some-access-secret",
							"--aws-region", "some-region",
							"--state-dir", tempDirectory,
							"unsupported-deploy-bosh-on-aws-for-concourse",
							"--lb-type", "concourse",
						}

						executeCommand(args, 0)
						Expect(fakeBOSH.CloudConfig).To(MatchYAML(string(contents)))
					})

					By("invoking bbl without the lb-type", func() {
						args := []string{
							fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
							"--state-dir", tempDirectory,
							"unsupported-deploy-bosh-on-aws-for-concourse",
						}

						executeCommand(args, 0)
						Expect(fakeBOSH.CloudConfig).To(MatchYAML(string(contents)))
					})
				})
			})
		})
	})
})

func writeStateJson(state storage.State, tempDirectory string) {
	buf, err := json.Marshal(state)
	Expect(err).NotTo(HaveOccurred())

	ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), buf, os.ModePerm)
}

func readStateJson(tempDirectory string) storage.State {
	buf, err := ioutil.ReadFile(filepath.Join(tempDirectory, "state.json"))
	Expect(err).NotTo(HaveOccurred())

	var state storage.State
	err = json.Unmarshal(buf, &state)
	Expect(err).NotTo(HaveOccurred())

	return state
}

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
