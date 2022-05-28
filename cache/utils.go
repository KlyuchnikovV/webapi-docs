package cache

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/pkg"
	"github.com/KlyuchnikovV/webapi-docs/types"
)

func UnwrapImportedType(s types.ImportedType) (types.Type, error) {
	if err := ParsePackage(s.Package); err != nil {
		return nil, err
	}

	return FindModelByName(s.Name())
}

func ParseDirectory(path, localPath string) (map[string]pkg.Package, error) {
	paths, err := getDirectoriesPaths(path)
	if err != nil {
		return nil, err
	}

	var (
		packages = make(map[string]pkg.Package)
		fileSet  token.FileSet
	)

	for _, path := range paths {
		pkgs, err := parser.ParseDir(&fileSet, path, nil, parser.AllErrors)
		if err != nil {
			return nil, err
		}

		for name, pak := range pkgs {
			var packagePath = localPath

			if name != "main" {
				packagePath = fmt.Sprintf(
					"%s/%s",
					localPath,
					name,
				)
			}

			packages[packagePath] = pkg.NewPackage(*pak)
		}
	}

	return packages, nil
}

func ParseFile(path, localPath string) (map[string]pkg.Package, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return ParseDirectory(path, localPath)
	}

	if !strings.HasSuffix(path, ".go") {
		return nil, nil
	}

	astFile, err := parser.ParseFile(&token.FileSet{}, "", file, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	return map[string]pkg.Package{
		"": pkg.NewPackage(ast.Package{
			Files: map[string]*ast.File{"": astFile},
		}),
	}, nil
}

func getDirectoriesPaths(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, nil
	}

	paths, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	var directoriesPaths = []string{path}

	for _, innerPath := range paths {
		subPaths, err := getDirectoriesPaths(filepath.Join(path, innerPath))
		if err != nil {
			return nil, err
		}

		directoriesPaths = append(directoriesPaths, subPaths...)
	}

	return directoriesPaths, nil
}
