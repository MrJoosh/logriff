package main

import (
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

type fileOpenerFn func(name string) (io.ReadCloser, error)

func (f fileOpenerFn) Open(name string) (io.ReadCloser, error) { return f(name) }

type fixedStdout struct{ w io.Writer }

func (s fixedStdout) Stdout() io.Writer { return s.w }

type fixedStdin struct{ r io.Reader }

func (s fixedStdin) Stdin() io.Reader { return s.r }

type fixedStderr struct{ w io.Writer }

func (s fixedStderr) Stderr() io.Writer { return s.w }

type testRuntime struct {
	FileOpener
	StdinProvider
	StdoutProvider
	StderrProvider
}

func newTestRt() testRuntime {
	return testRuntime{
		StdinProvider:  fixedStdin{r: strings.NewReader("")},
		StdoutProvider: fixedStdout{w: io.Discard},
		StderrProvider: fixedStderr{w: io.Discard},
		FileOpener: fileOpenerFn(func(string) (io.ReadCloser, error) {
			return nil, errors.New("unexpected Open call")
		})}

}

func Test_run(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		setup func(*testRuntime)
		argv  []string
		want  ExitCode
	}{
		{
			name:  "invalid cli args",
			setup: nil,
			argv:  []string{"-s", "bar"},
			want:  ExitValidationError,
		},
		{
			name: "file not found",
			setup: func(rt *testRuntime) {
				rt.FileOpener = fileOpenerFn(func(name string) (io.ReadCloser, error) {
					return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
				})
			},
			argv: []string{"bar"},
			want: ExitIOError,
		},
		{
			name: "file access denied",
			setup: func(rt *testRuntime) {
				rt.FileOpener = fileOpenerFn(func(name string) (io.ReadCloser, error) {
					return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
				})
			},
			argv: []string{"foo"},
			want: ExitIOError,
		},
		{
			name: "strict mode",
			setup: func(rt *testRuntime) {
				rt.StdinProvider = fixedStdin{r: strings.NewReader("Malformed Line!\nMalformed Line!\n\n")}
			},
			argv: []string{"--strict"},
			want: ExitStrictError,
		},
		{
			name: "non-strict mode",
			setup: func(rt *testRuntime) {
				rt.StdinProvider = fixedStdin{r: strings.NewReader("Malformed Line!\nMalformed Line!\n\n")}
			},
			argv: []string{},
			want: ExitSuccess,
		},
		{
			name: "file open and read",
			setup: func(rt *testRuntime) {
				rt.FileOpener = fileOpenerFn(func(name string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(`192.168.1.10 - - [06/Nov/2026:08:12:01 +0000] "GET /api/v1/health HTTP/1.1" 200 42 1.2 "Go-http-client/1.1"`)), nil
				})
			},
			argv: []string{},
			want: ExitSuccess,
		},
		{
			name: "scanner read error",
			setup: func(rt *testRuntime) {
				rt.StdinProvider = fixedStdin{r: iotest.ErrReader(errors.New("stdin read error"))}
			},
			argv: []string{},
			want: ExitIOError,
		},
		{
			name: "stdout broken pipe",
			setup: func(rt *testRuntime) {
				rt.StdinProvider = fixedStdin{r: strings.NewReader(`192.168.1.10 - - [06/Nov/2026:08:12:01 +0000] "GET /api/v1/health HTTP/1.1" 200 42 1.2 "Go-http-client/1.1"`)}
				rt.StdoutProvider = fixedStdout{w: brokenWriter{err: errors.New("write /dev/stdout: broken pipe")}}
			},
			argv: []string{},
			want: ExitIOError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := newTestRt()
			if tt.setup != nil {
				tt.setup(&rt)
			}
			got := run(tt.argv, rt)
			if got != tt.want {
				t.Errorf("run() = %v, want %v", got, tt.want)
			}
		})
	}
}
