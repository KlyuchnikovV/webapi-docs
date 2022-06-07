package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/KlyuchnikovV/webapi-docs/parser"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TODO: review error responses and add builtin packages support
// TODO: add vendor and mod folders support

var (
	app = kingpin.New("wadocs", "")
	// verbose    = app.Flag("verbose", "Verbose mode.").Short('v').Bool()
	// privacyMode = app.Flag("privacy", "Privacy mode.").Default("none").String() // TODO.

	parseCmd        = app.Command("parse", "")
	inputParserPath = parseCmd.Arg("input", "Path to work with.").Required().String()
	outputPath      = parseCmd.Arg("output", "Output path.").Required().String()

	lintCmd = app.Command("lint", "")
	// inputLinterPath = *lintCmd.Arg("input", "Path to work with.").Required().String()
	// verbose    = lintCmd.Flag("level", "").Short('l').Bool()
)

func main() {
	var (
		cmd = kingpin.MustParse(
			app.Parse(os.Args[1:]),
		)
		err error
	)

	switch cmd {
	case parseCmd.FullCommand():
		p, _, err := parser.Parse(*inputParserPath)
		if err != nil {
			panic(err)
		}

		bytes, err := json.MarshalIndent(p, "", "\t")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(bytes))

		// err = parseInput()
	case lintCmd.FullCommand():
		// err = lintInput()
	default:
		err = fmt.Errorf("unknown command '%s'", cmd)
	}

	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	fmt.Printf("Successfully parsed '%s'", *inputParserPath)
	os.Exit(0)
}

/*
func parseInput() error {
	parser, err := parser.NewParser(*inputParserPath)
	if err != nil {
		return err
	}

	spec, err := parser.GenerateDocs()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(spec, "", "\t")
	if err != nil {
		return err
	}

	output, err := os.OpenFile(*outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = output.Write(bytes)

	return err
}

func lintInput() error {
	return fmt.Errorf("linter mode currently unsupported") // TODO:
}
*/
