package main

import (
	"strings"
)

func filter(d lineData, f args) bool {
	if f.statusMin != 0 && d.status < f.statusMin {
		return false
	}
	if f.statusMax != 0 && d.status > f.statusMax {
		return false
	}
	if f.method != "" && d.method != f.method {
		return false
	}
	if f.path != "" && !strings.HasPrefix(d.path, f.path) {
		return false
	}
	if f.minDuration != 0 && d.responseTimeMs < f.minDuration {
		return false
	}
	return true
}
