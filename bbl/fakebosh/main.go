package main

import (
	"crypto/md5"
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
	if os.Args[1] == "interpolate" {
		writeVariablesToFile()
		postArgsToBackendServer(os.Args[1], os.Args[1:])
		fmt.Fprintf(os.Stderr, "bosh director name: %s\n", extractDirectorName(os.Args))
	}

	if os.Args[1] == "create-env" {
		if checkFastFail(os.Args[1]) {
			log.Fatal("failed to bosh")
		}
		oldArgsChecksum := getOldArgMD5()
		argsChecksum := calculateArgMD5(os.Args[1:])

		postArgsToBackendServer(os.Args[1], os.Args[1:])
		writeStateToFile(fmt.Sprintf(`{"key":"value", "md5checksum": "%s"}`, argsChecksum))
		writeVariablesToFile()

		if oldArgsChecksum == argsChecksum {
			fmt.Println("No new changes, skipping deployment...")
		}

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fmt.Printf("working directory: %s\n", dir)
		fmt.Printf("bosh %s/n", removeBrackets(fmt.Sprintf("%+v", os.Args)))
	}

	if os.Args[1] == "delete-env" {
		if checkFastFail(os.Args[1]) {
			log.Fatal("failed to bosh")
		}
	}
}

func getOldArgMD5() string {
	contents, err := ioutil.ReadFile("state.json")
	if err != nil {
		return ""
	}

	var stateJson map[string]string
	err = json.Unmarshal(contents, &stateJson)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return stateJson["md5checksum"]
}

func writeVariablesToFile() {
	variables := `
admin_password: rhkj9ys4l9guqfpc9vmp
director_ssl:
  certificate: some-certificate
  private_key: some-private-key
  ca: some-ca
`
	err := ioutil.WriteFile("variables.yml", []byte(variables), os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func writeStateToFile(stateContents string) {
	err := ioutil.WriteFile("state.json", []byte(stateContents), os.ModePerm)
	if err != nil {
		panic(err)
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

func calculateArgMD5(args []string) string {
	var argString string
	path := strings.Trim(args[1], "manifest.yml")
	for _, arg := range args {
		arg = strings.Replace(arg, path, "", 1)
		argString = fmt.Sprintf("%s %s", argString, arg)
	}

	return fmt.Sprintf("%x", md5.Sum([]byte(argString)))
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

	if resp.StatusCode == http.StatusInternalServerError {
		writeStateToFile(`{"partial":"bosh-state"}`)
	}

	return resp.StatusCode == http.StatusInternalServerError
}

func extractDirectorName(args []string) string {
	for _, arg := range args {
		if strings.HasPrefix(arg, "director_name=") {
			return strings.TrimLeft(arg, "director_name=")
		}
	}

	return ""
}
