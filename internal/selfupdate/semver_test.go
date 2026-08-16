package selfupdate

import (
	"testing"
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input     string
		wantErr   bool
		wantMajor int
		wantMinor int
		wantPatch int
		wantPre   string
	}{
		{input: "v4.3.2", wantMajor: 4, wantMinor: 3, wantPatch: 2, wantPre: ""},
		{input: "4.3.2", wantMajor: 4, wantMinor: 3, wantPatch: 2, wantPre: ""},
		{input: "V1.0.0", wantMajor: 1, wantMinor: 0, wantPatch: 0, wantPre: ""},
		{input: "v2.5.0-rc.1", wantMajor: 2, wantMinor: 5, wantPatch: 0, wantPre: "rc.1"},
		{input: "v3.0.0+build.123", wantMajor: 3, wantMinor: 0, wantPatch: 0, wantPre: ""},
		{input: "1.2", wantMajor: 1, wantMinor: 2, wantPatch: 0, wantPre: ""},
		{input: "5", wantMajor: 5, wantMinor: 0, wantPatch: 0, wantPre: ""},
		{input: "", wantErr: true},
		{input: "invalid", wantErr: true},
		{input: "1.2.3.4", wantErr: true},
		{input: "-1.0.0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := ParseSemver(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSemver(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if err == nil {
				if v.Major != tt.wantMajor || v.Minor != tt.wantMinor || v.Patch != tt.wantPatch || v.Prerelease != tt.wantPre {
					t.Errorf("ParseSemver(%q) = %+v, want Major=%d, Minor=%d, Patch=%d, Pre=%q",
						tt.input, v, tt.wantMajor, tt.wantMinor, tt.wantPatch, tt.wantPre)
				}
				if v.String() == "" {
					t.Errorf("v.String() should not be empty")
				}
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		v1      string
		v2      string
		want    int
		isNewer bool
	}{
		{v1: "v4.4.0", v2: "v4.3.2", want: 1, isNewer: true},
		{v1: "4.3.2", v2: "v4.3.2", want: 0, isNewer: false},
		{v1: "v4.3.1", v2: "v4.3.2", want: -1, isNewer: false},
		{v1: "v5.0.0", v2: "v4.9.9", want: 1, isNewer: true},
		{v1: "v4.3.2", v2: "v4.3.2-rc.1", want: 1, isNewer: true},
		{v1: "v4.3.2-rc.1", v2: "v4.3.2", want: -1, isNewer: false},
		{v1: "v4.3.2-rc.2", v2: "v4.3.2-rc.1", want: 1, isNewer: true},
		{v1: "v4.3.2-alpha", v2: "v4.3.2-beta", want: -1, isNewer: false},
		{v1: "v4.3.2-rc.1.1", v2: "v4.3.2-rc.1", want: 1, isNewer: true},
		{v1: "v4.4.0", v2: "dev", want: 1, isNewer: true},
		{v1: "dev", v2: "v4.4.0", want: -1, isNewer: false},
		{v1: "dirty", v2: "dirty", want: 0, isNewer: false},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			got := Compare(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
			if gotNewer := IsNewer(tt.v1, tt.v2); gotNewer != tt.isNewer {
				t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.v1, tt.v2, gotNewer, tt.isNewer)
			}
		})
	}
}
