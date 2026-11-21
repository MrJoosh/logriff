package main

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

func validateFields(l lineFields) (lineData, error) {
	if _, err := netip.ParseAddr(l.addr); err != nil {
		return lineData{}, fmt.Errorf("invalid remote address: %s", l.addr)
	}
	method := l.method
	if _, ok := httpMethods[method]; !ok {
		return lineData{}, fmt.Errorf("invalid method: %s", method)
	}
	path, _, _ := strings.Cut(l.path, "?")
	status, err := strconv.Atoi(l.status)
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert status to int: %s", l.status)
	}
	if status < 100 || status > 599 {
		return lineData{}, fmt.Errorf("invalid http status: %v", status)
	}
	bytesSent, err := strconv.Atoi(l.bytes)
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert bytes sent to int: %s", l.bytes)
	}
	if bytesSent < 0 {
		return lineData{}, fmt.Errorf("negative bytes sent is invalid: %v", bytesSent)
	}
	responseTimeMs, err := strconv.ParseFloat(l.response, 64)
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert response time to float: %s", l.response)
	}
	if responseTimeMs < 0 {
		return lineData{}, fmt.Errorf("negative response time is invalid: %v", responseTimeMs)
	}

	return lineData{
		remoteAddr:     l.addr,
		timestamp:      l.timestamp,
		method:         method,
		path:           path,
		protocol:       l.protocol,
		status:         status,
		bytesSent:      bytesSent,
		responseTimeMs: responseTimeMs,
		userAgent:      l.userAgent,
	}, nil
}
