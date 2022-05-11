package objects

import "fmt"

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

func (r Reference) NameParam() string {
	return r.name
}

func (r Reference) Type() string {
	return "ref"
}

func (r Reference) SchemaType() string {
	return "reference"
}

func (r Reference) EqualTo(p interface{}) bool {
	typed, ok := p.(Reference)
	if !ok {
		return false
	}

	return typed.Ref == r.Ref
}
