package main

import (
	"fmt"
	"regexp"
	"strings"
)

var lineRegexStr []string = []string{
	// start of line
	`^`,
	// ip
	`(?P<remote_addr>(?:(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.){3}`,
	`(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])|`,
	`(?:(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|`,
	`(?:[0-9a-fA-F]{1,4}:){1,7}:|`,
	`(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|`,
	`(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|`,
	`(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|`,
	`(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|`,
	`(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|`,
	`[0-9a-fA-F]{1,4}:(?:(?::[0-9a-fA-F]{1,4}){1,6})|`,
	`:(?:(?::[0-9a-fA-F]{1,4}){1,7}|:)))`,
	// extra
	` - - \[`,
	// timestamp
	`(?P<timestamp>[\w/: +]+)`,
	// extra
	`\] "`,
	// method
	`(?P<method>[A-Z]{3,7})`,
	// extra
	` `,
	// path
	`(?P<path>[\w/.?=%-]+)`,
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
	`(?P<response_time_ms>[\d.]+)`,
	// extra
	` "`,
	// user agent
	`(?P<user_agent>.+)`,
	// extra
	`"`,
	// end of line
	`$`,
}

func parseLine(line string) ([]string, error) {
	// Build the regex
	var b strings.Builder
	for _, s := range lineRegexStr {
		b.WriteString(s)
	}
	re := regexp.MustCompile(b.String())
	// Parse the line using the regex
	matches := re.FindAllStringSubmatch(line, -1)
	if len(matches) != 1 {
		return []string{}, fmt.Errorf("failed to parse input line: %s", line)
	}
	if len(matches[0]) < 9 {
		return []string{}, fmt.Errorf("failed to parse input line: %s", line)
	}
	return matches[0][1:], nil
}
