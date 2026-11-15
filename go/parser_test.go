package main

import (
	"reflect"
	"testing"
)

func Test_parseLine(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		line    string
		want    []string
		wantErr bool
	}{
		{
			name:    "extracts all fields",
			line:    genLog(lineFields{}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "preserves spaces in user agent",
			line:    genLog(lineFields{userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/1.1", "200", "42", "1.2", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"},
			wantErr: false,
		},
		{
			name:    "accepts zero bytes",
			line:    genLog(lineFields{bytes: "0"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/1.1", "200", "0", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "accepts large byte counts",
			line:    genLog(lineFields{bytes: "10000000000"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/1.1", "200", "10000000000", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "accepts http 3",
			line:    genLog(lineFields{protocol: "HTTP/3"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/3", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "reject obvious malform blank",
			line:    "",
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "reject obvious malform text",
			line:    "not an access log line",
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "reject obvious malform truncated",
			line:    "10.0.0.1 - - [06/Nov/2026:08:12:01 +0000]",
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "reject obvious malform truncated user agent",
			line:    "10.0.0.1 - - [06/Nov/2026:08:12:01 +0000] \"GET / HTTP/1.1\" 200 42 1.2",
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "ipv6 is accepted",
			line:    genLog(lineFields{addr: "2001:db8::1"}),
			want:    []string{"2001:db8::1", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "query string in path",
			line:    genLog(lineFields{path: "/api/v1/items?page=2"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/items?page=2", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "invalid ipv4 all octets",
			line:    genLog(lineFields{addr: "999.999.999.999"}),
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid ipv4 one octet",
			line:    genLog(lineFields{addr: "256.1.1.1"}),
			want:    []string{},
			wantErr: true,
		},
		{
			name:    "common urls chars accepted dash",
			line:    genLog(lineFields{path: "/files/hello-world"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/files/hello-world", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "common urls chars accepted url encode",
			line:    genLog(lineFields{path: "/files/a%20b.txt"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/files/a%20b.txt", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "common url chars accepted query string",
			line:    genLog(lineFields{path: "/api/v1/items?page=2&foo=bar&user=@me"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/items?page=2&foo=bar&user=@me", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "ISO-like timestamp support",
			line:    genLog(lineFields{timestamp: "2026-11-15T08:08:19.12345Z"}),
			want:    []string{"192.168.1.10", "2026-11-15T08:08:19.12345Z", "GET", "/api/v1/health", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "IPv6 IPv4 mapped",
			line:    genLog(lineFields{addr: "::ffff:192.168.1.1"}),
			want:    []string{"::ffff:192.168.1.1", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "IPv6 zone scoped",
			line:    genLog(lineFields{addr: "fe80::1%eth0"}),
			want:    []string{"fe80::1%eth0", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/1.1", "200", "42", "1.2", "Go-http-client/1.1"},
			wantErr: false,
		},
		{
			name:    "empty quoted user-agent",
			line:    genLog(lineFields{userAgent: "<blank>"}),
			want:    []string{"192.168.1.10", "06/Nov/2026:08:12:01 +0000", "GET", "/api/v1/health", "HTTP/1.1", "200", "42", "1.2", ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseLine(tt.line)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseLine() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseLine() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
