package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type Env struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Open   func(name string) (io.ReadCloser, error)
}

func DefaultEnv() Env {
	return Env{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Open: func(name string) (io.ReadCloser, error) {
			return os.Open(name)
		},
	}
}

func run(argv []string, env Env) ExitCode {
	l := log.New(env.Stderr, "", 0)
	args, err, help := parseArgs(argv)
	if err != nil {
		l.Println(err)
		usage(l)
		return ExitValidationError
	} else if help == true {
		usage(l)
		return ExitSuccess
	}

	var in io.Reader
	if args.file == "-" {
		in = env.Stdin
	} else {
		f, err := env.Open(args.file)
		if err != nil {
			l.Println(err)
			return ExitIOError
		}
		defer f.Close()
		in = f
	}

	stats := stats{statusCodes: make(map[int]int), routeDistribution: make(map[string]int)}
	re := getParseRegex()
	scanner := bufio.NewScanner(in)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 100*1024*1024)

	for scanner.Scan() {
		stats.lines++
		parsed, err := parseLine(scanner.Text(), re)
		if err != nil {
			if !args.strict {
				stats.error++
				continue
			}
			l.Println(err)
			return ExitStrictError
		}
		validated, err := validator(parsed)
		if err != nil {
			if !args.strict {
				stats.error++
				continue
			}
			l.Println(err)
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
		l.Println(err)
		return ExitIOError
	}

	switch args.format {
	case "json":
		_, err := fmt.Fprintln(env.Stdout, aggregate(stats).toJson(args.topRoutes))
		if err != nil {
			l.Println(err)
			return ExitIOError
		}
	case "text":
		_, err := fmt.Fprintln(env.Stdout, aggregate(stats).toText(args.topRoutes))
		if err != nil {
			l.Println(err)
			return ExitIOError
		}
	}
	return ExitSuccess
}

func main() {
	rc := run(os.Args[1:], DefaultEnv())
	os.Exit(int(rc))
}
