package fileio

import (
	"io/ioutil"
	"os"
)

type FileIOAdapter struct{}

func (*FileIOAdapter) TempFile(dir, prefix string) (f *os.File, err error) {
	return ioutil.TempFile(dir, prefix)
}

func (*FileIOAdapter) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func (*FileIOAdapter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func (*FileIOAdapter) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (*FileIOAdapter) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (*FileIOAdapter) Remove(name string) error {
	return os.Remove(name)
}

func (*FileIOAdapter) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (*FileIOAdapter) RemoveAll(path string) error {
	return os.RemoveAll(path)
}
