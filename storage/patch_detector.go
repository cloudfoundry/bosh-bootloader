package storage

import (
	"os"
	"path/filepath"
)

type detectorLogger interface {
	Printf(message string, a ...interface{})
	Println(message string)
}

type PatchDetector struct {
	logger detectorLogger
	walker *walker
}

func NewPatchDetector(path string, logger detectorLogger) PatchDetector {
	return PatchDetector{logger: logger, walker: &walker{path: path}}
}

func (p PatchDetector) Find() error {
	collectedPaths, err := p.walker.walk()
	if err != nil {
		return err
	}
	if len(collectedPaths) > 0 {
		p.logger.Println("\nyou've supplied the following files to bbl:\n")
	}
	for _, path := range collectedPaths {
		p.logger.Printf("\t%s\n", path)
	}
	if len(collectedPaths) > 0 {
		p.logger.Println("\nthey will be used by \"bbl up\".\n")
	}
	return nil
}

type walker struct {
	path           string
	collectedPaths []string
}

func (w *walker) walk() ([]string, error) {
	err := filepath.Walk(w.path, w.printIfForeign)
	return w.collectedPaths, err
}

func (w *walker) printIfForeign(path string, info os.FileInfo, err error) error {
	if err != nil { // untested
		return err
	}

	if path == w.path {
		return nil
	}

	relPath, err := filepath.Rel(w.path, path)
	if err != nil { // untested
		return err
	}

	if info.IsDir() {
		return nil
	}

	if isUserManaged(relPath) && !isBBLManaged(relPath) {
		w.collectedPaths = append(w.collectedPaths, relPath)
	}

	return nil
}
