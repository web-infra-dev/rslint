// TestCatchErrorNameExtras locks in branches and edge shapes that the upstream
// test suite does not exercise. Upstream-migrated cases live in
// catch_error_name_upstream_test.go.
package catch_error_name_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/catch_error_name"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCatchErrorNameExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&catch_error_name.CatchErrorNameRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: destructuring parameters are not identifiers ----
			{Code: "promise.catch(({message}) => message)"},
			// ---- Dimension 4: optional invocation is excluded ----
			{Code: "promise.catch?.(bad => bad)"},
			// N/A: literals and computed keys are not parameter-name shapes.
			// ---- Real-user: issue #1075 unreproducible missing-variable shape ----
			{Code: "try {} catch (_) {}"},
			// ---- Real-user: nested descriptive errors remain readable ----
			{Code: "try {} catch (outerError) { try {} catch (innerError) {} }"},
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream isPromiseCatchParameter() arm 1: catch callback.
			invalid("promise.catch(function (bad) { return bad })", "promise.catch(function (error) { return error })", "bad", "error"),
			// Locks in upstream isPromiseCatchParameter() arm 2: then rejection callback.
			invalid("promise.then(ok => ok, bad => bad)", "promise.then(ok => ok, error => error)", "bad", "error"),
			// Locks in upstream collision arm: append an underscore.
			invalid("const error = 1; try {} catch (bad) { use(bad, error) }", "const error = 1; try {} catch (error_) { use(error_, error) }", "bad", "error_"),
			// Locks in renameVariable() shorthand expansion.
			invalid("try {} catch (bad) { use({bad}) }", "try {} catch (error) { use({bad: error}) }", "bad", "error"),
		},
	)
}
