package main

import (
	"fmt"
	"regexp"
	"strings"
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

var lineRegexStr []string = []string{
	// start of line
	`^`,
	// ip
	`(?P<remote_addr>[A-Za-z0-9.:%]+)`,
	// extra
	` - - \[`,
	// timestamp
	`(?P<timestamp>[\w/: +TZ.-]+)`,
	// extra
	`\] "`,
	// method
	`(?P<method>[A-Z]{3,7})`,
	// extra
	` `,
	// path
	`(?P<path>[\w\p{L}/.?=%&@:#\[\]@!$'\(\)*+,;~\|-]+)`,
	// extra
	` `,
	// protocol
	`(?P<protocol>[A-Z0-9./]+)`,
	// extra
	`" `,
	// status
	`(?P<status>\d{3})`,
	// extra
	` `,
	// bytes
	`(?P<bytes_sent>\d+)`,
	// extra
	` `,
	// response time
	`(?P<response_time_ms>[\d.e+-]+)`,
	// extra
	` "`,
	// user agent
	`(?P<user_agent>.*)`,
	// extra
	`"`,
	// end of line
	`$`,
}

func parseLine(line string, re *regexp.Regexp) ([]string, error) {
	m := re.FindStringSubmatch(line)
	if m == nil || len(m)-1 != fieldCount {
		return nil, fmt.Errorf("failed to parse input line: %s", line)
	}
	return m[1:], nil
}

func getParseRegex() *regexp.Regexp {
	var b strings.Builder
	for _, s := range lineRegexStr {
		b.WriteString(s)
	}
	return regexp.MustCompile(b.String())
}
