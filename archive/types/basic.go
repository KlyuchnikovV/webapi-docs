package types

import (
	"fmt"
	"go/ast"

	"github.com/KlyuchnikovV/webapi-docs/utils"
)

type BasicType struct {
	*typeBase

	BasicType string
	Ident     *ast.Ident
}

func NewBasic(file *ast.File, name string, ident *ast.Ident, tag *ast.BasicLit) Type {
	if name == "" {
		name = ident.Name
	}

	return BasicType{
		typeBase:  newTypeBase(file, name, tag),
		BasicType: ident.Name,

		Ident: ident,
	}
}

func NewFromObject(file *ast.File, name string, obj *ast.Object, tag *ast.BasicLit) Type {
	switch typed := obj.Decl.(type) {
	case *ast.TypeSpec:
		return NewType(file, name, &typed.Type, tag)
	case *ast.AssignStmt:
		for i, variable := range typed.Lhs {
			v, ok := variable.(*ast.Ident)
			if !ok {
				continue
			}

			if obj.Name != v.Name {
				continue
			}

			if len(typed.Rhs) <= i {
				return NewType(file, name, &typed.Rhs[len(typed.Rhs)-1], tag)
			}

			return NewType(file, name, &typed.Rhs[i], tag)
		}
	case *ast.ValueSpec:
		if len(typed.Values) == 0 {
			return NewType(file, name, &typed.Type, tag)
		}

		return NewType(file, name, &typed.Values[0], tag)
	case *ast.Field:
		return NewType(file, name, &typed.Type, tag)
	case *ast.FuncDecl:
		return NewFunc(file, *typed, name, tag)
	default:
		panic("not ok")
	}

	return nil
}

func NewBasicFromBasicLit(file *ast.File, name string, basic, tag *ast.BasicLit) BasicType {
	if name == "" {
		name = basic.Value
	}

	return BasicType{
		typeBase:  newTypeBase(file, name, tag),
		BasicType: basic.Value,
	}
}

func NewSimpleBasicType(name string) BasicType {
	return BasicType{
		typeBase: &typeBase{
			name: name,
		},

		BasicType: name,
	}
}

func (b BasicType) EqualTo(t Type) bool {
	basic, ok := t.(BasicType)
	if !ok {
		return false
	}

	if b.BasicType != basic.BasicType {
		return false
	}

	return b.typeBase.EqualTo(basic)
}

func (b BasicType) Schema() Schema {
	return FieldSchema{
		Type:   utils.ConvertFieldType(b.name),
		Format: utils.GetFieldTypeFormat(b.name),
	}
}

type FieldSchema struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
}

func (f FieldSchema) EqualTo(s Schema) bool {
	fs, ok := s.(FieldSchema)
	if !ok {
		return false
	}

	if f.Type != fs.Type {
		return false
	}

	return f.Format == fs.Format
}

func (f FieldSchema) SchemaType() string {
	return f.Type
}

type Reference struct {
	Ref string `json:"$ref"`

	name string
}

func NewReference(name, where string) *Reference {
	return &Reference{
		Ref:  fmt.Sprintf("#/components/%s/%s", where, name),
		name: name,
	}
}

func (r Reference) EqualTo(ref interface{}) bool {
	rf, ok := ref.(Reference)
	if !ok {
		return false
	}

	return r.Ref == rf.Ref
}

func (r Reference) NameParam() string {
	return r.name
}

func (Reference) Type() string {
	return "ref"
}

type StringType struct {
	*typeBase

	Data  string
	Basic *ast.BasicLit
}

func NewString(basic *ast.BasicLit) StringType {
	return StringType{
		typeBase: newTypeBase(nil, "", nil),
		Data:     basic.Value,

		Basic: basic,
	}
}

func (s StringType) Name() string {
	return "string"
}

func (s StringType) Schema() Schema {
	return StringSchema{
		Type: "string",
	}
}

type StringSchema struct {
	Type string `json:"type"`
}

func (s StringSchema) EqualTo(sh Schema) bool {
	rf, ok := sh.(StringSchema)
	if !ok {
		return false
	}

	return s == rf
}

func (s StringSchema) SchemaType() string {
	return "string"
}

// func (r StringSchema) NameParam() string {
// 	return ""
// }

// func (StringSchema) Type() string {
// 	return "string"
// }
