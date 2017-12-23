package fileio

import "os"

type FileWriter interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type TempFiler interface {
	TempFile(dir, prefix string) (f *os.File, err error)
}

type Stater interface {
	Stat(name string) (os.FileInfo, error)
}

type FileIO interface {
	FileWriter
	FileReader
	TempFiler
	Stater
}
