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
	Describe("bbl -h", func() {
		It("prints out the usage", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, "-h"), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("usage"))
		})
	})

	Describe("bbl -v", func() {
		It("prints out the current version", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, "-v"), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("bbl 0.0.1"))
		})
	})

	Describe("bbl unsupported-print-concourse-aws-template", func() {
		It("prints a CloudFomation template", func() {
			session, err := gexec.Start(exec.Command(pathToBBL, "unsupported-print-concourse-aws-template"), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			buf, err := ioutil.ReadFile("../cloudformation/fixtures/cloudformation.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(session.Out.Contents()).To(MatchJSON(string(buf)))
		})
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
				"--aws-access-key-id=some-access-key",
				"--aws-secret-access-key=some-access-secret",
				"--aws-region=some-region",
				"unsupported-create-bosh-aws-keypair",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(wasCalled).To(BeTrue())
		})
	})

	It("prints an error when an unknown flag is provided", func() {
		session, err := gexec.Start(exec.Command(pathToBBL, "--some-unknown-flag"), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err.Contents()).To(ContainSubstring("flag provided but not defined: -some-unknown-flag"))
		Expect(session.Out.Contents()).To(ContainSubstring("Usage"))
	})

	It("prints an error when an unknown command is provided", func() {
		session, err := gexec.Start(exec.Command(pathToBBL, "some-unknown-command"), GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err.Contents()).To(ContainSubstring("unknown command: some-unknown-command"))
		Expect(session.Out.Contents()).To(ContainSubstring("Usage"))
	})
})
