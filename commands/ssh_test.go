package commands_test

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSH", func() {
	var (
		ssh          commands.SSH
		sshCmd       *fakes.SSHCmd
		sshKeyGetter *fakes.SSHKeyGetter
		fileIO       *fakes.FileIO
	)

	BeforeEach(func() {
		sshCmd = &fakes.SSHCmd{}
		sshKeyGetter = &fakes.SSHKeyGetter{}
		fileIO = &fakes.FileIO{}

		ssh = commands.NewSSH(sshCmd, sshKeyGetter, fileIO)
	})

	Describe("CheckFastFails", func() {
		Context("--jumpbox", func() {
			Context("where there is a jumpbox url", func() {
				It("does not return an error", func() {
					err := ssh.CheckFastFails([]string{}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "some-jumpbox",
						},
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("where there is no jumpbox url", func() {
				It("does not return an error", func() {
					err := ssh.CheckFastFails([]string{}, storage.State{})
					Expect(err).To(MatchError("Invalid"))
				})
			})
		})
	})

	Describe("Execute", func() {
		var jumpboxPrivateKeyPath string
		BeforeEach(func() {
			sshKeyGetter.GetCall.Returns.PrivateKey = "some-private-key"
			fileIO.TempDirCall.Returns.Name = "some-temp-dir"
			jumpboxPrivateKeyPath = filepath.Join("some-temp-dir", "jumpbox-private-key")
		})

		Context("jumpbox", func() {
			Context("success", func() {
				It("calls ssh with appropriate arguments", func() {
					err := ssh.Execute([]string{}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "jumpboxURL:22",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(sshKeyGetter.GetCall.Receives.Deployment).To(Equal("jumpbox"))

					Expect(fileIO.WriteFileCall.CallCount).To(Equal(1))
					Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(jumpboxPrivateKeyPath))
					Expect(fileIO.WriteFileCall.Receives[0].Contents).To(Equal([]byte("some-private-key")))
					Expect(fileIO.WriteFileCall.Receives[0].Mode).To(Equal(os.FileMode(0600)))

					Expect(sshCmd.RunCall.Receives.Args).To(ConsistOf(
						"-o", "StrictHostKeyChecking=no",
						"-o", "ServerAliveInterval=300",
						"jumpbox@jumpboxURL",
						"-i", jumpboxPrivateKeyPath,
					))
				})
			})

			Context("failures", func() {
				It("contextualizes a failure to get the ssh private key", func() {
					sshKeyGetter.GetCall.Returns.Error = errors.New("fig")

					err := ssh.Execute([]string{}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "jumpboxURL",
						},
					})

					Expect(err).To(MatchError("Get jumpbox private key: fig"))
				})

				It("contextualizes a failure to create the temp directory", func() {
					fileIO.TempDirCall.Returns.Error = errors.New("date")

					err := ssh.Execute([]string{}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "jumpboxURL",
						},
					})

					Expect(err).To(MatchError("Create temp directory: date"))
				})

				It("contextualizes a failure to create the temp directory", func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{
						Error: errors.New("boisenberry"),
					}}

					err := ssh.Execute([]string{}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "jumpboxURL",
						},
					})

					Expect(err).To(MatchError("Write private key file: boisenberry"))
				})
			})
		})
	})
})
