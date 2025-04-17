package openstack_test

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pmezard/go-difflib/difflib"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/openstack"
)

var _ = Describe("TemplateGenerator", func() {
	var (
		templateGenerator openstack.TemplateGenerator
		expectedTemplate  string
	)

	BeforeEach(func() {
		templateGenerator = openstack.NewTemplateGenerator()
	})

	Describe("Generate", func() {
		BeforeEach(func() {
			expectedTemplate = expectTemplate("provider-vars", "provider", "resources-outputs", "resources-vars", "resources")
		})
		It("uses openstack templates", func() {
			template := templateGenerator.Generate(storage.State{})
			checkTemplate(template, expectedTemplate)
		})
	})
})

func expectTemplate(parts ...string) string {
	var contents []string
	for _, p := range parts {
		content, err := os.ReadFile(fmt.Sprintf("templates/%s.tf", p))
		Expect(err).NotTo(HaveOccurred())
		contents = append(contents, string(content))
	}
	return strings.Join(contents, "\n")
}

func checkTemplate(actual, expected string) {
	if actual != expected {
		diff, _ := difflib.GetContextDiffString(difflib.ContextDiff{ //nolint:errcheck
			A:        difflib.SplitLines(actual),
			B:        difflib.SplitLines(expected),
			FromFile: "actual",
			ToFile:   "expected",
			Context:  10,
		})
		fmt.Println(diff)
	}

	Expect(actual).To(Equal(expected))
}
