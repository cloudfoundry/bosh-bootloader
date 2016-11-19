package terraform

import (
	"io/ioutil"
	"os"
)

func SetTempDir(f func(dir, prefix string) (string, error)) {
	tempDir = f
}

func ResetTempDir() {
	tempDir = ioutil.TempDir
}

func SetWriteFile(f func(file string, data []byte, perm os.FileMode) error) {
	writeFile = f
}

func ResetWriteFile() {
	writeFile = ioutil.WriteFile
}

func SetReadFile(f func(filename string) ([]byte, error)) {
	readFile = f
}

func ResetReadFile() {
	readFile = ioutil.ReadFile
}
