package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/sha3"
)

func ParseMd(content string) string {
	cleanned := strings.ReplaceAll(content, "*", "")
	cleanned = strings.ReplaceAll(cleanned, "`", "")
	cleanned = strings.ReplaceAll(cleanned, "# ", "")
	cleanned = strings.ReplaceAll(cleanned, "#", "")
	cleanned = strings.ReplaceAll(cleanned, "[", "<")
	cleanned = strings.ReplaceAll(cleanned, "]", ">")
	return cleanned
}

func main() {
	filename := os.Args[1]
	operation := os.Args[2]
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	parsedContent := ParseMd(string(content))
	if operation == "parse" {
		fmt.Println(parsedContent)
		os.Exit(0)
	} else if operation == "hash" {
		hashObject := sha3.New512()
		hashObject.Write([]byte(parsedContent))
		digest := hashObject.Sum(nil)
		fmt.Printf("%x\n", digest)
	} else {
		fmt.Println("Unknown operation")
		os.Exit(1)
	}
}
