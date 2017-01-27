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
		log.Fatal("failed to bosh")
	}

	if os.Args[1] == "create-env" {
		postArgs, err := json.Marshal(os.Args[1:])
		if err != nil {
			panic(err)
		}

		_, err = http.Post(fmt.Sprintf("%s/args", backendURL), "application/json", strings.NewReader(string(postArgs)))
		if err != nil {
			panic(err)
		}

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		variables := `
admin_password: rhkj9ys4l9guqfpc9vmp
director_ssl:
  certificate: some-certificate
  private_key: some-private-key
  ca: some-ca
`
		err = ioutil.WriteFile("variables.yml", []byte(variables), os.ModePerm)
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile("state.json", []byte(`{"key":"value"}`), os.ModePerm)
		if err != nil {
			panic(err)
		}

		fmt.Printf("working directory: %s\n", dir)
		fmt.Printf("bosh %s/n", removeBrackets(fmt.Sprintf("%+v", os.Args)))
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
