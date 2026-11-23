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

	if math.IsNaN(d[0]) {
		return math.NaN()
	}

	r := 0.95 * float64(l-1)
	k := int(r)
	f := r - float64(k)

	if f == 0 || d[k] == d[k+1] {
		return d[k]
	}

	delta := d[k+1] - d[k]
	if math.IsInf(delta, 1) {
		return (1-f)*d[k] + f*d[k+1]
	}
	return d[k] + f*delta
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

func aggregate(s stats) output {
	// Sort response times
	rt := slices.Clone(s.responseTimes)
	slices.Sort(rt)

	lat := latency{}
	if len(rt) > 0 {
		lat.Min = latValue(rt[0])
		lat.Mean = latValue(mean(rt))
		lat.P95 = latValue(p95(rt))
		lat.Max = latValue(rt[len(rt)-1])
	}

	// Sort and format status codes
	keys := make([]int, 0, len(s.statusCodes))
	for k := range s.statusCodes {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	statusCounts := make([]statusCount, 0, len(keys))
	for _, k := range keys {
		statusCounts = append(statusCounts, statusCount{Status: k, Count: s.statusCodes[k]})
	}

	// Sort and format routes
	routes := make([]routeCount, 0, len(s.routeDistribution))
	for path, count := range s.routeDistribution {
		routes = append(routes, routeCount{Path: path, Count: count})
	}
	slices.SortFunc(routes, func(a, b routeCount) int {
		if i := cmp.Compare(b.Count, a.Count); i != 0 {
			return i
		}
		return strings.Compare(a.Path, b.Path)
	})

	return output{
		Summary: summary{
			Total:   s.lines,
			Matched: s.matched,
			Error:   s.error,
		},
		Traffic:     traffic{TotalBytes: s.bytesVolume},
		Latency:     lat,
		StatusCodes: statusCounts,
		TopRoutes:   routes,
	}
}

func (s *stats) merge(other stats) {
	s.lines += other.lines
	s.matched += other.matched
	s.error += other.error
	s.bytesVolume += other.bytesVolume

	s.responseTimes = append(s.responseTimes, other.responseTimes...)
	for code, count := range other.statusCodes {
		s.statusCodes[code] += count
	}
	for route, count := range other.routeDistribution {
		s.routeDistribution[route] += count
	}
}
