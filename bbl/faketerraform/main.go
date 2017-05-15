package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
)

var (
	backendURL string
)

func main() {
	if checkFastFail() {
		log.Fatal("failed to terraform")
	}

	if testhelpers.Contains(os.Args, "region=fail-to-terraform") {
		err := ioutil.WriteFile("terraform.tfstate", []byte(`{"key":"partial-apply"}`), os.ModePerm)
		if err != nil {
			panic(err)
		}

		log.Fatal("failed to terraform")
	}

	if os.Args[1] == "version" {
		resp, err := http.Get(fmt.Sprintf("%s/version", backendURL))
		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			panic(err)
		}

		fmt.Printf("some-text v%s some-more-text", string(body))
	}

	if os.Args[1] == "output" {
		resp, err := http.Get(fmt.Sprintf("%s/output/%s", backendURL, os.Args[2]))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			fmt.Print(string(body))
		case http.StatusInternalServerError:
			fmt.Fprintf(os.Stderr, "Returning error in fake terraform.")
			os.Exit(1)
		}
	}

	if os.Args[1] == "apply" || os.Args[1] == "destroy" {
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
