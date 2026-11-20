package main

import (
	"io"
	"os"
)

type ExitCode int

const (
	ExitSuccess ExitCode = iota
	ExitValidationError
	ExitIOError
	ExitStrictError
)

var httpMethods = map[string]bool{
	"CONNECT": true,
	"DELETE":  true,
	"GET":     true,
	"HEAD":    true,
	"OPTIONS": true,
	"PATCH":   true,
	"POST":    true,
	"PUT":     true,
	"TRACE":   true,
}

type args struct {
	statusMin   int
	statusMax   int
	method      string
	path        string
	minDuration float64
	topRoutes   int
	format      string
	strict      bool
	file        string
}

type lineData struct {
	remoteAddr     string
	timestamp      string
	method         string
	path           string
	protocol       string
	status         int
	bytesSent      int
	responseTimeMs float64
	userAgent      string
}

type stats struct {
	lines             int
	matched           int
	error             int
	bytesVolume       int
	responseTimes     []float64
	statusCodes       map[int]int
	routeDistribution map[string]int
}

type output struct {
	summary     summary
	traffic     int
	latency     latency
	statusCodes []statusCount
	topRoutes   []routeCount
}

type summary struct {
	total   int
	matched int
	error   int
}

type latency struct {
	min  float64
	mean float64
	p95  float64
	max  float64
}

type routeCount struct {
	path  string
	count int
}

type statusCount struct {
	status int
	count  int
}

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

type result struct {
	st  stats
	err error
}

type exitInfo struct {
	rc  ExitCode
	err error
}
