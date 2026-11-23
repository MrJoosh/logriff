package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strconv"
)

type ExitCode int

const (
	ExitSuccess ExitCode = iota
	ExitValidationError
	ExitIOError
	ExitStrictError
)

var httpMethods = map[string]struct{}{
	"CONNECT": {},
	"DELETE":  {},
	"GET":     {},
	"HEAD":    {},
	"OPTIONS": {},
	"PATCH":   {},
	"POST":    {},
	"PUT":     {},
	"TRACE":   {},
}

type args struct {
	statusMin   *int
	statusMax   *int
	method      string
	path        string
	minDuration *float64
	topRoutes   *int
	format      string
	strict      bool
	file        string
}

type lineFields struct {
	addr, timestamp, method, path, protocol, status, bytes, response, userAgent string
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
	Summary     summary      `json:"summary"`
	Traffic     traffic      `json:"traffic"`
	Latency     latency      `json:"latency_ms"`
	StatusCodes statusStats  `json:"status_codes"`
	TopRoutes   []routeCount `json:"top_routes"`
}

type traffic struct {
	TotalBytes int `json:"total_bytes"`
}

type summary struct {
	Total   int `json:"total_lines"`
	Matched int `json:"matched_lines"`
	Error   int `json:"error_lines"`
}

type latValue float64

func (l latValue) MarshalJSON() ([]byte, error) {
	s := strconv.FormatFloat(float64(l), 'f', 2, 64)
	return []byte(s), nil
}

type latency struct {
	Min  latValue `json:"min"`
	Mean latValue `json:"mean"`
	P95  latValue `json:"p95"`
	Max  latValue `json:"max"`
}

type routeCount struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

type statusStats []statusCount

func (s statusStats) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, status := range s {
		if i > 0 {
			buf.WriteByte(',')
		}

		key, err := json.Marshal(strconv.Itoa(status.Status))
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')

		val, err := json.Marshal(status.Count)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

type statusCount struct {
	Status int
	Count  int
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

type exitInfo struct {
	rc  ExitCode
	err error
}
