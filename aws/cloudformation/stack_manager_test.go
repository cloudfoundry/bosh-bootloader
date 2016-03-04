package cloudformation_test

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackManager", func() {
	var (
		cloudformationClient *fakes.CloudFormationClient
		logger               *fakes.Logger
		manager              cloudformation.StackManager
	)

	BeforeEach(func() {
		cloudformationClient = &fakes.CloudFormationClient{}
		logger = &fakes.Logger{}
		manager = cloudformation.NewStackManager(logger)
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

			template := templates.Template{
				Description: "testing template",
			}

			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudformationClient.CreateStackCall.Receives.Input).To(Equal(&awscloudformation.CreateStackInput{
				StackName:    aws.String("some-stack-name"),
				Capabilities: []*string{aws.String("CAPABILITY_IAM")},
				TemplateBody: aws.String(string(templateJson)),
			}))

			Expect(logger.StepCall.Receives.Message).To(Equal("creating cloudformation stack"))
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

			template := templates.Template{
				Description: "testing template",
			}

			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudformationClient.UpdateStackCall.Receives.Input).To(Equal(&awscloudformation.UpdateStackInput{
				StackName:    aws.String("some-stack-name"),
				TemplateBody: aws.String(string(templateJson)),
			}))

			Expect(logger.StepCall.Receives.Message).To(Equal("updating cloudformation stack"))
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

			template := templates.Template{
				Description: "testing template",
			}

			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudformationClient.UpdateStackCall.Receives.Input).To(Equal(&awscloudformation.UpdateStackInput{
				StackName:    aws.String("some-stack-name"),
				TemplateBody: aws.String(string(templateJson)),
			}))
		})

		Context("failure cases", func() {
			It("returns an error when the stack cannot be described", func() {
				cloudformationClient.DescribeStacksCall.Returns.Error = errors.New("error describing stack")

				template := templates.Template{
					Description: "testing template",
				}

				err := manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
				Expect(err).To(MatchError("error describing stack"))
			})

			It("returns an error when the stack cannot be created", func() {
				cloudformationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
				cloudformationClient.CreateStackCall.Returns.Error = errors.New("error creating stack")

				template := templates.Template{
					Description: "testing template",
				}

				err := manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
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

				template := templates.Template{
					Description: "testing template",
				}

				err := manager.CreateOrUpdate(cloudformationClient, "some-stack-name", template)
				Expect(err).To(MatchError("error updating stack"))
			})
		})
	})

	Describe("WaitForCompletion", func() {
		DescribeTable("waiting for a done state", func(startState, endState string, callCount int) {
			cloudformationClient.DescribeStacksCall.Stub = func(input *awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error) {
				status := startState
				if cloudformationClient.DescribeStacksCall.CallCount > 2 {
					status = endState
				}

				return &awscloudformation.DescribeStacksOutput{
					Stacks: []*awscloudformation.Stack{
						{
							StackName:   aws.String("some-stack-name"),
							StackStatus: aws.String(status),
						},
					},
				}, nil
			}

			err := manager.WaitForCompletion(cloudformationClient, "some-stack-name", 0*time.Millisecond)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudformationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))
			Expect(cloudformationClient.DescribeStacksCall.CallCount).To(Equal(callCount))
			Expect(logger.DotCall.CallCount).To(Equal(callCount - 1))
			Expect(logger.StepCall.Receives.Message).To(Equal("finished applying cloudformation template"))
		},

			Entry("create succeeded",
				awscloudformation.StackStatusCreateInProgress, awscloudformation.StackStatusCreateComplete, 3),

			Entry("create failed",
				awscloudformation.StackStatusCreateInProgress, awscloudformation.StackStatusCreateFailed, 3),

			Entry("rollback complete",
				awscloudformation.StackStatusCreateInProgress, awscloudformation.StackStatusRollbackComplete, 3),

			Entry("rollback failed",
				awscloudformation.StackStatusCreateInProgress, awscloudformation.StackStatusRollbackFailed, 3),

			Entry("update succeeded",
				awscloudformation.StackStatusUpdateInProgress, awscloudformation.StackStatusUpdateComplete, 3),

			Entry("update failed, rollback succeeded",
				awscloudformation.StackStatusUpdateInProgress, awscloudformation.StackStatusUpdateRollbackComplete, 3),

			Entry("update failed, rollback failed",
				awscloudformation.StackStatusUpdateInProgress, awscloudformation.StackStatusUpdateRollbackFailed, 3))
	})

	Context("failures cases", func() {
		Context("when the describe stacks call fails", func() {
			It("returns an error", func() {
				cloudformationClient.DescribeStacksCall.Returns.Error = errors.New("failed to describe stack")

				err := manager.WaitForCompletion(cloudformationClient, "some-stack-name", 0*time.Millisecond)
				Expect(err).To(MatchError("failed to describe stack"))
			})
		})
	})
})
