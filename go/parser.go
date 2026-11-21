package main

import (
	"fmt"
	"regexp"
)

const (
	iRemoteAddr = iota
	iTimestamp
	iMethod
	iPath
	iProtocol
	iStatus
	iBytesSent
	iRespTime
	iUserAgent
	fieldCount
)

var lineRE = regexp.MustCompile(
	// start of line
	`^` +
		// ip
		`(?P<remote_addr>[A-Za-z0-9.:%]+)` +
		// extra
		` - - \[` +
		// timestamp
		`(?P<timestamp>[\w/: +TZ.-]+)` +
		// extra
		`\] "` +
		// method
		`(?P<method>[A-Z]{3,7})` +
		// extra
		` ` +
		// path
		`(?P<path>[\w\p{L}/.?=%&@:#\[\]@!$'\(\)*+,;~\|-]+)` +
		// extra
		` ` +
		// protocol
		`(?P<protocol>[A-Z0-9./]+)` +
		// extra
		`" ` +
		// status
		`(?P<status>\d{3})` +
		// extra
		` ` +
		// bytes
		`(?P<bytes_sent>\d+)` +
		// extra
		` ` +
		// response time
		`(?P<response_time_ms>[\d.e+-]+)` +
		// extra
		` "` +
		// user agent
		`(?P<user_agent>.*)` +
		// extra
		`"` +
		// end of line
		`$`,
)

func parseLine(line string) (lineFields, error) {
	m := lineRE.FindStringSubmatch(line)
	if m == nil || len(m)-1 != fieldCount {
		return lineFields{}, fmt.Errorf("failed to parse input line: %s", line)
	}
	fields := m[1:]
	return lineFields{
		addr:      fields[iRemoteAddr],
		timestamp: fields[iTimestamp],
		method:    fields[iMethod],
		path:      fields[iPath],
		protocol:  fields[iProtocol],
		status:    fields[iStatus],
		bytes:     fields[iBytesSent],
		response:  fields[iRespTime],
		userAgent: fields[iUserAgent],
	}, nil
}
