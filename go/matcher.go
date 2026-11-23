package main

import (
	"strings"
)

func match(rec lineData, opts args) bool {
	if opts.statusMin != nil && rec.status < *opts.statusMin {
		return false
	}
	if opts.statusMax != nil && rec.status > *opts.statusMax {
		return false
	}
	if opts.method != "" && rec.method != opts.method {
		return false
	}
	if opts.path != "" && !strings.HasPrefix(rec.path, opts.path) {
		return false
	}
	if opts.minDuration != nil && rec.responseTimeMs < *opts.minDuration {
		return false
	}
	return true
}
