package parser

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	"github.com/KlyuchnikovV/webapi-docs/constants"
	"github.com/KlyuchnikovV/webapi-docs/service"
	"github.com/KlyuchnikovV/webapi-docs/types"
)

type Parser struct {
	services map[string]types.Type

	Spec *types.OpenAPISpec

	gopath    string
	localPath string
	apiPrefix string
}

func NewParser(localPath, gopath string) Parser {
	return Parser{
		Spec:      types.NewOpenAPISpec(),
		gopath:    gopath,
		localPath: localPath,
		apiPrefix: "api",
		services:  make(map[string]types.Type),
	}
}

func (parser *Parser) GenerateDocs(path string) (*types.OpenAPISpec, error) {
	cache.Init(parser.gopath, parser.localPath, path)

	var packages = cache.GetPackages()

	parser.extractEngineData(packages)

	if err := parser.ParseServices(packages); err != nil {
		return nil, err
	}

	return parser.Spec, nil
}

func (parser *Parser) extractEngineData(pkgs map[string]types.Package) {
	for _, pkg := range pkgs {
		for _, fun := range pkg.Functions {
			parser.getEngineInfo(fun)
		}

		for _, t := range pkg.Types {
			for _, fun := range t.Constructors() {
				parser.getEngineInfo(fun)
			}

			if !t.Implements(constants.RoutersInterface()) {
				continue
			}

			parser.getServiceSelector(t)
		}
	}
}

func (parser *Parser) getEngineInfo(fun types.FuncType) {
	var (
		file       = fun.File()
		engineNew  = types.NewSimpleImported("New", constants.WebapiPath)
		withPrefix = types.NewSimpleImported("WithPrefix", constants.WebapiPath)
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
			if imp == nil {
				return true
			}

			if !imp.EqualTo(engineNew) {
				if imp.EqualTo(withPrefix) {
					parser.apiPrefix = strings.Trim(call.Args[0].(*ast.BasicLit).Value, "\"")
				}

				return true
			}

			var url = strings.Trim(call.Args[0].(*ast.BasicLit).Value, "\"")

			if len(url) > 0 && url[0] == ':' {
				url = fmt.Sprintf("http://localhost%s", url)
			}

			parser.Spec.Servers = append(parser.Spec.Servers,
				types.ServerInfo{
					URL: url,
				},
			)

			return true
		})
	}
}

func (parser *Parser) ParseServices(pkgs map[string]types.Package) error {
	for prefix, model := range parser.services {
		for _, pkg := range pkgs {
			if _, ok := pkg.Types[model.Name()]; !ok {
				continue
			}

			var srv = service.New(pkg, model, prefix)
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
	}

	return nil
}

func (parser *Parser) getServiceSelector(serviceType types.Type) {
	for _, constructor := range serviceType.Constructors() {
		for _, ret := range constructor.ReturnStatements() {
			for _, res := range ret.Results {
				unary, ok := res.(*ast.UnaryExpr)
				if !ok {
					continue
				}

				composite, ok := unary.X.(*ast.CompositeLit)
				if !ok {
					continue
				}

				parser.parseServiceConstructor(serviceType, composite.Elts)
			}
		}
	}
}

func (parser *Parser) parseServiceConstructor(serviceType types.Type, exprs []ast.Expr) {
	for _, elt := range exprs {
		keyValue, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		star, ok := keyValue.Value.(*ast.StarExpr)
		if !ok {
			continue
		}

		call, ok := star.X.(*ast.CallExpr)
		if !ok {
			continue
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		if types.NewImported(serviceType.File(), sel, nil).EqualTo(
			types.NewSimpleImported("NewService", constants.WebapiPath),
		) {
			parser.services[strings.Trim(call.Args[1].(*ast.BasicLit).Value, "\"")] = serviceType
		}
	}
}
