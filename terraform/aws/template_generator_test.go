package aws_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/aws"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateGenerator", func() {
	var (
		templateGenerator aws.TemplateGenerator
		expectedTemplate  string
		lb                storage.LB
	)

	BeforeEach(func() {
		templateGenerator = aws.NewTemplateGenerator()
	})

	Describe("Generate", func() {
		Context("when no lb type is provided", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("base", "iam", "vpc")
			})

			It("uses the base template", func() {
				template := templateGenerator.Generate(storage.State{})
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a concourse lb type is provided", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("base", "iam", "vpc", "lb_subnet", "concourse_lb")
				lb = storage.LB{
					Type: "concourse",
				}
			})
			It("adds the lb subnet and concourse lb to the base template", func() {
				template := templateGenerator.Generate(storage.State{LB: lb})
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a CF lb type is provided with no system domain", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("base", "iam", "vpc", "lb_subnet", "cf_lb", "cf_lb_common", "ssl_certificate", "iso_segments")
				lb = storage.LB{
					Type: "cf",
				}
			})
			It("adds the lb subnet, cf lb, ssl cert and iso seg to the base template", func() {
				template := templateGenerator.Generate(storage.State{LB: lb})
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a CF lb type is provided with a system domain", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("base", "iam", "vpc", "lb_subnet", "cf_lb", "cf_lb_common", "ssl_certificate", "iso_segments", "cf_dns")
				lb = storage.LB{
					Type:   "cf",
					Domain: "some-domain",
				}
			})
			It("adds the domain", func() {
				template := templateGenerator.Generate(storage.State{LB: lb})
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a nlb type is provided with no system domain", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("base", "iam", "vpc", "lb_subnet", "cf_nlb", "cf_lb_common", "ssl_certificate", "iso_segments")
				lb = storage.LB{
					Type: "nlb",
				}
			})
			It("adds the nlb lb templates without DNS", func() {
				template := templateGenerator.Generate(storage.State{LB: lb})
				checkTemplate(template, expectedTemplate)
			})
		})

		Context("when a nlb type is provided with a system domain", func() {
			BeforeEach(func() {
				expectedTemplate = expectTemplate("base", "iam", "vpc", "lb_subnet", "cf_nlb", "cf_lb_common", "ssl_certificate", "iso_segments", "cf_nlb_dns")
				lb = storage.LB{
					Type:   "nlb",
					Domain: "some-domain",
				}
			})
			It("adds the nlb DNS template that references aws_lb resources", func() {
				template := templateGenerator.Generate(storage.State{LB: lb})
				checkTemplate(template, expectedTemplate)
			})
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
