package objects

import (
	"go/ast"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	"github.com/KlyuchnikovV/webapi-docs/cache/types"
)

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

type Content struct {
	Schema Reference `json:"schema"`
}

type Schema interface {
	// TODO: review methods list
	SchemaType() string
	EqualTo(interface{}) bool
}

func (c *Components) NewSchema(typeSpec ast.TypeSpec) (Schema, error) {
	var (
		schema Schema
		err    error
	)

	switch typed := typeSpec.Type.(type) {
	case *ast.StructType:
		schema, err = c.NewObject(*typed)
	case *ast.ArrayType:
		schema, err = c.NewArray(*typed)
	}

	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (c *Components) NewSchema2(typeSpec types.Type) (Schema, error) {
	var (
		schema Schema
		err    error
	)

	switch typed := typeSpec.(type) {
	case types.ArrayType:
		schema, err = c.NewArray(*typed.ArrayType)
	case types.BasicType:
		schema = c.NewField(*typed.Ident)
	case types.ImportedType:
		t, err := cache.UnwrapImportedType(typed)
		if err != nil {
			return nil, err
		}

		return c.NewSchema2(t)
	case types.StructType:
		schema, err = c.NewObject(*typed.StructType)
		// case types.InterfaceType:
		// case types.MapType:
	default:
		return nil, nil
		panic("not ok")
	}

	return schema, err
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
