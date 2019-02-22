package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/afero"
)

const (
	tfBinDataAssetName = "terraform"
	bblTfBinaryName    = "bbl-terraform"
	terraformBinary    = "./binary_dist"
	terraformModTime   = "terraform-mod-time"
)

type tfBinaryPathFs interface {
	GetTempDir(string) string
	Exists(string) (bool, error)
	Stat(string) (os.FileInfo, error)
	WriteFile(string, []byte, os.FileMode) error
	Chtimes(name string, atime time.Time, mtime time.Time) error
}

type Binary struct {
	Box  *packr.Box
	FS   tfBinaryPathFs
	Path string
}

func NewBinary() *Binary {
	fs := afero.Afero{Fs: afero.NewOsFs()}
	return &Binary{
		Box:  packr.New("terraform", terraformBinary),
		FS:   fs,
		Path: filepath.Join(fs.GetTempDir(""), bblTfBinaryName),
	}
}

func (binary *Binary) BinaryPath() (string, error) {
	exists, err := binary.FS.Exists(binary.Path)
	if err != nil {
		return "", err
	}

	if !exists {
		return binary.Path, binary.installTfBinary()
	}

	foundBinaryFileInfo, err := binary.FS.Stat(binary.Path)
	if err != nil {
		return "", err
	}

	distModTime, err := binary.RetrieveModTime()
	if err != nil {
		return "", err
	}

	foundBinaryModTime := foundBinaryFileInfo.ModTime()
	if !distModTime.After(foundBinaryModTime) {
		return binary.Path, nil
	}

	return binary.Path, binary.installTfBinary()
}

func (binary *Binary) installTfBinary() error {
	terraBytes, err := binary.Box.Find(tfBinDataAssetName)
	if err != nil {
		return err
	}

	err = binary.FS.WriteFile(binary.Path, terraBytes, os.ModePerm)
	if err != nil {
		return err
	}

	m, err := binary.RetrieveModTime()
	if err != nil {
		return err
	}

	return binary.FS.Chtimes(binary.Path, time.Now(), m)
}

func (binary *Binary) RetrieveModTime() (time.Time, error) {
	timeStr, err := binary.Box.FindString(terraformModTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not find %s file", terraformModTime)
	}

	tmNum, err := strconv.ParseInt(strings.TrimSpace(timeStr), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("incorrect format of time in terraform-mod-time: %s", err)
	}

	return time.Unix(tmNum, 0), nil
}
