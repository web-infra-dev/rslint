package error_boundaries

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestErrorBoundariesUpstream migrates the full valid/invalid suite available
// from upstream React Compiler fixtures for validateNoJSXInTryStatement 1:1.
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in error_boundaries_extras_test.go.

var errorBoundariesUpstreamValid = []rule_tester.ValidTestCase{
	// ---- Upstream docs: ErrorBoundary wrapper is the intended pattern. ----
	{Code: `function Parent(){ return <ErrorBoundary><ChildComponent /></ErrorBoundary>; }`, Tsx: true},
	// ---- Upstream docs: catch JSX is allowed when not nested in an outer try. ----
	{Code: `function Parent(){ try { doSomething(); } catch { return <div>Error occurred</div>; } }`, Tsx: true},
	// ---- Upstream compiler: lowercase utility functions are skipped. ----
	{Code: `function helper(){ try { return <Child />; } catch { return null; } }`, Tsx: true},
	// ---- Upstream compiler: module-level try/catch is skipped. ----
	{Code: `try { const el = <Child />; } catch {}`, Tsx: true},
	// ---- Upstream compiler TODO fixture: finally does not emit this lint category. ----
	{Code: `function Component(){ try { foo(); } finally { return <Child />; } }`, Tsx: true},
	// ---- Upstream docs: use() in try/catch is handled by rules-of-hooks, not this rule. ----
	{Code: `function Component({promise}){ try { const data = use(promise); return <div>{data}</div>; } catch { return null; } }`, Tsx: true, Skip: true},
}

var errorBoundariesUpstreamInvalid = []rule_tester.InvalidTestCase{
	// ---- Upstream docs: JSX returned directly from a try block. ----
	{
		Code: `function Parent(){ try { return <ChildComponent />; } catch (error) { return <div>Error occurred</div>; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Message: errorBoundariesMessage, Line: 1, Column: 33, EndLine: 1, EndColumn: 51},
		},
	},
	// ---- Upstream fixture: invalid-jsx-in-try-with-catch.js. ----
	{
		Code: `function Component(){ let el; try { el = <div />; } catch { return null; } return el; }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 42, EndLine: 1, EndColumn: 49},
		},
	},
	// ---- Upstream fixture: invalid-jsx-in-catch-in-outer-try-with-catch.js. ----
	{
		Code: `function Component(props){ let el; try { let value; try { value = identity(props.foo); } catch { el = <div value={value} />; } } catch { return null; } return el; }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 103, EndLine: 1, EndColumn: 124},
		},
	},
	// ---- Upstream compiler: fragments are JsxFragment diagnostics. ----
	{
		Code: `function Component(){ try { return <>x</>; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 36, EndLine: 1, EndColumn: 42},
		},
	},
}

func TestErrorBoundariesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ErrorBoundariesRule,
		errorBoundariesUpstreamValid,
		errorBoundariesUpstreamInvalid,
	)
}
