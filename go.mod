module github.com/KlyuchnikovV/webapi-docs

go 1.17

require (
	github.com/teris-io/cli v1.0.1
	github.com/KlyuchnikovV/stack v0.0.0-20210427103552-f6f21f8b4227
	github.com/KlyuchnikovV/webapi v0.0.0-20220325204625-c03b5b6e4f1d
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace (
	github.com/KlyuchnikovV/webapi => ../webapi
)
