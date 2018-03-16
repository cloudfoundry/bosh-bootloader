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
		sshKeyGetter *fakes.FancySSHKeyGetter
		fileIO       *fakes.FileIO
		randomPort   *fakes.RandomPort
	)

	BeforeEach(func() {
		sshCmd = &fakes.SSHCmd{}
		sshKeyGetter = &fakes.FancySSHKeyGetter{}
		fileIO = &fakes.FileIO{}
		randomPort = &fakes.RandomPort{}

		ssh = commands.NewSSH(sshCmd, sshKeyGetter, fileIO, randomPort)
	})

	Describe("CheckFastFails", func() {
		Context("--jumpbox", func() {
			Context("where there is a jumpbox url", func() {
				It("does not return an error", func() {
					err := ssh.CheckFastFails([]string{"--jumpbox"}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "some-jumpbox",
						},
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("where there is no jumpbox url", func() {
				It("does not return an error", func() {
					err := ssh.CheckFastFails([]string{"--jumpbox"}, storage.State{})
					Expect(err).To(MatchError("Invalid"))
				})
			})
		})
	})

	Describe("Execute", func() {
		var jumpboxPrivateKeyPath string
		BeforeEach(func() {
			sshKeyGetter.JumpboxGetCall.Returns.PrivateKey = "jumpbox-private-key"
			fileIO.TempDirCall.Returns.Name = "some-temp-dir"
			jumpboxPrivateKeyPath = filepath.Join("some-temp-dir", "jumpbox-private-key")
		})

		Context("director", func() {
			var (
				state                  storage.State
				directorPrivateKeyPath string
			)

			BeforeEach(func() {
				state = storage.State{
					Jumpbox: storage.Jumpbox{
						URL: "jumpboxURL:22",
					},
					BOSH: storage.BOSH{
						DirectorAddress: "https://directorURL:25",
					},
				}

				sshKeyGetter.DirectorGetCall.Returns.PrivateKey = "director-private-key"
				directorPrivateKeyPath = filepath.Join("some-temp-dir", "director-private-key")

				randomPort.GetPortCall.Returns.Port = "60000"
			})

			Context("success", func() {
				It("calls ssh with appropriate arguments", func() {
					err := ssh.Execute([]string{"--director"}, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(sshKeyGetter.JumpboxGetCall.CallCount).To(Equal(1))
					Expect(sshKeyGetter.DirectorGetCall.CallCount).To(Equal(1))

					Expect(fileIO.WriteFileCall.CallCount).To(Equal(2))
					Expect(fileIO.WriteFileCall.Receives).To(ConsistOf(
						fakes.WriteFileReceive{
							Filename: jumpboxPrivateKeyPath,
							Contents: []byte("jumpbox-private-key"),
							Mode:     os.FileMode(0600),
						},
						fakes.WriteFileReceive{
							Filename: directorPrivateKeyPath,
							Contents: []byte("director-private-key"),
							Mode:     os.FileMode(0600),
						},
					))

					Expect(sshCmd.RunCall.Receives[0].Args).To(ConsistOf(
						"-4", "-D", "60000",
						"-fNC", "jumpbox@jumpboxURL",
						"-i", jumpboxPrivateKeyPath,
					))

					Expect(sshCmd.RunCall.Receives[1].Args).To(ConsistOf(
						"-o", "ProxyCommand=nc -x localhost:60000 %h %p",
						"-i", directorPrivateKeyPath,
						"jumpbox@directorURL",
					))
				})
			})

			Context("failure", func() {
				It("contextualizes a failure to get the ssh private key", func() {
					sshKeyGetter.DirectorGetCall.Returns.Error = errors.New("fig")

					err := ssh.Execute([]string{"--director"}, state)

					Expect(err).To(MatchError("Get director private key: fig"))
				})

				It("contextualizes a failure to create the temp directory", func() {
					fileIO.TempDirCall.Returns.Error = errors.New("date")

					err := ssh.Execute([]string{"--director"}, state)

					Expect(err).To(MatchError("Create temp directory: date"))
				})

				It("contextualizes a failure to write the private key", func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{
						Error: errors.New("boisenberry"),
					}}

					err := ssh.Execute([]string{"--director"}, state)

					Expect(err).To(MatchError("Write private key file: boisenberry"))
				})

				It("contextualizes a failure to find a random open port", func() {
					randomPort.GetPortCall.Returns.Error = errors.New("prune")

					err := ssh.Execute([]string{"--director"}, state)

					Expect(err).To(MatchError("Open proxy port: prune"))
				})

				It("contextualizes a failure to open a tunnel to the jumpbox", func() {
					sshCmd.RunCall.Returns = []fakes.SSHRunReturn{
						fakes.SSHRunReturn{
							Error: errors.New("lignonberry"),
						},
					}

					err := ssh.Execute([]string{"--director"}, state)

					Expect(err).To(MatchError("Open tunnel to jumpbox: lignonberry"))
				})
			})
		})

		Context("jumpbox", func() {
			Context("success", func() {
				It("calls ssh with appropriate arguments", func() {
					err := ssh.Execute([]string{"--jumpbox"}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "jumpboxURL:22",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(sshKeyGetter.JumpboxGetCall.CallCount).To(Equal(1))

					Expect(fileIO.WriteFileCall.CallCount).To(Equal(1))
					Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(jumpboxPrivateKeyPath))
					Expect(fileIO.WriteFileCall.Receives[0].Contents).To(Equal([]byte("jumpbox-private-key")))
					Expect(fileIO.WriteFileCall.Receives[0].Mode).To(Equal(os.FileMode(0600)))

					Expect(sshCmd.RunCall.Receives[0].Args).To(ConsistOf(
						"-o", "StrictHostKeyChecking=no",
						"-o", "ServerAliveInterval=300",
						"jumpbox@jumpboxURL",
						"-i", jumpboxPrivateKeyPath,
					))
				})
			})

			Context("failures", func() {
				It("contextualizes a failure to get the ssh private key", func() {
					sshKeyGetter.JumpboxGetCall.Returns.Error = errors.New("fig")

					err := ssh.Execute([]string{"--jumpbox"}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "jumpboxURL",
						},
					})

					Expect(err).To(MatchError("Get jumpbox private key: fig"))
				})

				It("contextualizes a failure to create the temp directory", func() {
					fileIO.TempDirCall.Returns.Error = errors.New("date")

					err := ssh.Execute([]string{"--jumpbox"}, storage.State{
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

					err := ssh.Execute([]string{"--jumpbox"}, storage.State{
						Jumpbox: storage.Jumpbox{
							URL: "jumpboxURL",
						},
					})

					Expect(err).To(MatchError("Write private key file: boisenberry"))
				})
			})
		})

		Context("without director or jumpbox flags", func() {
			It("returns an error", func() {
				err := ssh.Execute([]string{}, storage.State{Jumpbox: storage.Jumpbox{URL: "no"}})

				Expect(err).To(MatchError("ssh expects --jumpbox or --director"))
			})
		})

		Context("with invalid flags", func() {
			It("returns an error", func() {
				err := ssh.Execute([]string{"--bogus-flag"}, storage.State{Jumpbox: storage.Jumpbox{URL: "no"}})

				Expect(err).To(MatchError("flag provided but not defined: -bogus-flag"))
			})
		})
	})
})
