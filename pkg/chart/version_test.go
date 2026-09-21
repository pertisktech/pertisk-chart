package chart

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int // sign only
	}{
		{"0.2.10", "0.2.6", 1},
		{"0.2.6", "0.2.10", -1},
		{"0.2.10", "0.2.10", 0},
		{"v0.2.10", "0.2.6", 1},
		{"1.0.0", "0.9.9", 1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"0.2.9", "0.2.10", -1},
		{"2.0", "2.0.0", 0},
		{"10.0.0", "9.9.9", 1},
	}

	for _, tt := range tests {
		got := CompareVersions(tt.a, tt.b)
		if sign(got) != sign(tt.want) {
			t.Errorf("CompareVersions(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func sign(n int) int {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}
