package main

import (
	"bufio"
	"fmt"
	"os"
)

func run(argv []string) ExitCode {
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
		in, err = os.Open(args.file)
		if err != nil {
			fmt.Println(err)
			return ExitIOError
		}
		defer in.Close()
	}
	stats := stats{statusCodes: make(map[int]int), routeDistribution: make(map[string]int)}
	re := getParseRegex()
	scanner := bufio.NewScanner(in)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		stats.lines++
		parsed, err := parseLine(scanner.Text(), re)
		if err != nil {
			if !args.strict {
				stats.error++
				continue
			}
			fmt.Println(err)
			return ExitStrictError
		}
		validated, err := validator(parsed)
		if err != nil {
			if !args.strict {
				stats.error++
				continue
			}
			fmt.Println(err)
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
		fmt.Println(err)
		return ExitIOError
	}
	switch args.format {
	case "json":
		_, err := fmt.Println(aggregate(stats).toJson(args.topRoutes))
		if err != nil {
			fmt.Println(err)
			return ExitIOError
		}
	case "text":
		_, err := fmt.Println(aggregate(stats).toText(args.topRoutes))
		if err != nil {
			fmt.Println(err)
			return ExitIOError
		}
	}
	return ExitSuccess
}

func main() {
	argv := os.Args[1:]
	rc := run(argv)
	os.Exit(int(rc))
}
