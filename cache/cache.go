package cache

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

// TODO: cache of all files for searching for models and methods

var innerCache *cache

type cache struct {
	gopath    string
	localPath string

	newPackages map[string]types.Package
}

func Init(gopath, localPath string, path string) {
	packages, err := ParseDirectory(path, localPath)
	if err != nil {
		panic(err)
	}

	innerCache = &cache{
		gopath:      gopath,
		localPath:   localPath,
		newPackages: packages,
	}
}

func GetPackages() map[string]types.Package {
	return innerCache.newPackages
}

func GetNewPackage(pkg string) types.Package {
	return innerCache.newPackages[pkg]
}

func ParsePackage(path string) error {
	absolutePath := filepath.Join(innerCache.gopath, "src", path)

	pkgs, err := parser.ParseDir(&token.FileSet{}, absolutePath, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	if len(path) == 0 {
		return nil
	}

	for name, pkg := range pkgs {
		var (
			index   = strings.Index(name, "/")
			newPath = path
		)

		if index != -1 {
			newPath = fmt.Sprintf("%s/%s", path, name[index+1:])
		}

		innerCache.newPackages[newPath] = types.NewPackage(*pkg)
	}

	return nil
}

func FindAliasOfWebAPIInFile(file ast.File) string {
	var result string

	for _, imp := range file.Imports {
		if imp.Path.Value != "\"github.com/KlyuchnikovV/webapi\"" {
			continue
		}

		if imp.Name != nil {
			result = imp.Name.Name
			break
		}

		result = strings.Trim(
			imp.Path.Value[strings.LastIndex(imp.Path.Value, "/")+1:],
			"\"",
		)

		break
	}

	return result
}

func FindModelByName(name string) (types.Type, error) {
	var obj types.Type

	for _, pkg := range innerCache.newPackages {
		for _, t := range pkg.Types {
			if t.Name() == name {
				obj = t
				break
			}
		}

		if obj != nil {
			break
		}
	}

	if obj == nil {
		return nil, fmt.Errorf("no model found for name: '%s'", name)
	}

	if imp, ok := obj.(types.ImportedType); ok {
		return UnwrapImportedType(imp)
	}

	return obj, nil
}

func FindModel(selector ast.SelectorExpr) types.Type {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return nil
	}

	pkg, ok := innerCache.newPackages[ident.Name]
	if !ok {
		return nil
	}

	return pkg.Types[selector.Sel.Name]
}

func FindMethod2(selector ast.SelectorExpr) types.FuncType {
	var model types.Type

	switch typed := selector.X.(type) {
	case *ast.SelectorExpr:
		model = FindModel(*typed)
	case *ast.CallExpr:
		return FindMethod2(*typed.Fun.(*ast.SelectorExpr))
	default:
		panic("not ok")
	}

	if model == nil {
		panic("!ok")
	}

	return *model.Method(selector.Sel.Name)
}
