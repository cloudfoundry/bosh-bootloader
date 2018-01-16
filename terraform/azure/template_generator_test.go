package azure_test

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/azure"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pmezard/go-difflib/difflib"
)

var _ = Describe("TemplateGenerator", func() {
	var (
		templateGenerator azure.TemplateGenerator
		expectedTemplate  string
		lb                storage.LB
	)

	BeforeEach(func() {
		templateGenerator = azure.NewTemplateGenerator()
	})

	Describe("Generate", func() {
		Context("when no lb type is provided", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("vars", "resource_group", "network", "storage", "network_security_group", "output", "tls")
			})
			It("uses the base template", func() {
				template := templateGenerator.Generate(storage.State{})
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a CF lb type is provided with no system domain", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("vars", "resource_group", "network", "storage", "network_security_group", "output", "tls", "cf_lb")
				lb = storage.LB{
					Type: "cf",
				}
			})
			It("adds the lb subnet, cf lb, ssl cert and iso seg to the base template", func() {
				template := templateGenerator.Generate(storage.State{LB: lb})
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a concourse lb type is provided", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("vars", "resource_group", "network", "storage", "network_security_group", "output", "tls", "concourse_lb")
				lb = storage.LB{
					Type: "concourse",
				}
			})

			It("adds the lb subnet, concourse lb and iso seg to the base template", func() {
				template := templateGenerator.Generate(storage.State{LB: lb})
				checkTemplate(template, expectedTemplate)
			})
		})
	})
})

func expectTemplate(parts ...string) string {
	var contents []string
	for _, p := range parts {
		content, err := ioutil.ReadFile(fmt.Sprintf("templates/%s.tf", p))
		Expect(err).NotTo(HaveOccurred())
		contents = append(contents, string(content))
	}
	return strings.Join(contents, "\n")
}

func checkTemplate(actual, expected string) {
	if actual != string(expected) {
		diff, _ := difflib.GetContextDiffString(difflib.ContextDiff{
			A:        difflib.SplitLines(actual),
			B:        difflib.SplitLines(string(expected)),
			FromFile: "actual",
			ToFile:   "expected",
			Context:  10,
		})
		fmt.Println(diff)
	}

	Expect(actual).To(Equal(string(expected)))
}
