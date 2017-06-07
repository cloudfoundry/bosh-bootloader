package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
)

var (
	backendURL string
)

func main() {
	if os.Args[1] == "interpolate" {
		writeVariablesToFile()
		postArgsToBackendServer(os.Args[1], os.Args[1:])
		fmt.Fprintf(os.Stderr, "bosh director name: %s\n", extractDirectorName(os.Args))

		if callRealInterpolate() {
			fmt.Fprintf(os.Stderr, "running real interpolate")
			runRealInterpolate(os.Args[1:])
		}
	}

	if os.Args[1] == "create-env" {
		incrementCallCountOnBackendServer(os.Args[1])
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

	if os.Args[1] == "-v" {
		resp, err := http.Get(fmt.Sprintf("%s/version", backendURL))
		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			panic(err)
		}

		fmt.Printf("some-text version %s some-more-text", string(body))
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
	variables := fmt.Sprintf(`
admin_password: rhkj9ys4l9guqfpc9vmp
director_ssl:
  certificate: some-certificate
  private_key: some-private-key
  ca: some-ca
jumpbox_ssh:
  private_key: |
    %s
  public_key: |
    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDfyo+svyrc4/yTS1Z1ov2WFFHaZsytgGcmzaYOET+vyyGZUd4aFxhhY0UanWhmJOCrqudP1CtLEm+uCptRuqCG8ZP63Dd5sSoxVK5pZuDMkCgqVFL4aG++LecAzWjrW0txfRNtWHB+O2gbSPgYmWUEwGDP1jdoSNcvdzPXoWjVkpPzCTk1AlCGn+dGQTasFRxGTLSZcuY2vuK6bnRnffQg2MjbgH3hSk87eST6sUyVwgOxsVej50lc6Grc0Px/6t151Zu/erXxaoZJpNF4dwRHOsGfPg/YMnT9dBttfGOVWtUcBDvxvnRpWLEOejNQLVG1VaYi0fGlKQretG8LTLlZ
`, strings.Replace(testhelpers.JUMPBOX_SSH_KEY, "\n", "\n    ", -1))
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

func incrementCallCountOnBackendServer(command string) {
	_, err := http.Get(fmt.Sprintf("%s/%s/call-count", backendURL, command))
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
		if strings.Contains(arg, "deployment-vars") {
			contents, err := ioutil.ReadFile(arg)
			if err != nil {
				panic(err)
			}

			variables := strings.Split(string(contents), "\n")
			for _, variable := range variables {
				if strings.HasPrefix(variable, "director_name") {
					return strings.TrimLeft(variable, "director_name: ")
				}
			}
		}
	}
	return ""
}

func callRealInterpolate() bool {
	resp, err := http.Get(fmt.Sprintf("%s/call-real-interpolate", backendURL))
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body) == "true"
}

func runRealInterpolate(args []string) {
	originalPath := getOriginalPath()
	os.Setenv("PATH", originalPath)
	cmd := exec.Command("bosh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func getOriginalPath() string {
	resp, err := http.Get(fmt.Sprintf("%s/path", backendURL))
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body)
}
