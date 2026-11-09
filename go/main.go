package main

import (
	"fmt"
	"os"
)

func run() ExitCode {
	// TODO: parse flags, stream input, apply filters, aggregate metrics, and emit output.
	// See README.md for the full CLI specification and output formats.
	argv := os.Args[1:]
	args, err := parseArgs(argv)
	if err != nil {
		fmt.Println(err)
		usage()
		return ExitValidationError
	}
	fmt.Println(args)
	return ExitSuccess
}

func main() {
	rc := run()
	os.Exit(int(rc))
}
