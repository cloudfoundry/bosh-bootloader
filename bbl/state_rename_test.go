package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("state rename", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error

		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when state.json exists", func() {
		It("renames to bbl-state.json", func() {
			state := []byte("{}")
			err := ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), state, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			args := []string{
				"--state-dir", tempDirectory,
				"create-lbs",
			}

			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session.Out.Contents).Should(ContainSubstring("renaming state.json to bbl-state.json"))
			Expect(filepath.Join(tempDirectory, "state.json")).NotTo(BeAnExistingFile())
			Expect(filepath.Join(tempDirectory, "bbl-state.json")).To(BeAnExistingFile())
		})
	})
})
