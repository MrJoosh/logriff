package main

import (
	"cmp"
	"math"
	"slices"
	"strings"
)

func p95(d []float64) (p float64) {
	l := len(d)
	if l == 0 {
		return 0
	}

	if l == 1 {
		return d[0]
	}

	s := slices.Clone(d)
	slices.Sort(s)

	if math.IsNaN(s[0]) {
		return math.NaN()
	}

	r := 0.95 * float64(l-1)
	k := int(r)
	f := r - float64(k)

	if f == 0 || s[k] == s[k+1] {
		return s[k]
	}

	delta := s[k+1] - s[k]
	if math.IsInf(delta, 1) {
		return (1-f)*s[k] + f*s[k+1]
	} else {
		return s[k] + f*delta
	}
}

func mean(d []float64) float64 {
	if len(d) == 0 {
		return 0
	}
	var s float64
	for _, v := range d {
		s += v
	}
	return s / float64(len(d))
}

func sliceMin(d []float64) float64 {
	if len(d) == 0 {
		return 0
	}
	return slices.Min(d)
}

func sliceMax(d []float64) float64 {
	if len(d) == 0 {
		return 0
	}
	return slices.Max(d)
}

func aggregate(s stats) output {
	// Sort response times
	rt := slices.Clone(s.responseTimes)
	slices.Sort(rt)

	// Sort and format status codes
	keys := make([]int, 0, len(s.statusCodes))
	for k := range s.statusCodes {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	statusCounts := make([]statusCount, 0, len(keys))
	for _, k := range keys {
		statusCounts = append(statusCounts, statusCount{status: k, count: s.statusCodes[k]})
	}

	// Sort and format routes
	routes := make([]routeCount, 0, len(s.routeDistribution))
	for path, count := range s.routeDistribution {
		routes = append(routes, routeCount{path: path, count: count})
	}
	slices.SortFunc(routes, func(a, b routeCount) int {
		if i := cmp.Compare(a.count, b.count); i != 0 {
			return i
		}
		return strings.Compare(b.path, a.path)
	})
	slices.Reverse(routes)

	return output{
		summary: summary{
			total:   s.lines,
			matched: s.matched,
			error:   s.error,
		},
		traffic: s.bytesVolume,
		latency: latency{
			min:  sliceMin(rt),
			mean: mean(rt),
			p95:  p95(rt),
			max:  sliceMax(rt),
		},
		statusCodes: statusCounts,
		topRoutes:   routes,
	}
}

func mergeStats(m *stats, r stats) {
	m.lines += r.lines
	m.matched += r.matched
	m.error += r.error
	m.bytesVolume += r.bytesVolume

	m.responseTimes = append(m.responseTimes, r.responseTimes...)
	for code, count := range r.statusCodes {
		m.statusCodes[code] += count
	}
	for route, count := range r.routeDistribution {
		m.routeDistribution[route] += count
	}
}
