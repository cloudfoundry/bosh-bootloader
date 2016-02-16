package unsupported_test

import (
	"bytes"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeTemplateBuilder struct {
	BuildCall struct {
		Returns struct {
			Template cloudformation.Template
		}
	}
}

func (b FakeTemplateBuilder) Build() cloudformation.Template {
	return b.BuildCall.Returns.Template
}

var _ = Describe("PrintConcourseAWSTemplate", func() {
	Describe("Execute", func() {
		var (
			stdout  *bytes.Buffer
			builder *FakeTemplateBuilder
			command unsupported.PrintConcourseAWSTemplate
		)

		BeforeEach(func() {
			stdout = bytes.NewBuffer([]byte{})
			builder = &FakeTemplateBuilder{}
			command = unsupported.NewPrintConcourseAWSTemplate(stdout, builder)

			builder.BuildCall.Returns.Template = cloudformation.Template{
				AWSTemplateFormatVersion: "some-template-version",
				Description:              "some-description",
				Parameters: map[string]cloudformation.Parameter{
					"some-parameter": {
						Type:        "some-type",
						Default:     "some-default",
						Description: "some-description",
					},
				},
				Mappings: map[string]interface{}{
					"some-mapping": 42,
				},
				Resources: map[string]cloudformation.Resource{
					"some-resource": {
						Type:           "some-type",
						Properties:     "some-properties",
						DependsOn:      "some-dependency",
						CreationPolicy: "some-creation-policy",
						UpdatePolicy:   "some-update-policy",
						DeletionPolicy: "some-deletion-policy",
					},
				},
			}
		})

		It("prints a CloudFormation template", func() {
			err := command.Execute(commands.GlobalFlags{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.String()).To(MatchJSON(`{
				"AWSTemplateFormatVersion": "some-template-version",
				"Description": "some-description",
				"Parameters": {
					"some-parameter": {
						"Type": "some-type",
						"Default": "some-default",
						"Description": "some-description"
					}
				},
				"Mappings": {
					"some-mapping": 42
				},
				"Resources": {
					"some-resource": {
						"Type": "some-type",
						"Properties": "some-properties",
						"DependsOn": "some-dependency",
						"CreationPolicy": "some-creation-policy",
						"UpdatePolicy": "some-update-policy",
						"DeletionPolicy": "some-deletion-policy"
					}
				}
			}`))
		})

		Context("failure cases", func() {
			Context("when the template cannot be marshaled", func() {
				It("returns an error", func() {
					builder.BuildCall.Returns.Template = cloudformation.Template{
						AWSTemplateFormatVersion: "some-template-version",
						Description:              "some-description",
						Mappings: map[string]interface{}{
							"some-mapping": func() string {
								return "I cannot be marshaled"
							},
						},
					}

					err := command.Execute(commands.GlobalFlags{})
					Expect(err).To(MatchError(ContainSubstring("unsupported type: func() string")))
				})
			})
		})
	})
})
