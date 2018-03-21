package commands

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type SSH struct {
	cmd           sshCmd
	keyGetter     sshKeyGetter
	tempDirWriter tempDirWriter
	randomPort    randomPort
}

type sshCmd interface {
	Run([]string) error
}

type randomPort interface {
	GetPort() (string, error)
}

type tempDirWriter interface {
	fileio.FileWriter
	fileio.TempDirer
}

func NewSSH(
	sshCmd sshCmd,
	sshKeyGetter sshKeyGetter,
	tempDirWriter tempDirWriter,
	randomPort randomPort,
) SSH {
	return SSH{
		cmd:           sshCmd,
		keyGetter:     sshKeyGetter,
		tempDirWriter: tempDirWriter,
		randomPort:    randomPort,
	}
}

func (s SSH) CheckFastFails(subcommandFlags []string, state storage.State) error {
	if len(state.Jumpbox.URL) == 0 {
		return errors.New("Invalid")
	}
	return nil
}

func (s SSH) Execute(args []string, state storage.State) error {
	var (
		jumpbox  bool
		director bool
	)
	sshFlags := flags.New("ssh")
	sshFlags.Bool(&jumpbox, "jumpbox")
	sshFlags.Bool(&director, "director")
	err := sshFlags.Parse(args)
	if err != nil {
		return err
	}

	jumpboxPrivateKey, err := s.keyGetter.Get("jumpbox")
	if err != nil {
		return fmt.Errorf("Get jumpbox private key: %s", err)
	}

	tempDir, err := s.tempDirWriter.TempDir("", "")
	if err != nil {
		return fmt.Errorf("Create temp directory: %s", err)
	}

	jumpboxPrivateKeyPath := filepath.Join(tempDir, "jumpbox-private-key")
	err = s.tempDirWriter.WriteFile(jumpboxPrivateKeyPath, []byte(jumpboxPrivateKey), 0600)
	if err != nil {
		return fmt.Errorf("Write private key file: %s", err)
	}

	jumpboxURL := strings.Split(state.Jumpbox.URL, ":")[0]

	if jumpbox {
		return s.cmd.Run([]string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "ServerAliveInterval=300",
			fmt.Sprintf("jumpbox@%s", jumpboxURL),
			"-i", jumpboxPrivateKeyPath,
		})
	} else if director {
		directorPrivateKey, err := s.keyGetter.Get("director")
		if err != nil {
			return fmt.Errorf("Get director private key: %s", err)
		}

		directorPrivateKeyPath := filepath.Join(tempDir, "director-private-key")
		err = s.tempDirWriter.WriteFile(directorPrivateKeyPath, []byte(directorPrivateKey), 0600)
		if err != nil {
			return fmt.Errorf("Write private key file: %s", err)
		}

		port, err := s.randomPort.GetPort()
		if err != nil {
			return fmt.Errorf("Open proxy port: %s", err)
		}

		err = s.cmd.Run([]string{
			"-4", "-D", port,
			"-fNC", fmt.Sprintf("jumpbox@%s", jumpboxURL),
			"-i", jumpboxPrivateKeyPath,
		})
		if err != nil {
			return fmt.Errorf("Open tunnel to jumpbox: %s", err)
		}

		directorURL := strings.Split(strings.TrimPrefix(state.BOSH.DirectorAddress, "https://"), ":")[0]

		return s.cmd.Run([]string{
			"-o", fmt.Sprintf("ProxyCommand=nc -x localhost:%s %s", port, "%h %p"),
			"-i", directorPrivateKeyPath,
			fmt.Sprintf("jumpbox@%s", directorURL),
		})
	}
	return errors.New("ssh expects --jumpbox or --director")
}
