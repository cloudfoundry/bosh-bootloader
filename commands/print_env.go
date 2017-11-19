package commands

import (
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type PrintEnv struct {
	stateValidator   stateValidator
	logger           logger
	sshKeyGetter     sshKeyGetter
	terraformManager terraformManager
}

type envSetter interface {
	Set(key, value string) error
}

func NewPrintEnv(logger logger, stateValidator stateValidator, sshKeyGetter sshKeyGetter, terraformManager terraformManager) PrintEnv {
	return PrintEnv{
		stateValidator:   stateValidator,
		logger:           logger,
		sshKeyGetter:     sshKeyGetter,
		terraformManager: terraformManager,
	}
}

func (p PrintEnv) CheckFastFails(subcommandFlags []string, state storage.State) error {
	err := p.stateValidator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (p PrintEnv) Execute(args []string, state storage.State) error {
	if state.NoDirector {
		terraformOutputs, err := p.terraformManager.GetOutputs()
		if err != nil {
			return err
		}

		p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=https://%s:25555", terraformOutputs.GetString("external_ip")))
		return nil
	}

	p.logger.Println(fmt.Sprintf("export BOSH_CLIENT=%s", state.BOSH.DirectorUsername))
	p.logger.Println(fmt.Sprintf("export BOSH_CLIENT_SECRET=%s", state.BOSH.DirectorPassword))
	p.logger.Println(fmt.Sprintf("export BOSH_ENVIRONMENT=%s", state.BOSH.DirectorAddress))
	p.logger.Println(fmt.Sprintf("export BOSH_CA_CERT='%s'", state.BOSH.DirectorSSLCA))

	portNumber, err := p.getPort()
	if err != nil {
		// not tested
		return err
	}

	dir, err := ioutil.TempDir("", "bosh-jumpbox")
	if err != nil {
		// not tested
		return err
	}

	privateKeyPath := filepath.Join(dir, "bosh_jumpbox_private.key")

	privateKeyContents, err := p.sshKeyGetter.Get("jumpbox")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(privateKeyPath, []byte(privateKeyContents), 0600)
	if err != nil {
		// not tested
		return err
	}

	jumpboxURL := strings.Split(state.Jumpbox.URL, ":")[0]

	p.logger.Println(fmt.Sprintf("export BOSH_ALL_PROXY=socks5://localhost:%s", portNumber))
	p.logger.Println(fmt.Sprintf("export JUMPBOX_PRIVATE_KEY=%s", privateKeyPath))
	p.logger.Println(fmt.Sprintf("ssh -f -N -o StrictHostKeyChecking=no -o ServerAliveInterval=300 -D %s jumpbox@%s -i $JUMPBOX_PRIVATE_KEY", portNumber, jumpboxURL))

	return nil
}

func (p PrintEnv) getPort() (string, error) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	defer l.Close()

	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return "", err
	}

	return port, nil
}
