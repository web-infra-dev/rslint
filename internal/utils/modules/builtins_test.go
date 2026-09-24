package modules

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/semver"
)

func TestIsNodeBuiltinAtVersion(t *testing.T) {
	for _, tc := range []struct {
		name, version string
		want          bool
	}{
		{"fs", "0.1.0", true},
		{"node:fs", "14.17.0", false},
		{"node:fs", "14.18.0", true},
		{"node:fs", "15.0.0", false},
		{"node:fs", "16.0.0", true},
		{"fs/promises", "10.0.0", true},
		{"fs/promises", "10.1.0", false},
		{"fs/promises", "13.14.0", false},
		{"fs/promises", "14.0.0", true},
		{"assert/strict", "14.18.0", false},
		{"assert/strict", "15.0.0", true},
		{"node:assert/strict", "15.0.0", false},
		{"node:assert/strict", "16.0.0", true},
		{"node:stream/web", "16.4.0", false},
		{"node:stream/web", "16.5.0", true},
		{"test", "22.0.0", false},
		{"node:test", "16.16.0", false},
		{"node:test", "16.17.0", true},
		{"node:test", "17.0.0", false},
		{"node:test", "18.0.0", true},
		{"test/reporters", "19.9.0", true},
		{"test/reporters", "20.2.0", false},
		{"node:test/reporters", "20.2.0", true},
		{"_stream_wrap", "25.0.0", true},
		{"_stream_wrap", "26.0.0", false},
		{"node:_stream_wrap", "26.0.0", false},
		{"_debugger", "7.9.0", true},
		{"_debugger", "8.0.0", false},
		{"node:_debugger", "7.9.0", false},
		{"node:wasi", "18.16.0", false},
		{"node:wasi", "18.17.0", true},
		{"node:wasi", "19.0.0", false},
		{"node:wasi", "20.0.0", true},
		{"node:ffi", "26.8.0", false},
		{"node:ffi", "26.9.0", true},
		{"unknown", "22.0.0", false},
		{"node:unknown", "22.0.0", false},
		{"fs/not-a-builtin", "22.0.0", false},
		{"node:node:fs", "22.0.0", false},
	} {
		t.Run(tc.name+"/"+tc.version, func(t *testing.T) {
			if got := IsNodeBuiltinAtVersion(tc.name, semver.MustParse(tc.version)); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
