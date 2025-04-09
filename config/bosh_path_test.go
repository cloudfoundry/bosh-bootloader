package config_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/bosh-bootloader/config"
)

var _ = Describe("GetBOSHPath", func() {
	var (
		originalPath string
		pathToBOSH   string
	)

	BeforeEach(func() {
		originalPath = os.Getenv("PATH")

		tempDir, err := os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		pathToBOSH = filepath.Join(tempDir, "bosh")

		err = os.WriteFile(pathToBOSH, []byte("fake"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath) //nolint:errcheck
		gexec.CleanupBuildArtifacts()
	})

	Context("when a user has bosh", func() {
		It("returns bosh", func() {
			os.Setenv("PATH", filepath.Dir(pathToBOSH)) //nolint:errcheck

			boshPath, err := config.GetBOSHPath()
			Expect(boshPath).To(Equal("bosh"))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when a user has bosh2", func() {
		var pathToBOSH2 string
		BeforeEach(func() {
			pathToBOSH2 = filepath.Join(filepath.Dir(pathToBOSH), "bosh2")
			err := os.Rename(pathToBOSH, pathToBOSH2)
			Expect(err).NotTo(HaveOccurred())

			err = os.Setenv("PATH", filepath.Dir(pathToBOSH2))
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns bosh2", func() {
			boshPath, err := config.GetBOSHPath()
			Expect(boshPath).To(Equal(pathToBOSH2))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
