package main

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"testing/iotest"
)

type brokenWriter struct {
	err error
}

func (b brokenWriter) Write(p []byte) (int, error) {
	return 0, b.err
}

func Test_run(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		env  Env
		argv []string
		want ExitCode
	}{
		{
			name: "invalid cli args",
			env: Env{
				Stdin:  &stdin,
				Stdout: &stdout,
				Stderr: &stderr,
			},
			argv: []string{"-s", "bar"},
			want: ExitValidationError,
		},
		{
			name: "file not found",
			env: Env{
				Stdin:  &stdin,
				Stdout: &stdout,
				Stderr: &stderr,
				Open: func(name string) (io.ReadCloser, error) {
					return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
				},
			},
			argv: []string{"bar"},
			want: ExitIOError,
		},
		{
			name: "file access denied",
			env: Env{
				Stdin:  &stdin,
				Stdout: &stdout,
				Stderr: &stderr,
				Open: func(name string) (io.ReadCloser, error) {
					return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
				},
			},
			argv: []string{"foo"},
			want: ExitIOError,
		},
		{
			name: "strict mode",
			env: Env{
				Stdin:  strings.NewReader("Malformed Line!\nMalformed Line!\n\n"),
				Stdout: &stdout,
				Stderr: &stderr,
			},
			argv: []string{"--strict"},
			want: ExitStrictError,
		},
		{
			name: "non-strict mode",
			env: Env{
				Stdin:  strings.NewReader("Malformed Line!\nMalformed Line!\n\n"),
				Stdout: &stdout,
				Stderr: &stderr,
			},
			argv: []string{},
			want: ExitSuccess,
		},
		{
			name: "file open and read",
			env: Env{
				Stdin:  &stdin,
				Stdout: &stdout,
				Stderr: &stderr,
				Open: func(name string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(`192.168.1.10 - - [06/Nov/2026:08:12:01 +0000] "GET /api/v1/health HTTP/1.1" 200 42 1.2 "Go-http-client/1.1"`)), nil
				},
			},
			argv: []string{},
			want: ExitSuccess,
		},
		{
			name: "scanner read error",
			env: Env{
				Stdin:  iotest.ErrReader(errors.New("stdin read error")),
				Stdout: &stdout,
				Stderr: &stderr,
			},
			argv: []string{},
			want: ExitIOError,
		},
		{
			name: "stdout broken pipe",
			env: Env{
				Stdin:  strings.NewReader(`192.168.1.10 - - [06/Nov/2026:08:12:01 +0000] "GET /api/v1/health HTTP/1.1" 200 42 1.2 "Go-http-client/1.1"`),
				Stdout: brokenWriter{err: errors.New("write /dev/stdout: broken pipe")},
				Stderr: &stderr,
			},
			argv: []string{},
			want: ExitIOError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := run(tt.argv, tt.env)
			if got != tt.want {
				t.Errorf("run() = %v, want %v", got, tt.want)
			}
		})
	}
}
