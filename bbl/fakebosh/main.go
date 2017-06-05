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
	variables := `
admin_password: rhkj9ys4l9guqfpc9vmp
director_ssl:
  certificate: some-certificate
  private_key: some-private-key
  ca: some-ca
jumpbox_ssh:
  private_key: |
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEA38qPrL8q3OP8k0tWdaL9lhRR2mbMrYBnJs2mDhE/r8shmVHe
    GhcYYWNFGp1oZiTgq6rnT9QrSxJvrgqbUbqghvGT+tw3ebEqMVSuaWbgzJAoKlRS
    +Ghvvi3nAM1o61tLcX0TbVhwfjtoG0j4GJllBMBgz9Y3aEjXL3cz16Fo1ZKT8wk5
    NQJQhp/nRkE2rBUcRky0mXLmNr7ium50Z330INjI24B94UpPO3kk+rFMlcIDsbFX
    o+dJXOhq3ND8f+rdedWbv3q18WqGSaTReHcERzrBnz4P2DJ0/XQbbXxjlVrVHAQ7
    8b50aVixDnozUC1RtVWmItHxpSkK3rRvC0y5WQIDAQABAoIBAQCNVfKzWPCLHPmx
    VM0/8jZRiHfBhVcS5JtA6HRNQhuEvLd1izzIIXnmV7mW+36ps/SotoDr68WD3hrm
    QhCh50nmr7+TmWz30CojiaW1L6Idz5VuVl8oP10DMR5JZXEz4y6ceC/CyS4SqxYu
    1UDK2GXyQEVkPZg0pnwwoAn/zxLUflXC6+I25RsT/X5vDPW9l8VpFE4QjWT9XeOu
    ElfiX66IW5V6/y/SEll1pzueDM+M6ec1p/pS24iiLuNHWA4feyVJTQ8nzxQuYz1q
    yI4uQZLHIMiFCNY2nhoK2EZtocKPFo8DUuWph1k0uycv8GGm8gOd8spg1na27ja9
    iXJxEyF5AoGBAOnC1wvtIUzTmARtBY5S7s0ho3NueR5GUi5ku5S3TdFGCmUaA3zE
    2VYJ7h0eioabtJwY1tzsKxahJNeKkRgW7SUDBxnEltbN7WA/mDgu5d+NAT59Ho6T
    7A1DZIKCTo8hBvHzgSIang63fqoFco6WOv9JG4V7BHIegYISpweJbT/7AoGBAPUU
    6OB1GJTOgFgKNc04Xy5LlFE05mAGeaMaeGmsNSDJ5UAs+w0NiMObJ7DNqdM26rSn
    RQkt7sU8GECc9PtiY9TQTW9VNyXALC/vap5Ee/8bAuZ3dM08LtkuCpeIplOD5RAJ
    r7Buh+GbHPZ6LqL8f/5hWeeIUMXXc09HCP0Cc2e7AoGBAIP1nXf6EQZRnEtDUBOb
    9XqPNrn+7xiMEfBmpQ26vI8avtt75+QTK61KRcTibMi4NSi5TPHB0EEiDq4uZuH2
    b0CpiOSe+ZehABOJUuDEeLfN3Zns/8b08hg6pw6ViMt7lXQYRhl+dSNRqotIL/cW
    D4/1MTgUzdmuJuXKqcezaJzpAoGALnyj24d6fSdaQtjU8bNCopZlcK3XENnJkr1/
    n5OxlCGXoX+mswghK/EvKyMnlk+xX0jnGGGlC7ZlZ0QeV9yG0SQdvANu7XMxLnp8
    P77/whjOiQaZmiBTRpCsI6gg3HCFL3CW6aFdltaEPOBaHkJEyOyQUBGUOKKwVZZE
    xzEC0OcCgYBo70a2s0VzC8BR7BwcuNgVQp8e90+AeHHzgx1Z/yjxzkRw533Tlpj8
    eYYHipQy737P1l9vz2BX6YHn8Kos0Y1pzG7CqRjwrCfGRDM4DcA64P4oR55RULWa
    e2rINGOsVkW6atdh+5XwGMLS8QDccwaPMpcqdVbdo4c0YcfGRWgB3w==
    -----END RSA PRIVATE KEY-----
  public_key: |
    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDfyo+svyrc4/yTS1Z1ov2WFFHaZsytgGcmzaYOET+vyyGZUd4aFxhhY0UanWhmJOCrqudP1CtLEm+uCptRuqCG8ZP63Dd5sSoxVK5pZuDMkCgqVFL4aG++LecAzWjrW0txfRNtWHB+O2gbSPgYmWUEwGDP1jdoSNcvdzPXoWjVkpPzCTk1AlCGn+dGQTasFRxGTLSZcuY2vuK6bnRnffQg2MjbgH3hSk87eST6sUyVwgOxsVej50lc6Grc0Px/6t151Zu/erXxaoZJpNF4dwRHOsGfPg/YMnT9dBttfGOVWtUcBDvxvnRpWLEOejNQLVG1VaYi0fGlKQretG8LTLlZ
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
