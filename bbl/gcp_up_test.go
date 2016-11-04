package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl up gcp", func() {
	var tempDirectory string

	BeforeEach(func() {
		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("writes iaas: gcp to state", func() {
		args := []string{
			"--state-dir", tempDirectory,
			"up",
			"--iaas", "gcp",
		}

		executeCommand(args, 0)

		state := readStateJson(tempDirectory)
		Expect(state.IAAS).To(Equal("gcp"))
	})

	Context("when bbl-state.json contains iaas: gcp", func() {
		BeforeEach(func() {
			buf, err := json.Marshal(storage.State{
				IAAS: "gcp",
			})
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), buf, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not require --iaas flag and exits 0", func() {
			args := []string{
				"--state-dir", tempDirectory,
				"up",
			}

			executeCommand(args, 0)
		})

		Context("when called with --iaas aws", func() {
			It("exits 1 and prints error message", func() {
				session := upAWS("", tempDirectory, 1)

				Expect(session.Err.Contents()).To(ContainSubstring("the iaas provided must match the iaas in bbl-state.json"))
			})
		})
	})
})
