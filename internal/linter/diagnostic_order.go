package linter

import (
	"cmp"
	"slices"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// StableSortDiagnosticsByFileAndStart orders a completed diagnostic set by
// caller-visible file path and start byte offset. Diagnostics with the same
// file and start retain their emission order.
func StableSortDiagnosticsByFileAndStart(diagnostics []rule.RuleDiagnostic) {
	slices.SortStableFunc(diagnostics, func(a, b rule.RuleDiagnostic) int {
		if c := strings.Compare(a.FilePath, b.FilePath); c != 0 {
			return c
		}
		return cmp.Compare(a.Range.Pos(), b.Range.Pos())
	})
}
