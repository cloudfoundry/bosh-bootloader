package main_test

import (
	"fmt"
	"os/exec"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl", func() {
	Describe("version", func() {
		Describe("bbl version", func() {
			It("prints out the version 'dev' if not built with an ldflag", func() {
				session, err := gexec.Start(exec.Command(pathToBBL, "version"), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("bbl dev"))
			})

			Context("bbl provided version", func() {
				var (
					pathToBBL string
				)

				BeforeEach(func() {
					var err error
					pathToBBL, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl",
						"--ldflags", "-X main.Version=1.2.3")
					Expect(err).NotTo(HaveOccurred())
				})

				It("prints out the version passed into the build process via LDFlags", func() {
					session, err := gexec.Start(exec.Command(pathToBBL, "version"), GinkgoWriter, GinkgoWriter)

					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0))
					Expect(session.Out.Contents()).To(ContainSubstring("bbl 1.2.3"))
					Expect(session.Out.Contents()).To(ContainSubstring(fmt.Sprintf("(%s/%s)", runtime.GOOS, runtime.GOARCH)))
				})
			})
		})

		Describe("bbl --version", func() {
			It("prints out the version 'dev' if not built with an ldflag", func() {
				session, err := gexec.Start(exec.Command(pathToBBL, "--version"), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(ContainSubstring("bbl dev"))
			})

			Context("bbl provided version", func() {
				var (
					pathToBBL string
				)

				BeforeEach(func() {
					var err error
					pathToBBL, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl",
						"--ldflags", "-X main.Version=1.2.3")
					Expect(err).NotTo(HaveOccurred())
				})

				It("prints out the version passed into the build process via LDFlags", func() {
					session, err := gexec.Start(exec.Command(pathToBBL, "--version"), GinkgoWriter, GinkgoWriter)

					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0))
					Expect(session.Out.Contents()).To(ContainSubstring("bbl 1.2.3"))
					Expect(session.Out.Contents()).To(ContainSubstring(fmt.Sprintf("(%s/%s)", runtime.GOOS, runtime.GOARCH)))
				})
			})

		})

	})
})
