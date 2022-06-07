package types

import "encoding/json"

type BuiltIn string

func NewBuiltIn(name string) BuiltIn {
	return BuiltIn(name)
}

func (b BuiltIn) GetName() string {
	return string(b)
}

func (b BuiltIn) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}
