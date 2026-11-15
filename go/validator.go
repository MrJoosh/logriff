package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func validator(f []string) (lineData, error) {
	if net.ParseIP(f[0]) == nil {
		return lineData{}, fmt.Errorf("invalid remote address: %s", f[0])
	}
	remoteAddr := f[0]
	timestamp := f[1]
	method := f[2]
	if !httpMethods[method] {
		return lineData{}, fmt.Errorf("invalid method: %s", f[2])
	}
	path, _, _ := strings.Cut(f[3], "?")
	protocol := f[4]
	status, err := strconv.Atoi(f[5])
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert status to int: %s", f[0])
	}
	if status < 100 || status > 599 {
		return lineData{}, fmt.Errorf("invalid http status: %v", status)
	}
	bytesSent, err := strconv.Atoi(f[6])
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert bytes sent to int: %s", f[6])
	}
	if bytesSent < 0 {
		return lineData{}, fmt.Errorf("bytes sent shouldn't be negative: %v", bytesSent)
	}
	responseTimeMs, err := strconv.ParseFloat(f[7], 64)
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert response time to float: %s", f[6])
	}
	userAgent := f[8]

	return lineData{
		remoteAddr:     remoteAddr,
		timestamp:      timestamp,
		method:         method,
		path:           path,
		protocol:       protocol,
		status:         status,
		bytesSent:      bytesSent,
		responseTimeMs: responseTimeMs,
		userAgent:      userAgent,
	}, nil
}
