package types

import (
	"encoding/json"
	"fmt"
	"go/ast"
)

type BasicType struct {
	*typeBase

	BasicType string
	Ident     *ast.Ident

	Format string `json:"format,omitempty"`
}

func (o BasicType) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type        SchemaType `json:"type"`
		Description string     `json:"description,omitempty"`
		Example     string     `json:"example,omitempty"`
		Required    bool       `json:"required,omitempty"`
		Format      string     `json:"format,omitempty"`
	}{
		Type:        o.Type,
		Description: o.Description,
		Example:     o.Example,
		Required:    o.Required,
		Format:      o.Format,
	})
}

func NewBasic(file *ast.File, name string, ident *ast.Ident, tag *ast.BasicLit) Type {
	if name == "" {
		name = ident.Name
	}

	return BasicType{
		typeBase:  newTypeBase(file, name, tag, ConvertFieldType(name)),
		BasicType: ident.Name,

		Ident:  ident,
		Format: GetFieldTypeFormat(name),
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
		return NewFunc(file, typed.Type, name, typed.Body.List, tag)
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
		typeBase:  newTypeBase(file, name, tag, ConvertFieldType(name)),
		BasicType: basic.Value,
	}
}

func NewSimpleBasicType(name SchemaType) BasicType {
	return BasicType{
		typeBase: &typeBase{
			name: string(name),
			Type: name,
		},

		BasicType: string(name),
	}
}

func (o BasicType) EqualTo(t Type) bool {
	basic, ok := t.(BasicType)
	if !ok {
		return false
	}

	if o.BasicType != basic.BasicType {
		return false
	}

	if o.Format != basic.Format {
		return false
	}

	return o.typeBase.EqualTo(basic.typeBase)
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
		typeBase: newTypeBase(nil, "", nil, StringSchemaType),
		Data:     basic.Value,

		Basic: basic,
	}
}

func (s StringType) Name() string {
	return "string"
}

func (s StringType) EqualTo(sch Type) bool {
	rf, ok := sch.(StringType)
	if !ok {
		return false
	}

	if rf.Data != s.Data {
		return false
	}

	return s.typeBase.EqualTo(rf.typeBase)
}
