package storage

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var ObjectNotFoundError = errors.New("Object not found")

type Version struct {
	Name    string    `json:"name"`
	Ref     string    `json:"ref"`
	Updated time.Time `json:"updated"`
}

// public only because []Object != []ObjectImpl :(
type Object interface {
	NewReader() (io.ReadCloser, error)
	NewWriter() io.WriteCloser
	Version() (Version, error)
}

type Bucket interface {
	GetAllObjects() ([]Object, error)
	Delete() error // test only
}

type tarrer interface {
	Write(io.Writer, []string) error
	Read(io.Reader, string) error
}

type Storage struct {
	Name     string
	Bucket   Bucket
	Object   Object
	Archiver tarrer
}

func (s Storage) GetAllNewerVersions(watermark Version) ([]Version, error) {
	objects, err := s.Bucket.GetAllObjects()
	if err != nil {
		return nil, err
	}
	versions := []Version{}
	for _, object := range objects {
		version, err := object.Version()
		if err != nil {
			return nil, err
		}
		if version.Updated.Before(watermark.Updated) {
			continue
		}
		versions = append(versions, version)
	}
	return versions, nil
}

func (s Storage) Version() (Version, error) {
	return s.Object.Version()
}

func (s Storage) Download(targetDir string) (Version, error) {
	reader, err := s.Object.NewReader()
	if err != nil {
		if err == ObjectNotFoundError {
			return s.Upload(targetDir)
		}
		return Version{}, err
	}
	defer reader.Close() // what happens if this errors?

	err = os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return Version{}, err
	}

	err = s.Archiver.Read(reader, targetDir)
	if err != nil {
		return Version{}, err
	}

	return s.Version()
}

func (s Storage) Upload(filePath string) (Version, error) {
	writer := s.Object.NewWriter()
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return Version{}, err
	}

	paths := []string{}
	for _, file := range files {
		paths = append(paths, filepath.Join(filePath, file.Name()))
	}

	err = s.Archiver.Write(writer, paths)
	if err != nil {
		return Version{}, err
	}

	err = writer.Close()
	if err != nil {
		return Version{}, err
	}

	return s.Version()
}

// test cleanup only
func (s Storage) DeleteBucket() error {
	return s.Bucket.Delete()
}
