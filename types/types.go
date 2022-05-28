package types

import (
	"encoding/json"
	"go/ast"
	"strings"
)

type (
	Schema interface {
		// TODO: review methods list
		SchemaType() string
		EqualTo(Schema) bool
	}

	OpenAPISpec struct {
		Openapi    string                      `json:"openapi"`
		Info       Info                        `json:"info"`
		Servers    []ServerInfo                `json:"servers"`
		Components Components                  `json:"components"`
		Paths      map[string]map[string]Route `json:"paths"`
	}

	Info struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	}

	ServerInfo struct {
		URL         string `json:"url"`
		Description string `json:"description"`
	}
)

func NewOpenAPISpec(servers ...ServerInfo) *OpenAPISpec {
	return &OpenAPISpec{
		Openapi: "3.0.3",
		Info: Info{
			Version: "3.0.3",
		},
		Servers:    servers,
		Paths:      make(map[string]map[string]Route),
		Components: NewComponents(),
	}
}

type Type interface {
	Name() string
	Tag() string
	AddMethod(FuncType)
	AddConstructor(FuncType)

	Field(string) Type
	Method(string) *FuncType
	Fields() map[string]Type
	Constructors() []FuncType
	File() *ast.File

	EqualTo(t Type) bool
	Implements(InterfaceType) bool
	Schema() Schema
}

func NewType(file *ast.File, name string, ts *ast.Expr, tag *ast.BasicLit) Type {
	var result Type

	switch typed := (*ts).(type) {
	case *ast.ArrayType:
		result = NewArray(file, name, typed, &typed.Elt, tag)
	case *ast.StructType:
		result = NewStruct(file, name, typed, tag)
	case *ast.InterfaceType:
		result = NewInterface(file, name, typed, tag)
	case *ast.Ident:
		result = NewBasic(file, name, typed, tag)
	case *ast.SelectorExpr:
		result = NewImported(file, typed, tag)
	case *ast.StarExpr:
		result = NewType(file, name, &typed.X, tag)
	case *ast.MapType:
		result = NewMap(file, name, typed, tag)
	case *ast.FuncType:
		result = NewFuncFromType(file, typed, name)
	case *ast.Ellipsis:
		result = NewType(file, name, &typed.Elt, tag)
	case *ast.ChanType:
		result = NewType(file, name, &typed.Value, tag)
	case *ast.UnaryExpr:
		result = NewType(file, name, &typed.X, tag)
	case *ast.CompositeLit:
		result = NewType(file, name, &typed.Type, tag)
	case *ast.CallExpr:
		result = NewType(file, name, &typed.Fun, tag)
	case *ast.IndexExpr:
		result = NewType(file, name, &typed.X, tag)
	case *ast.FuncLit:
		result = NewFuncFromType(file, typed.Type, name)
	case *ast.BinaryExpr:
		result = NewType(file, name, &typed.X, tag)
	case *ast.BasicLit:
		result = NewBasicFromBasicLit(file, name, typed, tag)
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

	var t = NewType(file, name, &field.Type, field.Tag)
	if name == "" {
		name = t.Name()
	}

	return name, t
}

type typeBase struct {
	name   string
	fields map[string]Type
	file   *ast.File
	tag    string

	constructors []FuncType
}

func (tb typeBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   string
		Fields map[string]Type
	}{
		Name:   tb.name,
		Fields: tb.fields,
	})
}

func newTypeBase(file *ast.File, name string, tag *ast.BasicLit) *typeBase {
	var t string
	if tag != nil {
		t = strings.Trim(strings.TrimPrefix(tag.Value, "`json:"), "\"`")
	}

	return &typeBase{
		name:         name,
		fields:       make(map[string]Type),
		file:         file,
		constructors: make([]FuncType, 0),
		tag:          t,
	}
}

func (tb typeBase) Name() string {
	return tb.name
}

func (tb typeBase) AddMethod(f FuncType) {
	tb.fields[f.Name()] = f
}

func (tb *typeBase) AddConstructor(f FuncType) {
	tb.constructors = append(tb.constructors, f)
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
	var fields = make(map[string]Type)

	for name, field := range tb.fields {
		if _, ok := field.(FuncType); !ok {
			fields[name] = field
		}
	}

	return fields
}

func (tb typeBase) Constructors() []FuncType {
	return tb.constructors
}

func (tb typeBase) File() *ast.File {
	return tb.file
}

