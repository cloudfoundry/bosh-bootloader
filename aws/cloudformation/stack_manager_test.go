package cloudformation_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackManager", func() {
	var (
		cloudFormationClient *fakes.CloudFormationClient
		awsClientProvider    *fakes.AWSClientProvider
		logger               *fakes.Logger
		manager              cloudformation.StackManager
	)

	BeforeEach(func() {
		awsClientProvider = &fakes.AWSClientProvider{}
		cloudFormationClient = &fakes.CloudFormationClient{}
		logger = &fakes.Logger{}
		awsClientProvider.GetCloudFormationClientCall.Returns.CloudFormationClient = cloudFormationClient
		manager = cloudformation.NewStackManager(awsClientProvider, logger)
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
			stack, err := manager.Describe("some-stack-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(awsClientProvider.GetCloudFormationClientCall.CallCount).To(Equal(1))

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
				It("returns an error when the RequestFailure response is not a 'StackNotFound'", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("ValidationError", "something bad happened", errors.New("")), 400, "0")
					_, err := manager.Describe("some-stack-name")
					Expect(err).To(MatchError(ContainSubstring("something bad happened")))
				})

				It("returns an error when the response is an unknown error", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Error = errors.New("an unknown error occurred")
					_, err := manager.Describe("some-stack-name")
					Expect(err).To(MatchError(ContainSubstring("an unknown error occurred")))
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

					_, err := manager.Describe("some-stack-name")
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

					stack, err := manager.Describe("some-stack-name")
					Expect(err).NotTo(HaveOccurred())

					Expect(stack.Outputs["first-key"]).To(Equal(""))
					Expect(stack.Outputs["second-key"]).To(Equal("second-value"))
				})
			})

			It("returns a StackNotFound error when the stack doesn't exist", func() {
				stackName := fmt.Sprintf("some-stack-name-%d", rand.Int())
				cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("ValidationError", fmt.Sprintf("Stack with id %s does not exist", stackName), errors.New("")), 400, "0")
				_, err := manager.Describe(stackName)
				Expect(err).To(MatchError(cloudformation.StackNotFound))
			})

			It("returns a StackNotFound error when the stack name is empty", func() {
				cloudFormationClient.DescribeStacksCall.Returns.Error = awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 400, "0")
				_, err := manager.Describe("")
				Expect(err).To(MatchError(cloudformation.StackNotFound))

				Expect(cloudFormationClient.DescribeStacksCall.CallCount).To(Equal(0))
			})

			Context("when the api returns no stacks", func() {
				It("returns a StackNotFound error", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Output = &awscloudformation.DescribeStacksOutput{
						Stacks: []*awscloudformation.Stack{},
					}

					_, err := manager.Describe("some-stack-name")
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
					stack, err := manager.Describe("some-stack-name")
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
					stack, err := manager.Describe("some-stack-name")
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
					stack, err := manager.Describe("some-stack-name")
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

	Describe("Update", func() {
		var (
			template     templates.Template
			templateJson []byte
			tags         cloudformation.Tags
		)

		BeforeEach(func() {
			var err error

			template = templates.Template{
				Description: "testing template",
			}

			templateJson, err = json.Marshal(&template)
			Expect(err).NotTo(HaveOccurred())

			tags = cloudformation.Tags{
				{
					Key:   "bbl-env-id",
					Value: "some-env-id",
				},
			}
		})

		It("updates the stack if the stack exists", func() {
			err := manager.Update("some-stack-name", template, tags)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClient.UpdateStackCall.Receives.Input).To(Equal(&awscloudformation.UpdateStackInput{
				StackName:    aws.String("some-stack-name"),
				Capabilities: []*string{aws.String("CAPABILITY_IAM"), aws.String("CAPABILITY_NAMED_IAM")},
				TemplateBody: aws.String(string(templateJson)),
				Tags: []*awscloudformation.Tag{
					{
						Key:   aws.String("bbl-env-id"),
						Value: aws.String("some-env-id"),
					},
				},
			}))

			Expect(logger.StepCall.Receives.Message).To(Equal("updating cloudformation stack"))
		})

		It("does not return an error when no updates are to be performed", func() {
			cloudFormationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(awserr.New("ValidationError", "No updates are to be performed.", errors.New("")), 400, "0")

			err := manager.Update("some-stack-name", template, cloudformation.Tags{})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			Context("when update stack returns a validation error other than 'No updates to be performed'", func() {
				It("returns error", func() {
					cloudFormationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(awserr.New("ValidationError", "something bad happened", errors.New("")), 400, "0")

					err := manager.Update("some-stack-name", template, cloudformation.Tags{})
					Expect(err).To(MatchError(ContainSubstring("something bad happened")))
				})
			})

			Context("when update stack returns an unknown error", func() {
				It("returns error", func() {
					cloudFormationClient.UpdateStackCall.Returns.Error = errors.New("an unknown error has occurred")

					err := manager.Update("some-stack-name", template, cloudformation.Tags{})
					Expect(err).To(MatchError("an unknown error has occurred"))
				})
			})

			Context("when the stack does not exist", func() {
				It("returns a StackNotFound error", func() {
					stackName := fmt.Sprintf("some-stack-name-%d", rand.Int())
					cloudFormationClient.UpdateStackCall.Returns.Error = awserr.NewRequestFailure(awserr.New("ValidationError", fmt.Sprintf("Stack [%s] does not exist", stackName), errors.New("")), 400, "0")

					err := manager.Update(stackName, template, cloudformation.Tags{})
					Expect(err).To(Equal(cloudformation.StackNotFound))
				})
			})
		})
	})

	Describe("WaitForCompletion", func() {
		var stubDescribeStacksCall = func(startState string, endState string) {
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
			stubDescribeStacksCall(startState, endState)

			err := manager.WaitForCompletion("some-stack-name", 0*time.Millisecond, action)
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
			stubDescribeStacksCall(startState, endState)

			err := manager.WaitForCompletion("some-stack-name", 0*time.Millisecond, action)
			Expect(err).To(MatchError(`CloudFormation failure on stack 'some-stack-name'.
Check the AWS console for error events related to this stack,
and/or open a GitHub issue at https://github.com/cloudfoundry/bosh-bootloader/issues.`))
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
				cloudFormationClient.DescribeStacksCall.Returns.Error = cloudformation.StackNotFound

				err := manager.WaitForCompletion("some-stack-name", 0*time.Millisecond, "foo")
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.StepCall.Receives.Message).To(Equal("finished foo"))
			})
		})

		Context("failures cases", func() {
			Context("when the describe stacks call fails", func() {
				It("returns an error", func() {
					cloudFormationClient.DescribeStacksCall.Returns.Error = errors.New("failed to describe stack")

					err := manager.WaitForCompletion("some-stack-name", 0*time.Millisecond, "foo")
					Expect(err).To(MatchError("failed to describe stack"))
				})
			})
		})
	})

	Describe("Delete", func() {
		It("deletes the stack", func() {
			err := manager.Delete("some-stack-name")
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

					err := manager.Delete("some-stack-name")
					Expect(err).To(MatchError("failed to delete stack"))
				})
			})
		})
	})

	Describe("GetPhysicalIDForResource", func() {
		It("gets the physical resource id for the given stack resource", func() {
			cloudFormationClient.DescribeStackResourceCall.Returns.Output = &awscloudformation.DescribeStackResourceOutput{
				StackResourceDetail: &awscloudformation.StackResourceDetail{
					PhysicalResourceId: aws.String("some-physical-resource-id"),
				},
			}
			physicalID, err := manager.GetPhysicalIDForResource("some-stack-name", "some-logical-resource-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(cloudFormationClient.DescribeStackResourceCall.Receives.Input).To(Equal(&awscloudformation.DescribeStackResourceInput{
				StackName:         aws.String("some-stack-name"),
				LogicalResourceId: aws.String("some-logical-resource-id"),
			}))
			Expect(physicalID).To(Equal("some-physical-resource-id"))
		})

		Context("failure cases", func() {
			It("returns an error in case the DescribeStackResource Call fails", func() {
				cloudFormationClient.DescribeStackResourceCall.Returns.Error = errors.New("DescribeStackResource Call Failed")

				_, err := manager.GetPhysicalIDForResource("some-stack-name", "some-logical-resource-id")
				Expect(err).To(MatchError("DescribeStackResource Call Failed"))
			})
		})
	})
})
