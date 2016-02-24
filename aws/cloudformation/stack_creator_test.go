package cloudformation_test

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackCreator", func() {
	var (
		cloudformationClient *fakes.CloudFormationClient
		creator              cloudformation.StackCreator
		templateBuilder      cloudformation.TemplateBuilder
	)

	BeforeEach(func() {
		cloudformationClient = &fakes.CloudFormationClient{}
		creator = cloudformation.NewStackCreator()
		templateBuilder = cloudformation.NewTemplateBuilder()
	})

	Describe("Create", func() {
		It("creates a stack given a name and cloudformation template", func() {
			template := templateBuilder.Build()
			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = creator.Create(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.CreateStackCall.Receives.CreateStackInput).To(Equal(&awscloudformation.CreateStackInput{
				StackName:    aws.String("some-stack-name"),
				TemplateBody: aws.String(string(templateJson)),
			}))
		})

		Context("error cases", func() {
			It("returns an error when the stack cannot be created", func() {
				cloudformationClient.CreateStackCall.Returns.Error = errors.New("something bad happened")

				err := creator.Create(cloudformationClient, "some-stack-name", templateBuilder.Build())
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
