package main

import (
	"github.com/KlyuchnikovV/webapi-docs/parser"
)

func generateDocumentation() (string, error) {
	for name, pkg := range parser.Pkgs {
		for _, fun := range pkg.Functions {
			parseFunction( name, fun)
		}
	}

	return "", nil
}

func parseFunction(pkgName string, decl *parser.FuncDecl) {
	var vars = decl.Variables(pkgName)

	for range vars {

		// for _, spec := range stmt.Rhs {
		// call, ok := spec.(*ast.CallExpr)
		// if !ok {
		// 	continue
		// }

		// sel, ok := call.Fun.(*ast.SelectorExpr)
		// if !ok {
		// 	continue
		// }

		// _, err := types.NewImported(sel, decl.File)
		// if err != nil {
		// 	continue
		// }

		// if !imp.EqualTo(engineNew) {
		// 	if imp.EqualTo(withPrefix) {
		// 		parser.apiPrefix = strings.Trim(call.Args[0].(*ast.BasicLit).Value, "\"")
		// 	}

		// 	return true
		// }

		// var url = strings.Trim(call.Args[0].(*ast.BasicLit).Value, "\"")

		// if len(url) > 0 && url[0] == ':' {
		// 	url = fmt.Sprintf("http://localhost%s", url)
		// }

		// parser.Spec.Servers = append(parser.Spec.Servers,
		// 	types.ServerInfo{
		// 		URL: url,
		// 	},
		// )

		// return true
		// }

	}
}

func parseRegisterServices() {

}
