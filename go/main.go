package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

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

func processLine(re *regexp.Regexp, args args, ctx context.Context, jobs chan string, results chan<- result, wg *sync.WaitGroup, reportFatal func(ExitCode, error)) {
	defer wg.Done()
	local := stats{statusCodes: make(map[int]int), routeDistribution: make(map[string]int)}
	for line := range jobs {
		// Parse, Validate and statify line
		select {
		case <-ctx.Done():
			results <- result{local, ctx.Err()}
			return
		default:
		}

		parsed, err := parseLine(line, re)
		local.lines++
		if err != nil {
			if !args.strict {
				local.error++
				continue
			}
			reportFatal(ExitStrictError, err)
			results <- result{local, ctx.Err()}
			return
		}
		validated, err := validator(parsed)
		if err != nil {
			if !args.strict {
				local.error++
				continue
			}
			reportFatal(ExitStrictError, err)
			results <- result{local, ctx.Err()}
			return
		}
		if !filter(validated, args) {
			continue
		}
		local.matched++
		local.bytesVolume += validated.bytesSent
		local.responseTimes = append(local.responseTimes, validated.responseTimeMs)
		local.statusCodes[validated.status]++
		local.routeDistribution[validated.path]++
	}

	results <- result{local, nil}
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

	re := getParseRegex()
	reader := bufio.NewReader(in)

	numWorkers := runtime.NumCPU()

	jobs := make(chan string, 1024)
	results := make(chan result, numWorkers)
	exitStatus := make(chan exitInfo, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reportFatal := func(rc ExitCode, err error) {
		cancel()
		select {
		case exitStatus <- exitInfo{rc: rc, err: err}:
		default:
		}
	}

	go func() {
		defer close(jobs)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				select {
				case jobs <- strings.TrimRight(line, "\r\n"):
				case <-ctx.Done():
					return
				}
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				reportFatal(ExitIOError, err)
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go processLine(re, args, ctx, jobs, results, &wg, reportFatal)
	}

	wg.Wait()

	select {
	case e := <-exitStatus:
		l.Println(e.err)
		return e.rc
	default:
	}

	merged := stats{
		statusCodes:       make(map[int]int),
		routeDistribution: make(map[string]int),
	}

	for range numWorkers {
		r := <-results

		merged.lines += r.st.lines
		merged.matched += r.st.matched
		merged.error += r.st.error
		merged.bytesVolume += r.st.bytesVolume

		merged.responseTimes = append(merged.responseTimes, r.st.responseTimes...)
		for code, count := range r.st.statusCodes {
			merged.statusCodes[code] += count
		}
		for route, count := range r.st.routeDistribution {
			merged.routeDistribution[route] += count
		}
	}

	if err := writeOutput(rt, args.format, aggregate(merged), args.topRoutes); err != nil {
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
