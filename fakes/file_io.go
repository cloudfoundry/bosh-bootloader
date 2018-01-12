package fakes

import "os"

type FileIO struct {
	TempFileCall struct {
		CallCount int
		Receives  struct {
			Dir    string
			Prefix string
		}
		Returns struct {
			File  *os.File
			Error error
		}
	}

	ReadFileCall struct {
		CallCount int
		Receives  struct {
			Filename string
		}
		Returns struct {
			Contents []byte
			Error    error
		}
	}

	WriteFileCall struct {
		CallCount int
		Receives  []WriteFileReceive
		Returns   []WriteFileReturn
	}

	StatCall struct {
		CallCount int
		Receives  struct {
			Name string
		}
		Returns struct {
			FileInfo os.FileInfo
			Error    error
		}
	}

	RenameCall struct {
		CallCount int
		Receives  struct {
			Oldpath string
			Newpath string
		}
		Returns struct {
			Error error
		}
	}

	RemoveCall struct {
		CallCount int
		Receives  struct {
			Name string
		}
		Returns struct {
			Error error
		}
	}

	RemoveAllCall struct {
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Error error
		}
	}

	ReadDirCall struct {
		CallCount int
		Receives  struct {
			Dirname string
		}
		Returns struct {
			FileInfos []os.FileInfo
			Error     error
		}
	}
}

type WriteFileReceive struct {
	Filename string
	Contents []byte
}

type WriteFileReturn struct {
	Error error
}

func (f *FileIO) TempFile(dir, prefix string) (*os.File, error) {
	f.TempFileCall.CallCount++
	f.TempFileCall.Receives.Dir = dir
	f.TempFileCall.Receives.Prefix = prefix
	return f.TempFileCall.Returns.File, f.TempFileCall.Returns.Error
}

func (f *FileIO) ReadFile(filename string) ([]byte, error) {
	f.ReadFileCall.CallCount++
	f.ReadFileCall.Receives.Filename = filename
	return f.ReadFileCall.Returns.Contents, f.ReadFileCall.Returns.Error
}

func (f *FileIO) WriteFile(filename string, contents []byte, perm os.FileMode) error {
	f.WriteFileCall.CallCount++

	f.WriteFileCall.Receives = append(f.WriteFileCall.Receives, WriteFileReceive{
		Filename: filename,
		Contents: contents,
	})

	if len(f.WriteFileCall.Returns) < f.WriteFileCall.CallCount {
		return nil
	}

	return f.WriteFileCall.Returns[f.WriteFileCall.CallCount-1].Error
}

func (f *FileIO) Stat(name string) (os.FileInfo, error) {
	f.StatCall.CallCount++
	f.StatCall.Receives.Name = name
	return f.StatCall.Returns.FileInfo, f.StatCall.Returns.Error
}

func (f *FileIO) Rename(oldpath, newpath string) error {
	f.RenameCall.CallCount++
	f.RenameCall.Receives.Oldpath = oldpath
	f.RenameCall.Receives.Newpath = newpath
	return f.RenameCall.Returns.Error
}

func (f *FileIO) Remove(name string) error {
	f.RemoveCall.CallCount++
	f.RemoveCall.Receives.Name = name
	return f.RemoveCall.Returns.Error
}

func (f *FileIO) RemoveAll(path string) error {
	f.RemoveAllCall.CallCount++
	f.RemoveAllCall.Receives.Path = path
	return f.RemoveAllCall.Returns.Error
}
func (f *FileIO) ReadDir(dirname string) ([]os.FileInfo, error) {
	f.ReadDirCall.CallCount++
	f.ReadDirCall.Receives.Dirname = dirname
	return f.ReadDirCall.Returns.FileInfos, f.ReadDirCall.Returns.Error
}
