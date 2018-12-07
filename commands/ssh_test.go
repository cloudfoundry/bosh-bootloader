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
		sshCLI       *fakes.SSHCLI
		pathFinder   *fakes.PathFinder
		sshKeyGetter *fakes.FancySSHKeyGetter
		fileIO       *fakes.FileIO
		randomPort   *fakes.RandomPort
		logger       *fakes.Logger
	)

	BeforeEach(func() {
		sshCLI = &fakes.SSHCLI{}
		sshKeyGetter = &fakes.FancySSHKeyGetter{}
		pathFinder = &fakes.PathFinder{}
		fileIO = &fakes.FileIO{}
		randomPort = &fakes.RandomPort{}
		logger = &fakes.Logger{}

		ssh = commands.NewSSH(logger, sshCLI, sshKeyGetter, pathFinder, fileIO, randomPort)
	})

	Describe("CheckFastFails", func() {
		It("checks the bbl state for the jumpbox url", func() {
			err := ssh.CheckFastFails([]string{""}, storage.State{Jumpbox: storage.Jumpbox{URL: "some-jumpbox"}})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("where there is no jumpbox url", func() {
			It("returns an error", func() {
				err := ssh.CheckFastFails([]string{""}, storage.State{})
				Expect(err).To(MatchError("Invalid bbl state for bbl ssh."))
			})
		})
	})

	Describe("Execute", func() {
		var (
			jumpboxPrivateKeyPath string
			state                 storage.State
		)
		BeforeEach(func() {
			fileIO.TempDirCall.Returns.Name = "some-temp-dir"
			sshKeyGetter.JumpboxGetCall.Returns.PrivateKey = "jumpbox-private-key"
			jumpboxPrivateKeyPath = filepath.Join("some-temp-dir", "jumpbox-private-key")
			pathFinder.CommandExistsCall.Returns.Exists = false

			state = storage.State{
				Jumpbox: storage.Jumpbox{
					URL: "jumpboxURL:22",
				},
				BOSH: storage.BOSH{
					DirectorAddress: "https://directorURL:25",
				},
			}
		})

		Context("--director", func() {
			var directorPrivateKeyPath string

			BeforeEach(func() {
				sshKeyGetter.DirectorGetCall.Returns.PrivateKey = "director-private-key"
				directorPrivateKeyPath = filepath.Join("some-temp-dir", "director-private-key")
				randomPort.GetPortCall.Returns.Port = "60000"
			})

			It("preemptively sshes to confirm host keys", func() {
				err := ssh.Execute([]string{"--director"}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintlnCall.Messages).To(Equal([]string{
					"checking host key",
					"opening a tunnel through your jumpbox",
				}))

				Expect(sshCLI.RunCall.Receives[0]).To(ConsistOf(
					"-T",
					"jumpbox@jumpboxURL",
					"-i",
					"some-temp-dir/jumpbox-private-key",
					"echo",
					"host key confirmed",
				))
			})

			It("opens a tunnel through the jumpbox", func() {
				err := ssh.Execute([]string{"--director"}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintlnCall.Messages).To(Equal([]string{
					"checking host key",
					"opening a tunnel through your jumpbox",
				}))

				Expect(sshKeyGetter.JumpboxGetCall.CallCount).To(Equal(1))

				Expect(fileIO.WriteFileCall.Receives).To(ContainElement(
					fakes.WriteFileReceive{
						Filename: jumpboxPrivateKeyPath,
						Contents: []byte("jumpbox-private-key"),
						Mode:     os.FileMode(0600),
					},
				))

				Expect(sshCLI.StartCall.Receives[0]).To(ConsistOf(
					"-4", "-D", "60000", "-nNC", "jumpbox@jumpboxURL", "-i", jumpboxPrivateKeyPath,
				))
			})

			It("sshes through the tunnel", func() {
				err := ssh.Execute([]string{"--director"}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(sshKeyGetter.DirectorGetCall.CallCount).To(Equal(1))

				Expect(fileIO.WriteFileCall.Receives).To(ContainElement(
					fakes.WriteFileReceive{
						Filename: directorPrivateKeyPath,
						Contents: []byte("director-private-key"),
						Mode:     os.FileMode(0600),
					},
				))

				Expect(sshCLI.RunCall.Receives[1]).To(ConsistOf(
					"-tt",
					"-o", "StrictHostKeyChecking=no",
					"-o", "ServerAliveInterval=300",
					"-o", "ProxyCommand=nc -x localhost:60000 %h %p",
					"-i", directorPrivateKeyPath,
					"jumpbox@directorURL",
				))
			})

			It("executes a command on the director through the ssh tunnel", func() {
				cmd := "echo hello"
				err := ssh.Execute([]string{"--director", "--cmd", cmd}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(sshKeyGetter.DirectorGetCall.CallCount).To(Equal(1))

				Expect(fileIO.WriteFileCall.Receives).To(ContainElement(
					fakes.WriteFileReceive{
						Filename: directorPrivateKeyPath,
						Contents: []byte("director-private-key"),
						Mode:     os.FileMode(0600),
					},
				))

				Expect(sshCLI.RunCall.Receives[1]).To(ConsistOf(
					"-tt",
					"-o", "StrictHostKeyChecking=no",
					"-o", "ServerAliveInterval=300",
					"-o", "ProxyCommand=nc -x localhost:60000 %h %p",
					"-i", directorPrivateKeyPath,
					"jumpbox@directorURL",
					cmd,
				))
			})

			Context("when connect-proxy is found", func() {
				BeforeEach(func() {
					pathFinder.CommandExistsCall.Returns.Exists = true
				})

				It("uses connect-proxy instead of netcat", func() {
					err := ssh.Execute([]string{"--director"}, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(pathFinder.CommandExistsCall.Receives.Command).To(Equal("connect-proxy"))

					Expect(sshCLI.RunCall.Receives[1]).To(ConsistOf(
						"-tt",
						"-o", "StrictHostKeyChecking=no",
						"-o", "ServerAliveInterval=300",
						"-o", "ProxyCommand=connect-proxy -S localhost:60000 %h %p",
						"-i", directorPrivateKeyPath,
						"jumpbox@directorURL",
					))
				})
			})

			Context("failure cases", func() {
				Context("when ssh key getter fails to get director key", func() {
					It("returns the error", func() {
						sshKeyGetter.DirectorGetCall.Returns.Error = errors.New("fig")

						err := ssh.Execute([]string{"--director"}, state)

						Expect(err).To(MatchError("Get director private key: fig"))
					})
				})

				Context("when fileio fails to create a temp dir", func() {
					It("returns the error", func() {
						fileIO.TempDirCall.Returns.Error = errors.New("date")

						err := ssh.Execute([]string{"--director"}, state)

						Expect(err).To(MatchError("Create temp directory: date"))
					})
				})

				Context("when fileio fails to create a temp dir", func() {
					It("contextualizes a failure to write the private key", func() {
						fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("boisenberry")}}

						err := ssh.Execute([]string{"--director"}, state)

						Expect(err).To(MatchError("Write private key file: boisenberry"))
					})
				})

				Context("when random port fails to return a port", func() {
					It("returns the error", func() {
						randomPort.GetPortCall.Returns.Error = errors.New("prune")

						err := ssh.Execute([]string{"--director"}, state)

						Expect(err).To(MatchError("Open proxy port: prune"))
					})
				})

				Context("when the ssh command fails to open a tunnel to the jumpbox", func() {
					It("returns the error", func() {
						sshCLI.StartCall.Returns = []fakes.SSHStartReturn{{Error: errors.New("lignonberry")}}

						err := ssh.Execute([]string{"--director"}, state)

						Expect(err).To(MatchError("Open tunnel to jumpbox: lignonberry"))
					})
				})
			})
		})

		Context("--jumpbox", func() {
			It("calls ssh with appropriate arguments", func() {
				err := ssh.Execute([]string{"--jumpbox"}, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(sshKeyGetter.JumpboxGetCall.CallCount).To(Equal(1))

				Expect(fileIO.WriteFileCall.CallCount).To(Equal(1))
				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(jumpboxPrivateKeyPath))
				Expect(fileIO.WriteFileCall.Receives[0].Contents).To(Equal([]byte("jumpbox-private-key")))
				Expect(fileIO.WriteFileCall.Receives[0].Mode).To(Equal(os.FileMode(0600)))

				Expect(sshCLI.RunCall.Receives[0]).To(ConsistOf(
					"-tt",
					"-o", "ServerAliveInterval=300",
					"jumpbox@jumpboxURL",
					"-i", jumpboxPrivateKeyPath,
				))
			})

			Context("when ssh key getter fails to get the jumpbox ssh private key", func() {
				It("returns the error", func() {
					sshKeyGetter.JumpboxGetCall.Returns.Error = errors.New("fig")

					err := ssh.Execute([]string{"--jumpbox"}, state)

					Expect(err).To(MatchError("Get jumpbox private key: fig"))
				})
			})

			Context("when fileio fails to write the jumpbox private key", func() {
				It("returns the error", func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("boisenberry")}}

					err := ssh.Execute([]string{"--jumpbox"}, state)

					Expect(err).To(MatchError("Write private key file: boisenberry"))
				})
			})
		})

		Context("when the user does not provide a flag", func() {
			It("returns an error", func() {
				err := ssh.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("This command requires the --jumpbox or --director flag."))
			})
		})

		Context("when the user provides a command to execute on the jumpbox", func() {
			It("returns an error", func() {
				err := ssh.Execute([]string{"--jumpbox", "--cmd", "bogus"}, storage.State{})
				Expect(err).To(MatchError("Executing commands on jumpbox not supported (only on director)."))
			})
		})

		Context("when the user provides invalid flags", func() {
			It("returns an error", func() {
				err := ssh.Execute([]string{"--bogus-flag"}, storage.State{})
				Expect(err).To(MatchError("flag provided but not defined: -bogus-flag"))
			})
		})
	})
})
