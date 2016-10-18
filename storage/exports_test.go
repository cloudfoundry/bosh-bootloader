package storage

import (
	"io"
	"os"
)

func SetEncode(f func(io.Writer, interface{}) error) {
	encode = f
}

func ResetEncode() {
	encode = encodeFile
}

func SetRename(f func(string, string) error) {
	rename = f
}

func ResetRename() {
	rename = os.Rename
}
