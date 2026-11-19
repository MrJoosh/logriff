package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type FileOpener interface {
	Open(name string) (io.ReadCloser, error)
}

type StdinProvider interface {
	Stdin() io.Reader
}

type StdoutProvider interface {
	Stdout() io.Writer
}

type StderrProvider interface {
	Stderr() io.Writer
}

type Runtime interface {
	StdinProvider
	StdoutProvider
	StderrProvider
	FileOpener
}

type osRuntime struct{}

func (osRuntime) Stdin() io.Reader  { return os.Stdin }
func (osRuntime) Stdout() io.Writer { return os.Stdout }
func (osRuntime) Stderr() io.Writer { return os.Stderr }
func (osRuntime) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func writeOutput(out StdoutProvider, format string, output output, numRoutes int) error {
	switch format {
	case "json":
		_, err := fmt.Fprintln(out.Stdout(), output.toJson(numRoutes))
		if err != nil {
			return err
		}
	case "text":
		_, err := fmt.Fprintln(out.Stdout(), output.toText(numRoutes))
		if err != nil {
			return err
		}
	}
	return nil
}

func run(argv []string, rt Runtime) ExitCode {
	l := log.New(rt.Stderr(), "", 0)
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
		in = rt.Stdin()
	} else {
		f, err := rt.Open(args.file)
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

	if err := writeOutput(rt, args.format, aggregate(stats), args.topRoutes); err != nil {
		l.Println(err)
		return ExitIOError
	}

	return ExitSuccess
}

func main() {
	rc := run(os.Args[1:], osRuntime{})
	os.Exit(int(rc))
}
