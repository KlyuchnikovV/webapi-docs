package types

import (
	"go/ast"
)

type Type interface {
	Name() string
	AddMethod(FuncType)

	Field(string) Type
	Method(string) *FuncType
	Fields() map[string]Type
	File() *ast.File

	EqualTo(t Type) bool
	Implements(InterfaceType) bool
}

func NewType(file *ast.File, name string, ts *ast.Expr) Type {
	var result Type

	switch typed := (*ts).(type) {
	case *ast.ArrayType:
		result = NewArray(file, name, typed, &typed.Elt)
	case *ast.StructType:
		result = NewStruct(file, name, typed)
	case *ast.InterfaceType:
		result = NewInterface(file, name, typed)
	case *ast.Ident:
		result = NewBasic(file, name, typed)
	case *ast.SelectorExpr:
		result = NewImported(file, typed)
	case *ast.StarExpr:
		result = NewType(file, name, &typed.X)
	case *ast.MapType:
		result = NewMap(file, name, typed)
	default:
		panic(typed)
	}

	return result
}

func NewTypeFromField(file *ast.File, field *ast.Field) (string, Type) {
	var name string
	if len(field.Names) != 0 {
		name = field.Names[0].Name
	}

	var t = NewType(file, name, &field.Type)
	if name == "" {
		name = t.Name()
	}

	return name, t
}

type typeBase struct {
	name   string
	fields map[string]Type
	file   *ast.File
}

func newTypeBase(file *ast.File, name string) typeBase {
	return typeBase{
		name:   name,
		fields: make(map[string]Type),
		file:   file,
	}
}

func (tb typeBase) Name() string {
	return tb.name
}

func (tb typeBase) AddMethod(f FuncType) {
	tb.fields[f.Name()] = f
}

func (tb typeBase) Field(name string) Type {
	var field = tb.fields[name]

	if _, ok := field.(FuncType); ok {
		return nil
	}

	return field
}

func (tb typeBase) Method(name string) *FuncType {
	var field = tb.fields[name]

	if fun, ok := field.(FuncType); ok {
		return &fun
	}

	return nil
}

func (tb typeBase) Fields() map[string]Type {
	return tb.fields
}

func (tb typeBase) File() *ast.File {
	return tb.file
}

func (tb typeBase) EqualTo(t Type) bool {
	if tb.name != t.Name() {
		return false
	}

	if tb.file != t.File() {
		return false
	}

	for name, field := range tb.fields {
		if !t.Field(name).EqualTo(field) {
			return false
		}
	}

	return true
}

func (tb typeBase) Implements(it InterfaceType) bool {
	for name, field := range it.Fields() {
		method, ok := field.(FuncType)
		if !ok {
			continue
		}

		tbField, ok := tb.fields[name]
		if !ok {
			return false
		}

		tbMethod, ok := tbField.(FuncType)
		if !ok {
			continue
		}

		if !tbMethod.EqualTo(method) {
			return false
		}
	}

	return true
}
