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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl", func() {
	var tempDir string
	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("bbl unsupported-create-bosh-aws-keypair", func() {
		It("generates a RSA key and uploads it to AWS", func() {
			var wasCalled bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				wasCalled = true
				Expect(r.Method).To(Equal("POST"))

				body, err := ioutil.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(body)).To(ContainSubstring("Action=ImportKeyPair"))
				Expect(string(body)).To(ContainSubstring("KeyName=keypair-"))
			}))

			args := []string{
				fmt.Sprintf("--endpoint-override=%s", server.URL),
				"--aws-access-key-id", "some-access-key",
				"--aws-secret-access-key", "some-access-secret",
				"--aws-region", "some-region",
				"--state-dir", tempDir,
				"unsupported-create-bosh-aws-keypair",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(wasCalled).To(BeTrue())
		})

		Describe("when new AWS credentials are provided", func() {
			It("stores the credentials", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				}))

				By("supplying the aws credentials", func() {
					args := []string{
						fmt.Sprintf("--endpoint-override=%s", server.URL),
						"--aws-access-key-id", "some-access-key",
						"--aws-secret-access-key", "some-access-secret",
						"--aws-region", "some-region",
						"--state-dir", tempDir,
						"unsupported-create-bosh-aws-keypair",
					}

					session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0))
				})

				By("clearing out the keypair information", func() {
					buf, err := ioutil.ReadFile(filepath.Join(tempDir, "state.json"))
					Expect(err).NotTo(HaveOccurred())

					var state map[string]interface{}
					err = json.Unmarshal(buf, &state)
					Expect(err).NotTo(HaveOccurred())

					delete(state, "keyPair")

					buf, err = json.Marshal(state)
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(tempDir, "state.json"), buf, os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
				})

				By("reading the aws credentials from the state dir", func() {
					args := []string{
						fmt.Sprintf("--endpoint-override=%s", server.URL),
						"--state-dir", tempDir,
						"unsupported-create-bosh-aws-keypair",
					}

					session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0))
				})
			})
		})

		Describe("when AWS credentials have not been provided", func() {
			It("errors", func() {
				tempDir, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				args := []string{
					"--state-dir", tempDir,
					"unsupported-create-bosh-aws-keypair",
				}

				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("aws access key id must be provided"))
			})
		})
	})
})