func (tb typeBase) EqualTo(t Type) bool {
	if tb.name != t.Name() {
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

func (tb typeBase) Schema() Schema {
	return nil
}

func (tb typeBase) Tag() string {
	if tb.tag == "" {
		return tb.name
	}
	return tb.tag
}

type Components struct {
	Schemas       map[string]Schema      `json:"schemas,omitempty"`
	Parameters    map[string]IParameter  `json:"parameters,omitempty"`
	RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`
	Responses     map[string]Response    `json:"responses,omitempty"`

	loopController map[string]struct{}
}

func NewComponents() Components {
	return Components{
		Schemas:        make(map[string]Schema),
		Parameters:     make(map[string]IParameter),
		Responses:      make(map[string]Response),
		RequestBodies:  make(map[string]RequestBody),
		loopController: make(map[string]struct{}),
	}
}

func (c *Components) Add(comp Components) {
	for name, schema := range comp.Schemas {
		c.Schemas[name] = schema
	}

	for name, parameter := range comp.Parameters {
		c.Parameters[name] = parameter
	}

	for name, body := range comp.RequestBodies {
		c.RequestBodies[name] = body
	}

	for name, response := range comp.Responses {
		c.Responses[name] = response
	}
}

type Response struct {
	Description string             `json:"description"`
	Content     map[string]Content `json:"content,omitempty"`
}

type RequestBody struct {
	Description string             `json:"description,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Content     map[string]Content `json:"content"`
}

func NewRequestBody(ref Reference) RequestBody {
	return RequestBody{
		Content: map[string]Content{
			// TODO: multiple schemas
			"application/json": {
				Schema: ref,
			},
		},
	}
}

func NewResponse(desc string, refs ...Reference) *Response {
	var response = Response{
		Description: desc,
	}

	if len(refs) > 0 {
		response.Content = make(map[string]Content)
	}

	for _, ref := range refs {
		cnt := response.Content["application/json"]
		cnt.Schema = ref
		response.Content["application/json"] = cnt
	}

	return &response
}

func NewErrorResponse(desc string, refs ...Reference) *Response {
	var response = Response{
		Description: desc,
	}

	if len(refs) > 0 {
		response.Content = make(map[string]Content)
	}

	for _, ref := range refs {
		cnt := response.Content["text/plain"]
		cnt.Schema = ref
		response.Content["text/plain"] = cnt
	}

	return &response
}

type Content struct {
	Schema Reference `json:"schema"`
}

type Route struct {
	Summary     string             `json:"summary,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	Parameters  []IParameter       `json:"parameters,omitempty"`
	RequestBody *Reference         `json:"requestBody,omitempty"`
	Responses   map[int]*Reference `json:"responses"`
}

func NewRoute(tags ...string) *Route {
	return &Route{
		Parameters: make([]IParameter, 0),
		Responses:  make(map[int]*Reference),
		Tags:       tags,
	}
}

type IParameter interface {
	NameParam() string
	Type() string
	EqualTo(interface{}) bool
}

func getBaseTypeAlias(expr ast.Expr, i int) string {
	switch typed := expr.(type) {
	case *ast.SelectorExpr:
		return getBaseTypeAlias(typed.X, i)
	case *ast.Ident:
		if typed.Obj != nil {
			return getBaseTypeAliasFromObj(*typed.Obj, i)
		}

		return typed.Name
	}

	return ""
}

func getBaseTypeAliasFromObj(obj ast.Object, i int) string {
	switch typed := obj.Decl.(type) {
	case *ast.AssignStmt:
		for i, variable := range typed.Lhs {
			ident, ok := variable.(*ast.Ident)
			if !ok {
				continue
			}

			if obj.Name != ident.Name {
				continue
			}

			if len(typed.Rhs) <= i {
				return getBaseTypeAlias(typed.Rhs[len(typed.Rhs)-1], i)
			}

			return getBaseTypeAlias(typed.Rhs[i], i)
		}
	case *ast.ValueSpec:
		return getBaseTypeAlias(typed.Values[i], 0)
	case *ast.Field:
		return getBaseTypeAlias(typed.Type, i)
	case *ast.TypeSpec:
		return getBaseTypeAlias(typed.Type, i)
	default:
		panic("not ok")
	}

	return ""
}

func NewValue(expr ast.Expr, targetType Type) ReturnStatement {
	var result ReturnStatement

	switch typed := expr.(type) {
	// case *ast.ArrayType:
	// 	result = NewArray(file, name, typed, &typed.Elt, tag)
	// case *ast.StructType:
	// 	result = NewStruct(file, name, typed, tag)
	// case *ast.InterfaceType:
	// 	result = NewInterface(file, name, typed, tag)
	// case *ast.Ident:
	// 	result = NewBasic(file, name, typed, tag)
	case *ast.SelectorExpr:
		result = ReturnStatement{
			Type: targetType,
		}
	// 	result = NewImported(file, typed, tag)
	// case *ast.StarExpr:
	// 	result = NewType(file, name, &typed.X, tag)
	// case *ast.MapType:
	// 	result = NewMap(file, name, typed, tag)
	// case *ast.FuncType:
	// 	result = NewFuncFromType(file, typed, name)
	// case *ast.Ellipsis:
	// 	result = NewType(file, name, &typed.Elt, tag)
	// case *ast.ChanType:
	// 	result = NewType(file, name, &typed.Value, tag)
	case *ast.UnaryExpr:
		return NewValue(typed.X, targetType)
	case *ast.CompositeLit:
		result = ReturnStatement{
			Type: targetType,

			Value: typed,
		}
	// 	result = NewType(file, name, &typed.Type, tag)
	case *ast.CallExpr:
		return NewValue(typed.Fun, targetType)
	// 	result = NewType(file, name, &typed.Fun, tag)
	// case *ast.IndexExpr:
	// 	result = NewType(file, name, &typed.X, tag)
	// case *ast.FuncLit:
	// 	result = NewFuncFromType(file, typed.Type, name)
	// case *ast.BinaryExpr:
	// 	result = NewType(file, name, &typed.X, tag)
	// case *ast.BasicLit:
	// 	result = NewBasicFromBasicLit(file, name, typed, tag)
	default:
		panic(typed)
	}

	return result
}
