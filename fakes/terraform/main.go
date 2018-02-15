package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var (
	backendURL string
	stateFile  *string
)

func main() {
	if checkFastFail() {
		log.Fatal("failed to terraform")
	}

	if os.Args[1] == "apply" {
		flagSet := flag.NewFlagSet("apply", flag.PanicOnError)
		stateFile = flagSet.String("state", "fake-terraform.tfstate", "output tfvars")
		flagSet.Parse(os.Args[2:])
	} else {
		stateFile = flag.String("state", "fake-terraform.tfstate", "output tfvars")
		flag.Parse()
	}

	if contains(os.Args, "region=fail-to-terraform") {
		fmt.Printf("received args: %+v\n", os.Args)
		err := ioutil.WriteFile(*stateFile, []byte(`{"key":"partial-apply"}`), storage.StateMode)
		if err != nil {
			panic(err)
		}

		log.Fatal("failed to terraform")
	}

	if os.Args[1] == "apply" {
		postArgs, err := json.Marshal(os.Args[1:])
		if err != nil {
			panic(err)
		}

		_, err = http.Post(fmt.Sprintf("%s/args", backendURL), "application/json", strings.NewReader(string(postArgs)))
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(*stateFile, []byte(`{"key":"value"}`), storage.StateMode)
		if err != nil {
			panic(err)
		}

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		fmt.Printf("working directory: %s\n", dir)
		fmt.Printf("data directory: %s\n", os.Getenv("TF_DATA_DIR"))
		fmt.Printf("terraform %s/n", removeBrackets(fmt.Sprintf("%+v", os.Args)))
	}
}

func removeBrackets(contents string) string {
	contents = strings.Replace(contents, "[", "", -1)
	contents = strings.Replace(contents, "]", "", -1)
	return contents
}

func checkFastFail() bool {
	resp, err := http.Get(fmt.Sprintf("%s/fastfail", backendURL))
	if err != nil {
		panic(err)
	}

	return resp.StatusCode == http.StatusInternalServerError
}

func contains(slice []string, word string) bool {
	for _, item := range slice {
		if item == word {
			return true
		}
	}
	return false
}
