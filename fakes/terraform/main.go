package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	backendURL string
)

func main() {
	if checkFastFail() {
		log.Fatal("failed to terraform")
	}

	if contains(os.Args, "region=fail-to-terraform") {
		err := ioutil.WriteFile("terraform.tfstate", []byte(`{"key":"partial-apply"}`), os.ModePerm)
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

		err = ioutil.WriteFile("terraform.tfstate", []byte(`{"key":"value"}`), os.ModePerm)
		if err != nil {
			panic(err)
		}

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		fmt.Printf("working directory: %s\n", dir)
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
