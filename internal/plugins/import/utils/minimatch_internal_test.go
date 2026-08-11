package utils

import "testing"

func TestNormalizeMinimatchPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		pattern            string
		windows            bool
		allowWindowsEscape bool
		want               string
	}{
		{name: "Unix preserves escapes", pattern: `pkg\{a,b\}`, want: `pkg\{a,b\}`},
		{name: "Windows separator", pattern: `pkg\nested\*`, windows: true, want: "pkg/nested/*"},
		{name: "Windows escape opt-in", pattern: `pkg\{a,b\}`, windows: true, allowWindowsEscape: true, want: `pkg\{a,b\}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeMinimatchPattern(test.pattern, test.windows, test.allowWindowsEscape); got != test.want {
				t.Fatalf("normalizeMinimatchPattern(%q) = %q, want %q", test.pattern, got, test.want)
			}
		})
	}
}
