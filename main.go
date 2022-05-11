package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/parser"
)

// TODO: service prefix

// TODO: body parsing
// TODO: merge objects and types
// TODO: move on our type definitions and cache
// TODO: move types from cache
// TODO: refactor parser
// TODO: create cli
// TODO: review error responses and add builtin packages support
// TODO: add vendor and mod folders support
// TODO: add privacy configurations
// TODO: add linters mode

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("File path must be provided\n")
		os.Exit(1)
	}

	var (
		path                  = os.Args[1]
		basePath, gopath, err = getBasePath(path)
		parser                = parser.NewParser(basePath, gopath)
	)

	if err != nil {
		panic(err)
	}

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

func getBasePath(path string) (string, string, error) {
	var (
		gopath     = os.Getenv("GOPATH")
		srcDirPath = filepath.Join(gopath, "src/")
	)

	if len(gopath) == 0 {
		// TODO: disable error in private mode
		return "", "", fmt.Errorf("GOPATH must be provided")
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", "", err
	}

	return absolutePath[strings.LastIndex(absolutePath, srcDirPath)+len(srcDirPath):], gopath, nil
}
