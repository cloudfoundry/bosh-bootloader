package storage

import "io"

func SetEncode(f func(io.Writer, interface{}) error) {
	encode = f
}

func ResetEncode() {
	encode = encodeFile
}
