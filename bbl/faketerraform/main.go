package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	if os.Args[1] == "fast-fail" {
		log.Fatal("failed to terraform")
	}

	err := ioutil.WriteFile("terraform.tfstate", []byte("hello-world"), os.ModePerm)
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

func removeBrackets(contents string) string {
	contents = strings.Replace(contents, "[", "", -1)
	contents = strings.Replace(contents, "]", "", -1)
	return contents
}
