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
	if path == w.path {
		return nil
	}

	relPath, err := filepath.Rel(w.path, path)
	if err != nil {
		return err
	}

	if info.IsDir() && isBBLManaged(relPath) {
		return filepath.SkipDir
	}

	if info.IsDir() {
		return nil
	}

	if isPatchRelevant(relPath) && !isBBLManaged(relPath) {
		w.collectedPaths = append(w.collectedPaths, relPath)
	}

	return nil
}

var bblManaged = []string{
	"bbl-state.json",
	"create-jumpbox.sh",
	"create-director.sh",
	"delete-jumpbox.sh",
	"delete-director.sh",

	// vars
	"vars/bbl.tfvars",
	"vars/bosh-state.json",
	"vars/cloud-config-vars.yml",
	"vars/director-vars-file.yml",
	"vars/director-vars-store.yml",
	"vars/jumpbox-state.json",
	"vars/jumpbox-vars-file.yml",
	"vars/jumpbox-vars-store.yml",
	"vars/terraform.tfstate",
	"vars/terraform.tfstate.backup",

	// terraform files
	"terraform/bbl-template.tf",
	"terraform/.terraform",

	// cloud config
	"cloud-config/cloud-config.yml",
	"cloud-config/ops.yml",

	// directories
	"jumpbox-deployment",
	"bosh-deployment",
	"bbl-ops-files",
}

// relPath must be from same dir as patterns above
func isBBLManaged(relPath string) bool {
	return matchesGlobList(relPath, bblManaged)
}

var patchRelevant = []string{
	"create-jumpbox-override.sh",
	"create-director-override.sh",
	"delete-jumpbox-override.sh",
	"delete-director-override.sh",

	"vars/*.tfvars",
	"terraform/*.tf",
	"cloud-config/*.yml",
}

func isPatchRelevant(relPath string) bool {
	return matchesGlobList(relPath, patchRelevant)
}

func matchesGlobList(relPath string, globList []string) bool {
	for _, pattern := range globList {
		matches, err := filepath.Match(pattern, relPath)
		if err != nil {
			panic(err) // only errors for malformed patterns
		}
		if matches {
			return true
		}
	}
	return false
}
