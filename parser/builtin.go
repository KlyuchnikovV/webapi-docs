package parser

import "encoding/json"

type BuiltIn string

func NewBuiltIn(name string) BuiltIn {
	return BuiltIn(name)
}

func (b BuiltIn) GetName() string {
	return string(b)
}

func (b BuiltIn) EqualTo(t Type) bool {
	typed, ok := t.(*BuiltIn)
	if !ok {
		return false
	}

	return b == *typed
}

func (b BuiltIn) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}
