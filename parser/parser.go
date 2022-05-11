package parser

import (
	"fmt"
	"go/ast"
	"path/filepath"

	"strings"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	cacheTypes "github.com/KlyuchnikovV/webapi-docs/cache/types"
	"github.com/KlyuchnikovV/webapi-docs/constants"
	"github.com/KlyuchnikovV/webapi-docs/objects"
	"github.com/KlyuchnikovV/webapi-docs/service"
	"github.com/KlyuchnikovV/webapi-docs/types"
	"github.com/KlyuchnikovV/webapi-docs/utils"
)

type Parser struct {
	variableName string
	services     []ast.SelectorExpr

	notFoundImports []string

	Spec *types.OpenAPISpec

	gopath         string
	localPath      string
	apiPrefix      string
	loopController map[string]struct{}
}

func NewParser(localPath, gopath string) Parser {
	return Parser{
		notFoundImports: make([]string, 0),
		Spec:            types.NewOpenAPISpec(),
		gopath:          gopath,
		localPath:       localPath,
		loopController:  make(map[string]struct{}),
		apiPrefix:       "api",
	}
}

func (parser *Parser) GenerateDocs(path string) (*types.OpenAPISpec, error) {
	cache.Init2(parser.gopath, parser.localPath, path)

	if err := parser.extractEngineData(); err != nil {
		return nil, err
	}

	if err := parser.ParseServices(); err != nil {
		return nil, err
	}

	return parser.Spec, nil
}

func (parser *Parser) extractEngineData() error {
	for _, pkg := range cache.GetPackages() {

		// for _, fun := range pkg.Functions {
		// 	for _, stmt := range fun.Body { // TODO: lookup methods for body parsing
		// 		ast.Inspect(stmt, func(n ast.Node) bool {
		// 			call, ok := n.(*ast.CallExpr)
		// 			if !ok {
		// 				return true
		// 			}

		// 			fun, ok := call.Fun.(*ast.SelectorExpr)
		// 			if !ok {
		// 				return true
		// 			}

		// 			method := cache.FindMethod2(*fun)
		// 			fmt.Print(method)

		// 			return true
		// 		})
		// 	}
		// }

		for _, file := range pkg.Pkg.Files {
			var webapiPkgAlias, _ = utils.FindImportWithPath(*file, "github.com/KlyuchnikovV/webapi")

			if webapiPkgAlias == "" {
				// Do not parse files that are not related to webapi.
				continue
			}

			ast.Inspect(file, func(n ast.Node) bool {
				parser.getVarName(n, webapiPkgAlias)

				parser.getAPIPrefix(n)

				parser.getServiceSelectors(n)

				return true
			})
		}
	}

	if len(parser.services) != 0 {
		return nil
	}

	for _, pkg := range cache.GetPackages() {
		// for _, file := range pkg.Pkg.Files {
		if err := parser.getServiceSelector(pkg); err != nil {
			return err
		}
		// }
	}

	return nil
}

func (parser *Parser) ParseServices() error {
	// for _, selector := range parser.services {
	// 	if selector.Sel == nil {
	// 		continue
	// 	}

	for _, pkg := range cache.GetPackages() {
		// fmt.Printf("%#v\n", cache.FindMethod2(selector))

		// for _, file := range pkg.Files {
		// obj := file.Scope.Lookup(selector.Sel.Name)
		// if obj == nil {
		// 	continue
		// }

		// funcDecl, ok := obj.Decl.(*ast.FuncDecl)
		// if !ok {
		// 	continue
		// }

		for _, model := range pkg.Types {
			if fun := model.Method("Routers"); fun == nil {
				continue
			}

			var srv = service.New(pkg, model, "")
			if err := srv.Parse(); err != nil {
				return err
			}

			// if err := srv.Parse(*file, *funcDecl); err != nil {
			// 	return err
			// }

			for name, schema := range srv.Components.Schemas {
				parser.Spec.Components.Schemas[name] = schema
			}

			for name, parameter := range srv.Components.Parameters {
				parser.Spec.Components.Parameters[name] = parameter
			}

			for name, body := range srv.Components.RequestBodies {
				parser.Spec.Components.RequestBodies[name] = body
			}

			for name, response := range srv.Components.Responses {
				parser.Spec.Components.Responses[name] = response
			}

			for path, paths := range srv.Paths {
				path = filepath.Join("/", parser.apiPrefix, path)

				if _, ok := parser.Spec.Paths[path]; !ok {
					parser.Spec.Paths[path] = make(map[string]objects.Route)
				}

				for method, handler := range paths {
					parser.Spec.Paths[path][method] = handler
				}
			}

			// var srv = NewService(parser, *pkg, *funcDecl)
			// if err := srv.ParseService(*file, *funcDecl); err != nil {
			// 	return err
			// }
		}

	}

	// }
	// }

	return nil
}

func (parser *Parser) getVarName(n ast.Node, alias string) {
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

	parser.variableName = name

	url := strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")

	if len(url) > 0 && url[0] == ':' {
		url = fmt.Sprintf("http://localhost%s", url)
	}

	parser.Spec.Servers = append(parser.Spec.Servers, types.ServerInfo{
		URL: url,
	})
}

func (parser *Parser) getAPIPrefix(n ast.Node) {
	if n == nil {
		return
	}

	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}

	if !IsMethod(*callExpr, NewSelector(parser.variableName, "WithPrefix")) {
		return
	}

	parser.apiPrefix = strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")
}

func (parser *Parser) getServiceSelectors(n ast.Node) {
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
		// TODO: check type - using tags
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

		parser.services = append(parser.services, *selector)
	}
}

func (parser *Parser) getServiceSelector(pkg cacheTypes.Package) error {
	var (
		// typeSpecs = make([]*ast.StarExpr, 0)
		result = make([]cacheTypes.Type, 0)
	)

	for _, t := range pkg.Types {
		alias, imp := utils.FindImportWithPath(*t.File(), "github.com/KlyuchnikovV/webapi")
		if imp == nil {
			continue
		}

		if t.Implements(constants.RoutersInterface(alias, imp.Path.Value)) {
			result = append(result, t)
		}
	}

	// ast.Inspect(&file, func(n ast.Node) bool {
	// 	fun, ok := n.(*ast.FuncDecl)
	// 	if !ok {
	// 		return true
	// 	}

	// 	if fun.Recv == nil || len(fun.Recv.List) == 0 {
	// 		return true
	// 	}

	// 	if err := CheckFuncDeclaration(*fun, "Routers", nil, CheckRoutersResultType); err != nil {
	// 		return true
	// 	}

	// 	ts, ok := fun.Recv.List[0].Type.(*ast.StarExpr)
	// 	if !ok {
	// 		return true
	// 	}

	// 	typeSpecs = append(typeSpecs, ts)

	// 	return true
	// })

	// ast.Inspect(&file, func(n ast.Node) bool {
	// 	funcDecl, ok := n.(*ast.FuncDecl)
	// 	if !ok {
	// 		return true
	// 	}

	// 	if len(funcDecl.Type.Results.List) != 1 {
	// 		return true
	// 	}

	// 	for _, typeSpec := range typeSpecs {
	// 		if SameNodes(typeSpec, funcDecl.Type.Results.List[0].Type) {
	// 			parser.services = append(parser.services, NewSelector("", funcDecl.Name.Name))
	// 			break
	// 		}
	// 	}

	// 	return true
	// })

	return nil
}
