package linter

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestStableSortDiagnosticsByFileAndStartPreservesEqualKeyOrder(t *testing.T) {
	diagnostics := []rule.RuleDiagnostic{
		{FilePath: "b.ts", Range: core.NewTextRange(0, 1), RuleName: "b"},
		{FilePath: "a.ts", Range: core.NewTextRange(9, 10), RuleName: "later"},
		{FilePath: "a.ts", Range: core.NewTextRange(2, 8), RuleName: "same-start-first"},
		{FilePath: "a.ts", Range: core.NewTextRange(2, 3), RuleName: "same-start-second"},
	}

	StableSortDiagnosticsByFileAndStart(diagnostics)

	got := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		got = append(got, diagnostic.RuleName)
	}
	want := []string{"same-start-first", "same-start-second", "later", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diagnostic order = %v, want %v", got, want)
	}
}
