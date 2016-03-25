package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("fakeboshinit", func() {
	Describe("main", func() {
		var (
			pathToFake   string
			tempDir      string
			manifestData []byte
		)
		BeforeEach(func() {
			var err error
			pathToFake, err = gexec.Build("github.com/pivotal-cf-experimental/bosh-bootloader/bbl/fakeboshinit")
			Expect(err).NotTo(HaveOccurred())

			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			manifestData = []byte(`---
name:
  bosh-init
`)
		})

		It("reads and prints the bosh state", func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("{}"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir)
			Expect(session.Out.Contents()).To(ContainSubstring(`bosh-state.json: {}`))
		})

		It("skips deployment on a second deploy with same manifest and bosh state", func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("{}"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir)
			session = runFakeBoshInit(pathToFake, tempDir)

			Expect(session.Out.Contents()).To(ContainSubstring("No new changes, skipping deployment..."))
		})

		It("updates deployment on a second deploy with different manifest", func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("{}"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir)

			manifestData = []byte(`---
name:
  bosh-init-updated
`)
			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session = runFakeBoshInit(pathToFake, tempDir)

			Expect(session.Out.Contents()).NotTo(ContainSubstring("No new changes, skipping deployment..."))
		})
	})
})

func runFakeBoshInit(pathToFake, tempDir string) *gexec.Session {
	cmd := &exec.Cmd{
		Path: pathToFake,
		Args: []string{
			filepath.Base(pathToFake),
			"deploy",
			"bosh.yml",
		},
		Dir:    tempDir,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(0))

	return session
}
