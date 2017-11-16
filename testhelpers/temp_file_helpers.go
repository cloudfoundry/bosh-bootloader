package testhelpers

import (
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

func WriteContentsToTempFile(contents string) (string, error) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	err = ioutil.WriteFile(tempFile.Name(), []byte(contents), storage.StateMode)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), err
}
