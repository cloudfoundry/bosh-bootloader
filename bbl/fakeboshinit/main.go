package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

var (
	FailFast = "false"
)

func main() {
	fmt.Printf("bosh-init was called with %+v\n", os.Args)

	if FailFast == "true" {
		fmt.Fprintln(os.Stderr, "failing fast...")
		os.Exit(1)
	}

	contents, err := ioutil.ReadFile("bosh-state.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("bosh-state.json: %s\n", contents)

	var stateJson map[string]string
	err = json.Unmarshal(contents, &stateJson)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	oldManifestChecksum := stateJson["md5checksum"]

	manifestContents, err := ioutil.ReadFile("bosh.yml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("bosh director name: %s\n", extractDirectorName(manifestContents))
	newManifestChecksum := fmt.Sprintf("%x", md5.Sum(manifestContents))
	stateContents := fmt.Sprintf(`{"key": "value", "md5checksum": %q}`, newManifestChecksum)

	err = ioutil.WriteFile("bosh-state.json", []byte(stateContents), os.FileMode(0644))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if oldManifestChecksum == newManifestChecksum {
		fmt.Println("No new changes, skipping deployment...")
	}
}

func extractDirectorName(manifestContents []byte) string {
	var manifest struct {
		Jobs []struct {
			Properties struct {
				Director struct {
					Name string
				}
			}
		}
	}

	err := yaml.Unmarshal(manifestContents, &manifest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(manifest.Jobs) == 0 {
		return ""
	}

	return manifest.Jobs[0].Properties.Director.Name
}
