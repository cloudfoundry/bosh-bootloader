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
		Receives  struct {
			Filename string
			Contents []byte
		}
		Returns struct {
			Error error
		}
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
}

func (fake *FileIO) TempFile(dir, prefix string) (f *os.File, err error) {
	fake.TempFileCall.CallCount++
	fake.TempFileCall.Receives.Dir = dir
	fake.TempFileCall.Receives.Prefix = prefix
	return fake.TempFileCall.Returns.File, fake.TempFileCall.Returns.Error
}

func (fake *FileIO) ReadFile(filename string) ([]byte, error) {
	fake.ReadFileCall.CallCount++
	fake.ReadFileCall.Receives.Filename = filename
	return fake.ReadFileCall.Returns.Contents, fake.ReadFileCall.Returns.Error
}

func (fake *FileIO) WriteFile(filename string, contents []byte, perm os.FileMode) error {
	fake.WriteFileCall.CallCount++
	fake.WriteFileCall.Receives.Filename = filename
	fake.WriteFileCall.Receives.Contents = contents
	return fake.WriteFileCall.Returns.Error
}

func (fake *FileIO) Stat(name string) (os.FileInfo, error) {
	fake.StatCall.CallCount++
	fake.StatCall.Receives.Name = name
	return fake.StatCall.Returns.FileInfo, fake.StatCall.Returns.Error
}
