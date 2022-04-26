package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Parser struct {
	variableName string
	services     []ast.SelectorExpr
	servers      []Server

	packages map[string]ast.Package

	notFoundImports []string

	Spec *SwaggerSpec

	gopath         string
	localPath      string
	apiPrefix      string
	loopController map[string]struct{}
	fset           *token.FileSet
}

func NewParser(localPath, gopath string) Parser {
	return Parser{
		notFoundImports: make([]string, 0),
		Spec:            NewSwaggerSpec(),
		gopath:          gopath,
		localPath:       localPath,
		loopController:  make(map[string]struct{}),
		fset:            &token.FileSet{},
		packages:        make(map[string]ast.Package),
		apiPrefix:       "api",
	}
}

func (p *Parser) GenerateDocs(path string) (*SwaggerSpec, error) {
	if err := p.ParsePackages(path); err != nil {
		return nil, err
	}

	p.ExtractEngineData()

	if err := p.ParseServices(); err != nil {
		return nil, err
	}

	return p.Spec, nil
}

func (p *Parser) ParsePackages(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return p.parseFile(path, file, info)
	}

	return p.parseDir(path, file, info)
}

func (p *Parser) parseDir(path string, file *os.File, info os.FileInfo) error {
	if !info.IsDir() {
		return p.parseFile(path, file, info)
	}

	pkgs, err := parser.ParseDir(p.fset, path, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}

	var packageName = strings.Trim(path, "./")

	for _, pkg := range pkgs {
		if p.packages[packageName].Files == nil {
			pkg := p.packages[packageName]
			pkg.Files = make(map[string]*ast.File)
			p.packages[packageName] = pkg
		}

		for name, file := range pkg.Files {
			p.packages[packageName].Files[name] = file
		}
	}

	paths, err := file.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, innerPath := range paths {
		var filePath = filepath.Join(path, innerPath)

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := p.parseDir(filePath, file, info); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Parser) parseFile(path string, file *os.File, info os.FileInfo) error {
	if info.IsDir() {
		return p.parseDir(path, file, info)
	}

	if !strings.HasSuffix(path, ".go") {
		return nil
	}

	astFile, err := parser.ParseFile(p.fset, "", file, parser.AllErrors)
	if err != nil {
		return err
	}

	p.packages[""] = ast.Package{
		Files: map[string]*ast.File{
			"": astFile,
		},
	}

	var typeSpec *ast.StarExpr

	ast.Inspect(astFile, func(n ast.Node) bool {
		fun, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if fun.Recv == nil || len(fun.Recv.List) == 0 {
			return true
		}

		if err := CheckFuncDeclaration(*fun, "Routers", nil, CheckRoutersResultType); err != nil {
			return true
		}

		typeSpec, ok = fun.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			err = fmt.Errorf("receiver of 'Routers' is not a pointer")
			return false
		}

		return false
	})

	if err != nil {
		return err
	}

	ast.Inspect(astFile, func(n ast.Node) bool {
		fun, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		for _, result := range fun.Type.Results.List {
			if SameNodes(typeSpec, result.Type) {
				if err := p.ParseService("", *astFile, *fun); err != nil {
					return true
				}
				break
			}
		}

		return true
	})

	return nil
}

func (p *Parser) ExtractEngineData() {
	for _, pkg := range p.packages {
		for _, file := range pkg.Files {
			var webapiPkgAlias = p.findWebapiImport(*file)

			if webapiPkgAlias == "" {
				// Do not parse files that are not related to webapi.
				continue
			}

			ast.Inspect(file, func(n ast.Node) bool {
				p.getVarName(n, webapiPkgAlias)

				p.getAPIPrefix(n)

				p.getServiceSelectors(n)

				return true
			})
		}
	}
}

func (p *Parser) findWebapiImport(file ast.File) string {
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

func (p *Parser) getVarName(n ast.Node, alias string) {
	if n == nil {
		return
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
			return
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
		return
	}

	callExpr, ok := typed.(*ast.CallExpr)
	if !ok {
		return
	}

	if !IsMethod(*callExpr, NewSelector(alias, "New")) {
		return
	}

	p.variableName = name

	url := strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")

	if len(url) > 0 && url[0] == ':' {
		url = fmt.Sprintf("http://localhost%s", url)
	}

	p.servers = append(p.servers, Server{
		URL: url,
	})
}

func (p *Parser) getAPIPrefix(n ast.Node) {
	if n == nil {
		return
	}

	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}

	if !IsMethod(*callExpr, NewSelector(p.variableName, "WithPrefix")) {
		return
	}

	p.apiPrefix = strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")

	return
}

func (p *Parser) getServiceSelectors(n ast.Node) {
	if n == nil {
		return
	}

	call, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	if selector.Sel.Name != "RegisterServices" {
		// TODO:
		return
	}

	for _, arg := range call.Args {
		call, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		if selector.Sel == nil {
			return
		}

		p.services = append(p.services, *selector)
	}
}
