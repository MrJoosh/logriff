package main

import (
	"strings"
)

func match(rec lineData, opts args) bool {
	if opts.statusMin != 0 && rec.status < opts.statusMin {
		return false
	}
	if opts.statusMax != 0 && rec.status > opts.statusMax {
		return false
	}
	if opts.method != "" && rec.method != opts.method {
		return false
	}
	if opts.path != "" && !strings.HasPrefix(rec.path, opts.path) {
		return false
	}
	if opts.minDuration != 0 && rec.responseTimeMs < opts.minDuration {
		return false
	}
	return true
}
