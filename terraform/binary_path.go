package terraform

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
)

const (
	tfBinDataAssetName = "terraform"
	bblTfBinaryName    = "bbl-terraform"
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
	FS              tfBinaryPathFs
	EmbedData       embed.FS
	Path            string
	TerraformBinary string
}

//go:embed binary_dist
var content embed.FS

func NewBinary(terraformBinary string) *Binary {
	fs := afero.Afero{Fs: afero.NewOsFs()}
	return &Binary{
		FS:              fs,
		Path:            "binary_dist",
		EmbedData:       content,
		TerraformBinary: terraformBinary,
	}
}

func (binary *Binary) BinaryPath() (string, error) {
	// if user sets a terraform binary use it
	if binary.TerraformBinary != "" {
		exists, err := binary.FS.Exists(binary.TerraformBinary)
		if err == nil && exists {
			return binary.TerraformBinary, nil
		}
	}

	destinationPath := fmt.Sprintf("%s/%s", binary.FS.GetTempDir(os.TempDir()), bblTfBinaryName)
	exists, err := binary.FS.Exists(destinationPath)
	if err != nil {
		return "", err
	}

	if !exists {
		return destinationPath, binary.installTfBinary()
	}

	foundBinaryFileInfo, err := binary.FS.Stat(destinationPath)
	if err != nil {
		return "", err
	}

	distModTime, err := binary.RetrieveModTime()
	if err != nil {
		return "", err
	}

	foundBinaryModTime := foundBinaryFileInfo.ModTime()
	if !distModTime.After(foundBinaryModTime) {
		return destinationPath, nil
	}

	return destinationPath, binary.installTfBinary()
}

func (binary *Binary) installTfBinary() error {
	destinationPath := fmt.Sprintf("%s/%s", binary.FS.GetTempDir(os.TempDir()), bblTfBinaryName)
	sourcePath := fmt.Sprintf("%s/%s", binary.Path, tfBinDataAssetName)
	terraBytes, err := binary.EmbedData.ReadFile(sourcePath)
	if err != nil {
		return errors.New("missing terraform")
	}

	err = binary.FS.WriteFile(destinationPath, terraBytes, os.ModePerm)
	if err != nil {
		return err
	}

	m, err := binary.RetrieveModTime()
	if err != nil {
		return err
	}

	return binary.FS.Chtimes(destinationPath, time.Now(), m)
}

func (binary *Binary) RetrieveModTime() (time.Time, error) {
	timeStr, err := binary.EmbedData.ReadFile(fmt.Sprintf("%s/%s", binary.Path, terraformModTime))
	if err != nil {
		return time.Time{}, fmt.Errorf("could not find %s file", terraformModTime)
	}

	tmNum, err := strconv.ParseInt(strings.TrimSpace(string(timeStr)), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("incorrect format of time in terraform-mod-time: %s", err)
	}

	return time.Unix(tmNum, 0), nil
}
