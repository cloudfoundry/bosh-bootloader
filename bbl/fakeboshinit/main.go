package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	fmt.Printf("bosh-init was called with %+v\n", os.Args)

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
	newManifestChecksum := fmt.Sprintf("%x", md5.Sum(manifestContents))
	stateContents := fmt.Sprintf(`{"key": "value", "md5checksum": "%s"}`, newManifestChecksum)

	err = ioutil.WriteFile("bosh-state.json", []byte(stateContents), os.FileMode(0644))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if oldManifestChecksum == newManifestChecksum {
		fmt.Println("No new changes, skipping deployment...")
	}
}
