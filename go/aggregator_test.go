package main

import (
	"math"
	"reflect"
	"testing"
)

func Test_aggregate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    stats
		want output
	}{
		{
			name: "valid input",
			s: stats{
				lines:         5,
				matched:       5,
				error:         0,
				bytesVolume:   250,
				responseTimes: []float64{1.2, 1.6, 1.4, 1.5, 1.3},
				statusCodes: map[int]int{
					200: 3,
					401: 1,
					500: 1,
				},
				routeDistribution: map[string]int{
					"/api/health":     3,
					"/assets/foo.css": 1,
					"/index.html":     1,
				},
			},
			want: output{
				summary: summary{
					total:   5,
					matched: 5,
					error:   0,
				},
				traffic: 250,
				latency: latency{
					min:  1.2,
					mean: 1.4,
					p95:  1.58,
					max:  1.6,
				},
				statusCodes: []statusCount{
					{status: 200, count: 3},
					{status: 401, count: 1},
					{status: 500, count: 1},
				},
				topRoutes: []routeCount{
					{path: "/api/health", count: 3},
					{path: "/assets/foo.css", count: 1},
					{path: "/index.html", count: 1},
				},
			},
		},
		{
			name: "empty input",
			s:    stats{},
			want: output{
				summary: summary{
					total:   0,
					matched: 0,
					error:   0,
				},
				traffic: 0,
				latency: latency{
					min:  0,
					mean: 0,
					p95:  0,
					max:  0,
				},
				statusCodes: []statusCount{},
				topRoutes:   []routeCount{},
			},
		},
		{
			name: "status sorting",
			s: stats{
				lines:         5,
				matched:       5,
				error:         0,
				bytesVolume:   250,
				responseTimes: []float64{1.2, 1.6, 1.4, 1.5, 1.3},
				statusCodes: map[int]int{
					500: 1,
					401: 1,
					200: 3,
				},
				routeDistribution: map[string]int{
					"/api/health":     3,
					"/assets/foo.css": 1,
					"/index.html":     1,
				},
			},
			want: output{
				summary: summary{
					total:   5,
					matched: 5,
					error:   0,
				},
				traffic: 250,
				latency: latency{
					min:  1.2,
					mean: 1.4,
					p95:  1.58,
					max:  1.6,
				},
				statusCodes: []statusCount{
					{status: 200, count: 3},
					{status: 401, count: 1},
					{status: 500, count: 1},
				},
				topRoutes: []routeCount{
					{path: "/api/health", count: 3},
					{path: "/assets/foo.css", count: 1},
					{path: "/index.html", count: 1},
				},
			},
		},
		{
			name: "routes sorting",
			s: stats{
				lines:         5,
				matched:       5,
				error:         0,
				bytesVolume:   250,
				responseTimes: []float64{1.2, 1.6, 1.4, 1.5, 1.3},
				statusCodes: map[int]int{
					200: 3,
					401: 1,
					500: 1,
				},
				routeDistribution: map[string]int{
					"/index.html":     1,
					"/assets/foo.css": 1,
					"/api/health":     3,
				},
			},
			want: output{
				summary: summary{
					total:   5,
					matched: 5,
					error:   0,
				},
				traffic: 250,
				latency: latency{
					min:  1.2,
					mean: 1.4,
					p95:  1.58,
					max:  1.6,
				},
				statusCodes: []statusCount{
					{status: 200, count: 3},
					{status: 401, count: 1},
					{status: 500, count: 1},
				},
				topRoutes: []routeCount{
					{path: "/api/health", count: 3},
					{path: "/assets/foo.css", count: 1},
					{path: "/index.html", count: 1},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := aggregate(tt.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("aggregate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_p95(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		d    []float64
		want float64
	}{
		{
			name: "nil input",
			d:    nil,
			want: float64(0),
		},
		{
			name: "empty input",
			d:    []float64{},
			want: float64(0),
		},
		{
			name: "single element",
			d:    []float64{42.5},
			want: float64(42.5),
		},
		{
			name: "two elements",
			d:    []float64{10, 20},
			want: float64(19.5),
		},
		{
			name: "exact index",
			d:    genFloats(0, 20),
			want: float64(19),
		},
		{
			name: "interpolated",
			d:    genFloats(1, 20),
			want: float64(19.05),
		},
		{
			name: "dupes at boundary",
			d:    []float64{1.0, 2.0, 3.0, 50.0, 50.0},
			want: float64(50.0),
		},
		{
			name: "extreme values",
			d:    []float64{-math.MaxFloat64, math.MaxFloat64},
			want: float64(1.617923821376084e+308),
		},
		{
			name: "infinity values",
			d:    []float64{1.0, math.Inf(1)},
			want: math.Inf(1),
		},
		{
			name: "NaN values",
			d:    []float64{1.0, math.NaN()},
			want: math.NaN(),
		},
		{
			name: "negative values",
			d:    []float64{-1.0, -1.0},
			want: float64(-1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p95(tt.d)
			if !(math.IsNaN(got) && math.IsNaN(tt.want)) && got != tt.want {
				t.Errorf("p95() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mean(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		d    []float64
		want float64
	}{
		{
			name: "nil slice",
			d:    nil,
			want: float64(0),
		},
		{
			name: "empty input",
			d:    []float64{},
			want: float64(0),
		},
		{
			name: "single element",
			d:    []float64{1},
			want: float64(1),
		},
		{
			name: "positive floats",
			d:    []float64{10.0, 20.0, 30.0},
			want: float64(20.0),
		},
		{
			name: "negative and zero values",
			d:    []float64{-10.0, 0.0, 10.0},
			want: float64(0.0),
		},
		{
			name: "non-terminating decimal",
			d:    []float64{1.0, 2.0, 4.0},
			want: 7.0 / 3.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mean(tt.d)
			if got != tt.want {
				t.Errorf("mean() = %v, want %v", got, tt.want)
			}
		})
	}
}
