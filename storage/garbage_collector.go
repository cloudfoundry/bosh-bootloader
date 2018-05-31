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

	g.fs.RemoveAll(filepath.Join(dir, "bosh-deployment"))
	g.fs.RemoveAll(filepath.Join(dir, "jumpbox-deployment"))
	g.fs.RemoveAll(filepath.Join(dir, "bbl-ops-files"))

	tfDir := filepath.Join(dir, "terraform")
	g.fs.Remove(filepath.Join(tfDir, "bbl-template.tf")) // terraform/bbl-template
	g.fs.RemoveAll(filepath.Join(tfDir, ".terraform"))
	g.fs.Remove(tfDir)

	ccDir := filepath.Join(dir, "cloud-config")
	g.fs.Remove(filepath.Join(ccDir, "cloud-config.yml"))
	g.fs.Remove(filepath.Join(ccDir, "ops.yml"))
	g.fs.Remove(ccDir)

	vFiles, _ := g.fs.ReadDir(filepath.Join(dir, "vars"))
	for _, f := range vFiles {
		varPath := filepath.Join("vars", f.Name())
		if isBBLManaged(varPath) {
			_ = g.fs.Remove(filepath.Join(dir, varPath))
		}
	}
	g.fs.Remove(filepath.Join(dir, "vars"))

	g.fs.RemoveAll(filepath.Join(dir, ".terraform"))
	g.fs.Remove(filepath.Join(dir, "create-jumpbox.sh"))
	g.fs.Remove(filepath.Join(dir, "create-director.sh"))
	g.fs.Remove(filepath.Join(dir, "delete-jumpbox.sh"))
	g.fs.Remove(filepath.Join(dir, "delete-director.sh"))

	return nil
}
