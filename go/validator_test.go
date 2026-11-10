package main

import "testing"

func genFields(f lineFields) []string {
	if f.addr == "" {
		f.addr = "192.168.1.10"
	}
	if f.timestamp == "" {
		f.timestamp = "06/Nov/2026:08:12:01 +0000"
	}
	if f.method == "" {
		f.method = "GET"
	}
	if f.path == "" {
		f.path = "/api/v1/health"
	}
	if f.protocol == "" {
		f.protocol = "HTTP/1.1"
	}
	if f.status == "" {
		f.status = "200"
	}
	if f.bytes == "" {
		f.bytes = "42"
	}
	if f.response == "" {
		f.response = "1.2"
	}
	if f.userAgent == "" {
		f.userAgent = "Go-http-client/1.1"
	}

	return []string{f.addr, f.timestamp, f.method, f.path, f.protocol, f.status, f.bytes, f.response, f.userAgent}

}

func Test_validator(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		f       []string
		want    lineData
		wantErr bool
	}{
		{
			name:    "invalid status not int",
			f:       genFields(lineFields{status: "20x"}),
			want:    lineData{},
			wantErr: true,
		},
		{
			name:    "invalid status too small",
			f:       genFields(lineFields{status: "99"}),
			want:    lineData{},
			wantErr: true,
		},
		{
			name:    "invalid status too big",
			f:       genFields(lineFields{status: "999"}),
			want:    lineData{},
			wantErr: true,
		},
		{
			name:    "negative bytes",
			f:       genFields(lineFields{bytes: "-1"}),
			want:    lineData{},
			wantErr: true,
		},
		{
			name:    "response time not float word",
			f:       genFields(lineFields{response: "slow"}),
			want:    lineData{},
			wantErr: true,
		},
		{
			name:    "response time not float semvar",
			f:       genFields(lineFields{response: "1.2.3"}),
			want:    lineData{},
			wantErr: true,
		},
		{
			name:    "invalid method",
			f:       genFields(lineFields{response: "FLY"}),
			want:    lineData{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := validator(tt.f)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validator() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validator() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("validator() = %v, want %v", got, tt.want)
			}
		})
	}
}
