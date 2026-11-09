package main

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
