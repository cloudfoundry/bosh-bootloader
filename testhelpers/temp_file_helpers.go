package testhelpers

import (
	"os"
)

func WriteByteContentsToTempFile(contents []byte) (string, error) {
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	err = os.WriteFile(tempFile.Name(), contents, os.ModePerm)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), err
}

func WriteContentsToTempFile(contents string) (string, error) {
	return WriteByteContentsToTempFile([]byte(contents))
}
