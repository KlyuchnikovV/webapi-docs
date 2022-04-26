package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	parser2 "github.com/KlyuchnikovV/webapi-docs/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("File path must be provided\n")
		os.Exit(1)
	}

	var (
		path             = os.Args[1]
		basePath, gopath = getBasePath(path)
		parser           = parser2.NewParser(basePath, gopath)
	)

	spec, err := parser.GenerateDocs(path)
	if err != nil {
		panic(err)
	}

	bytes, err := json.MarshalIndent(spec, "", "\t")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", string(bytes))
}

func getBasePath(path string) (string, string) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		log.Println(err)
		return "", ""
	}

	srcDirPath := os.Getenv("GOPATH") + "/src/"

	return absolutePath[strings.LastIndex(absolutePath, srcDirPath)+len(srcDirPath):], os.Getenv("GOPATH")
}
