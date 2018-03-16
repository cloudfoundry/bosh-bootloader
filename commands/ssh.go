package commands

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type SSH struct {
	cmd           sshCmd
	keyGetter     sshKeyGetter
	tempDirWriter tempDirWriter
}

type sshCmd interface {
	Run([]string) error
}

type tempDirWriter interface {
	fileio.FileWriter
	fileio.TempDirer
}

func NewSSH(
	sshCmd sshCmd,
	sshKeyGetter sshKeyGetter,
	tempDirWriter tempDirWriter,
) SSH {
	return SSH{
		cmd:           sshCmd,
		keyGetter:     sshKeyGetter,
		tempDirWriter: tempDirWriter,
	}
}

func (s SSH) CheckFastFails(subcommandFlags []string, state storage.State) error {
	if len(state.Jumpbox.URL) == 0 {
		return errors.New("Invalid")
	}
	return nil
}

func (s SSH) Execute(args []string, state storage.State) error {
	privateKey, err := s.keyGetter.Get("jumpbox")
	if err != nil {
		return fmt.Errorf("Get jumpbox private key: %s", err)
	}

	tempDir, err := s.tempDirWriter.TempDir("", "")
	if err != nil {
		return fmt.Errorf("Create temp directory: %s", err)
	}

	jumpboxPrivateKeyPath := filepath.Join(tempDir, "jumpbox-private-key")
	err = s.tempDirWriter.WriteFile(jumpboxPrivateKeyPath, []byte(privateKey), 0600)
	if err != nil {
		return fmt.Errorf("Write private key file: %s", err)
	}

	jumpboxURL := strings.Split(state.Jumpbox.URL, ":")[0]

	return s.cmd.Run([]string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "ServerAliveInterval=300",
		fmt.Sprintf("jumpbox@%s", jumpboxURL),
		"-i", jumpboxPrivateKeyPath,
	})
}
