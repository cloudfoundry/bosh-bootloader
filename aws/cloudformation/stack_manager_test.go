package cloudformation_test

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackManager", func() {
	var (
		cloudformationClient *fakes.CloudFormationClient
		manager              cloudformation.StackManager
		templateBuilder      cloudformation.TemplateBuilder
	)

	BeforeEach(func() {
		cloudformationClient = &fakes.CloudFormationClient{}
		manager = cloudformation.NewStackManager()
		templateBuilder = cloudformation.NewTemplateBuilder()
	})

	Describe("Create", func() {
		It("creates a stack given a name and cloudformation template", func() {
			template := templateBuilder.Build()
			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.Create(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.CreateStackCall.Receives.CreateStackInput).To(Equal(&awscloudformation.CreateStackInput{
				StackName:    aws.String("some-stack-name"),
				TemplateBody: aws.String(string(templateJson)),
			}))
		})

		Context("error cases", func() {
			It("returns an error when the stack cannot be created", func() {
				cloudformationClient.CreateStackCall.Returns.Error = errors.New("something bad happened")

				err := manager.Create(cloudformationClient, "some-stack-name", templateBuilder.Build())
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})

	Describe("Update", func() {
		var template cloudformation.Template

		BeforeEach(func() {
			template = templateBuilder.Build()
		})

		It("updates a stack with the given name", func() {
			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.Update(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.UpdateStackCall.Receives.UpdateStackInput).To(Equal(&awscloudformation.UpdateStackInput{
				StackName:    aws.String("some-stack-name"),
				TemplateBody: aws.String(string(templateJson)),
			}))
		})

		Context("when the stack does not need to be updated", func() {
			It("does not return an error", func() {
				cloudformationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
				err := manager.Update(cloudformationClient, "some-stack-name", template)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("error cases", func() {
			It("returns an error when the stack cannot be updated", func() {
				cloudformationClient.UpdateStackCall.Returns.Error = errors.New("something bad happened")

				err := manager.Update(cloudformationClient, "some-stack-name", templateBuilder.Build())
				Expect(err).To(MatchError("something bad happened"))
			})

			Context("when there is a response error", func() {
				It("returns an error when the response code is not a 400", func() {
					cloudformationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("something bad happened")), 500, "0")
					err := manager.Update(cloudformationClient, "some-stack-name", template)
					Expect(err).To(MatchError(ContainSubstring("something bad happened")))
				})
			})
		})
	})

	Describe("Describe", func() {
		It("describes the stack with the given name", func() {
			cloudformationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
				Stacks: []*awscloudformation.Stack{{
					StackName:   aws.String("some-stack-name"),
					StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
				}},
			}
			stack, err := manager.Describe(cloudformationClient, "some-stack-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))
			Expect(stack).To(Equal(cloudformation.Stack{
				Name:   "some-stack-name",
				Status: "UPDATE_COMPLETE",
			}))
		})

		Context("failure cases", func() {
			Context("when there is a response error", func() {
				It("returns an error when the response code is not a 400", func() {
					cloudformationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("something bad happened")), 500, "0")
					_, err := manager.Describe(cloudformationClient, "some-stack-name")
					Expect(err).To(MatchError(ContainSubstring("something bad happened")))
				})
			})

			It("returns a StackNotFound error when the stack doesn't exist", func() {
				cloudformationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
				_, err := manager.Describe(cloudformationClient, "some-stack-name")
				Expect(err).To(MatchError(cloudformation.StackNotFound))
			})

			Context("when the api returns no stacks", func() {
				It("returns a StackNotFound error", func() {
					cloudformationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
						Stacks: []*awscloudformation.Stack{},
					}

					_, err := manager.Describe(cloudformationClient, "some-stack-name")
					Expect(err).To(MatchError(cloudformation.StackNotFound))
				})
			})
			Context("when the api returns more than one stack", func() {
				It("returns the correct stack", func() {
					cloudformationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
						Stacks: []*awscloudformation.Stack{
							{
								StackName:   aws.String("some-other-stack-name"),
								StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
							},
							{
								StackName:   aws.String("some-stack-name"),
								StackStatus: aws.String(awscloudformation.StackStatusCreateComplete),
							},
						},
					}
					stack, err := manager.Describe(cloudformationClient, "some-stack-name")
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
						StackName: aws.String("some-stack-name"),
					}))
					Expect(stack).To(Equal(cloudformation.Stack{
						Name:   "some-stack-name",
						Status: "CREATE_COMPLETE",
					}))
				})
			})

			Context("when the api returns a stack missing a name", func() {
				It("doesn't explode", func() {
					cloudformationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
						Stacks: []*awscloudformation.Stack{
							{
								StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
							},
							{
								StackName:   aws.String("some-stack-name"),
								StackStatus: aws.String(awscloudformation.StackStatusCreateComplete),
							},
						},
					}
					stack, err := manager.Describe(cloudformationClient, "some-stack-name")
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
						StackName: aws.String("some-stack-name"),
					}))
					Expect(stack).To(Equal(cloudformation.Stack{
						Name:   "some-stack-name",
						Status: "CREATE_COMPLETE",
					}))
				})
			})

			Context("when the api returns a stack missing a status", func() {
				It("doesn't explode", func() {
					cloudformationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
						Stacks: []*awscloudformation.Stack{
							{
								StackName:   aws.String("some-other-stack-name"),
								StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
							},
							{
								StackName: aws.String("some-stack-name"),
							},
						},
					}
					stack, err := manager.Describe(cloudformationClient, "some-stack-name")
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
						StackName: aws.String("some-stack-name"),
					}))
					Expect(stack).To(Equal(cloudformation.Stack{
						Name:   "some-stack-name",
						Status: "UNKNOWN",
					}))
				})
			})
		})
	})

	Describe("CreateOrUpdate", func() {
		It("creates a stack if the stack does not exist", func() {
			cloudformationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")

			template := templateBuilder.Build()
			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudformationClient.CreateStackCall.Receives.CreateStackInput).To(Equal(&awscloudformation.CreateStackInput{
				StackName:    aws.String("some-stack-name"),
				TemplateBody: aws.String(string(templateJson)),
			}))
		})

		It("updates the stack if the stack exists", func() {
			cloudformationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
				Stacks: []*awscloudformation.Stack{
					{
						StackName:   aws.String("some-stack-name"),
						StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
					},
				},
			}

			template := templateBuilder.Build()
			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudformationClient.UpdateStackCall.Receives.UpdateStackInput).To(Equal(&awscloudformation.UpdateStackInput{
				StackName:    aws.String("some-stack-name"),
				TemplateBody: aws.String(string(templateJson)),
			}))
		})

		It("does not return an error when no updates are to be performed", func() {
			cloudformationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
			cloudformationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
				Stacks: []*awscloudformation.Stack{
					{
						StackName:   aws.String("some-stack-name"),
						StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
					},
				},
			}

			template := templateBuilder.Build()
			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudformationClient.UpdateStackCall.Receives.UpdateStackInput).To(Equal(&awscloudformation.UpdateStackInput{
				StackName:    aws.String("some-stack-name"),
				TemplateBody: aws.String(string(templateJson)),
			}))
		})

		Context("failure cases", func() {
			It("returns an error when the stack cannot be described", func() {
				cloudformationClient.DescribeStacksCall.Returns.Error = errors.New("error describing stack")

				err := manager.CreateOrUpdate(cloudformationClient, "some-stack-name", templateBuilder.Build())
				Expect(err).To(MatchError("error describing stack"))
			})

			It("returns an error when the stack cannot be created", func() {
				cloudformationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
				cloudformationClient.CreateStackCall.Returns.Error = errors.New("error creating stack")

				err := manager.CreateOrUpdate(cloudformationClient, "some-stack-name", templateBuilder.Build())
				Expect(err).To(MatchError("error creating stack"))
			})

			It("returns an error when the stack cannot be updated", func() {
				cloudformationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
					Stacks: []*awscloudformation.Stack{
						{
							StackName:   aws.String("some-stack-name"),
							StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
						},
					},
				}
				cloudformationClient.UpdateStackCall.Returns.Error = errors.New("error updating stack")

				err := manager.CreateOrUpdate(cloudformationClient, "some-stack-name", templateBuilder.Build())
				Expect(err).To(MatchError("error updating stack"))
			})
		})
	})
})
