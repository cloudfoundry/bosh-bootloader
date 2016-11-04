package commands_test

import (
	"errors"
	"os"

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

		fakeAWSUp *fakes.AWSUp
		fakeGCPUp *fakes.GCPUp
	)

	describeAWSEnvVars := func(state storage.State) {
		Context("when aws args are provided through environment variables", func() {
			BeforeEach(func() {
				os.Setenv("BBL_AWS_ACCESS_KEY_ID", "access-key-id-from-env")
				os.Setenv("BBL_AWS_SECRET_ACCESS_KEY", "secret-access-key-from-env")
				os.Setenv("BBL_AWS_REGION", "region-from-env")
			})

			AfterEach(func() {
				os.Setenv("BBL_AWS_ACCESS_KEY_ID", "")
				os.Setenv("BBL_AWS_SECRET_ACCESS_KEY", "")
				os.Setenv("BBL_AWS_REGION", "")
			})

			It("uses the aws args provided by environment variables", func() {
				err := command.Execute([]string{
					"--iaas", "aws",
				}, storage.State{Version: 999})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig).To(Equal(commands.AWSUpConfig{
					AWSAccessKeyID:     "access-key-id-from-env",
					AWSSecretAccessKey: "secret-access-key-from-env",
					AWSRegion:          "region-from-env",
				}))
				Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(storage.State{Version: 999}))
			})

			DescribeTable("gives precedence to arguments passed as command line args", func(args []string, expectedConfig commands.AWSUpConfig) {
				args = append(args, "--iaas", "aws")

				err := command.Execute(args, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig).To(Equal(expectedConfig))
				Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(state))
			},
				Entry("precedence to aws access key id",
					[]string{"--aws-access-key-id", "access-key-id-from-args"},
					commands.AWSUpConfig{
						AWSAccessKeyID:     "access-key-id-from-args",
						AWSSecretAccessKey: "secret-access-key-from-env",
						AWSRegion:          "region-from-env",
					},
				),
				Entry("precedence to aws secret access key",
					[]string{"--aws-secret-access-key", "secret-access-key-from-args"},
					commands.AWSUpConfig{
						AWSAccessKeyID:     "access-key-id-from-env",
						AWSSecretAccessKey: "secret-access-key-from-args",
						AWSRegion:          "region-from-env",
					},
				),
				Entry("precedence to aws region",
					[]string{"--aws-region", "region-from-args"},
					commands.AWSUpConfig{
						AWSAccessKeyID:     "access-key-id-from-env",
						AWSSecretAccessKey: "secret-access-key-from-env",
						AWSRegion:          "region-from-args",
					},
				),
			)
		})
	}

	BeforeEach(func() {
		fakeAWSUp = &fakes.AWSUp{Name: "aws"}
		fakeGCPUp = &fakes.GCPUp{Name: "gcp"}

		command = commands.NewUp(fakeAWSUp, fakeGCPUp)
	})

	Describe("Execute", func() {
		Context("when state does not contain an iaas", func() {
			Context("when desired iaas is gcp", func() {
				It("executes the GCP up", func() {
					err := command.Execute([]string{"--iaas", "gcp"}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{}))
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
						AWSAccessKeyID:     "some-access-key-id",
						AWSSecretAccessKey: "some-secret-access-key",
						AWSRegion:          "some-region",
					}))
					Expect(fakeAWSUp.ExecuteCall.Receives.State).To(Equal(storage.State{}))
				})

				describeAWSEnvVars(storage.State{})
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
					Expect(err).To(MatchError(`"bad-iaas" is invalid; supported values: [gcp, aws]`))
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
					}))
				})

				describeAWSEnvVars(state)
			})

			Context("when iaas is GCP", func() {
				It("executes the GCP up", func() {
					err := command.Execute([]string{}, storage.State{IAAS: "gcp"})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{IAAS: "gcp"}))
				})
			})

			Context("when iaas specified is different than the iaas in state", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--iaas", "aws"}, storage.State{IAAS: "gcp"})
					Expect(err).To(MatchError("the iaas provided must match the iaas in bbl-state.json"))
				})
			})
		})
	})
})
