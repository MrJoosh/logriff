package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
)

func run(argv []string, rt Runtime) ExitCode {
	args, err := parseArgs(argv)
	switch {
	case errors.Is(err, errHelp):
		usage(rt.Stdout())
		return ExitSuccess
	case err != nil:
		fmt.Fprintln(rt.Stderr(), err)
		usage(rt.Stderr())
		return ExitValidationError
	}

	var in io.Reader
	if args.file == "-" {
		in = rt.Stdin()
	} else {
		f, err := rt.Open(args.file)
		if err != nil {
			fmt.Fprintln(rt.Stderr(), err)
			return ExitIOError
		}
		defer f.Close()
		in = f
	}

	inReader := bufio.NewReader(in)

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

	go reader(inReader, jobs, ctx, reportFatal)

	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go processLine(args, ctx, jobs, results, &wg, reportFatal, rt)
	}

	wg.Wait()

	select {
	case e := <-exitStatus:
		fmt.Fprintln(rt.Stderr(), e.err)
		return e.rc
	default:
	}

	merged := stats{
		statusCodes:       make(map[int]int),
		routeDistribution: make(map[string]int),
	}

	for range numWorkers {
		merged.merge(<-results)
	}

	if err := writeOutput(rt, args.format, aggregate(merged), args.topRoutes); err != nil {
		fmt.Fprintln(rt.Stderr(), err)
		return ExitIOError
	}

	return ExitSuccess
}

func processLine(
	args args,
	ctx context.Context,
	jobs <-chan string,
	results chan<- stats,
	wg *sync.WaitGroup,
	reportFatal func(ExitCode, error),
	rt Runtime,
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

		parsed, err := parseLine(line)
		local.lines++
		if err != nil {
			if !args.strict {
				local.error++
				fmt.Fprintln(rt.Stderr(), err)
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
				fmt.Fprintln(rt.Stderr(), err)
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

func reader(r *bufio.Reader, jobs chan<- string, ctx context.Context, reportFatal func(ExitCode, error)) {
	defer close(jobs)
	for {
		line, err := r.ReadString('\n')
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

}
