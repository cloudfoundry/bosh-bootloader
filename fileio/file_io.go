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

type Renamer interface {
	Rename(oldpath, newpath string) error
}

type Remover interface {
	Remove(name string) error
}

type DirReader interface {
	ReadDir(dirname string) ([]os.FileInfo, error)
}

type AllRemover interface {
	RemoveAll(path string) error
}

type FileIO interface {
	FileWriter
	FileReader
	DirReader
	TempFiler
	Stater
	Renamer
	Remover
	AllRemover
}
