package terraform

import (
	"io/ioutil"
)

func SetReadFile(f func(filename string) ([]byte, error)) {
	readFile = f
}

func ResetReadFile() {
	readFile = ioutil.ReadFile
}
