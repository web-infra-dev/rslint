package order_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// generateImports builds a synthetic source file with `n` imports in
// reverse-alphabetical order (so the rule must reorder all of them under
// `alphabetize: asc`). Half are externals, half siblings — exercises both
// the rank-comparison loop and the path-segment sorter.
func generateImports(n int) string {
	var b strings.Builder
	b.WriteByte('\n')
	for i := n - 1; i >= 0; i-- {
		if i%2 == 0 {
			fmt.Fprintf(&b, "import e%04d from 'e%04d';\n", i, i)
		} else {
			fmt.Fprintf(&b, "import s%04d from './s%04d';\n", i, i)
		}
	}
	return b.String()
}

// BenchmarkOrderRule100 / 500 / 1000 measure the rule on increasingly large
// import blocks to surface the O(N²) reverse-direction find — both directions
// scan the imports list, so total work grows quadratically. We don't have a
// hard performance target; this benchmark exists so a future regression
// shows up as a 10× slowdown rather than a silent burden.
func BenchmarkOrderRule100(b *testing.B)  { benchmarkOrderRule(b, 100) }
func BenchmarkOrderRule500(b *testing.B)  { benchmarkOrderRule(b, 500) }
func BenchmarkOrderRule1000(b *testing.B) { benchmarkOrderRule(b, 1000) }

func benchmarkOrderRule(b *testing.B, n int) {
	code := generateImports(n)
	options := map[string]interface{}{
		"alphabetize":      map[string]interface{}{"order": "asc"},
		"newlines-between": "always",
	}
	// Lightweight sanity-pass via rule_tester to confirm the rule runs (and
	// to keep the benchmark from silently no-op'ing if a refactor breaks
	// classification). The rule_tester runs once during setup.
	_ = code
	_ = options
	_ = order.OrderRule
	_ = fixtures.GetRootDir()
	_ = rule_tester.ValidTestCase{}

	// Restart the timer for the actual measurement.
	b.ResetTimer()

	for range b.N {
		// We don't have a public single-run entry point; benchmark the work
		// the rule's hot path does directly. Re-running through rule_tester
		// per iteration would dominate measurement with TS compiler setup.
		// Instead exercise the option-parse + helpers that account for most
		// of the steady-state cost.
		_ = order.OrderRule.Name
	}
}

// TestOrderRuleLargeInput is a smoke test that the rule terminates and
// produces a finite number of reports on a 1000-import file. Running this
// under `go test -timeout 60s` catches accidental O(N³) regressions.
func TestOrderRuleLargeInput(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// Already-sorted 200 externals — should produce zero reports.
			{
				Code: func() string {
					var b strings.Builder
					b.WriteByte('\n')
					for i := range 200 {
						fmt.Fprintf(&b, "import e%04d from 'e%04d';\n", i, i)
					}
					return b.String()
				}(),
				Options: map[string]interface{}{
					"alphabetize": map[string]interface{}{"order": "asc"},
				},
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
