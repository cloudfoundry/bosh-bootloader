package certs

import (
	"io"
	"io/ioutil"
	"os"
)

func SetReadAll(f func(r io.Reader) ([]byte, error)) {
	readAll = f
}

func ResetReadAll() {
	readAll = ioutil.ReadAll
}

func SetStat(f func(name string) (os.FileInfo, error)) {
	stat = f
}

func ResetStat() {
	stat = os.Stat
}
