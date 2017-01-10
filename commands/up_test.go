package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Up", func() {
	var (
		command commands.Up

		fakeAWSUp          *fakes.AWSUp
		fakeGCPUp          *fakes.GCPUp
		fakeEnvGetter      *fakes.EnvGetter
		fakeEnvIDGenerator *fakes.EnvIDGenerator
		state              storage.State
	)

	BeforeEach(func() {
		fakeAWSUp = &fakes.AWSUp{Name: "aws"}
		fakeGCPUp = &fakes.GCPUp{Name: "gcp"}
		fakeEnvGetter = &fakes.EnvGetter{}

		fakeEnvIDGenerator = &fakes.EnvIDGenerator{}
		fakeEnvIDGenerator.GenerateCall.Returns.EnvID = "bbl-lake-time:stamp"

		command = commands.NewUp(fakeAWSUp, fakeGCPUp, fakeEnvGetter, fakeEnvIDGenerator)
	})

	Describe("Execute", func() {
		Context("when aws args are provided through environment variables", func() {
			BeforeEach(func() {
				fakeEnvGetter.Values = map[string]string{
					"BBL_AWS_ACCESS_KEY_ID":     "access-key-id-from-env",
					"BBL_AWS_SECRET_ACCESS_KEY": "secret-access-key-from-env",
					"BBL_AWS_REGION":            "region-from-env",
				}
			})

			It("uses the aws args provided by environment variables", func() {
				err := command.Execute([]string{
					"--iaas", "aws",
				}, storage.State{Version: 999})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig).To(Equal(commands.AWSUpConfig{
					AccessKeyID:     "access-key-id-from-env",
					SecretAccessKey: "secret-access-key-from-env",
					Region:          "region-from-env",
				}))
				Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(storage.State{
					Version: 999,
					EnvID:   "bbl-lake-time:stamp",
				}))
			})

			DescribeTable("gives precedence to arguments passed as command line args", func(args []string, expectedConfig commands.AWSUpConfig) {
				args = append(args, "--iaas", "aws")
				err := command.Execute(args, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig).To(Equal(expectedConfig))
				Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(storage.State{
					EnvID: "bbl-lake-time:stamp",
				}))
			},
				Entry("precedence to aws access key id",
					[]string{"--aws-access-key-id", "access-key-id-from-args"},
					commands.AWSUpConfig{
						AccessKeyID:     "access-key-id-from-args",
						SecretAccessKey: "secret-access-key-from-env",
						Region:          "region-from-env",
					},
				),
				Entry("precedence to aws secret access key",
					[]string{"--aws-secret-access-key", "secret-access-key-from-args"},
					commands.AWSUpConfig{
						AccessKeyID:     "access-key-id-from-env",
						SecretAccessKey: "secret-access-key-from-args",
						Region:          "region-from-env",
					},
				),
				Entry("precedence to aws region",
					[]string{"--aws-region", "region-from-args"},
					commands.AWSUpConfig{
						AccessKeyID:     "access-key-id-from-env",
						SecretAccessKey: "secret-access-key-from-env",
						Region:          "region-from-args",
					},
				),
			)
		})

		Context("env id", func() {
			Context("when the env id doesn't exist", func() {
				It("populates a new bbl env id", func() {
					fakeEnvIDGenerator.GenerateCall.Returns.EnvID = "bbl-lake-time:stamp"

					err := command.Execute([]string{
						"--iaas", "aws",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeEnvIDGenerator.GenerateCall.CallCount).To(Equal(1))
					Expect(fakeAWSUp.ExecuteCall.Receives.State.EnvID).To(Equal("bbl-lake-time:stamp"))
				})
			})

			Context("when the env id exists", func() {
				It("does not modify the state", func() {
					incomingState := storage.State{
						EnvID: "bbl-lake-time:stamp",
					}

					err := command.Execute([]string{
						"--iaas", "aws",
					}, incomingState)
					Expect(err).NotTo(HaveOccurred())

					state := fakeAWSUp.ExecuteCall.Receives.State
					Expect(state.EnvID).To(Equal("bbl-lake-time:stamp"))
				})
			})

			Context("when the user provides the name flag", func() {
				It("uses the name flag instead of generating one", func() {
					fakeEnvIDGenerator.GenerateCall.Returns.EnvID = "bbl-lake-time:stamp"

					err := command.Execute([]string{
						"--iaas", "aws",
						"--name", "a-better-name",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeEnvIDGenerator.GenerateCall.CallCount).To(Equal(0))
					Expect(fakeAWSUp.ExecuteCall.Receives.State.EnvID).To(Equal("a-better-name"))
				})
			})

			Context("failure cases", func() {
				It("returns an error when env id generator fails", func() {
					fakeEnvIDGenerator.GenerateCall.Returns.Error = errors.New("env id generation failed")

					err := command.Execute([]string{
						"--iaas", "aws",
					}, storage.State{})
					Expect(err).To(MatchError("env id generation failed"))
				})

				It("returns an error when name is passed for an existing env", func() {
					err := command.Execute([]string{
						"--iaas", "aws",
						"--name", "a-bad-name",
					}, storage.State{
						EnvID: "a-name",
					})
					Expect(err).To(MatchError("The director name cannot be changed for an existing environment. Current name is a-name."))
				})
			})
		})

		Context("when gcp args are provided through environment variables", func() {
			BeforeEach(func() {
				fakeEnvGetter.Values = map[string]string{
					"BBL_GCP_SERVICE_ACCOUNT_KEY": "some-service-account-key-env",
					"BBL_GCP_PROJECT_ID":          "some-project-id-env",
					"BBL_GCP_ZONE":                "some-zone-env",
					"BBL_GCP_REGION":              "some-region-env",
				}
			})

			It("uses the gcp args provided by environment variables", func() {
				err := command.Execute([]string{
					"--iaas", "gcp",
				}, storage.State{Version: 999})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeGCPUp.ExecuteCall.Receives.GCPUpConfig).To(Equal(commands.GCPUpConfig{
					ServiceAccountKeyPath: "some-service-account-key-env",
					ProjectID:             "some-project-id-env",
					Zone:                  "some-zone-env",
					Region:                "some-region-env",
				}))
				Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{
					Version: 999,
					EnvID:   "bbl-lake-time:stamp",
				}))
			})

			DescribeTable("gives precedence to arguments passed as command line args", func(args []string, expectedConfig commands.GCPUpConfig) {
				args = append(args, "--iaas", "gcp")

				err := command.Execute(args, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeGCPUp.ExecuteCall.Receives.GCPUpConfig).To(Equal(expectedConfig))
				Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{
					EnvID: "bbl-lake-time:stamp",
				}))
			},
				Entry("precedence to service account key",
					[]string{"--gcp-service-account-key", "some-service-account-key-from-args"},
					commands.GCPUpConfig{
						ServiceAccountKeyPath: "some-service-account-key-from-args",
						ProjectID:             "some-project-id-env",
						Zone:                  "some-zone-env",
						Region:                "some-region-env",
					},
				),
				Entry("precedence to project id",
					[]string{"--gcp-project-id", "some-project-id-from-args"},
					commands.GCPUpConfig{
						ServiceAccountKeyPath: "some-service-account-key-env",
						ProjectID:             "some-project-id-from-args",
						Zone:                  "some-zone-env",
						Region:                "some-region-env",
					},
				),
				Entry("precedence to zone",
					[]string{"--gcp-zone", "some-zone-from-args"},
					commands.GCPUpConfig{
						ServiceAccountKeyPath: "some-service-account-key-env",
						ProjectID:             "some-project-id-env",
						Zone:                  "some-zone-from-args",
						Region:                "some-region-env",
					},
				),
				Entry("precedence to region",
					[]string{"--gcp-region", "some-region-from-args"},
					commands.GCPUpConfig{
						ServiceAccountKeyPath: "some-service-account-key-env",
						ProjectID:             "some-project-id-env",
						Zone:                  "some-zone-env",
						Region:                "some-region-from-args",
					},
				),
			)
		})

		Context("when state does not contain an iaas", func() {
			It("uses the iaas from the env var", func() {
				fakeEnvGetter.Values = map[string]string{
					"BBL_IAAS": "gcp",
				}
				err := command.Execute([]string{
					"--gcp-service-account-key", "some-service-account-key",
					"--gcp-project-id", "some-project-id",
					"--gcp-zone", "some-zone",
					"--gcp-region", "some-region",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
				Expect(fakeAWSUp.ExecuteCall.CallCount).To(Equal(0))
			})

			It("uses the iaas from the args over the env var", func() {
				fakeEnvGetter.Values = map[string]string{
					"BBL_IAAS": "aws",
				}
				err := command.Execute([]string{
					"--iaas", "gcp",
					"--gcp-service-account-key", "some-service-account-key",
					"--gcp-project-id", "some-project-id",
					"--gcp-zone", "some-zone",
					"--gcp-region", "some-region",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
				Expect(fakeAWSUp.ExecuteCall.CallCount).To(Equal(0))
			})

			Context("when desired iaas is gcp", func() {
				It("executes the GCP up with gcp details from args", func() {
					err := command.Execute([]string{
						"--iaas", "gcp",
						"--gcp-service-account-key", "some-service-account-key",
						"--gcp-project-id", "some-project-id",
						"--gcp-zone", "some-zone",
						"--gcp-region", "some-region",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeGCPUp.ExecuteCall.Receives.GCPUpConfig).To(Equal(commands.GCPUpConfig{
						ServiceAccountKeyPath: "some-service-account-key",
						ProjectID:             "some-project-id",
						Zone:                  "some-zone",
						Region:                "some-region",
					}))
					Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{
						EnvID: "bbl-lake-time:stamp",
					}))
				})

				It("executes the GCP up with gcp details from env vars", func() {
					fakeEnvGetter.Values = map[string]string{
						"BBL_GCP_SERVICE_ACCOUNT_KEY": "some-service-account-key",
						"BBL_GCP_PROJECT_ID":          "some-project-id",
						"BBL_GCP_ZONE":                "some-zone",
						"BBL_GCP_REGION":              "some-region",
					}
					err := command.Execute([]string{
						"--iaas", "gcp",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeGCPUp.ExecuteCall.Receives.GCPUpConfig).To(Equal(commands.GCPUpConfig{
						ServiceAccountKeyPath: "some-service-account-key",
						ProjectID:             "some-project-id",
						Zone:                  "some-zone",
						Region:                "some-region",
					}))
					Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{
						EnvID: "bbl-lake-time:stamp",
					}))
				})
			})

			Context("when desired iaas is aws", func() {
				It("executes the AWS up", func() {
					err := command.Execute([]string{
						"--iaas", "aws",
						"--aws-access-key-id", "some-access-key-id",
						"--aws-secret-access-key", "some-secret-access-key",
						"--aws-region", "some-region",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeAWSUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig).To(Equal(commands.AWSUpConfig{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
						Region:          "some-region",
					}))
					Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(storage.State{
						EnvID: "bbl-lake-time:stamp",
					}))
				})

			})

			Context("when iaas is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{}, storage.State{})
					Expect(err).To(MatchError("--iaas [gcp, aws] must be provided"))
				})
			})

			Context("when an invalid iaas is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--iaas", "bad-iaas"}, storage.State{})
					Expect(err).To(MatchError(`"bad-iaas" is an invalid iaas type, supported values are: [gcp, aws]`))
				})
			})

			Context("failure cases", func() {
				It("returns an error when the desired up command fails", func() {
					fakeAWSUp.ExecuteCall.Returns.Error = errors.New("failed execution")
					err := command.Execute([]string{"--iaas", "aws"}, storage.State{})
					Expect(err).To(MatchError("failed execution"))
				})

				It("returns an error when undefined flags are passed", func() {
					err := command.Execute([]string{"--foo", "bar"}, storage.State{})
					Expect(err).To(MatchError("flag provided but not defined: -foo"))
				})
			})
		})

		Context("when state contains an iaas", func() {
			Context("when iaas is AWS", func() {
				var state storage.State

				BeforeEach(func() {
					state = storage.State{
						IAAS: "aws",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-region",
						},
					}
				})

				It("executes the AWS up", func() {
					err := command.Execute([]string{}, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeAWSUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig).To(Equal(commands.AWSUpConfig{}))
					Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(storage.State{
						IAAS: "aws",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-region",
						},
						EnvID: "bbl-lake-time:stamp",
					}))
				})

			})

			Context("when iaas is GCP", func() {
				It("executes the GCP up", func() {
					err := command.Execute([]string{}, storage.State{IAAS: "gcp"})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{
						IAAS:  "gcp",
						EnvID: "bbl-lake-time:stamp",
					}))
				})
			})

			Context("when iaas specified is different than the iaas in state", func() {
				It("returns an error when the iaas is provided via args", func() {
					err := command.Execute([]string{"--iaas", "aws"}, storage.State{IAAS: "gcp"})
					Expect(err).To(MatchError("The iaas type cannot be changed for an existing environment. The current iaas type is gcp."))
				})

				It("returns an error when the iaas is provided via env vars", func() {
					fakeEnvGetter.Values = map[string]string{
						"BBL_IAAS": "aws",
					}
					err := command.Execute([]string{}, storage.State{IAAS: "gcp"})
					Expect(err).To(MatchError("The iaas type cannot be changed for an existing environment. The current iaas type is gcp."))
				})
			})
		})
	})
})
