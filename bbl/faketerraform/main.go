package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if os.Args[1] == "fast-fail" {
		log.Fatal("failed to terraform")
	}

	fmt.Printf("terraform %s/n", removeBrackets(fmt.Sprintf("%+v", os.Args)))
}

func removeBrackets(contents string) string {
	contents = strings.Replace(contents, "[", "", -1)
	contents = strings.Replace(contents, "]", "", -1)
	return contents
}
