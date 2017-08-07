package commands_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Up", func() {
	var (
		command commands.Up

		fakeAWSUp       *fakes.AWSUp
		fakeGCPUp       *fakes.GCPUp
		fakeEnvGetter   *fakes.EnvGetter
		fakeBOSHManager *fakes.BOSHManager
	)

	BeforeEach(func() {
		fakeAWSUp = &fakes.AWSUp{Name: "aws"}
		fakeGCPUp = &fakes.GCPUp{Name: "gcp"}
		fakeEnvGetter = &fakes.EnvGetter{}
		fakeBOSHManager = &fakes.BOSHManager{}
		fakeBOSHManager.VersionCall.Returns.Version = "2.0.24"

		command = commands.NewUp(fakeAWSUp, fakeGCPUp, fakeEnvGetter, fakeBOSHManager)
	})

	Describe("CheckFastFails", func() {
		Context("when the version of BOSH is a dev build", func() {
			It("does not fail", func() {
				fakeBOSHManager.VersionCall.Returns.Error = bosh.NewBOSHVersionError(errors.New("BOSH version could not be parsed"))

				err := command.CheckFastFails([]string{
					"--iaas", "aws",
				}, storage.State{Version: 999})

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the version of BOSH is lower than 2.0.24", func() {
			It("returns a helpful error message when bbling up with a director", func() {
				fakeBOSHManager.VersionCall.Returns.Version = "1.9.1"
				err := command.CheckFastFails([]string{
					"--iaas", "aws",
				}, storage.State{Version: 999})

				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})

			Context("when the no-director flag is specified", func() {
				It("does not return an error", func() {
					fakeBOSHManager.VersionCall.Returns.Version = "1.9.1"
					err := command.CheckFastFails([]string{
						"--iaas", "aws",
						"--no-director",
					}, storage.State{Version: 999})

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when the version of BOSH cannot be retrieved", func() {
			It("returns an error", func() {
				fakeBOSHManager.VersionCall.Returns.Error = errors.New("BOOM")
				err := command.CheckFastFails([]string{
					"--iaas", "aws",
				}, storage.State{Version: 999})

				Expect(err.Error()).To(ContainSubstring("BOOM"))
			})
		})

		Context("when the version of BOSH is invalid", func() {
			It("returns an error", func() {
				fakeBOSHManager.VersionCall.Returns.Version = "X.5.2"
				err := command.CheckFastFails([]string{
					"--iaas", "aws",
				}, storage.State{Version: 999})

				Expect(err.Error()).To(ContainSubstring("invalid syntax"))
			})
		})

		Context("when bbl-state contains an env-id", func() {
			var (
				name  = "some-name"
				state = storage.State{EnvID: name}
			)

			Context("when the passed in name matches the env-id", func() {
				It("returns no error", func() {
					err := command.CheckFastFails([]string{
						"--iaas", "aws",
						"--name", name,
					}, state)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the passed in name does not match the env-id", func() {
				It("returns an error", func() {
					err := command.CheckFastFails([]string{
						"--iaas", "aws",
						"--name", "some-other-name",
					}, state)
					Expect(err).To(MatchError(fmt.Sprintf("The director name cannot be changed for an existing environment. Current name is %s.", name)))
				})
			})
		})
	})

	Describe("Execute", func() {
		Context("when an ops-file is provided via command line flag", func() {
			It("populates the aws config with the correct ops-file path", func() {
				err := command.Execute([]string{
					"--iaas", "aws",
					"--ops-file", "some-ops-file-path",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig).To(Equal(commands.AWSUpConfig{
					OpsFilePath: "some-ops-file-path",
				}))
			})

			It("populates the gcp config with the correct ops-file path", func() {
				err := command.Execute([]string{
					"--iaas", "gcp",
					"--ops-file", "some-ops-file-path",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeGCPUp.ExecuteCall.Receives.GCPUpConfig).To(Equal(commands.GCPUpConfig{
					OpsFilePath: "some-ops-file-path",
				}))
			})
		})

		Context("when state does not contain an iaas", func() {
			Context("when desired iaas is gcp", func() {
				It("executes the GCP up with gcp details from args", func() {
					err := command.Execute([]string{
						"--iaas", "gcp",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
				})

				Context("when the --jumpbox flag is specified", func() {
					It("executes the GCP up with gcp details from args", func() {
						err := command.Execute([]string{
							"--iaas", "gcp",
							"--jumpbox",
						}, storage.State{})
						Expect(err).NotTo(HaveOccurred())

						Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
						Expect(fakeGCPUp.ExecuteCall.Receives.GCPUpConfig).To(Equal(commands.GCPUpConfig{
							Jumpbox: true,
						}))
					})
				})
			})

			Context("when desired iaas is aws", func() {
				It("executes the AWS up", func() {
					err := command.Execute([]string{
						"--iaas", "aws",
					}, storage.State{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeAWSUp.ExecuteCall.CallCount).To(Equal(1))
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
			})

			Context("when iaas is GCP", func() {
				It("executes the GCP up", func() {
					err := command.Execute([]string{}, storage.State{IAAS: "gcp"})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeGCPUp.ExecuteCall.CallCount).To(Equal(1))
					Expect(fakeGCPUp.ExecuteCall.Receives.State).To(Equal(storage.State{
						IAAS: "gcp",
					}))
				})
			})
		})

		Context("when the user provides the name flag", func() {
			It("passes the name flag in the up config", func() {
				err := command.Execute([]string{
					"--iaas", "aws",
					"--name", "a-better-name",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig.Name).To(Equal("a-better-name"))
			})
		})

		Context("when the user provides the no-director flag", func() {
			It("passes no-director as true in the up config", func() {
				err := command.Execute([]string{
					"--iaas", "aws",
					"--no-director",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeAWSUp.ExecuteCall.Receives.AWSUpConfig.NoDirector).To(Equal(true))
			})

			It("passes no-director as true in the up config", func() {
				err := command.Execute([]string{
					"--iaas", "gcp",
					"--no-director",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeGCPUp.ExecuteCall.Receives.GCPUpConfig.NoDirector).To(Equal(true))
			})
		})

		Context("when the user provides the jumpbox flag", func() {
			It("passes jumpbox as true in the up config", func() {
				err := command.Execute([]string{
					"--iaas", "gcp",
					"--jumpbox",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeGCPUp.ExecuteCall.Receives.GCPUpConfig.Jumpbox).To(Equal(true))
			})
		})
	})
})
