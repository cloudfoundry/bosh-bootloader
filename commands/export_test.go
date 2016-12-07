package commands

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

func SetTempDir(f func(dir, prefix string) (string, error)) {
	tempDir = f
}

func ResetTempDir() {
	tempDir = ioutil.TempDir
}

func SetWriteFile(f func(filename string, data []byte, perm os.FileMode) error) {
	writeFile = f
}

func ResetWriteFile() {
	writeFile = ioutil.WriteFile
}

func SetMarshal(f func(in interface{}) (out []byte, err error)) {
	marshal = f
}

func ResetMarshal() {
	marshal = yaml.Marshal
}
