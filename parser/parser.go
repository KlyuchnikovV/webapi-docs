package parser

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/constants"
	"github.com/KlyuchnikovV/webapi-docs/service"
	"github.com/KlyuchnikovV/webapi-docs/types"
)

type Parser struct {
	GOPATH      string
	InitialPath string

	Spec      *types.OpenAPISpec
	apiPrefix string

	packages map[string]Package
}

func New(path, gopath string) (*Parser, error) {
	packages, err := ParseDirectory(path, gopath)
	if err != nil {
		return nil, err
	}

	return &Parser{
		GOPATH:      gopath,
		InitialPath: path,
		Spec:        types.NewOpenAPISpec(),
		packages:    packages,
		apiPrefix:   "api",
	}, nil
}

func (parser *Parser) GetPackages() map[string]Package {
	return parser.packages
}

func (parser *Parser) GetPackage(name string) (*Package, error) {
	pkg, ok := parser.packages[name]
	if !ok {
		return nil, fmt.Errorf("package '%s' not found", name)
	}

	return &pkg, nil
}

func (parser *Parser) UnwrapImportedType(s types.ImportedType) (types.Type, error) {
	packages, err := ParsePackage(parser.GOPATH, s.Package)
	if err != nil {
		return nil, err
	}

	for key, value := range packages {
		parser.packages[key] = value
	}

	pkg, err := parser.GetPackage(s.Package)
	if err != nil {
		return nil, err
	}

	return pkg.FindModelByName(s.Name(), parser.UnwrapImportedType)
}

func (parser *Parser) FindModel(selector ast.SelectorExpr) types.Type {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return nil
	}

	pkg, ok := parser.packages[ident.Name]
	if !ok {
		return nil
	}

	return pkg.Types[selector.Sel.Name]
}

func (parser *Parser) FindMethod(selector ast.SelectorExpr) types.FuncType {
	var model types.Type

	switch typed := selector.X.(type) {
	case *ast.SelectorExpr:
		model = parser.FindModel(*typed)
	case *ast.CallExpr:
		sel, ok := typed.Fun.(*ast.SelectorExpr)
		if !ok {
			panic("not ok")
		}

		return parser.FindMethod(*sel)
	default:
		panic("not ok")
	}

	if model == nil {
		panic("!ok")
	}

	return *model.Method(selector.Sel.Name)
}

func (parser *Parser) FindCall(fun types.FuncType, t types.Type) *types.Call {
	var result *types.Call

	for _, stmt := range fun.Body {
		ast.Inspect(stmt, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			call := types.NewCall(fun.File(), "", callExpr, nil)
			if !call.Call.EqualTo(t) {
				return true
			}

			result = &call

			return false
		})

		if result != nil {
			return result
		}
	}

	return nil
}

func (parser *Parser) GenerateDocs(path string) (*types.OpenAPISpec, error) {
	var services = parser.getServices()
	for prefix, model := range services {
		if err := parser.ParseService(prefix, model); err != nil {
			return nil, err
		}
	}

	return parser.Spec, nil
}

func (parser *Parser) ParseService(prefix string, model types.Type) error {
	for _, pkg := range parser.GetPackages() {
		if _, ok := pkg.Types[model.Name()]; !ok {
			continue
		}

		var srv = service.New(parser, model, prefix)
		if err := srv.Parse(); err != nil {
			return err
		}

		parser.Spec.Components.Add(srv.Components)

		for path, paths := range srv.Paths {
			path = filepath.Join("/", parser.apiPrefix, path)

			if _, ok := parser.Spec.Paths[path]; !ok {
				parser.Spec.Paths[path] = make(map[string]types.Route)
			}

			for method, handler := range paths {
				parser.Spec.Paths[path][method] = handler
			}
		}
	}

	return nil
}

func (parser *Parser) getServices() map[string]types.Type {
	var services = make(map[string]types.Type)

	for _, pkg := range parser.GetPackages() {
		for _, fun := range pkg.Functions {
			parser.getEngineInfo(fun)
		}

		for key, value := range pkg.GetServices() {
			services[key] = value
		}
	}

	return services
}

func (parser *Parser) getInfo() {
	var engineNew = types.NewSimpleImported("New", constants.WebapiPath)

	for _, pkg := range parser.GetPackages() {
		for _, fun := range pkg.Functions {
			var call = parser.FindCall(fun, engineNew)
			if call == nil {
				continue
			}

			// call.Parameters[0]
		}
	}
}

// TODO: replace
func (parser *Parser) getEngineInfo(fun types.FuncType) {
	var (
		file           = fun.File()
		engineNew      = types.NewSimpleImported("New", constants.WebapiPath)
		withPrefix     = types.NewSimpleImported("WithPrefix", constants.WebapiPath)
		responseAsJSON = types.NewSimpleImported("ResponseAsJSON", constants.WebapiPath)
		responseAsXML  = types.NewSimpleImported("ResponseAsXML", constants.WebapiPath)
	)

	for _, stmt := range fun.Body {
		ast.Inspect(stmt, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			imp := types.NewImported(file, sel, nil)

			if len(call.Args) < 1 {
				return true
			}

			switch {
			case imp.EqualTo(responseAsJSON):
				var t = types.NewType(nil, "", call.Args[0], nil)

				// TODO:
				fmt.Print(t)
			case imp.EqualTo(responseAsXML):
				// TODO:
			case imp.EqualTo(withPrefix):
				lit, ok := call.Args[0].(*ast.BasicLit)
				if !ok {
					return true
				}

				parser.apiPrefix = strings.Trim(lit.Value, "\"")
			case imp.EqualTo(engineNew):
				lit, ok := call.Args[0].(*ast.BasicLit)
				if !ok {
					return true
				}

				var url = strings.Trim(lit.Value, "\"")

				if len(url) > 0 && url[0] == ':' {
					url = fmt.Sprintf("http://localhost%s", url)
				}

				parser.Spec.Servers = append(parser.Spec.Servers,
					types.ServerInfo{
						URL: url,
					},
				)
			}

			return true
		})
	}
}
