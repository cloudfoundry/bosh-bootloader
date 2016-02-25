package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

var _ = Describe("bbl", func() {
	Describe("unsupported-provision-aws-for-concourse", func() {
		Context("when the cloudformation stack does not exist", func() {
			It("creates and applies a cloudformation template", func() {
				tempDir, err := ioutil.TempDir("", "")

				state := storage.State{
					KeyPair: &storage.KeyPair{
						Name: "some-keypair-name",
					},
				}

				buf, err := json.Marshal(state)
				Expect(err).NotTo(HaveOccurred())

				ioutil.WriteFile(filepath.Join(tempDir, "state.json"), buf, os.ModePerm)

				responses := []string{}

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))

					body, err := ioutil.ReadAll(r.Body)
					Expect(err).NotTo(HaveOccurred())

					responses = append(responses, string(body))
				}))

				args := []string{
					fmt.Sprintf("--endpoint-override=%s", server.URL),
					"--aws-access-key-id", "some-access-key",
					"--aws-secret-access-key", "some-access-secret",
					"--aws-region", "some-region",
					"--state-dir", tempDir,
					"unsupported-provision-aws-for-concourse",
				}
				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(responses).To(ContainElement(ContainSubstring("Action=CreateStack")))
				Expect(responses).To(ContainElement(ContainSubstring("StackName=concourse")))
			})
		})

		Context("when the cloudformation stack already exists", func() {
			It("updates the stack with the cloudformation template", func() {
				tempDir, err := ioutil.TempDir("", "")

				state := storage.State{
					KeyPair: &storage.KeyPair{
						Name: "some-keypair-name",
					},
				}

				buf, err := json.Marshal(state)
				Expect(err).NotTo(HaveOccurred())

				ioutil.WriteFile(filepath.Join(tempDir, "state.json"), buf, os.ModePerm)

				responses := []string{}

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))

					body, err := ioutil.ReadAll(r.Body)
					Expect(err).NotTo(HaveOccurred())

					responses = append(responses, string(body))
					if strings.Contains(string(body), "Action=DescribeStack") {
						w.Write([]byte(describeResponse))
					}
				}))

				args := []string{
					fmt.Sprintf("--endpoint-override=%s", server.URL),
					"--aws-access-key-id", "some-access-key",
					"--aws-secret-access-key", "some-access-secret",
					"--aws-region", "some-region",
					"--state-dir", tempDir,
					"unsupported-provision-aws-for-concourse",
				}
				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(responses).To(ContainElement(ContainSubstring("Action=UpdateStack")))
				Expect(responses).To(ContainElement(ContainSubstring("StackName=concourse")))
			})
		})
	})
})

var describeResponse = strings.TrimSpace(`
<DescribeStacksResponse xmlns="http://cloudformation.amazonaws.com/doc/2010-05-15/">
  <DescribeStacksResult>
    <Stacks>
      <member>
        <StackName>concourse</StackName>
        <StackId>arn:aws:cloudformation:us-east-1:123456789:stack/MyStack/aaf549a0-a413-11df-adb3-5081b3858e83</StackId>
        <CreationTime>2010-07-27T22:28:28Z</CreationTime>
        <StackStatus>CREATE_COMPLETE</StackStatus>
        <DisableRollback>false</DisableRollback>
        <Outputs>
          <member>
            <OutputKey>StartPage</OutputKey>
            <OutputValue>http://my-load-balancer.amazonaws.com:80/index.html</OutputValue>
          </member>
        </Outputs>
      </member>
    </Stacks>
  </DescribeStacksResult>
  <ResponseMetadata>
    <RequestId>b9b4b068-3a41-11e5-94eb-example</RequestId>
  </ResponseMetadata>
</DescribeStacksResponse>
`)
