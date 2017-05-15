package main_test

import (
	"bufio"
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
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/onsi/gomega/gexec"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	variables = `
admin_password: rhkj9ys4l9guqfpc9vmp
director_ssl:
  certificate: some-certificate
  private_key: some-private-key
  ca: some-ca
`
)

var _ = Describe("destroy", func() {
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
		fakeTerraformBackendServer.SetFakeBOSHServer(fakeBOSHServer.URL)

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

		fakeBOSHCLIBackendServer.ResetAll()
	})

	AfterEach(func() {
		fakeBOSHCLIBackendServer.ResetAll()
		fakeTerraformBackendServer.ResetAll()
	})

	Context("when the state file does not exist", func() {
		It("exits with status 0 if --skip-if-missing flag is provided", func() {
			tempDirectory, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"destroy",
				"--skip-if-missing",
			}
			cmd := exec.Command(pathToBBL, args...)

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))

			Expect(session.Out.Contents()).To(ContainSubstring("state file not found, and --skip-if-missing flag provided, exiting"))
		})

		It("exits with status 1 and outputs helpful error message", func() {
			tempDirectory, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"destroy",
			}
			cmd := exec.Command(pathToBBL, args...)

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(1))

			Expect(session.Err.Contents()).To(ContainSubstring(fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)))
		})
	})

	Context("when bosh director, infrastructure, certificate, and ec2 keypair exists", func() {
		BeforeEach(func() {
			state := storage.State{
				Version: 3,
				IAAS:    "aws",
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
					Variables: variables,
					State: map[string]interface{}{
						"new-key": "new-value",
					},
				},
			}

			writeStateJson(state, tempDirectory)
		})

		Context("asks for confirmation before it starts destroying things", func() {
			var (
				cmd    *exec.Cmd
				stdin  io.WriteCloser
				stdout io.ReadCloser
			)

			BeforeEach(func() {
				args := []string{
					fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
					"--state-dir", tempDirectory,
					"destroy",
				}
				cmd = exec.Command(pathToBBL, args...)

				var err error
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

		It("invokes bosh delete-env", func() {
			session := destroy(fakeAWSServer.URL, tempDirectory, 0)
			Expect(session.Out.Contents()).To(ContainSubstring("step: destroying bosh director"))
		})

		Context("when the bbl environment used terraform to create infrastructure", func() {
			var (
				state storage.State
			)
			BeforeEach(func() {
				state = storage.State{
					Version: 3,
					IAAS:    "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key",
						SecretAccessKey: "some-access-secret",
						Region:          "some-region",
					},
					TFState: "some-tf-state",
					BOSH: storage.BOSH{
						Variables: variables,
						State: map[string]interface{}{
							"new-key": "new-value",
						},
					},
				}

				writeStateJson(state, tempDirectory)
			})

			It("deletes the infrastructure", func() {
				session := destroy(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("step: destroying infrastructure"))
				Expect(session.Out.Contents()).To(ContainSubstring("step: finished destroying infrastructure"))

				Expect(session.Out.Contents()).To(ContainSubstring("terraform destroy"))
			})

			Context("when destroy fails to destroy terraform infrastructure", func() {
				BeforeEach(func() {
					state.AWS.Region = "fail-to-terraform"

					writeStateJson(state, tempDirectory)
				})

				It("removes bosh properties from the state", func() {
					session := destroy(fakeAWSServer.URL, tempDirectory, 1)
					Expect(session.Out.Contents()).To(ContainSubstring("step: destroying bosh director"))
					Expect(session.Out.Contents()).To(ContainSubstring("step: destroying infrastructure"))

					Expect(session.Out.Contents()).NotTo(ContainSubstring("step: finished destroying infrastructure"))
					state := readStateJson(tempDirectory)

					Expect(state.BOSH).To(Equal(storage.BOSH{}))
				})

				It("saves the tf state when terraform destroy fails with ManagerError", func() {
					destroy(fakeAWSServer.URL, tempDirectory, 1)

					state := readStateJson(tempDirectory)
					Expect(state.TFState).To(Equal(`{"key":"partial-apply"}`))
				})

				Context("when no --debug is provided", func() {
					It("returns a helpful error message", func() {
						session := destroyWithOptionalDebug(fakeAWSServer.URL, tempDirectory, 1, false)
						Expect(session.Err.Contents()).To(ContainSubstring("use --debug for additional debug output"))
					})
				})
			})
		})

		Context("when the bbl environment used cloudformation to create infrastructure", func() {
			It("deletes the stack", func() {
				session := destroy(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("step: deleting cloudformation stack"))
				Expect(session.Out.Contents()).To(ContainSubstring("step: finished deleting cloudformation stack"))

				_, ok := fakeAWS.Stacks.Get("some-stack-name")
				Expect(ok).To(BeFalse())
			})

			Context("when destroy fails to delete the stack", func() {
				It("removes bosh properties from the state", func() {
					fakeAWS.Stacks.SetDeleteStackReturnError(&awsfaker.ErrorResponse{
						HTTPStatusCode:  http.StatusBadRequest,
						AWSErrorCode:    "InvalidRequest",
						AWSErrorMessage: "failed to delete stack",
					})
					session := destroy(fakeAWSServer.URL, tempDirectory, 1)
					Expect(session.Out.Contents()).To(ContainSubstring("step: destroying bosh director"))
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
					Expect(session.Out.Contents()).To(ContainSubstring("No infrastructure found, skipping..."))
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
			})
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

		Context("when destroy fails to delete the keypair", func() {
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

		Context("when the bosh cli version is <2.0", func() {
			BeforeEach(func() {
				fakeBOSHCLIBackendServer.SetVersion("1.9.0")
			})

			It("fast fails with a helpful error message", func() {
				session := destroy(fakeAWSServer.URL, tempDirectory, 1)

				Expect(session.Err.Contents()).To(ContainSubstring("BOSH version must be at least v2.0.0"))
			})
		})

		Context("when bosh fails", func() {
			BeforeEach(func() {
				fakeBOSHCLIBackendServer.SetDeleteEnvFastFail(true)

				destroy(fakeAWSServer.URL, tempDirectory, 1)
			})

			It("stores a partial bosh state", func() {
				state := readStateJson(tempDirectory)
				Expect(state.BOSH.State).To(Equal(map[string]interface{}{
					"partial": "bosh-state",
				}))
			})
		})
	})

	Context("when the state contains empty TFState and Stack", func() {
		var state storage.State

		BeforeEach(func() {
			state = storage.State{
				Version: 3,
				IAAS:    "aws",
				AWS: storage.AWS{
					AccessKeyID:     "some-access-key",
					SecretAccessKey: "some-access-secret",
					Region:          "some-region",
				},
			}

			writeStateJson(state, tempDirectory)
		})

		It("succeeds and prints a helpful message", func() {
			session := destroy(fakeAWSServer.URL, tempDirectory, 0)

			Expect(session.Out.Contents()).To(ContainSubstring("No infrastructure found, skipping..."))
		})
	})
})

func destroy(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	return destroyWithOptionalDebug(serverURL, tempDirectory, exitCode, true)
}

func destroyWithOptionalDebug(serverURL string, tempDirectory string, exitCode int, debug bool) *gexec.Session {
	var args []string
	if debug {
		args = []string{
			fmt.Sprintf("--endpoint-override=%s", serverURL),
			"--debug",
			"--state-dir", tempDirectory,
			"destroy",
		}
	} else {
		args = []string{
			fmt.Sprintf("--endpoint-override=%s", serverURL),
			"--state-dir", tempDirectory,
			"destroy",
		}
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
