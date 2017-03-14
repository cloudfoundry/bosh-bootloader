package cloudconfig

import (
	"io/ioutil"
	"os"
)

func SetTempDir(f func(string, string) (string, error)) {
	tempDir = f
}

func ResetTempDir() {
	tempDir = ioutil.TempDir
}

func SetWriteFile(f func(string, []byte, os.FileMode) error) {
	writeFile = f
}

func ResetWriteFile() {
	writeFile = ioutil.WriteFile
}
