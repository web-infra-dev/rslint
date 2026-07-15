package compilerpath

import (
	"runtime"
	"strings"
	"testing"
)

func TestCanRepresentNativePOSIXBackslash(t *testing.T) {
	path := `/repo/a\b.ts`
	want := runtime.GOOS == "windows"
	if got := CanRepresent(path); got != want {
		t.Fatalf("CanRepresent(%q) = %v, want %v", path, got, want)
	}
}

func TestSourceSuffixPreservesCompoundSuffixSpelling(t *testing.T) {
	for _, test := range []struct {
		path string
		want string
	}{
		{path: "/repo/index.D.MTS", want: ".D.MTS"},
		{path: "/repo/index.d.cts", want: ".d.cts"},
		{path: "/repo/component.TSX", want: ".TSX"},
		{path: "/repo/style.css", want: ".css"},
	} {
		if got := sourceSuffix(test.path); got != test.want {
			t.Errorf("sourceSuffix(%q) = %q, want %q", test.path, got, test.want)
		}
	}
}

func TestAliasAvoidsReservedCompilerPathCollision(t *testing.T) {
	reserved := make(map[string]struct{})
	first := Alias(`/repo/a\b.D.TS`, nil, reserved)
	second := Alias(`/repo/a\b.D.TS`, nil, reserved)
	if first == second {
		t.Fatalf("reserved alias was reused: %q", first)
	}
	if !strings.HasSuffix(first, ".D.TS") || !strings.HasSuffix(second, ".D.TS") {
		t.Fatalf("compound source suffix was not retained: first=%q second=%q", first, second)
	}
	if !strings.Contains(first, "-0.D.TS") || !strings.Contains(second, "-1.D.TS") {
		t.Fatalf("collision salt did not advance deterministically: first=%q second=%q", first, second)
	}
}
