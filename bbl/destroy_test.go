package main_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bbl/awsbackend"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("destroy", func() {
	Context("asks for confirmation before it starts destroying things", func() {
		var (
			cmd            *exec.Cmd
			stdin          io.WriteCloser
			stdout         io.ReadCloser
			fakeAWS        *awsbackend.Backend
			fakeAWSServer  *httptest.Server
			fakeBOSH       *fakeBOSHDirector
			fakeBOSHServer *httptest.Server
			tempDirectory  string
			privateKey     string
		)

		BeforeEach(func() {
			fakeBOSH = &fakeBOSHDirector{}
			fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				fakeBOSH.ServeHTTP(responseWriter, request)
			}))

			fakeAWS = awsbackend.New(fakeBOSHServer.URL)
			fakeAWS.Stacks.Set(awsbackend.Stack{
				Name: "some-stack-name",
			})
			fakeAWS.KeyPairs.Set(awsbackend.KeyPair{
				Name: "some-keypair-name",
			})
			fakeAWS.Instances.Set([]awsbackend.Instance{
				{Name: "bosh/0", VPCID: "some-vpc-id"},
				{Name: "NAT", VPCID: "some-vpc-id"},
			})
			fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

			var err error
			tempDirectory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile("fixtures/key.pem")
			Expect(err).NotTo(HaveOccurred())
			privateKey = string(contents)

			buf, err := json.Marshal(storage.State{
				KeyPair: &storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: privateKey,
				},
				Stack: storage.Stack{
					Name: "some-stack-name",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), buf, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
				"--aws-access-key-id", "some-access-key",
				"--aws-secret-access-key", "some-access-secret",
				"--aws-region", "some-region",
				"--state-dir", tempDirectory,
				"destroy",
			}
			cmd = exec.Command(pathToBBL, args...)

			stdin, err = cmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			stdout, err = cmd.StdoutPipe()
			Expect(err).NotTo(HaveOccurred())

			err = cmd.Start()
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() (string, error) {
				bytes, err := bufio.NewReader(stdout).ReadBytes(':')
				if err != nil {
					return "", err
				}

				return string(bytes), nil
			}, "10s", "10s").Should(ContainSubstring("Are you sure you want to delete your infrastructure? This operation cannot be undone!"))
		})

		It("continues with the destruction if you agree", func() {
			_, err := stdin.Write([]byte("yes\n"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() (string, error) {
				bytes, err := bufio.NewReader(stdout).ReadBytes('\n')
				if err != nil {
					return "", err
				}

				return string(bytes), nil
			}, "10s", "10s").Should(ContainSubstring("step: destroying infrastructure"))

			Eventually(cmd.Wait).Should(Succeed())
		})

		It("does not destroy your infrastructure if you do not agree", func() {
			_, err := stdin.Write([]byte("no\n"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() (string, error) {
				bytes, err := bufio.NewReader(stdout).ReadBytes('\n')
				if err != nil {
					return "", err
				}

				return string(bytes), nil
			}, "10s", "10s").Should(ContainSubstring("step: exiting"))

			Eventually(cmd.Wait).Should(Succeed())
		})
	})

	Context("when the bosh director, cloudformation stack, and ec2 keypair exists", func() {
		var (
			fakeAWS        *awsbackend.Backend
			fakeAWSServer  *httptest.Server
			fakeBOSH       *fakeBOSHDirector
			fakeBOSHServer *httptest.Server
			tempDirectory  string
			privateKey     string
		)

		BeforeEach(func() {
			fakeBOSH = &fakeBOSHDirector{}
			fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				fakeBOSH.ServeHTTP(responseWriter, request)
			}))

			fakeAWS = awsbackend.New(fakeBOSHServer.URL)
			fakeAWS.Stacks.Set(awsbackend.Stack{
				Name: "some-stack-name",
			})
			fakeAWS.KeyPairs.Set(awsbackend.KeyPair{
				Name: "some-keypair-name",
			})
			fakeAWS.Instances.Set([]awsbackend.Instance{
				{Name: "NAT"},
				{Name: "bosh/0"},
			})
			fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

			var err error
			tempDirectory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile("fixtures/key.pem")
			Expect(err).NotTo(HaveOccurred())
			privateKey = string(contents)

			buf, err := json.Marshal(storage.State{
				KeyPair: &storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: privateKey,
				},
				Stack: storage.Stack{
					Name: "some-stack-name",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), buf, os.ModePerm)
		})

		It("invokes bosh-init delete", func() {
			session := destroy(fakeAWSServer.URL, tempDirectory, 0)
			Expect(session.Out.Contents()).To(ContainSubstring("step: destroying bosh director"))
			Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init delete bosh.yml]"))
			Expect(session.Out.Contents()).To(ContainSubstring(`bosh-state.json: {}`))
		})

		It("deletes the stack", func() {
			session := destroy(fakeAWSServer.URL, tempDirectory, 0)
			Expect(session.Out.Contents()).To(ContainSubstring("step: deleting cloudformation stack"))
			Expect(session.Out.Contents()).To(ContainSubstring("step: finished deleting cloudformation stack"))

			_, ok := fakeAWS.Stacks.Get("some-stack-name")
			Expect(ok).To(BeFalse())
		})

		It("deletes the keypair", func() {
			session := destroy(fakeAWSServer.URL, tempDirectory, 0)
			Expect(session.Out.Contents()).To(ContainSubstring("step: deleting keypair"))
			Expect(session.Out.Contents()).To(ContainSubstring("step: finished deleting cloudformation stack"))

			_, ok := fakeAWS.Stacks.Get("some-stack-name")
			Expect(ok).To(BeFalse())
		})

		It("deletes the state.json file", func() {
			destroy(fakeAWSServer.URL, tempDirectory, 0)

			_, err := os.Stat(filepath.Join(tempDirectory, "state.json"))
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})
})

func destroy(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
		"--state-dir", tempDirectory,
		"destroy",
	}
	cmd := exec.Command(pathToBBL, args...)
	stdin, err := cmd.StdinPipe()
	Expect(err).NotTo(HaveOccurred())

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	_, err = stdin.Write([]byte("yes\n"))
	Expect(err).NotTo(HaveOccurred())

	Eventually(session, 10*time.Second).Should(gexec.Exit(exitCode))

	return session
}
