package terraform

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/terraform/binary_dist"
	"github.com/spf13/afero"
)

const tfBinDataAssetName = "terraform"
const bblTfBinaryName = "bbl-terraform"

func BinaryPath() (string, error) {
	return BinaryPathInjected(afero.Afero{afero.NewOsFs()})
}

type tfBinaryPathFs interface {
	GetTempDir(string) string
	Exists(string) (bool, error)
	Stat(string) (os.FileInfo, error)
	WriteFile(string, []byte, os.FileMode) error
	Chtimes(name string, atime time.Time, mtime time.Time) error
}

func BinaryPathInjected(fs tfBinaryPathFs) (string, error) {
	path := filepath.Join(fs.GetTempDir(""), bblTfBinaryName)
	exists, err := fs.Exists(path)
	if err != nil {
		return "", err
	}

	if !exists {
		return path, installTfBinary(path, fs)
	}

	foundBinaryFileInfo, err := fs.Stat(path)
	if err != nil {
		return "", err
	}

	foundBinaryModTime := foundBinaryFileInfo.ModTime()
	distModTime := binary_dist.MustAssetInfo(tfBinDataAssetName).ModTime()
	if !distModTime.After(foundBinaryModTime) {
		return path, nil
	}

	return path, installTfBinary(path, fs)
}

func installTfBinary(path string, fs tfBinaryPathFs) error {
	err := fs.WriteFile(path, binary_dist.MustAsset(tfBinDataAssetName), os.ModePerm)
	if err != nil {
		return err
	}
	return fs.Chtimes(path, time.Now(), binary_dist.MustAssetInfo(tfBinDataAssetName).ModTime())
}
