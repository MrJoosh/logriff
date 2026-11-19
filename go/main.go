package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
		usage()
		return ExitValidationError
	} else if help == true {
		usage()
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
	reader := bufio.NewReader(in)

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\r\n")
			parsed, err := parseLine(line, re)
			stats.lines++
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
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			l.Println(err)
			return ExitIOError
		}
	}

	if err := writeOutput(rt, args.format, aggregate(stats), args.topRoutes); err != nil {
		l.Println(err)
		return ExitIOError
	}

	return ExitSuccess
}

func main() {
	signal.Ignore(syscall.SIGPIPE)
	rc := run(os.Args[1:], osRuntime{})
	os.Exit(int(rc))
}
