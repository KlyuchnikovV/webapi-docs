package main

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"regexp"
// 	"strings"

// 	"github.com/KlyuchnikovV/stack"
// 	"github.com/teris-io/cli"
// )

// func main2() {
// 	co := cli.NewCommand("parse", "parse given file or directory").
// 		WithShortcut("p").
// 		WithArg(cli.NewArg("path", "path to file being parsed")).
// 		WithAction(func(args []string, options map[string]string) int {
// 			// do something

// 			if err := parseDir(args[0]); err != nil {
// 				panic(err)
// 			}

// 			return 0
// 		})

// 	app := cli.New("webapi documentation parser").
// 		WithCommand(co)
// 		// no action attached, just print usage when executed

// 	os.Exit(
// 		app.Run(os.Args, os.Stdout),
// 	)
// }

// func parseDir(path string) error {
// 	files, err := ioutil.ReadDir(path)
// 	if err != nil {
// 		return err
// 	}

// 	for _, file := range files {
// 		if file.IsDir() {
// 			if err := parseDir(path + "/" + file.Name()); err != nil {
// 				return err
// 			}
// 		} else if strings.HasSuffix(file.Name(), ".go") {
// 			if err := parseFile(path + "/" + file.Name()); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// // func (e *Engine) parse(lines []types.Line) error {
// // 	var stack = stack.New(10)

// // 	// for line := 0; line < len(lines); line++ {
// // 	// 	for token := 0; token < len(lines[line]); token++ {
// // 	// 		switch
// // 	// 	}
// // 	// }

// // 	return nil
// // }

// func parseFile(path string) error {
// 	var e = new(Engine)

// 	bytes, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		return err
// 	}

// 	lines := ToLines(Tokenize(string(bytes)))

// 	e.Parse(lines)

// 	fmt.Println(e)

// 	return nil
// }

// type Engine struct {
// 	routes map[string]string

// 	imports   map[string]string
// 	types     map[string]string
// 	functions map[string]string
// }

// func (e *Engine) Parse(lines []Line) {
// 	e.parse(lines, 0, 0)
// }

// func (e *Engine) parse(lines []Line, line, token int) (int, int) {
// 	var (
// 		i = line
// 		j = token
// 	)

// 	for ; i < len(lines); i++ {
// 		for j = token; j < len(lines[i]); j++ {
// 			switch lines[i][j].value {
// 			case "import":
// 				i, j = e.parseImports(lines, i, j+1)
// 			case "func":
// 				matched, err := regexp.Match(
// 					"func \\(.+\\) Routers\\(\\) map\\[string\\].*RouterByPath{", []byte(lines[i].String()),
// 				)

// 				if err != nil {
// 					panic(err)
// 				}

// 				if matched {
// 					i, j = e.parseRoutes(lines, i+1, 0)
// 				}
// 			}
// 		}

// 		token = 0
// 	}

// 	return i, j
// }

// func (e *Engine) parseFunction()

// func (e *Engine) parseImports(lines []Line, line, token int) (int, int) {
// 	var (
// 		i = line
// 		j = token
// 	)

// 	if e.imports == nil {
// 		e.imports = make(map[string]string)
// 	}

// 	for ; i < len(lines); i++ {
// 		var alias string

// 		for j = token; j < len(lines[i]); j++ {
// 			switch lines[i][j].class {
// 			case OpeningBrace:
// 				if lines[i][j].value != "(" {
// 					return i, j
// 				}
// 			case ClosingBrace:
// 				return i, j
// 			case Symbols:
// 				if strings.HasPrefix(lines[i][j].value, "\"") {
// 					if alias == "" {
// 						alias = strings.Trim(lines[i][j].value[strings.LastIndex(lines[i][j].value, "/")+1:], "\"")
// 					}

// 					e.imports[alias] = lines[i][j].value
// 				} else {
// 					alias = lines[i][j].value
// 				}
// 			}
// 		}

// 		token = 0
// 	}

// 	return i, j
// }

// func (e *Engine) parseRoutes(lines []Line, line int) (int, int) {
// 	var (
// 		i       = line
// 		stack   = stack.New(10)
// 		started bool
// 	)

// 	stack.Push("{")

// 	if e.imports == nil {
// 		e.imports = make(map[string]string)
// 	}

// 	for ; i < len(lines); i++ {
// 		for j := 0; j < len(lines[i]); j++ {
// 			switch lines[i][j].class {
// 			case Symbols:
// 			case OpeningBrace:
// 			case ClosingBrace:
// 			}

// 			if started {

// 			}

// 			switch lines[i][j].value {
// 			case "return":
// 				matched, err := regexp.Match(
// 					"return map\\[string\\].*RouterByPath{", []byte(lines[i].String()),
// 				)

// 				if err != nil {
// 					panic(err)
// 				}

// 				if matched {
// 					started = true
// 					i++
// 					j = 0
// 					stack.Push("{")
// 				}
// 			}
// 		}
// 	}

// 	return i, j
// }
