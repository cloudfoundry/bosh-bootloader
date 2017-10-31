package commands_test

import (
	"errors"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plan", func() {
	var (
		command commands.Plan

		up                 *fakes.Up
		boshManager        *fakes.BOSHManager
		terraformManager   *fakes.TerraformManager
		cloudConfigManager *fakes.CloudConfigManager
		stateStore         *fakes.StateStore
		envIDManager       *fakes.EnvIDManager

		tempDir string
	)

	BeforeEach(func() {
		up = &fakes.Up{}
		boshManager = &fakes.BOSHManager{}
		boshManager.VersionCall.Returns.Version = "2.0.24"

		terraformManager = &fakes.TerraformManager{}
		cloudConfigManager = &fakes.CloudConfigManager{}
		stateStore = &fakes.StateStore{}
		envIDManager = &fakes.EnvIDManager{}

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		stateStore.GetBblDirCall.Returns.Directory = tempDir

		command = commands.NewPlan(up, boshManager, cloudConfigManager, stateStore, envIDManager, terraformManager)
	})

	Describe("Execute", func() {
		var (
			state       storage.State
			syncedState storage.State
		)

		BeforeEach(func() {
			state = storage.State{ID: "some-state-id"}
			syncedState = storage.State{ID: "synced-state-id"}
			envIDManager.SyncCall.Returns.State = syncedState
		})

		It("sets up the bbl state dir", func() {
			args := []string{"--ops-file"}
			err := command.Execute(args, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(up.ParseArgsCall.CallCount).To(Equal(1))
			Expect(up.ParseArgsCall.Receives.Args).To(Equal(args))
			Expect(up.ParseArgsCall.Receives.State).To(Equal(state))

			Expect(envIDManager.SyncCall.CallCount).To(Equal(1))
			Expect(envIDManager.SyncCall.Receives.State).To(Equal(state))

			Expect(stateStore.SetCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(syncedState))

			Expect(terraformManager.InitCall.CallCount).To(Equal(1))
			Expect(terraformManager.InitCall.Receives.BBLState).To(Equal(syncedState))

			Expect(boshManager.InitializeJumpboxCall.CallCount).To(Equal(1))
			Expect(boshManager.InitializeJumpboxCall.Receives.State).To(Equal(syncedState))
			Expect(boshManager.InitializeJumpboxCall.Receives.TerraformOutputs.Map).To(BeNil())

			Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.InitializeDirectorCall.Receives.State).To(Equal(syncedState))
			Expect(boshManager.InitializeDirectorCall.Receives.TerraformOutputs.Map).To(BeNil())

			Expect(cloudConfigManager.InitializeCall.CallCount).To(Equal(1))
			Expect(cloudConfigManager.InitializeCall.Receives.State).To(Equal(syncedState))
		})

		Context("when --no-director is passed", func() {
			It("sets no director on the state", func() {
				envIDManager.SyncCall.Returns.State = storage.State{NoDirector: true}
				up.ParseArgsCall.Returns.Config = commands.UpConfig{NoDirector: true}

				err := command.Execute([]string{"--no-director"}, storage.State{NoDirector: false})
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.InitializeJumpboxCall.CallCount).To(Equal(0))
				Expect(boshManager.InitializeDirectorCall.CallCount).To(Equal(0))
			})

			Context("but a director already exists", func() {
				It("returns a helpful error", func() {
					up.ParseArgsCall.Returns.Config = commands.UpConfig{NoDirector: true}

					err := command.Execute([]string{"--no-director"}, storage.State{
						BOSH: storage.BOSH{
							DirectorUsername: "admin",
						},
					})
					Expect(err).To(MatchError(`Director already exists, you must re-create your environment to use "--no-director"`))
				})
			})
		})

		Describe("failure cases", func() {
			It("returns an error if parse args fails", func() {
				up.ParseArgsCall.Returns.Error = errors.New("canteloupe")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("canteloupe"))
			})

			It("returns an error if state store set fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("peach")}}

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Save state: peach"))
			})

			It("returns an error if terraform manager init fails", func() {
				terraformManager.InitCall.Returns.Error = errors.New("pomegranate")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Terraform manager init: pomegranate"))
			})

			It("returns an error if bosh manager initialize jumpbox fails", func() {
				boshManager.InitializeJumpboxCall.Returns.Error = errors.New("tomato")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Bosh manager initialize jumpbox: tomato"))
			})

			It("returns an error if bosh manager initialize director fails", func() {
				boshManager.InitializeDirectorCall.Returns.Error = errors.New("tomatoe")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Bosh manager initialize director: tomatoe"))
			})

			It("returns an error if cloud config initialize fails", func() {
				cloudConfigManager.InitializeCall.Returns.Error = errors.New("potato")

				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("Cloud config manager initialize: potato"))
			})
		})
	})

	Describe("CheckFastFails", func() {
		It("returns CheckFastFails on Up", func() {
			up.CheckFastFailsCall.Returns.Error = errors.New("banana")
			err := command.CheckFastFails([]string{}, storage.State{Version: 999})

			Expect(err).To(MatchError("banana"))
			Expect(up.CheckFastFailsCall.Receives.SubcommandFlags).To(Equal([]string{}))
			Expect(up.CheckFastFailsCall.Receives.State).To(Equal(storage.State{Version: 999}))
		})
	})

	Describe("ParseArgs", func() {
		It("returns ParseArgs on Up", func() {
			up.ParseArgsCall.Returns.Config = commands.UpConfig{OpsFile: "some-path"}
			config, err := command.ParseArgs([]string{"--ops-file", "some-path"}, storage.State{ID: "some-state-id"})
			Expect(err).NotTo(HaveOccurred())

			Expect(up.ParseArgsCall.Receives.Args).To(Equal([]string{"--ops-file", "some-path"}))
			Expect(up.ParseArgsCall.Receives.State).To(Equal(storage.State{ID: "some-state-id"}))
			Expect(config.OpsFile).To(Equal("some-path"))
		})
	})
})
