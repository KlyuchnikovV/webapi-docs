package parser

// import (
// 	"go/ast"
// 	"go/parser"
// 	"go/token"
// 	"os"
// 	"strings"
// )

// type Package struct {
// 	ast.Package
// }

// func NewPackage(pkg ast.Package) *Package {
// 	var result = &Package{
// 		Package: pkg,
// 	}

// 	if result.Files == nil {
// 		result.Files = make(map[string]*ast.File)
// 	}

// 	return result
// }

// func ParseDirectory(path string) (map[string]*Package, error) {
// 	paths, err := getDirectoriesPaths(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var packages = make(map[string]*Package)

// 	for _, path := range paths {
// 		pkgs, err := parser.ParseDir(&token.FileSet{}, path, nil, parser.AllErrors)
// 		if err != nil {
// 			return nil, err
// 		}

// 		var packageName = strings.Trim(path, "./")

// 		for _, pkg := range pkgs {
// 			packages[packageName] = NewPackage(*pkg)
// 		}
// 	}

// 	return packages, nil
// }

// func ParseFile(path string) (map[string]*Package, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	info, err := file.Stat()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if info.IsDir() {
// 		return ParseDirectory(path)
// 	}

// 	if !strings.HasSuffix(path, ".go") {
// 		return nil, nil
// 	}

// 	astFile, err := parser.ParseFile(&token.FileSet{}, "", file, parser.AllErrors)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return map[string]*Package{
// 		"": NewPackage(ast.Package{
// 			Files: map[string]*ast.File{"": astFile},
// 		}),
// 	}, nil
// }

// func (pkg *Package) FindMethod(t ast.TypeSpec, selector ast.FuncDecl) *ast.FuncDecl {
// 	var result *ast.FuncDecl

// 	for _, file := range pkg.Files {
// 		ast.Inspect(file, func(n ast.Node) bool {
// 			funcDecl, ok := n.(*ast.FuncDecl)
// 			if !ok {
// 				return true
// 			}

// 			if !IsReceiverOfType(*funcDecl, t) {
// 				return true
// 			}

// 			if SameNodes(funcDecl, &selector) {
// 				result = funcDecl
// 				return false
// 			}

// 			return true
// 		})
// 	}

// 	return result
// }

// func (pkg *Package) Implements(t ast.TypeSpec, i ast.InterfaceType) bool {
// 	var (
// 		methods = make(map[string]*ast.FuncType)
// 		found   = make(map[string]struct{})
// 	)

// 	for _, method := range i.Methods.List {
// 		funcType, ok := method.Type.(*ast.FuncType)
// 		if !ok {
// 			continue
// 		}

// 		methods[method.Names[0].Name] = funcType
// 	}

// 	for _, file := range pkg.Files {
// 		ast.Inspect(file, func(n ast.Node) bool {
// 			funcDecl, ok := n.(*ast.FuncDecl)
// 			if !ok {
// 				return true
// 			}

// 			if !IsReceiverOfType(*funcDecl, t) {
// 				return true
// 			}

// 			interfaceFunc, ok := methods[funcDecl.Name.Name]
// 			if !ok {
// 				return true
// 			}

// 			if SameNodes(funcDecl.Type, interfaceFunc) {
// 				found[funcDecl.Name.Name] = struct{}{}
// 			}

// 			return true
// 		})
// 	}

// 	for key := range methods {
// 		if _, ok := found[key]; !ok {
// 			return false
// 		}
// 	}

// 	return true
// }

// func (parser *Parser) FindMethodByName(t ast.TypeSpec, selector string) *ast.FuncDecl {
// 	var result *ast.FuncDecl

// 	for _, pkg := range parser.packages {

// 		for _, file := range pkg.Files {
// 			ast.Inspect(file, func(n ast.Node) bool {
// 				funcDecl, ok := n.(*ast.FuncDecl)
// 				if !ok {
// 					return true
// 				}

// 				if !IsReceiverOfType(*funcDecl, t) {
// 					return true
// 				}

// 				if funcDecl.Name.Name == selector {
// 					result = funcDecl
// 					return false
// 				}

// 				return true
// 			})
// 		}
// 	}

// 	return result
// }
