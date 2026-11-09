package main

import (
	"fmt"
	"os"
)

func main() {
	// TODO: parse flags, stream input, apply filters, aggregate metrics, and emit output.
	// See README.md for the full CLI specification and output formats.
	cliArgs := os.Args[1:]
	input, err := parseArgs(cliArgs)
	if err != nil {
		fmt.Println(err)
		usage()
		os.Exit(1)
	}
	fmt.Println(input)
	os.Exit(0)
}
