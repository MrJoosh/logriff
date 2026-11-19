package main

import "testing"

func Test_validator(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		f       []string
		want    lineData
		wantErr bool
	}{
		{
			name:    "invalid ipv4 all octets",
			f:       genFields(lineFields{addr: "999.999.999.999"}),
			want:    lineData{},
			wantErr: true,
		},
		{
			name:    "invalid ipv4 one octet",
			f:       genFields(lineFields{addr: "256.1.1.1"}),
			want:    lineData{},
			wantErr: true,
		},
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
			name:    "response time negative",
			f:       genFields(lineFields{response: "-1.23"}),
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
