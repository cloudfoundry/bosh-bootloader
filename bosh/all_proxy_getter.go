package bosh

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
)

type AllProxyGetter struct {
	sshKeyGetter sshKeyGetter
	fs           allProxyFs
}

type allProxyFs interface {
	fileio.TempDirer
	fileio.FileWriter
}

func NewAllProxyGetter(sshKeyGetter sshKeyGetter, fs allProxyFs) AllProxyGetter {
	return AllProxyGetter{
		sshKeyGetter: sshKeyGetter,
		fs:           fs,
	}
}

func (a AllProxyGetter) GeneratePrivateKey() (string, error) {
	dir, err := a.fs.TempDir("", "bosh-jumpbox")
	if err != nil {
		return "", err
	}

	privateKeyPath := filepath.Join(dir, "bosh_jumpbox_private.key")

	privateKeyContents, err := a.sshKeyGetter.Get("jumpbox")
	if err != nil {
		return "", err
	}

	err = a.fs.WriteFile(privateKeyPath, []byte(privateKeyContents), 0600)
	if err != nil {
		return "", err
	}

	return privateKeyPath, err
}

func (a AllProxyGetter) BoshAllProxy(jumpboxURL, privateKeyPath string) string {
	return fmt.Sprintf("ssh+socks5://%s?private-key=%s", jumpboxURL, privateKeyPath)
}
