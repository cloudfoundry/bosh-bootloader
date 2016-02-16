package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl", func() {
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
				"unsupported-create-bosh-aws-keypair",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(wasCalled).To(BeTrue())
		})

		Describe("when AWS credentials are not provided", func() {
			PIt("errors", func() {
				session, err := gexec.Start(exec.Command(pathToBBL, "unsupported-create-bosh-aws-keypair"), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err.Contents()).To(ContainSubstring("aws credentials must be provided"))
				Expect(session.Out.Contents()).To(ContainSubstring("Usage"))
			})
		})
	})
})
