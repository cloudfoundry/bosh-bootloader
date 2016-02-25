package unsupported_test

import (
	"bytes"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrintConcourseAWSTemplate", func() {
	Describe("Execute", func() {
		var (
			stdout  *bytes.Buffer
			builder *fakes.TemplateBuilder
			command unsupported.PrintConcourseAWSTemplate
		)

		BeforeEach(func() {
			stdout = bytes.NewBuffer([]byte{})
			builder = &fakes.TemplateBuilder{}
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
			_, err := command.Execute(commands.GlobalFlags{}, state.State{})
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

		It("returns the given state unmodified", func() {
			s, err := command.Execute(commands.GlobalFlags{}, state.State{
				Version: 10,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal(state.State{
				Version: 10,
			}))
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

					_, err := command.Execute(commands.GlobalFlags{}, state.State{})
					Expect(err).To(MatchError(ContainSubstring("unsupported type: func() string")))
				})
			})
		})
	})
})
