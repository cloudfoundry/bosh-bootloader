package fileio

import (
	"os"

	"github.com/spf13/afero"
)

type FileWriter interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type TempFiler interface {
	TempFile(dir, prefix string) (f afero.File, err error)
}

type TempDirer interface {
	TempDir(dir, prefix string) (name string, err error)
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

type AllMkdirer interface {
	MkdirAll(dir string, perm os.FileMode) error
}
