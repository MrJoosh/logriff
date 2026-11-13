package main

import (
	"fmt"
	"strconv"
)

type lineFields struct {
	addr, timestamp, method, path, protocol, status, bytes, response, userAgent string
}

func genLog(f lineFields) string {
	if f.addr == "" {
		f.addr = "192.168.1.10"
	}
	if f.timestamp == "" {
		f.timestamp = "06/Nov/2026:08:12:01 +0000"
	}
	if f.method == "" {
		f.method = "GET"
	}
	if f.path == "" {
		f.path = "/api/v1/health"
	}
	if f.protocol == "" {
		f.protocol = "HTTP/1.1"
	}
	if f.status == "" {
		f.status = "200"
	}
	if f.bytes == "" {
		f.bytes = "42"
	}
	if f.response == "" {
		f.response = "1.2"
	}
	if f.userAgent == "" {
		f.userAgent = "Go-http-client/1.1"
	}

	return fmt.Sprintf("%s - - [%s] \"%s %s %s\" %s %s %s \"%s\"", f.addr, f.timestamp, f.method, f.path, f.protocol, f.status, f.bytes, f.response, f.userAgent)
}
func genFields(f lineFields) []string {
	if f.addr == "" {
		f.addr = "192.168.1.10"
	}
	if f.timestamp == "" {
		f.timestamp = "06/Nov/2026:08:12:01 +0000"
	}
	if f.method == "" {
		f.method = "GET"
	}
	if f.path == "" {
		f.path = "/api/v1/health"
	}
	if f.protocol == "" {
		f.protocol = "HTTP/1.1"
	}
	if f.status == "" {
		f.status = "200"
	}
	if f.bytes == "" {
		f.bytes = "42"
	}
	if f.response == "" {
		f.response = "1.2"
	}
	if f.userAgent == "" {
		f.userAgent = "Go-http-client/1.1"
	}

	return []string{f.addr, f.timestamp, f.method, f.path, f.protocol, f.status, f.bytes, f.response, f.userAgent}

}

func genLineData(f lineFields) lineData {
	var addr, timestamp, method, path, protocol, userAgent string
	var status, bytes int
	var response float64
	if f.addr == "" {
		f.addr = "192.168.1.10"
	}
	if f.timestamp == "" {
		f.timestamp = "06/Nov/2026:08:12:01 +0000"
	}
	if f.method == "" {
		f.method = "GET"
	}
	if f.path == "" {
		f.path = "/api/v1/health"
	}
	if f.protocol == "" {
		f.protocol = "HTTP/1.1"
	}
	if f.status == "" {
		f.status = "200"
	}
	if f.bytes == "" {
		f.bytes = "42"
	}
	if f.response == "" {
		f.response = "1.2"
	}
	if f.userAgent == "" {
		f.userAgent = "Go-http-client/1.1"
	}
	addr = f.addr
	timestamp = f.timestamp
	method = f.method
	path = f.path
	protocol = f.protocol
	status, _ = strconv.Atoi(f.status)
	bytes, _ = strconv.Atoi(f.bytes)
	response, _ = strconv.ParseFloat(f.response, 64)
	userAgent = f.userAgent

	return lineData{
		remoteAddr:     addr,
		timestamp:      timestamp,
		method:         method,
		path:           path,
		protocol:       protocol,
		status:         status,
		bytesSent:      bytes,
		responseTimeMs: response,
		userAgent:      userAgent,
	}

}

func genFloats(min, max int) []float64 {
	if min > max {
		return []float64{}
	}

	s := make([]float64, (max-min)+1)
	for n := range len(s) {
		s[n] = float64(n + min)
	}
	fmt.Println(s)
	return s
}
