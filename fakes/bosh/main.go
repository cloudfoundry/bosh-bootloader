package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	backendURL string
)

func main() {
	if os.Args[1] == "create-env" {
		if checkFastFail(os.Args[1]) {
			log.Fatal("failed to bosh")
		}

		postArgsToBackendServer(os.Args[1], os.Args[1:])

		fmt.Printf("bosh %s/n", removeBrackets(fmt.Sprintf("%+v", os.Args)))
	}
}

func postArgsToBackendServer(command string, args []string) {
	postArgs, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}

	_, err = http.Post(fmt.Sprintf("%s/%s/args", backendURL, command), "application/json", strings.NewReader(string(postArgs)))
	if err != nil {
		panic(err)
	}
}

func removeBrackets(contents string) string {
	contents = strings.Replace(contents, "[", "", -1)
	contents = strings.Replace(contents, "]", "", -1)
	return contents
}

func checkFastFail(command string) bool {
	resp, err := http.Get(fmt.Sprintf("%s/%s/fastfail", backendURL, command))
	if err != nil {
		panic(err)
	}

	return resp.StatusCode == http.StatusInternalServerError
}
