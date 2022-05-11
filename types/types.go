package types

import (
	"github.com/KlyuchnikovV/webapi-docs/objects"
)

type (
	OpenAPISpec struct {
		Openapi    string                              `json:"openapi"`
		Info       Info                                `json:"info"`
		Servers    []ServerInfo                        `json:"servers"`
		Components objects.Components                  `json:"components"`
		Paths      map[string]map[string]objects.Route `json:"paths"`
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
		Paths:      make(map[string]map[string]objects.Route),
		Components: objects.NewComponents(),
	}
}
