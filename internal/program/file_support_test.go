package program

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
)

func TestCompilerOptionsSupportFileName(t *testing.T) {
	tests := []struct {
		name     string
		options  *core.CompilerOptions
		fileName string
		want     bool
	}{
		{name: "nil options", fileName: "index.ts", want: false},
		{name: "TypeScript", options: &core.CompilerOptions{}, fileName: "index.ts", want: true},
		{name: "JavaScript disabled", options: &core.CompilerOptions{}, fileName: "index.js", want: false},
		{name: "JavaScript enabled", options: &core.CompilerOptions{AllowJs: core.TSTrue}, fileName: "index.js", want: true},
		{name: "JSON disabled", options: &core.CompilerOptions{ResolveJsonModule: core.TSFalse}, fileName: "data.json", want: false},
		{name: "JSON enabled", options: &core.CompilerOptions{ResolveJsonModule: core.TSTrue}, fileName: "data.json", want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CompilerOptionsSupportFileName(test.options, test.fileName); got != test.want {
				t.Fatalf("CompilerOptionsSupportFileName(%q) = %t, want %t", test.fileName, got, test.want)
			}
		})
	}
}
