package storage_test

import (
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("patch detector", func() {
	var (
		logger *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
	})

	It("logs the relevant files that it finds", func() {
		err := storage.NewPatchDetector("fixtures/patched", logger).Find()
		Expect(err).NotTo(HaveOccurred())
		Expect(logger.PrintlnCall.Messages).To(HaveLen(2))
		Expect(logger.PrintfCall.Messages).To(HaveLen(7))
		Expect(logger.PrintlnCall.Messages).To(ContainElement(ContainSubstring("you've supplied the following files to bbl:")))
		Expect(logger.PrintfCall.Messages).To(ContainElement(ContainSubstring("create-director-override.sh")))
		Expect(logger.PrintfCall.Messages).To(ContainElement(ContainSubstring("create-jumpbox-override.sh")))
		Expect(logger.PrintfCall.Messages).To(ContainElement(ContainSubstring("delete-director-override.sh")))
		Expect(logger.PrintfCall.Messages).To(ContainElement(ContainSubstring("delete-jumpbox-override.sh")))
		Expect(logger.PrintfCall.Messages).To(ContainElement(ContainSubstring("terraform/patched-terraform.tf")))
		Expect(logger.PrintfCall.Messages).To(ContainElement(ContainSubstring("vars/patched-vars.tfvars")))
		Expect(logger.PrintfCall.Messages).To(ContainElement(ContainSubstring("cloud-config/patch-cloud-config.yml")))
		Expect(logger.PrintlnCall.Messages).NotTo(ContainElement(ContainSubstring("UNRELATED_FILE.md")))
		Expect(logger.PrintlnCall.Messages).NotTo(ContainElement(ContainSubstring("fixtures")))
		Expect(logger.PrintlnCall.Messages).To(ContainElement(ContainSubstring("they will be used by \"bbl up\".")))
	})

	Context("given an unpatched directory", func() {
		It("is silent", func() {
			err := storage.NewPatchDetector("fixtures/unpatched", logger).Find()
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.PrintlnCall.Messages).To(BeEmpty())
		})
	})

	Context("given an unpatched, upped directory", func() {
		It("is silent", func() {
			err := storage.NewPatchDetector("fixtures/upped", logger).Find()
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.PrintlnCall.Messages).To(BeEmpty())
		})
	})

	Context("given an empty directory", func() {
		It("is silent", func() {
			err := storage.NewPatchDetector("fixtures/empty", logger).Find()
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.PrintlnCall.Messages).To(BeEmpty())
		})
	})
})
