package main

import (
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

	err = ioutil.WriteFile("bosh-state.json", []byte(`{"key": "value"}`), os.FileMode(0644))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
