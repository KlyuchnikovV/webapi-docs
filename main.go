package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/parser"
)

// TODO: apply middlewares as AsIsResponse
// TODO: Rename to smth like gogo
// TODO: merge objects and types
// TODO: move on our type definitions and cache
// TODO: move types from cache
// TODO: refactor parser
// TODO: create cli
// TODO: review error responses and add builtin packages support
// TODO: add vendor and mod folders support
// TODO: add privacy configurations
// TODO: add linters mode
// TODO: different parse models - with main, without, only file etc....

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("File path must be provided\n")
		os.Exit(1)
	}

	var path = os.Args[1]

	absPath, _, gopath, err := getBasePath(path)
	if err != nil {
		panic(err)
	}

	parser, err := parser.New(absPath, gopath)
	if err != nil {
		panic(err)
	}

	spec, err := parser.GenerateDocs(absPath)
	if err != nil {
		panic(err)
	}

	result, err := json.MarshalIndent(spec, "", "\t")
	if err != nil {
		panic(err)
	}

	if len(os.Args) <= 2 {
		fmt.Printf("%s", string(result))
		return
	}

	file, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		panic(err)
	}

	if _, err := file.Write(result); err != nil {
		panic(err)
	}

	outPath, err := filepath.Abs(os.Args[2])
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully generated file: %q", outPath)
}

func getBasePath(path string) (string, string, string, error) {
	var (
		gopath     = os.Getenv("GOPATH")
		srcDirPath = filepath.Join(gopath, "src/")
	)

	if len(gopath) == 0 {
		// TODO: disable error in private mode
		return "", "", "", fmt.Errorf("GOPATH must be provided")
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", "", "", err
	}

	var basePath = absolutePath[strings.LastIndex(absolutePath, srcDirPath)+len(srcDirPath):]

	return absolutePath, strings.TrimSuffix(basePath, strings.TrimLeft(path, ".")), gopath, nil
}
