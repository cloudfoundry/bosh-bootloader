package fakes

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

type FileIO struct {
	TempFileCall struct {
		CallCount int
		Receives  struct {
			Dir    string
			Prefix string
		}
		Returns struct {
			File  afero.File
			Error error
		}
	}

	TempDirCall struct {
		CallCount int
		Receives  struct {
			Dir    string
			Prefix string
		}
		Returns struct {
			Name  string
			Error error
		}
	}

	GetTempDirCall struct {
		CallCount int
		Receives  struct {
			Dir string
		}
		Returns struct {
			Name string
		}
	}

	ReadFileCall struct {
		CallCount int
		Fake      func(string) ([]byte, error)
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
		Fake      func(string) (os.FileInfo, error)
		Receives  struct {
			Name string
		}
		Returns struct {
			FileInfo os.FileInfo
			Error    error
		}
	}

	ExistsCall struct {
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Bool  bool
			Error error
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

	ChtimesCall struct {
		CallCount int
		Receives  struct {
			path    string
			ATime   time.Time
			ModTime time.Time
		}
		Returns struct {
			Error error
		}
	}

	RemoveCall struct {
		CallCount int
		Receives  []RemoveReceive
		Returns   []RemoveReturn
	}

	RemoveAllCall struct {
		CallCount int
		Receives  []RemoveAllReceive
		Returns   []RemoveAllReturn
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

	MkdirAllCall struct {
		CallCount int
		Receives  struct {
			Dir  string
			Perm os.FileMode
		}
		Returns struct {
			Error error
		}
	}
}

type WriteFileReceive struct {
	Filename string
	Contents []byte
	Mode     os.FileMode
}

type WriteFileReturn struct {
	Error error
}

type RemoveReceive struct {
	Name string
}

type RemoveReturn struct {
	Error error
}

type RemoveAllReceive struct {
	Path string
}

type RemoveAllReturn struct {
	Error error
}

func (f *FileIO) TempFile(dir, prefix string) (afero.File, error) {
	f.TempFileCall.CallCount++
	f.TempFileCall.Receives.Dir = dir
	f.TempFileCall.Receives.Prefix = prefix
	return f.TempFileCall.Returns.File, f.TempFileCall.Returns.Error
}

func (f *FileIO) TempDir(dir, prefix string) (string, error) {
	f.TempDirCall.CallCount++
	f.TempDirCall.Receives.Dir = dir
	f.TempDirCall.Receives.Prefix = prefix
	return f.TempDirCall.Returns.Name, f.TempDirCall.Returns.Error
}

func (f *FileIO) GetTempDir(dir string) string {
	f.GetTempDirCall.CallCount++
	f.GetTempDirCall.Receives.Dir = dir
	return f.GetTempDirCall.Returns.Name
}

func (f *FileIO) ReadFile(filename string) ([]byte, error) {
	f.ReadFileCall.CallCount++
	f.ReadFileCall.Receives.Filename = filename
	if f.ReadFileCall.Fake == nil {
		return f.ReadFileCall.Returns.Contents, f.ReadFileCall.Returns.Error
	}
	return f.ReadFileCall.Fake(filename)
}

func (f *FileIO) WriteFile(filename string, contents []byte, perm os.FileMode) error {
	f.WriteFileCall.CallCount++

	f.WriteFileCall.Receives = append(f.WriteFileCall.Receives, WriteFileReceive{
		Filename: filename,
		Contents: contents,
		Mode:     perm,
	})

	if len(f.WriteFileCall.Returns) < f.WriteFileCall.CallCount {
		return nil
	}

	return f.WriteFileCall.Returns[f.WriteFileCall.CallCount-1].Error
}

func (f *FileIO) Stat(name string) (os.FileInfo, error) {
	f.StatCall.CallCount++
	f.StatCall.Receives.Name = name
	if f.StatCall.Fake == nil {
		return f.StatCall.Returns.FileInfo, f.StatCall.Returns.Error
	}
	return f.StatCall.Fake(name)
}

func (f *FileIO) Exists(path string) (bool, error) {
	f.ExistsCall.CallCount++
	f.ExistsCall.Receives.Path = path
	return f.ExistsCall.Returns.Bool, f.ExistsCall.Returns.Error
}

func (f *FileIO) Rename(oldpath, newpath string) error {
	f.RenameCall.CallCount++
	f.RenameCall.Receives.Oldpath = oldpath
	f.RenameCall.Receives.Newpath = newpath
	return f.RenameCall.Returns.Error
}

func (f *FileIO) Chtimes(path string, atime, mtime time.Time) error {
	f.ChtimesCall.CallCount++
	f.ChtimesCall.Receives.ATime = atime
	f.ChtimesCall.Receives.ModTime = mtime
	return f.ChtimesCall.Returns.Error
}

func (f *FileIO) Remove(name string) error {
	f.RemoveCall.CallCount++

	f.RemoveCall.Receives = append(f.RemoveCall.Receives, RemoveReceive{
		Name: name,
	})

	if len(f.RemoveCall.Returns) < f.RemoveCall.CallCount {
		return nil
	}

	return f.RemoveCall.Returns[f.RemoveCall.CallCount-1].Error
}

func (f *FileIO) RemoveAll(path string) error {
	f.RemoveAllCall.CallCount++

	f.RemoveAllCall.Receives = append(f.RemoveAllCall.Receives, RemoveAllReceive{
		Path: path,
	})

	if len(f.RemoveAllCall.Returns) < f.RemoveAllCall.CallCount {
		return nil
	}

	return f.RemoveAllCall.Returns[f.RemoveAllCall.CallCount-1].Error
}
func (f *FileIO) ReadDir(dirname string) ([]os.FileInfo, error) {
	f.ReadDirCall.CallCount++
	f.ReadDirCall.Receives.Dirname = dirname
	return f.ReadDirCall.Returns.FileInfos, f.ReadDirCall.Returns.Error
}

func (f *FileIO) MkdirAll(dir string, perm os.FileMode) error {
	f.MkdirAllCall.CallCount++
	f.MkdirAllCall.Receives.Dir = dir
	f.MkdirAllCall.Receives.Perm = perm
	return f.MkdirAllCall.Returns.Error
}
