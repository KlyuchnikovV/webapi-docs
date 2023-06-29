package types

import (
	"encoding/json"
	"go/ast"
	"regexp"
	"strings"
)

type (
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
	Tag(string) string
	AddMethod(FuncType)
	Field(string) Type
	Method(string) *FuncType
	Fields() map[string]Type
	File() *ast.File

	EqualTo(t Type) bool
	Implements(InterfaceType) bool
	SchemaType() SchemaType
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
		result = NewFunc(file, typed, name, nil, tag)
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
		result = NewFunc(file, typed.Type, name, nil, tag)
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
	tags   map[string]string

	Type        SchemaType `json:"type"`
	Description string     `json:"description,omitempty"`
	Example     string     `json:"example,omitempty"`
	Required    bool       `json:"required,omitempty"`
}

func (tb typeBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type        SchemaType `json:"type"`
		Description string     `json:"description,omitempty"`
		Example     string     `json:"example,omitempty"`
		Required    bool       `json:"required,omitempty"`
	}{
		Type:        tb.Type,
		Description: tb.Description,
		Example:     tb.Example,
		Required:    tb.Required,
	})
}

var tagRegex = regexp.MustCompile(`\b.+?:".+?"`)

func newTypeBase(file *ast.File, name string, tags *ast.BasicLit, t SchemaType) *typeBase {
	var result = typeBase{
		name:   name,
		fields: make(map[string]Type),
		file:   file,
		tags:   make(map[string]string),
		Type:   t,
	}

	if tags == nil {
		return &result
	}

	p := tagRegex.FindAllString(strings.Trim(tags.Value, "`"), -1)
	for _, piece := range p {
		var pieces = strings.Split(piece, ":")

		result.tags[strings.Trim(pieces[0], `"`)] = strings.Trim(pieces[1], `"`)
	}

	tag, ok := result.tags["json"]
	if ok {
		result.Required = !strings.Contains(tag, "omitempty")
	}

	result.Description = result.tags["description"]
	result.Example = result.tags["example"]

	return &result
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
	var fields = make(map[string]Type)

	for name, field := range tb.fields {
		if _, ok := field.(FuncType); !ok {
			fields[name] = field
		}
	}

	return fields
}

func (tb typeBase) Methods() map[string]Type {
	var fields = make(map[string]Type)

	for name, field := range tb.fields {
		if _, ok := field.(FuncType); ok {
			fields[name] = field
		}
	}

	return fields
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

	base, ok := t.(*typeBase)
	if !ok {
		return false
	}

	return tb.Type == base.Type &&
		tb.Description == base.Description &&
		tb.Example == base.Example &&
		tb.Required == base.Required
}

func (tb typeBase) Implements(it InterfaceType) bool {
	var cache = make(map[string]bool)

	for method := range it.fields {
		cache[method] = false
	}

	for name, field := range it.Methods() {
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

		cache[tbMethod.name] = true
	}

	for _, ok := range cache {
		if !ok {
			return false
		}
	}

	return true
}

func (tb typeBase) SchemaType() SchemaType {
	return tb.Type
}

func (tb typeBase) Tag(name string) string {
	if value, ok := tb.tags[name]; ok {
		return value
	}

	return ""
}

type Components struct {
	Schemas       map[string]Type        `json:"schemas,omitempty"`
	Parameters    map[string]IParameter  `json:"parameters,omitempty"`
	RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`
	Responses     map[string]Response    `json:"responses,omitempty"`

	loopController map[string]struct{}
}

func NewComponents() Components {
	return Components{
		Schemas:        make(map[string]Type),
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
	Description string             `json:"description,omitempty"`
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
	case *ast.StarExpr:
		return getBaseTypeAlias(typed.X, i)
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

type SchemaType string

const (
	EmptySchemaType   SchemaType = ""
	ArraySchemaType   SchemaType = "array"
	NumberSchemaType  SchemaType = "number"
	ObjectSchemaType  SchemaType = "object"
	StringSchemaType  SchemaType = "string"
	IntegerSchemaType SchemaType = "integer"
	BooleanSchemaType SchemaType = "boolean"
)
