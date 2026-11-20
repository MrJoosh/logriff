package main

import (
	"errors"
	"testing"
)

func Test_parseArgsInvalid(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		argv     []string
		want     args
		wantErr  bool
		wantHelp bool
	}{
		{
			name:     "check non-int status-min short flag",
			argv:     []string{"-s", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check non-int status-min long flag",
			argv:     []string{"--status-min", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check non-int status-max short flag",
			argv:     []string{"-e", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check non-int status-max long flag",
			argv:     []string{"--status-max", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check non-int top-routes short flag",
			argv:     []string{"-n", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check non-int top-routes long flag",
			argv:     []string{"--top-routes", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check invalid method short flag",
			argv:     []string{"-m", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check invalid method long flag",
			argv:     []string{"--method", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check non-numeric min-duration long flag",
			argv:     []string{"--min-duration", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check invalid format short flag",
			argv:     []string{"-f", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check invalid format long flag",
			argv:     []string{"--format", "bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check short flag with no value",
			argv:     []string{"-s"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check long flag with no value",
			argv:     []string{"--status-min"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check unknown long flag",
			argv:     []string{"--not-a-real-flag"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "check multiple positionals",
			argv:     []string{"first.log", "second.log"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "status min greater than status max",
			argv:     []string{"-s", "500", "-e", "200", "/foo/bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
		{
			name:     "negative top routes",
			argv:     []string{"-n", "-10", "/foo/bar"},
			want:     args{},
			wantErr:  true,
			wantHelp: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseArgs(tt.argv)
			if gotErr != nil {
				if errors.Is(gotErr, errHelp) && !tt.wantHelp {
					t.Fatal("unexpectedly got help flag")
				}
				if !errors.Is(gotErr, errHelp) && tt.wantHelp {
					t.Fatal("didn't get expected help flag")
				}
				if !tt.wantErr {
					t.Errorf("parseArgs() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseArgs() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("parseArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseArgs(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		argv     []string
		want     args
		wantErr  bool
		wantHelp bool
	}{
		{
			name: "check option defaults",
			argv: []string{},
			want: args{
				statusMin:   0,
				statusMax:   0,
				method:      "",
				path:        "",
				minDuration: 0,
				topRoutes:   5,
				format:      "text",
				strict:      false,
				file:        "-",
			},
			wantErr:  false,
			wantHelp: false,
		},
		{
			name:     "help returns success short flag",
			argv:     []string{"-h"},
			want:     args{},
			wantErr:  false,
			wantHelp: true,
		},
		{
			name:     "help returns success long flag",
			argv:     []string{"--help"},
			want:     args{},
			wantErr:  false,
			wantHelp: true,
		},
		{
			name: "long form filters accepted",
			argv: []string{"--status-min", "300", "--format", "json", "/foo/bar"},
			want: args{
				statusMin:   300,
				statusMax:   0,
				method:      "",
				path:        "",
				minDuration: 0,
				topRoutes:   5,
				format:      "json",
				strict:      false,
				file:        "/foo/bar",
			},
			wantErr:  false,
			wantHelp: false,
		},
		{
			name: "method case insensitive",
			argv: []string{"-m", "get"},
			want: args{
				statusMin:   0,
				statusMax:   0,
				method:      "GET",
				path:        "",
				minDuration: 0,
				topRoutes:   5,
				format:      "text",
				strict:      false,
				file:        "-",
			},
			wantErr:  false,
			wantHelp: false,
		},
		{
			name: "long form equals type filters accepted",
			argv: []string{"--status-min=300", "--format=json", "/foo/bar"},
			want: args{
				statusMin:   300,
				statusMax:   0,
				method:      "",
				path:        "",
				minDuration: 0,
				topRoutes:   5,
				format:      "json",
				strict:      false,
				file:        "/foo/bar",
			},
			wantErr:  false,
			wantHelp: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseArgs(tt.argv)
			if gotErr != nil {
				if errors.Is(gotErr, errHelp) {
					if !tt.wantHelp {
						t.Fatal("got help unexpectedly")
					}
					return
				}
				if tt.wantHelp {
					t.Fatalf("parseArgs() error = %v, want help", gotErr)
				}
				if !tt.wantErr {
					t.Errorf("parseArgs() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseArgs() succeeded unexpectedly")
			}
			if tt.wantHelp {
				t.Fatal("didn't get expected help flag")
			}
			if got != tt.want {
				t.Errorf("parseArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
