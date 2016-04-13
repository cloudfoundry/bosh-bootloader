package cloudformation_test

import (
	"encoding/json"
	"errors"
	"fmt"
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
		cloudFormationClient *fakes.CloudFormationClient
		logger               *fakes.Logger
		manager              cloudformation.StackManager
	)

	BeforeEach(func() {
		cloudFormationClient = &fakes.CloudFormationClient{}
		logger = &fakes.Logger{}
		manager = cloudformation.NewStackManager(logger)
	})

	Describe("Describe", func() {
		It("describes the stack with the given name", func() {
			cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
				Stacks: []*awscloudformation.Stack{{
					StackName:   aws.String("some-stack-name"),
					StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
					Outputs: []*awscloudformation.Output{{
						OutputKey:   aws.String("some-output-key"),
						OutputValue: aws.String("some-output-value"),
					}},
				}},
			}
			stack, err := manager.Describe(cloudFormationClient, "some-stack-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))
			Expect(stack).To(Equal(cloudformation.Stack{
				Name:   "some-stack-name",
				Status: "UPDATE_COMPLETE",
				Outputs: map[string]string{
					"some-output-key": "some-output-value",
				},
			}))
		})

		Context("failure cases", func() {
			Context("when there is a response error", func() {
				It("returns an error when the response code is not a 400", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("something bad happened")), 500, "0")
					_, err := manager.Describe(cloudFormationClient, "some-stack-name")
					Expect(err).To(MatchError(ContainSubstring("something bad happened")))
				})
			})

			Context("when stack output key or value is nil", func() {
				It("returns an error when the key in nil", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
						Stacks: []*awscloudformation.Stack{
							{
								StackName:   aws.String("some-stack-name"),
								StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
								Outputs: []*awscloudformation.Output{{
									OutputKey:   nil,
									OutputValue: aws.String("some-value"),
								}},
							},
						},
					}

					_, err := manager.Describe(cloudFormationClient, "some-stack-name")
					Expect(err).To(MatchError("failed to parse outputs"))
				})

				It("assigns an empty string value when the value is nil", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
						Stacks: []*awscloudformation.Stack{
							{
								StackName:   aws.String("some-stack-name"),
								StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
								Outputs: []*awscloudformation.Output{
									{
										OutputKey:   aws.String("first-key"),
										OutputValue: nil,
									},
									{
										OutputKey:   aws.String("second-key"),
										OutputValue: aws.String("second-value"),
									},
								},
							},
						},
					}

					stack, err := manager.Describe(cloudFormationClient, "some-stack-name")
					Expect(err).NotTo(HaveOccurred())

					Expect(stack.Outputs["first-key"]).To(Equal(""))
					Expect(stack.Outputs["second-key"]).To(Equal("second-value"))
				})
			})

			It("returns a StackNotFound error when the stack doesn't exist", func() {
				cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
				_, err := manager.Describe(cloudFormationClient, "some-stack-name")
				Expect(err).To(MatchError(cloudformation.StackNotFound))
			})

			It("returns a StackNotFound error when the stack name is empty", func() {
				cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
				_, err := manager.Describe(cloudFormationClient, "")
				Expect(err).To(MatchError(cloudformation.StackNotFound))

				Expect(cloudFormationClient.DescribeStacksCall.CallCount).To(Equal(0))
			})

			Context("when the api returns no stacks", func() {
				It("returns a StackNotFound error", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
						Stacks: []*awscloudformation.Stack{},
					}

					_, err := manager.Describe(cloudFormationClient, "some-stack-name")
					Expect(err).To(MatchError(cloudformation.StackNotFound))
				})
			})
			Context("when the api returns more than one stack", func() {
				It("returns the correct stack", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
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
					stack, err := manager.Describe(cloudFormationClient, "some-stack-name")
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudFormationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
						StackName: aws.String("some-stack-name"),
					}))
					Expect(stack).To(Equal(cloudformation.Stack{
						Name:    "some-stack-name",
						Status:  "CREATE_COMPLETE",
						Outputs: map[string]string{},
					}))
				})
			})

			Context("when the api returns a stack missing a name", func() {
				It("doesn't explode", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
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
					stack, err := manager.Describe(cloudFormationClient, "some-stack-name")
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudFormationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
						StackName: aws.String("some-stack-name"),
					}))
					Expect(stack).To(Equal(cloudformation.Stack{
						Name:    "some-stack-name",
						Status:  "CREATE_COMPLETE",
						Outputs: map[string]string{},
					}))
				})
			})

			Context("when the api returns a stack missing a status", func() {
				It("doesn't explode", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
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
					stack, err := manager.Describe(cloudFormationClient, "some-stack-name")
					Expect(err).NotTo(HaveOccurred())

					Expect(cloudFormationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
						StackName: aws.String("some-stack-name"),
					}))
					Expect(stack).To(Equal(cloudformation.Stack{
						Name:    "some-stack-name",
						Status:  "UNKNOWN",
						Outputs: map[string]string{},
					}))
				})
			})
		})
	})

	Describe("CreateOrUpdate", func() {
		It("creates a stack if the stack does not exist", func() {
			cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")

			template := templates.Template{
				Description: "testing template",
			}

			templateJson, err := json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			err = manager.CreateOrUpdate(cloudFormationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudFormationClient.CreateStackCall.Receives.Input).To(Equal(&awscloudformation.CreateStackInput{
				StackName:    aws.String("some-stack-name"),
				Capabilities: []*string{aws.String("CAPABILITY_IAM")},
				TemplateBody: aws.String(string(templateJson)),
			}))

			Expect(logger.StepCall.Receives.Message).To(Equal("creating cloudformation stack"))
		})

		It("updates the stack if the stack exists", func() {
			cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
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

			err = manager.CreateOrUpdate(cloudFormationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudFormationClient.UpdateStackCall.Receives.Input).To(Equal(&awscloudformation.UpdateStackInput{
				StackName:    aws.String("some-stack-name"),
				Capabilities: []*string{aws.String("CAPABILITY_IAM")},
				TemplateBody: aws.String(string(templateJson)),
			}))

			Expect(logger.StepCall.Receives.Message).To(Equal("updating cloudformation stack"))
		})

		It("does not return an error when no updates are to be performed", func() {
			cloudFormationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(awserr.New("ValidationError", "No updates are to be performed.", errors.New("")), 400, "0")
			cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
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

			err = manager.CreateOrUpdate(cloudFormationClient, "some-stack-name", template)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(cloudFormationClient.UpdateStackCall.Receives.Input).To(Equal(&awscloudformation.UpdateStackInput{
				StackName:    aws.String("some-stack-name"),
				Capabilities: []*string{aws.String("CAPABILITY_IAM")},
				TemplateBody: aws.String(string(templateJson)),
			}))
		})

		Context("failure cases", func() {
			It("returns an error when the stack fails to update", func() {
				cloudFormationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(awserr.New("some-error", "something bad happened", errors.New("")), 400, "0")
				cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
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

				err := manager.CreateOrUpdate(cloudFormationClient, "some-stack-name", template)
				Expect(err).To(MatchError(ContainSubstring(("something bad happened"))))
			})

			It("returns an error when the stack cannot be described", func() {
				cloudFormationClient.DescribeStacksCall.Returns.Error = errors.New("error describing stack")

				template := templates.Template{
					Description: "testing template",
				}

				err := manager.CreateOrUpdate(cloudFormationClient, "some-stack-name", template)
				Expect(err).To(MatchError("error describing stack"))
			})

			It("returns an error when the stack cannot be created", func() {
				cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
				cloudFormationClient.CreateStackCall.Returns.Error = errors.New("error creating stack")

				template := templates.Template{
					Description: "testing template",
				}

				err := manager.CreateOrUpdate(cloudFormationClient, "some-stack-name", template)
				Expect(err).To(MatchError("error creating stack"))
			})

			It("returns an error when the stack cannot be updated", func() {
				cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
					Stacks: []*awscloudformation.Stack{
						{
							StackName:   aws.String("some-stack-name"),
							StackStatus: aws.String(awscloudformation.StackStatusUpdateComplete),
						},
					},
				}
				cloudFormationClient.UpdateStackCall.Returns.Error = errors.New("error updating stack")

				template := templates.Template{
					Description: "testing template",
				}

				err := manager.CreateOrUpdate(cloudFormationClient, "some-stack-name", template)
				Expect(err).To(MatchError("error updating stack"))
			})
		})
	})

	Describe("WaitForCompletion", func() {
		var stubDescribeStacksCall = func(startState string, endState string, cloudFormationClient *fakes.CloudFormationClient) {
			cloudFormationClient.DescribeStacksCall.Stub = func(input *awscloudformation.DescribeStacksInput) (*awscloudformation.DescribeStacksOutput, error) {
				status := startState
				if cloudFormationClient.DescribeStacksCall.CallCount > 2 {
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
		}

		DescribeTable("waiting for a done state", func(startState, endState string, action string) {
			stubDescribeStacksCall(startState, endState, cloudFormationClient)

			err := manager.WaitForCompletion(cloudFormationClient, "some-stack-name", 0*time.Millisecond, action)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClient.DescribeStacksCall.Receives.Input).To(Equal(&awscloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))
			Expect(cloudFormationClient.DescribeStacksCall.CallCount).To(Equal(3))
			Expect(logger.DotCall.CallCount).To(Equal(2))
			Expect(logger.StepCall.Receives.Message).To(Equal(fmt.Sprintf("finished %s", action)))
		},

			Entry("create succeeded",
				awscloudformation.StackStatusCreateInProgress, awscloudformation.StackStatusCreateComplete, "creating stack"),

			Entry("update succeeded",
				awscloudformation.StackStatusUpdateInProgress, awscloudformation.StackStatusUpdateComplete, "updating stack"),

			Entry("delete succeeded",
				awscloudformation.StackStatusDeleteInProgress, awscloudformation.StackStatusDeleteComplete, "deleting stack"),
		)

		DescribeTable("waiting for a done state", func(startState, endState string, action string) {
			stubDescribeStacksCall(startState, endState, cloudFormationClient)

			err := manager.WaitForCompletion(cloudFormationClient, "some-stack-name", 0*time.Millisecond, action)
			Expect(err).To(MatchError(`Failed creating cloudformation stack 'some-stack-name'.
			Check the AWS console for error events related to this stack,
			and/or open a GitHub issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues.`))
		},

			Entry("create failed",
				awscloudformation.StackStatusCreateInProgress, awscloudformation.StackStatusCreateFailed, "creating stack"),

			Entry("rollback complete",
				awscloudformation.StackStatusCreateInProgress, awscloudformation.StackStatusRollbackComplete, "creating stack"),

			Entry("rollback failed",
				awscloudformation.StackStatusCreateInProgress, awscloudformation.StackStatusRollbackFailed, "creating stack"),

			Entry("update failed, rollback succeeded",
				awscloudformation.StackStatusUpdateInProgress, awscloudformation.StackStatusUpdateRollbackComplete, "updating stack"),

			Entry("update failed, rollback failed",
				awscloudformation.StackStatusUpdateInProgress, awscloudformation.StackStatusUpdateRollbackFailed, "updating stack"),

			Entry("delete failed",
				awscloudformation.StackStatusDeleteInProgress, awscloudformation.StackStatusDeleteFailed, "deleting stack"),
		)

		Context("when the stack does not exist", func() {
			It("does not error", func() {
				cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")

				err := manager.WaitForCompletion(cloudFormationClient, "some-stack-name", 0*time.Millisecond, "foo")
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.StepCall.Receives.Message).To(Equal("finished foo"))
			})
		})

		Context("failures cases", func() {
			Context("when the describe stacks call fails", func() {
				It("returns an error", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Error = errors.New("failed to describe stack")

					err := manager.WaitForCompletion(cloudFormationClient, "some-stack-name", 0*time.Millisecond, "foo")
					Expect(err).To(MatchError("failed to describe stack"))
				})
			})
		})
	})

	Describe("Delete", func() {
		It("deletes the stack", func() {
			err := manager.Delete(cloudFormationClient, "some-stack-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClient.DeleteStackCall.Receives.Input).To(Equal(&awscloudformation.DeleteStackInput{
				StackName: aws.String("some-stack-name"),
			}))

			Expect(logger.StepCall.Receives.Message).To(Equal("deleting cloudformation stack"))
		})

		Context("failure cases", func() {
			Context("when the stack delete call fails", func() {
				It("returns an error", func() {
					cloudFormationClient.DeleteStackCall.Returns.Error = errors.New("failed to delete stack")

					err := manager.Delete(cloudFormationClient, "some-stack-name")
					Expect(err).To(MatchError("failed to delete stack"))
				})
			})
		})
	})
})
