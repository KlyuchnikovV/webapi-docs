package objects

type IParameter interface {
	NameParam() string
	Type() string
	EqualTo(interface{}) bool
}

type Parameter struct {
	In          string     `json:"in"`
	Name        string     `json:"name"`
	Required    bool       `json:"required"`
	Minimum     int        `json:"minimum,omitempty"`
	Description string     `json:"description,omitempty"`
	RequestBody *Reference `json:"requestBody,omitempty"`
	Schema      Schema     `json:"schema,omitempty"`
}
