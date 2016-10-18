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

	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/cloudfoundry/bosh-bootloader/boshinit"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	"github.com/onsi/gomega/gexec"
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

			buf, err := json.Marshal(storage.State{
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key",
					SecretAccessKey: "some-access-secret",
					Region:          "some-region",
				},
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: testhelpers.BBL_KEY,
				},
				Stack: storage.Stack{
					Name: "some-stack-name",
				},
				BOSH: storage.BOSH{
					DirectorName: "some-bosh-director-name",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), buf, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
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
			}, "10s", "10s").Should(ContainSubstring("Are you sure you want to delete infrastructure for"))
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
			}, "10s", "10s").Should(ContainSubstring("step: destroying bosh director"))

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

	Context("when the bosh director, cloudformation stack, certificate, and ec2 keypair exists", func() {
		var (
			fakeAWS        *awsbackend.Backend
			fakeAWSServer  *httptest.Server
			fakeBOSH       *fakeBOSHDirector
			fakeBOSHServer *httptest.Server
			tempDirectory  string
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
			fakeAWS.Certificates.Set(awsbackend.Certificate{
				Name: "some-certificate-name",
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

			buf, err := json.Marshal(storage.State{
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key",
					SecretAccessKey: "some-access-secret",
					Region:          "some-region",
				},
				Stack: storage.Stack{
					Name:            "some-stack-name",
					CertificateName: "some-certificate-name",
				},
				BOSH: storage.BOSH{
					State: boshinit.State{
						"key": "value",
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), buf, os.ModePerm)
		})

		It("invokes bosh-init delete", func() {
			session := destroy(fakeAWSServer.URL, tempDirectory, 0)
			Expect(session.Out.Contents()).To(ContainSubstring("step: destroying bosh director"))
			Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init delete bosh.yml]"))
			Expect(session.Out.Contents()).To(ContainSubstring(`bosh-state.json: {"key":"value"}`))
		})

		It("deletes the stack", func() {
			session := destroy(fakeAWSServer.URL, tempDirectory, 0)
			Expect(session.Out.Contents()).To(ContainSubstring("step: deleting cloudformation stack"))
			Expect(session.Out.Contents()).To(ContainSubstring("step: finished deleting cloudformation stack"))

			_, ok := fakeAWS.Stacks.Get("some-stack-name")
			Expect(ok).To(BeFalse())
		})

		It("deletes the certificate", func() {
			session := destroy(fakeAWSServer.URL, tempDirectory, 0)
			Expect(session.Out.Contents()).To(ContainSubstring("step: deleting certificate"))
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

		It("deletes the bbl state", func() {
			destroy(fakeAWSServer.URL, tempDirectory, 0)

			_, err := os.Stat(filepath.Join(tempDirectory, storage.StateFileName))
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		Context("reentrance", func() {
			Context("when destroy fails to delete the stack", func() {
				It("removes bosh properties from the state", func() {
					fakeAWS.Stacks.SetDeleteStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to delete stack",
					})
					session := destroy(fakeAWSServer.URL, tempDirectory, 1)
					Expect(session.Out.Contents()).To(ContainSubstring("step: destroying bosh director"))
					Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init delete bosh.yml]"))
					Expect(session.Out.Contents()).To(ContainSubstring(`bosh-state.json: {"key":"value"}`))
					Expect(session.Out.Contents()).To(ContainSubstring("step: deleting cloudformation stack"))

					Expect(session.Out.Contents()).NotTo(ContainSubstring("step: finished deleting cloudformation stack"))
					state := readStateJson(tempDirectory)

					Expect(state.BOSH).To(Equal(storage.BOSH{}))
				})
			})

			Context("when no bosh director exists", func() {
				BeforeEach(func() {
					fakeAWS.Stacks.SetDeleteStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to delete stack",
					})
					destroy(fakeAWSServer.URL, tempDirectory, 1)

					fakeAWS.Stacks.SetDeleteStackReturnError(nil)
				})

				It("skips deleting bosh director", func() {
					session := destroy(fakeAWSServer.URL, tempDirectory, 0)
					Expect(session.Out.Contents()).To(ContainSubstring("no BOSH director, skipping..."))
					Expect(session.Out.Contents()).To(ContainSubstring("step: finished deleting cloudformation stack"))

					Expect(session.Out.Contents()).NotTo(ContainSubstring("step: destroying bosh director"))
				})
			})

			Context("when destroy fails to delete the keypair", func() {
				It("removes the stack from the state", func() {
					fakeAWS.KeyPairs.SetDeleteKeyPairReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to delete keypair",
					})
					session := destroy(fakeAWSServer.URL, tempDirectory, 1)
					Expect(session.Out.Contents()).To(ContainSubstring("step: deleting cloudformation stack"))
					Expect(session.Out.Contents()).To(ContainSubstring("step: finished deleting cloudformation stack"))
					Expect(session.Out.Contents()).To(ContainSubstring("step: deleting keypair"))
					state := readStateJson(tempDirectory)

					Expect(state.Stack.Name).To(Equal(""))
					Expect(state.Stack.LBType).To(Equal(""))
				})

				It("removes the certificate from the state", func() {
					fakeAWS.KeyPairs.SetDeleteKeyPairReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to delete keypair",
					})
					session := destroy(fakeAWSServer.URL, tempDirectory, 1)
					Expect(session.Out.Contents()).To(ContainSubstring("step: deleting certificate"))
					state := readStateJson(tempDirectory)

					Expect(state.Stack.CertificateName).To(Equal(""))
				})
			})

			Context("when no stack exists", func() {
				BeforeEach(func() {
					fakeAWS.KeyPairs.SetDeleteKeyPairReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to delete keypair",
					})
					destroy(fakeAWSServer.URL, tempDirectory, 1)

					fakeAWS.KeyPairs.SetDeleteKeyPairReturnError(nil)
				})

				It("skips deleting aws stack", func() {
					session := destroy(fakeAWSServer.URL, tempDirectory, 0)
					Expect(session.Out.Contents()).To(ContainSubstring("no AWS stack, skipping..."))
				})
			})
		})
	})
})

func destroy(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
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
