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
name: bosh-init
jobs:
- name: bosh
  properties:
    director:
      name: my-bosh
  `)
		})

		It("fails fast if compiled with FailFast flag", func() {
			var err error
			pathToFake, err = gexec.Build("github.com/pivotal-cf-experimental/bosh-bootloader/bbl/fakeboshinit",
				"-ldflags",
				"-X main.FailFast=true")
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir, 1)
			Expect(session.Err.Contents()).To(ContainSubstring(`failing fast`))
		})

		It("reads and prints the bosh state", func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("{}"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir, 0)
			Expect(session.Out.Contents()).To(ContainSubstring(`bosh-state.json: {}`))
		})

		It("prints director name from bosh manifest", func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("{}"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir, 0)
			Expect(session.Out.Contents()).To(ContainSubstring(`bosh director name: my-bosh`))
		})

		It("prints no name for the bosh director if it doesn't exist", func() {
			manifestData = []byte(`---
name: bosh-init`)

			err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("{}"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir, 0)
			Expect(session.Out.Contents()).To(ContainSubstring(`bosh director name:`))
		})

		It("skips deployment on a second deploy with same manifest and bosh state", func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("{}"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir, 0)
			session = runFakeBoshInit(pathToFake, tempDir, 0)

			Expect(session.Out.Contents()).To(ContainSubstring("No new changes, skipping deployment..."))
		})

		It("updates deployment on a second deploy with different manifest", func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("{}"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session := runFakeBoshInit(pathToFake, tempDir, 0)

			manifestData = []byte(`---
name: bosh-init-updated
jobs:
- name: bosh
  properties:
    director:
      name: my-bosh
`)
			err = ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), manifestData, os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			session = runFakeBoshInit(pathToFake, tempDir, 0)

			Expect(session.Out.Contents()).NotTo(ContainSubstring("No new changes, skipping deployment..."))
		})
	})
})

func runFakeBoshInit(pathToFake, tempDir string, exitCode int) *gexec.Session {
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
	Eventually(session, 10*time.Second).Should(gexec.Exit(exitCode))

	return session
}
