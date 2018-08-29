package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type GarbageCollector struct {
	fs fs
}

func NewGarbageCollector(fs fs) GarbageCollector {
	return GarbageCollector{
		fs: fs,
	}
}

func (g GarbageCollector) Remove(dir string) error {
	bblStateJson := filepath.Join(dir, STATE_FILE)
	err := g.fs.Remove(bblStateJson)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Removing %s: %s", bblStateJson, err)
	}

	for _, relPath := range bblManaged {
		g.fs.RemoveAll(filepath.Join(dir, relPath))
	}

	for _, relPath := range bblManagedDirsWhichMayContainUserFiles {
		// this will not delete directories with files in them
		g.fs.Remove(filepath.Join(dir, relPath))
	}

	return nil
}
