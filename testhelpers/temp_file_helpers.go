package testhelpers

import (
	"io/ioutil"
	"os"
)

func WriteContentsToTempFile(contents string) (string, error) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	err = ioutil.WriteFile(tempFile.Name(), []byte(contents), os.ModePerm)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), err
}
