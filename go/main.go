package main

import (
	"bufio"
	"context"
	"errors"
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

func processLine(
	re *regexp.Regexp,
	args args,
	ctx context.Context,
	jobs <-chan string,
	results chan<- stats,
	wg *sync.WaitGroup,
	reportFatal func(ExitCode, error),
	l *log.Logger,
) {
	defer wg.Done()
	local := stats{statusCodes: make(map[int]int), routeDistribution: make(map[string]int)}
	for line := range jobs {
		// Parse, Validate and statify line
		select {
		case <-ctx.Done():
			results <- local
			return
		default:
		}

		parsed, err := parseLine(line, re)
		local.lines++
		if err != nil {
			if !args.strict {
				local.error++
				l.Println(err)
				continue
			}
			reportFatal(ExitStrictError, err)
			results <- local
			return
		}
		validated, err := validator(parsed)
		if err != nil {
			if !args.strict {
				local.error++
				l.Println(err)
				continue
			}
			reportFatal(ExitStrictError, err)
			results <- local
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

	results <- local
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
	results := make(chan stats, numWorkers)
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
		go processLine(re, args, ctx, jobs, results, &wg, reportFatal, l)
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

		merged.lines += r.lines
		merged.matched += r.matched
		merged.error += r.error
		merged.bytesVolume += r.bytesVolume

		merged.responseTimes = append(merged.responseTimes, r.responseTimes...)
		for code, count := range r.statusCodes {
			merged.statusCodes[code] += count
		}
		for route, count := range r.routeDistribution {
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
