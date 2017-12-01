package testhelpers

import (
	"io/ioutil"
	"os"
)

func WriteByteContentsToTempFile(contents []byte) (string, error) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	err = ioutil.WriteFile(tempFile.Name(), contents, os.ModePerm)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), err
}

func WriteContentsToTempFile(contents string) (string, error) {
	return WriteByteContentsToTempFile([]byte(contents))
}
