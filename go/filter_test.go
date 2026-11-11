package main

import "testing"

func Test_filter(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		d    lineData
		f    args
		want bool
	}{
		{
			name: "status min",
			d:    genLineData(lineFields{}),
			f:    args{statusMin: 200},
			want: true,
		},
		{
			name: "status min false",
			d:    genLineData(lineFields{}),
			f:    args{statusMin: 201},
			want: false,
		},
		{
			name: "status max true",
			d:    genLineData(lineFields{}),
			f:    args{statusMax: 201},
			want: true,
		},
		{
			name: "status max false",
			d:    genLineData(lineFields{}),
			f:    args{statusMax: 199},
			want: false,
		},
		{
			name: "method true",
			d:    genLineData(lineFields{}),
			f:    args{method: "GET"},
			want: true,
		},
		{
			name: "method false",
			d:    genLineData(lineFields{}),
			f:    args{method: "POST"},
			want: false,
		},
		{
			name: "path prefix true",
			d:    genLineData(lineFields{}),
			f:    args{path: "/api"},
			want: true,
		},
		{
			name: "path full true",
			d:    genLineData(lineFields{}),
			f:    args{path: "/api/v1/health"},
			want: true,
		},
		{
			name: "path prefix false",
			d:    genLineData(lineFields{}),
			f:    args{path: "/assets"},
			want: false,
		},
		{
			name: "duration true",
			d:    genLineData(lineFields{}),
			f:    args{minDuration: 1.2},
			want: true,
		},
		{
			name: "duration false",
			d:    genLineData(lineFields{}),
			f:    args{minDuration: 1.3},
			want: false,
		},
		{
			name: "multi filter true",
			d:    genLineData(lineFields{}),
			f: args{
				statusMin:   200,
				statusMax:   200,
				method:      "GET",
				path:        "/api",
				minDuration: 1.2,
			},
			want: true,
		},
		{
			name: "multi filter false",
			d:    genLineData(lineFields{}),
			f: args{
				statusMin:   201,
				statusMax:   200,
				method:      "POST",
				path:        "/assets",
				minDuration: 2.0,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter(tt.d, tt.f)
			if got != tt.want {
				t.Errorf("filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
