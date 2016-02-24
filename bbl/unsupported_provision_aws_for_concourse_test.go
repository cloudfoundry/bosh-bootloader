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
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"
)

var _ = Describe("bbl", func() {
	Describe("unsupported-provision-aws-for-concourse", func() {
		It("creates and applies a cloudformation template", func() {
			tempDir, err := ioutil.TempDir("", "")

			state := state.State{
				KeyPair: &state.KeyPair{
					Name: "some-keypair-name",
				},
			}

			buf, err := json.Marshal(state)
			Expect(err).NotTo(HaveOccurred())

			ioutil.WriteFile(filepath.Join(tempDir, "state.json"), buf, os.ModePerm)

			var wasCalled bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				wasCalled = true
				Expect(r.Method).To(Equal("POST"))

				body, err := ioutil.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(body)).To(ContainSubstring("Action=CreateStack"))
				Expect(string(body)).To(ContainSubstring("StackName=concourse"))
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
			Expect(wasCalled).To(BeTrue())
		})
	})
})
