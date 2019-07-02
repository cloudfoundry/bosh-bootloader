package commands

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/flags"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type SSH struct {
	logger        logger
	cli           sshCLI
	keyGetter     sshKeyGetter
	pathFinder    pathFinder
	tempDirWriter tempDirWriter
	randomPort    randomPort
}

type sshCLI interface {
	Run([]string) error
	Start([]string) (*exec.Cmd, error)
}

type pathFinder interface {
	CommandExists(string) bool
}

type randomPort interface {
	GetPort() (string, error)
}

type tempDirWriter interface {
	fileio.FileWriter
	fileio.TempDirer
}

func NewSSH(logger logger, sshCLI sshCLI, sshKeyGetter sshKeyGetter, pathFinder pathFinder, tempDirWriter tempDirWriter, randomPort randomPort) SSH {
	return SSH{
		logger:        logger,
		cli:           sshCLI,
		keyGetter:     sshKeyGetter,
		pathFinder:    pathFinder,
		tempDirWriter: tempDirWriter,
		randomPort:    randomPort,
	}
}

func (s SSH) CheckFastFails(subcommandFlags []string, state storage.State) error {
	if len(state.Jumpbox.URL) == 0 {
		return errors.New("Invalid bbl state for bbl ssh.")
	}

	return nil
}

func (s SSH) Execute(args []string, state storage.State) error {
	var (
		jumpbox  bool
		director bool
		cmd      string
	)
	sshFlags := flags.New("ssh")
	sshFlags.Bool(&jumpbox, "jumpbox")
	sshFlags.Bool(&director, "director")
	sshFlags.String(&cmd, "cmd", "")
	err := sshFlags.Parse(args)
	if err != nil {
		return err
	}

	if !jumpbox && !director {
		return fmt.Errorf("This command requires the --jumpbox or --director flag.")
	}

	if jumpbox && len(cmd) > 0 {
		return fmt.Errorf("Executing commands on jumpbox not supported (only on director).")
	}

	tempDir, err := s.tempDirWriter.TempDir("", "")
	if err != nil {
		return fmt.Errorf("Create temp directory: %s", err)
	}

	jumpboxKey, err := s.keyGetter.Get("jumpbox")
	if err != nil {
		return fmt.Errorf("Get jumpbox private key: %s", err)
	}

	jumpboxKeyPath := filepath.Join(tempDir, "jumpbox-private-key")

	err = s.tempDirWriter.WriteFile(jumpboxKeyPath, []byte(jumpboxKey), 0600)
	if err != nil {
		return fmt.Errorf("Write private key file: %s", err)
	}

	jumpboxURL := strings.Split(state.Jumpbox.URL, ":")[0]

	if jumpbox {
		return s.cli.Run([]string{
			"-tt",
			"-o", "ServerAliveInterval=300",
			fmt.Sprintf("jumpbox@%s", jumpboxURL),
			"-i", jumpboxKeyPath,
		})
	}

	directorPrivateKey, err := s.keyGetter.Get("director")
	if err != nil {
		return fmt.Errorf("Get director private key: %s", err)
	}

	directorKeyPath := filepath.Join(tempDir, "director-private-key")

	err = s.tempDirWriter.WriteFile(directorKeyPath, []byte(directorPrivateKey), 0600)
	if err != nil {
		return fmt.Errorf("Write private key file: %s", err)
	}

	port, err := s.randomPort.GetPort()
	if err != nil {
		return fmt.Errorf("Open proxy port: %s", err)
	}

	s.logger.Println("checking host key")
	err = s.cli.Run([]string{
		"-T",
		fmt.Sprintf("jumpbox@%s", jumpboxURL),
		"-i", jumpboxKeyPath,
		"echo", "host key confirmed",
	})
	if err != nil {
		return fmt.Errorf("unable to verify host key fingerprint: %s", err)
	}
	s.logger.Println("opening a tunnel through your jumpbox")
	backgroundTunnel, err := s.cli.Start([]string{
		"-4",
		"-D", port,
		"-nNC",
		fmt.Sprintf("jumpbox@%s", jumpboxURL),
		"-i", jumpboxKeyPath,
	})
	if err != nil {
		return fmt.Errorf("Open tunnel to jumpbox: %s", err)
	}
	defer func() {
		if backgroundTunnel != nil {
			backgroundTunnel.Process.Signal(syscall.SIGINT)
		} // removing this will break the acceptance test
	}()

	proxyCommandPrefix := "nc -x"
	if s.pathFinder.CommandExists("connect-proxy") {
		proxyCommandPrefix = "connect-proxy -S"
	}

	ip := strings.Split(strings.TrimPrefix(state.BOSH.DirectorAddress, "https://"), ":")[0]

	toExecute := []string{
		"-tt",
		"-o", "StrictHostKeyChecking=no",
		"-o", "ServerAliveInterval=300",
		"-o", fmt.Sprintf("ProxyCommand=%s localhost:%s %%h %%p", proxyCommandPrefix, port),
		"-i", directorKeyPath,
		fmt.Sprintf("jumpbox@%s", ip),
	}
	if len(cmd) > 0 {
		toExecute = append(toExecute, cmd)
		s.logger.Printf("executing command on director:\n%s\n", cmd)
	}

	time.Sleep(2 * time.Second) // make sure we give that tunnel a moment to open
	return s.cli.Run(toExecute)
}
