package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	parser2 "github.com/KlyuchnikovV/webapi-docs/parser"
)

var fset = token.NewFileSet()

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("File path must be provided\n")
		os.Exit(1)
	}

	var (
		path = os.Args[1] // positions are relative to fset
	)

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Println(err)
		return
	}

	if fileInfo.IsDir() {
		// file is a directory
		pkgs, err := ParseDir(path)
		if err != nil {
			log.Println(err)
			return
		}

		basePath, gopath := getBasePath(path)
		if err := parsePackages(pkgs, basePath, gopath); err != nil {
			log.Println(err)
			return
		}
	} else {
		// Parse src but stop after processing the imports.
		file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			log.Println(err)
			return
		}

		if err := parseFile(*file); err != nil {
			log.Println(err)
			return
		}
	}
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

func ParseDir(path string) (map[string]ast.Package, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, nil
	}

	var result = make(map[string]ast.Package)

	pkgs, err := parser.ParseDir(fset, path, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		var path = strings.Trim(path, "./")

		if result[path].Files == nil {
			pkg := result[path]
			pkg.Files = make(map[string]*ast.File)
			result[path] = pkg
		}

		for name, file := range pkg.Files {
			result[path].Files[name] = file
		}
	}

	paths, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	for _, innerPath := range paths {
		pkgs, err := ParseDir(path + "/" + innerPath)
		if err != nil {
			return nil, err
		}

		for name, files := range pkgs {
			result[name] = files
		}
	}

	return result, nil
}

type engineDef struct {
	variableName string
	prefix       string
	services     []ast.SelectorExpr
	imports      map[string]string

	servers []parser2.Server
}

func parsePackages(pkgs map[string]ast.Package, path, gopath string) error {
	var (
		engineDef = engineDef{
			services: make([]ast.SelectorExpr, 0),
			imports:  make(map[string]string),
			prefix:   "api",
		}
	)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			var webapiPkgAlias = engineDef.findWebapiImport(*file)

			if webapiPkgAlias == "" {
				// Do not parse files that are not related to webapi.
				continue
			}

			engineDef.getVarName(*file, webapiPkgAlias)

			engineDef.getApiPrefix(*file)

			engineDef.getServiceSelectors(*file)
		}
	}

	p := parser2.NewParser(path, gopath, engineDef.prefix, pkgs, engineDef.servers...)

	file, err := p.ParseServices(engineDef.services)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(file, "", "\t")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", string(bytes))

	return nil
}

func (d *engineDef) getServiceSelectors(file ast.File) {
	ast.Inspect(&file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if selector.Sel.Name != "RegisterServices" {
			// TODO:
			return true
		}

		for _, arg := range call.Args {
			call, ok := arg.(*ast.CallExpr)
			if !ok {
				continue
			}

			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if selector.Sel == nil {
				return true
			}

			d.services = append(d.services, *selector)
		}

		return true
	})
}

func (d *engineDef) findWebapiImport(file ast.File) string {
	var result string

	for i, imp := range file.Imports {
		var alias string

		if file.Imports[i].Name != nil {
			alias = file.Imports[i].Name.Name
		} else {
			alias = strings.Trim(
				file.Imports[i].Path.Value[strings.LastIndex(file.Imports[i].Path.Value, "/")+1:], "\"",
			)
		}

		if imp.Path.Value == "\"github.com/KlyuchnikovV/webapi\"" {
			result = alias
		}
	}

	return result
}

func (d *engineDef) getApiPrefix(file ast.File) {
	ast.Inspect(&file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if !parser2.IsMethod(*callExpr, parser2.NewSelector(d.variableName, "WithPrefix")) {
			return true
		}

		d.prefix = strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")
		return false
	})
}

func (d *engineDef) getVarName(file ast.File, alias string) {
	ast.Inspect(&file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		var (
			typed ast.Expr
			name  string
		)
		switch t := n.(type) {
		case *ast.AssignStmt:
			typed = t.Rhs[0]
			i, ok := t.Lhs[0].(*ast.Ident)
			if !ok {
				return true
			}
			name = i.Name
		case *ast.KeyValueExpr:
			typed = t.Value

			switch n := t.Key.(type) {
			case *ast.Ident:
				name = n.Name
			case *ast.BasicLit:
				name = n.Value
			}
		default:
			return true
		}

		callExpr, ok := typed.(*ast.CallExpr)
		if !ok {
			return true
		}

		if parser2.IsMethod(*callExpr, parser2.NewSelector(alias, "New")) {
			d.variableName = name

			url := strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")

			if len(url) > 0 && url[0] == ':' {
				url = fmt.Sprintf("http://localhost%s", url)
			}

			d.servers = append(d.servers, parser2.Server{
				Url: url,
			})

			return false
		}

		return true
	})

}

func parseFile(file ast.File) error {
	// fmt.Printf("found %#v\n", file.Scope.Objects)

	p := parser2.NewParser("", "", "", map[string]ast.Package{"": {
		Files: map[string]*ast.File{file.Name.Name: &file},
	}})

	for _, s := range file.Decls {
		switch typed := s.(type) {
		case *ast.FuncDecl:
			if typed.Recv == nil {
				continue
			}

			if len(typed.Recv.List) == 0 {
				continue
			}

			switch t := typed.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				fmt.Printf("params %s - %#v\n", typed.Name, t.X.(*ast.Ident).Obj.Decl.(*ast.TypeSpec).Type)
			}

			if err := p.ParseRouters(file, "api", *typed); err != nil {
				return err
			}

			bytes, err := json.MarshalIndent(p, "", "\t")
			if err != nil {
				return err
			}

			fmt.Printf("%s", string(bytes))
		}
	}

	return nil
}
