package fakes

import (
	"os"
	"time"
)

type FileInfo struct {
	FileName string
}

func (f FileInfo) Name() string {
	return f.FileName
}
func (f FileInfo) Size() int64 {
	return 0
}
func (f FileInfo) Mode() os.FileMode {
	return os.ModePerm
}
func (f FileInfo) ModTime() time.Time {
	return time.Now()
}
func (f FileInfo) IsDir() bool {
	return false
}
func (f FileInfo) Sys() interface{} {
	return nil
}
