package cache

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/cache/types"
	"github.com/KlyuchnikovV/webapi-docs/packages"
	"github.com/KlyuchnikovV/webapi-docs/utils"
)

// TODO: cache of all files for searching for models and methods

var innerCache *cache

type cache struct {
	packages  map[string]*packages.Package
	gopath    string
	localPath string

	newPackages map[string]types.Package
}

func Init(gopath, localPath string, packages map[string]*packages.Package) {
	innerCache = &cache{
		gopath:    gopath,
		localPath: localPath,
		packages:  packages,
	}
}

func Init2(gopath, localPath string, path string) {
	packages, err := packages.ParseDirectory(path)
	if err != nil {
		panic(err)
	}

	innerCache = &cache{
		gopath:      gopath,
		localPath:   localPath,
		packages:    packages,
		newPackages: make(map[string]types.Package),
	}

	for _, pkg := range packages {
		innerCache.newPackages[pkg.Name] = types.NewPackage(pkg.Package)
	}
}

func GetPackages() map[string]types.Package {
	return innerCache.newPackages
}

func GetNewPackage(pkg string) types.Package {
	return innerCache.newPackages[pkg]
}

func FindMethodByName(receiver ast.TypeSpec, name string) *ast.FuncDecl {
	var result *ast.FuncDecl

	var pkgName string

	for name, pkg := range innerCache.packages {
		for _, file := range pkg.Files {
			object := file.Scope.Lookup(receiver.Name.Name)
			if object == nil {
				continue
			}

			typeSpec, ok := object.Decl.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if packages.SameNodes(&receiver, typeSpec) {
				pkgName = name
				break
			}
		}
	}

	for _, file := range innerCache.packages[pkgName].Files {
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if funcDecl.Name.Name != name {
				return true
			}

			if !packages.IsReceiverOfType(*funcDecl, receiver) {
				return true
			}

			result = funcDecl
			return false
		})
	}

	return result
}

func FindMethod(receiver ast.TypeSpec, declaration ast.FuncDecl) *ast.FuncDecl {
	var method = FindMethodByName(receiver, declaration.Name.Name)

	if !utils.SameNodes(method, &declaration) {
		return nil
	}

	return method
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

		innerCache.packages[newPath] = packages.NewPackage(*pkg)
		innerCache.newPackages[newPath] = types.NewPackage(*pkg)
	}

	return nil
}

// func (c *cache) FindModelBySelector(name string) (*ast.Object, error) {
// 	var obj *ast.Object

// 	for _, pkg := range c.packages {
// 		for _, file := range pkg.Files {
// 			obj = file.Scope.Lookup(name)
// 			if obj != nil {
// 				break
// 			}
// 		}

// 		if obj != nil {
// 			break
// 		}
// 	}

// 	if obj == nil {
// 		return nil, fmt.Errorf("no model found for name: '%s'", name)
// 	}

// 	return obj, nil
// }

func FindAliasOfWebAPI(pkg, fileName string) string {
	return FindAliasOfWebAPIInFile(*innerCache.packages[pkg].Files[fileName])
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

		// ParsePackage()
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

func GetPackageByImportAlias(file ast.File, alias string) *packages.Package {
	var path string

	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name == alias {
				path = strings.Trim(imp.Path.Value, "\"")
				break
			}

			continue
		}

		if strings.HasSuffix(strings.Trim(imp.Path.Value, "\""), alias) {
			path = strings.Trim(imp.Path.Value, "\"")
			break
		}
	}

	pkg, ok := innerCache.packages[path]
	if ok {
		return pkg
	}

	if err := ParsePackage(path); err != nil {
		panic(err)
	}

	pkg, ok = innerCache.packages[path]
	if ok {
		return pkg
	}

	return nil
}
