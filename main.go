package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goParser "go/parser"
	"go/token"

	"github.com/KlyuchnikovV/webapi-docs/parser"
)

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	// Parse src but stop after processing the imports.
	f, err := goParser.ParseFile(fset, "../webapi/example/service/request.go", nil, goParser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return
	}

	file := parser.NewFile()

	for _, s := range f.Decls {
		switch typed := s.(type) {
		case *ast.FuncDecl:
			handler := parser.FuncHandlers[typed.Name.Name]
			if handler == nil {
				continue
			}

			if err := handler(file, *typed); err != nil {
				panic(err)
			}

			bytes, err := json.MarshalIndent(file, "", "\t")
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s", string(bytes))
		}
	}
}
