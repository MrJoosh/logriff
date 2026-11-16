package main

import "testing"

func Test_output_toJson(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		s stats
		// Named input parameters for target function.
		numRoutes int
		want      string
	}{
		{
			name: "json format sample",
			s: stats{
				lines:         5,
				matched:       5,
				error:         0,
				bytesVolume:   9330,
				responseTimes: []float64{1.2, 85.4, 14.8, 0.8, 210.6},
				statusCodes: map[int]int{
					200: 2,
					401: 1,
					304: 1,
					500: 1,
				},
				routeDistribution: map[string]int{
					"/api/v1/health":           1,
					"/api/v1/auth/login":       1,
					"/api/v1/items/4912":       1,
					"/assets/main.css":         1,
					"/api/v1/sessions/current": 1,
				},
			},
			numRoutes: 5,
			want:      `{"summary": {"total_lines": 5,"matched_lines": 5,"error_lines": 0},"traffic": {"total_bytes": 9330},"latency_ms": {"min": 0.8,"mean": 62.56,"p95": 185.55999999999997,"max": 210.6},"status_codes": {"200": 2,"304": 1,"401": 1,"500": 1},"top_routes": [{"path": "/api/v1/auth/login", "count": 1},{"path": "/api/v1/health", "count": 1},{"path": "/api/v1/items/4912", "count": 1},{"path": "/api/v1/sessions/current", "count": 1},{"path": "/assets/main.css", "count": 1}]}`,
		},
		{
			name:      "empty stats",
			s:         stats{},
			numRoutes: 5,
			want:      `{"summary": {"total_lines": 0,"matched_lines": 0,"error_lines": 0},"traffic": {"total_bytes": 0},"latency_ms": {"min": 0,"mean": 0,"p95": 0,"max": 0},"status_codes": {},"top_routes": []}`,
		},
		{
			name: "num routes limit",
			s: stats{
				lines:         5,
				matched:       5,
				error:         0,
				bytesVolume:   9330,
				responseTimes: []float64{1.2, 85.4, 14.8, 0.8, 210.6},
				statusCodes: map[int]int{
					200: 2,
					401: 1,
					304: 1,
					500: 1,
				},
				routeDistribution: map[string]int{
					"/api/v1/health":           1,
					"/api/v1/auth/login":       1,
					"/api/v1/items/4912":       1,
					"/assets/main.css":         1,
					"/api/v1/sessions/current": 1,
				},
			},
			numRoutes: 3,
			want:      `{"summary": {"total_lines": 5,"matched_lines": 5,"error_lines": 0},"traffic": {"total_bytes": 9330},"latency_ms": {"min": 0.8,"mean": 62.56,"p95": 185.55999999999997,"max": 210.6},"status_codes": {"200": 2,"304": 1,"401": 1,"500": 1},"top_routes": [{"path": "/api/v1/auth/login", "count": 1},{"path": "/api/v1/health", "count": 1},{"path": "/api/v1/items/4912", "count": 1}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := aggregate(tt.s)
			got := o.toJson(tt.numRoutes)
			if got != tt.want {
				t.Errorf("toJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_output_toText(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		s stats
		// Named input parameters for target function.
		numRoutes int
		want      string
	}{
		{
			name: "text format sample",
			s: stats{
				lines:         5,
				matched:       5,
				error:         0,
				bytesVolume:   9330,
				responseTimes: []float64{1.2, 85.4, 14.8, 0.8, 210.6},
				statusCodes: map[int]int{
					200: 2,
					401: 1,
					304: 1,
					500: 1,
				},
				routeDistribution: map[string]int{
					"/api/v1/health":           1,
					"/api/v1/auth/login":       1,
					"/api/v1/items/4912":       1,
					"/assets/main.css":         1,
					"/api/v1/sessions/current": 1,
				},
			},
			numRoutes: 5,
			want: `=== LogRiff Summary ===
Lines Processed : 5
Matched Records : 5
Skipped (Errors): 0

--- Traffic & Bandwidth ---
Total Volume    : 0.00 GB (9,330 bytes)

--- Latency (ms) ---
Min             : 0.80 ms
Mean            : 62.56 ms
p95             : 185.56 ms
Max             : 210.60 ms

--- Status Codes ---
200 :       2 (40.00%)
304 :       1 (20.00%)
401 :       1 (20.00%)
500 :       1 (20.00%)

--- Top Routes (Top 5) ---
1. /api/v1/auth/login                : 1
2. /api/v1/health                    : 1
3. /api/v1/items/4912                : 1
4. /api/v1/sessions/current          : 1
5. /assets/main.css                  : 1
========================`,
		},
		{
			name:      "empty stats",
			s:         stats{},
			numRoutes: 5,
			want: `=== LogRiff Summary ===
Lines Processed : 0
Matched Records : 0
Skipped (Errors): 0

--- Traffic & Bandwidth ---
Total Volume    : 0.00 GB (0 bytes)

--- Latency (ms) ---
Min             : 0.00 ms
Mean            : 0.00 ms
p95             : 0.00 ms
Max             : 0.00 ms

--- Status Codes ---


--- Top Routes (Top 5) ---

========================`,
		},
		{
			name: "num routes limit",
			s: stats{
				lines:         5,
				matched:       5,
				error:         0,
				bytesVolume:   9330,
				responseTimes: []float64{1.2, 85.4, 14.8, 0.8, 210.6},
				statusCodes: map[int]int{
					200: 2,
					401: 1,
					304: 1,
					500: 1,
				},
				routeDistribution: map[string]int{
					"/api/v1/health":           1,
					"/api/v1/auth/login":       1,
					"/api/v1/items/4912":       1,
					"/assets/main.css":         1,
					"/api/v1/sessions/current": 1,
				},
			},
			numRoutes: 3,
			want: `=== LogRiff Summary ===
Lines Processed : 5
Matched Records : 5
Skipped (Errors): 0

--- Traffic & Bandwidth ---
Total Volume    : 0.00 GB (9,330 bytes)

--- Latency (ms) ---
Min             : 0.80 ms
Mean            : 62.56 ms
p95             : 185.56 ms
Max             : 210.60 ms

--- Status Codes ---
200 :       2 (40.00%)
304 :       1 (20.00%)
401 :       1 (20.00%)
500 :       1 (20.00%)

--- Top Routes (Top 3) ---
1. /api/v1/auth/login                : 1
2. /api/v1/health                    : 1
3. /api/v1/items/4912                : 1
========================`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := aggregate(tt.s)
			got := o.toText(tt.numRoutes)
			if got != tt.want {
				t.Errorf("toText() = %v, want %v", got, tt.want)
			}
		})
	}
}
