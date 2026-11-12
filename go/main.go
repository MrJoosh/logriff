package main

import (
	"bufio"
	"fmt"
	"os"
)

func run() ExitCode {
	// TODO: parse flags, stream input, apply filters, aggregate metrics, and emit output.
	// See README.md for the full CLI specification and output formats.
	argv := os.Args[1:]
	args, err, help := parseArgs(argv)
	if err != nil {
		fmt.Println(err)
		usage()
		return ExitValidationError
	} else if help == true {
		usage()
		return ExitSuccess
	}
	var in *os.File
	if args.file == "-" {
		in = os.Stdin
	} else {
		in, err := os.Open(args.file)
		if err != nil {
			return ExitIOError
		}
		defer in.Close()
	}
	stats := stats{statusCodes: make(map[int]int), routeDistribution: make(map[string]int)}
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		stats.lines++
		parsed, err := parseLine(scanner.Text())
		if err != nil {
			if !args.strict {
				stats.error++
			}
			return ExitStrictError
		}
		validated, err := validator(parsed)
		if err != nil {
			if !args.strict {
				stats.error++
			}
			return ExitStrictError
		}
		if !filter(validated, args) {
			continue
		}
		stats.matched++
		stats.bytesVolume += validated.bytesSent
		stats.responseTimes = append(stats.responseTimes, validated.responseTimeMs)
		stats.statusCodes[validated.status]++
		stats.routeDistribution[validated.path]++
	}
	if err := scanner.Err(); err != nil {
		return ExitIOError
	}
	fmt.Println(stats)
	return ExitSuccess
}

func main() {
	rc := run()
	os.Exit(int(rc))
}
