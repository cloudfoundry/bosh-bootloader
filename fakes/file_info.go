package fakes

import (
	"os"
	"time"
)

type FileInfo struct {
	FileName string
	Modtime  *time.Time // this is an optional
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
	if f.Modtime == nil {
		return time.Now()
	}
	return *f.Modtime
}
func (f FileInfo) IsDir() bool {
	return false
}
func (f FileInfo) Sys() interface{} {
	return nil
}
