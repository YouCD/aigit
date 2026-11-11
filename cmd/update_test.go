package cmd

import "testing"

func TestVersionLess(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"v0.1.1", "v0.1.2", true},
		{"v0.1.2", "v0.1.1", false},
		{"v0.1.2", "v0.1.2", false},
		{"0.1.1", "v0.1.2", true},
		{"v0.0.9", "v0.0.10", true},
		{"v0.9.9", "v0.10.0", true},
		{"v0.1.1-7-ga98bc1b", "v0.1.2", true},
		{"v0.1.2-1-gb18cab4", "v0.1.2", false},
		{"v1.0.0-rc1", "v1.0.0", false},
		{"v0.1", "v0.1.1", true},
		{"v1.0.0", "v2.0.0", true},
	}
	for _, tt := range tests {
		if got := versionLess(tt.a, tt.b); got != tt.want {
			t.Errorf("versionLess(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestVersionSegments(t *testing.T) {
	tests := []struct {
		in   string
		want []int
	}{
		{"v0.1.2", []int{0, 1, 2}},
		{"0.1.2", []int{0, 1, 2}},
		{"v0.1.1-7-ga98bc1b", []int{0, 1, 1}},
		{"v1.0.0+build5", []int{1, 0, 0}},
		{" v2.3.4 ", []int{2, 3, 4}},
		{"garbage", []int{}},
	}
	for _, tt := range tests {
		got := versionSegments(tt.in)
		if len(got) != len(tt.want) {
			t.Errorf("versionSegments(%q) = %v, want %v", tt.in, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("versionSegments(%q) = %v, want %v", tt.in, got, tt.want)
				break
			}
		}
	}
}
