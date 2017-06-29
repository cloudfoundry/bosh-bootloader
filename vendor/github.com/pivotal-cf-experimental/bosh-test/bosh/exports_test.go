package bosh

import (
	"io"
	"io/ioutil"
)

func SetBodyReader(r func(io.Reader) ([]byte, error)) {
	bodyReader = r
}

func ResetBodyReader() {
	bodyReader = ioutil.ReadAll
}
