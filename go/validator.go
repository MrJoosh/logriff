package main

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

func validator(f []string) (lineData, error) {
	if len(f) != fieldCount {
		return lineData{}, fmt.Errorf("invalid field count: %d", len(f))
	}
	if _, err := netip.ParseAddr(f[iRemoteAddr]); err != nil {
		return lineData{}, fmt.Errorf("invalid remote address: %s", f[iRemoteAddr])
	}
	remoteAddr := f[iRemoteAddr]
	timestamp := f[iTimestamp]
	method := f[iMethod]
	if !httpMethods[method] {
		return lineData{}, fmt.Errorf("invalid method: %s", method)
	}
	path, _, _ := strings.Cut(f[iPath], "?")
	protocol := f[iProtocol]
	status, err := strconv.Atoi(f[iStatus])
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert status to int: %s", f[iStatus])
	}
	if status < 100 || status > 599 {
		return lineData{}, fmt.Errorf("invalid http status: %v", status)
	}
	bytesSent, err := strconv.Atoi(f[iBytesSent])
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert bytes sent to int: %s", f[iBytesSent])
	}
	if bytesSent < 0 {
		return lineData{}, fmt.Errorf("bytes sent shouldn't be negative: %v", bytesSent)
	}
	responseTimeMs, err := strconv.ParseFloat(f[iRespTime], 64)
	if err != nil {
		return lineData{}, fmt.Errorf("could not convert response time to float: %s", f[iRespTime])
	}
	if responseTimeMs < 0 {
		return lineData{}, fmt.Errorf("response time should not be negative: %v", responseTimeMs)
	}
	userAgent := f[iUserAgent]

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
