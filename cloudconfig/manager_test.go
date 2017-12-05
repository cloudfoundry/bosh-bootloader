package cloudconfig_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	var (
		logger             *fakes.Logger
		cmd                *fakes.BOSHCommand
		stateStore         *fakes.StateStore
		opsGenerator       *fakes.CloudConfigOpsGenerator
		boshClientProvider *fakes.BOSHClientProvider
		boshClient         *fakes.BOSHClient
		terraformManager   *fakes.TerraformManager
		sshKeyGetter       *fakes.SSHKeyGetter
		manager            cloudconfig.Manager

		cloudConfigDir string
		varsDir        string
		incomingState  storage.State

		baseCloudConfig []byte
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		cmd = &fakes.BOSHCommand{}
		stateStore = &fakes.StateStore{}
		opsGenerator = &fakes.CloudConfigOpsGenerator{}
		boshClient = &fakes.BOSHClient{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		terraformManager = &fakes.TerraformManager{}
		sshKeyGetter = &fakes.SSHKeyGetter{}

		boshClientProvider.ClientCall.Returns.Client = boshClient

		var err error
		cloudConfigDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		stateStore.GetCloudConfigDirCall.Returns.Directory = cloudConfigDir

		varsDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		stateStore.GetVarsDirCall.Returns.Directory = varsDir

		cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
			stdout.Write([]byte("some-cloud-config"))
			return nil
		}

		incomingState = storage.State{
			IAAS: "gcp",
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		}

		opsGenerator.GenerateCall.Returns.OpsYAML = "some-ops"
		opsGenerator.GenerateVarsCall.Returns.VarsYAML = "some-vars"

		baseCloudConfig, err = ioutil.ReadFile("fixtures/base-cloud-config.yml")
		Expect(err).NotTo(HaveOccurred())

		manager = cloudconfig.NewManager(logger, cmd, stateStore, opsGenerator, boshClientProvider, terraformManager, sshKeyGetter)
	})

	Describe("Initialize", func() {
		It("returns a cloud config yaml with variable placeholders", func() {
			err := manager.Initialize(incomingState)
			Expect(err).NotTo(HaveOccurred())

			cloudConfig, err := ioutil.ReadFile(fmt.Sprintf("%s/cloud-config.yml", cloudConfigDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudConfig).To(MatchYAML(baseCloudConfig))

			Expect(opsGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			ops, err := ioutil.ReadFile(fmt.Sprintf("%s/ops.yml", cloudConfigDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(ops)).To(Equal("some-ops"))
		})

		Context("failure cases", func() {
			Context("when getting the cloud config dir fails", func() {
				BeforeEach(func() {
					stateStore.GetCloudConfigDirCall.Returns.Error = errors.New("carrot")
				})

				It("returns an error", func() {
					err := manager.Initialize(storage.State{})
					Expect(err).To(MatchError("Get cloud config dir: carrot"))
				})
			})

			Context("when write file fails to write cloud-config.yml", func() {
				BeforeEach(func() {
					cloudconfig.SetWriteFile(func(filename string, body []byte, mode os.FileMode) error {
						if strings.Contains(filename, "cloud-config.yml") {
							return errors.New("failed to write file")
						}
						return nil
					})
				})

				AfterEach(func() {
					cloudconfig.ResetWriteFile()
				})

				It("returns an error", func() {
					err := manager.Initialize(storage.State{})
					Expect(err).To(MatchError("failed to write file"))
				})
			})

			Context("when ops generator fails to generate", func() {
				BeforeEach(func() {
					opsGenerator.GenerateCall.Returns.Error = errors.New("failed to generate")
				})

				It("returns an error", func() {
					err := manager.Initialize(storage.State{})
					Expect(err).To(MatchError("failed to generate"))
				})
			})

			Context("when write file fails to write the ops files", func() {
				BeforeEach(func() {
					cloudconfig.SetWriteFile(func(filename string, body []byte, mode os.FileMode) error {
						if strings.Contains(filename, "ops.yml") {
							return errors.New("failed to write file")
						}
						return nil
					})
				})

				AfterEach(func() {
					cloudconfig.ResetWriteFile()
				})

				It("returns an error", func() {
					err := manager.Initialize(storage.State{})
					Expect(err).To(MatchError("failed to write file"))
				})
			})
		})
	})

	Describe("GenerateVars", func() {
		It("writes cloud config vars to the vars dir", func() {
			err := manager.GenerateVars(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(opsGenerator.GenerateVarsCall.Receives.State).To(Equal(incomingState))

			vars, err := ioutil.ReadFile(fmt.Sprintf("%s/cloud-config-vars.yml", varsDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(vars)).To(Equal("some-vars"))
		})

		Context("failure cases", func() {
			Context("when getting the vars dir fails", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("eggplant")
				})

				It("returns an error", func() {
					err := manager.GenerateVars(storage.State{})
					Expect(err).To(MatchError("Get vars dir: eggplant"))
				})
			})

			Context("when ops generator fails to generate vars", func() {
				BeforeEach(func() {
					opsGenerator.GenerateVarsCall.Returns.Error = errors.New("raspberry")
				})

				It("returns an error", func() {
					err := manager.GenerateVars(storage.State{})
					Expect(err).To(MatchError("Generate cloud config vars: raspberry"))
				})
			})

			Context("when write file fails to write the vars file", func() {
				BeforeEach(func() {
					cloudconfig.SetWriteFile(func(filename string, body []byte, mode os.FileMode) error {
						if strings.Contains(filename, "cloud-config-vars.yml") {
							return errors.New("coconut")
						}
						return nil
					})
				})

				AfterEach(func() {
					cloudconfig.ResetWriteFile()
				})

				It("returns an error", func() {
					err := manager.GenerateVars(storage.State{})
					Expect(err).To(MatchError("Write cloud config vars: coconut"))
				})
			})
		})
	})

	Describe("IsPresentCloudConfig", func() {
		Context("when cloud config files exist in the state dir", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(cloudConfigDir, "cloud-config.yml"), []byte("some existing cloud config"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(cloudConfigDir, "ops.yml"), []byte("some existing ops"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true", func() {
				Expect(manager.IsPresentCloudConfig()).To(BeTrue())
			})
		})

		Context("when cloud config files do not exist in the state dir", func() {
			It("returns false", func() {
				Expect(manager.IsPresentCloudConfig()).To(BeFalse())
			})
		})

		Context("failure cases", func() {
			Context("when getting the cloud config dir fails", func() {
				BeforeEach(func() {
					stateStore.GetCloudConfigDirCall.Returns.Error = errors.New("carrot")
				})

				It("returns false", func() {
					Expect(manager.IsPresentCloudConfig()).To(BeFalse())
				})
			})
		})
	})

	Describe("IsPresentCloudConfigVars", func() {
		Context("when cloud config vars file exists in the vars dir", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(varsDir, "cloud-config-vars.yml"), []byte("some existing vars"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true", func() {
				Expect(manager.IsPresentCloudConfigVars()).To(BeTrue())
			})
		})

		Context("when cloud config vars file does not exist in the vars dir", func() {
			It("returns false", func() {
				Expect(manager.IsPresentCloudConfigVars()).To(BeFalse())
			})
		})

		Context("failure cases", func() {
			Context("when getting the vars dir fails", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("carrot")
				})

				It("returns false", func() {
					Expect(manager.IsPresentCloudConfigVars()).To(BeFalse())
				})
			})
		})
	})

	Describe("Interpolate", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(filepath.Join(cloudConfigDir, "shenanigans-ops.yml"), []byte("shenanigans"), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(cloudConfigDir, "ops.yml"), []byte("base-operations"), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(cloudConfigDir, "cloud-config.yml"), []byte("base"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a cloud config yaml provided a valid bbl state", func() {
			cloudConfigYAML, err := manager.Interpolate()
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCallCount()).To(Equal(1))
			_, workingDirectory, args := cmd.RunArgsForCall(0)
			Expect(workingDirectory).To(Equal(cloudConfigDir))
			Expect(args).To(Equal([]string{
				"interpolate", fmt.Sprintf("%s%ccloud-config.yml", cloudConfigDir, os.PathSeparator),
				"--vars-file", fmt.Sprintf("%s%ccloud-config-vars.yml", varsDir, os.PathSeparator),
				"-o", fmt.Sprintf("%s/ops.yml", cloudConfigDir),
				"-o", fmt.Sprintf("%s/shenanigans-ops.yml", cloudConfigDir),
			}))

			Expect(cloudConfigYAML).To(Equal("some-cloud-config"))
		})

		Context("failure cases", func() {
			Context("when getting the cloud config dir fails", func() {
				BeforeEach(func() {
					stateStore.GetCloudConfigDirCall.Returns.Error = errors.New("carrot")
				})

				It("returns an error", func() {
					_, err := manager.Interpolate()
					Expect(err).To(MatchError("Get cloud config dir: carrot"))
				})
			})

			Context("when getting the vars dir fails", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("eggplant")
				})

				It("returns an error", func() {
					_, err := manager.Interpolate()
					Expect(err).To(MatchError("Get vars dir: eggplant"))
				})
			})

			Context("when reading the cloud config dir fails", func() {
				BeforeEach(func() {
					stateStore.GetCloudConfigDirCall.Returns.Directory = "invalid-directory"
				})

				It("returns an error", func() {
					_, err := manager.Interpolate()
					Expect(err).To(MatchError("Read cloud config dir: open invalid-directory: no such file or directory"))
				})
			})

			Context("when command fails to run", func() {
				BeforeEach(func() {
					cmd.RunReturns(errors.New("Interpolate cloud config: failed to run"))
				})

				It("returns an error", func() {
					_, err := manager.Interpolate()
					Expect(err).To(MatchError("Interpolate cloud config: failed to run"))
				})
			})
		})
	})

	Describe("Update", func() {
		It("logs steps taken", func() {
			err := manager.Update(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.Messages).To(Equal([]string{
				"generating cloud config",
				"applying cloud config",
			}))
		})

		It("updates the bosh director with a cloud config provided a valid bbl state", func() {
			err := manager.Update(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
			Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

			Expect(boshClient.UpdateCloudConfigCall.Receives.Yaml).To(Equal([]byte("some-cloud-config")))
		})

		Context("failure cases", func() {
			Context("when manager generate's command fails to run", func() {
				BeforeEach(func() {
					cmd.RunReturns(errors.New("failed to run"))
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to run"))
				})
			})

			Context("when bosh client fails to update cloud config", func() {
				BeforeEach(func() {
					boshClient.UpdateCloudConfigCall.Returns.Error = errors.New("failed to update")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to update"))
				})
			})
		})
	})
})
