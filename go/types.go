package main

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
