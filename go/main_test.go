package main

import "testing"

func Test_run(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		argv []string
		want ExitCode
	}{
		{
			name: "invalid cli args",
			argv: []string{"-s", "bar"},
			want: ExitValidationError,
		},
		{
			name: "test file not found",
			argv: []string{"bar"},
			want: ExitIOError,
		},
		{
			name: "test access denied",
			argv: []string{"/etc/master.passwd"},
			want: ExitIOError,
		},
		{
			name: "test strict",
			argv: []string{"--strict", "../data/malformed_sample.log"},
			want: ExitStrictError,
		},
		{
			name: "test non-strict",
			argv: []string{"../data/malformed_sample.log"},
			want: ExitSuccess,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := run(tt.argv)
			if got != tt.want {
				t.Errorf("run() = %v, want %v", got, tt.want)
			}
		})
	}
}
