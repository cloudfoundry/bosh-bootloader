package integration

import "io/ioutil"

func SetTempDir(f func(string, string) (string, error)) {
	tempDir = f
}

func ResetTempDir() {
	tempDir = ioutil.TempDir
}
