package error_boundaries

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestErrorBoundariesExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / compiler heuristic it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.

var errorBoundariesExtrasValid = []rule_tester.ValidTestCase{
	// N/A: receiver / member-access wrappers are not inspected by this rule.
	// N/A: property key forms are not inspected by this rule.
	// N/A: autofix boundaries do not apply because the rule has no fixes.
	// ---- Dimension 4: anonymous default exports enable the file pre-pass but are not lint functions. ----
	{Code: `export default () => { try { return <Child />; } catch { return null; } }`, Tsx: true},
	// ---- Dimension 4: TS wrappers around function initializers are not peeled by the compiler pre-pass. ----
	{Code: `const Component = (() => { try { return <Child />; } catch { return null; } }) as React.FC;`, Tsx: true},
	{Code: `const Component = (() => { try { return <Child />; } catch { return null; } })!;`, Tsx: true},
	// ---- Dimension 4: nested function boundary inside try must not inherit the outer try. ----
	{Code: `function Component(){ try { const render = () => <Child />; return render; } catch { return null; } }`, Tsx: true},
	{Code: `function Component(){ try { function useInner(){ return <Child />; } return useInner; } catch { return null; } }`, Tsx: true},
	{Code: `function Component(){ try { foo(); } catch { try { foo(); } catch { return <Child />; } } }`, Tsx: true},
	// ---- Dimension 4: nested PascalCase component declarations are not compiler lint functions. ----
	{Code: `function Component(){ function Inner(){ try { return <Child />; } catch { return null; } } return <Inner />; }`, Tsx: true},
	// ---- Dimension 4: object/class method containers are not compiler lint functions. ----
	{Code: `function Sentinel(){ return null; } const obj = { Component(){ try { return <Child />; } catch { return null; } } };`, Tsx: true},
	{Code: `function Sentinel(){ return null; } class C { Component(){ try { return <Child />; } catch { return null; } } }`, Tsx: true},
	{Code: `function Sentinel(){ return null; } const obj = { Component: () => { try { return <Child />; } catch { return null; } } };`, Tsx: true},
	{Code: `function Sentinel(){ return null; } class C { Component = () => { try { return <Child />; } catch { return null; } } }`, Tsx: true},
	// ---- Dimension 4: assignment-created function names are not compiler lint functions. ----
	{Code: `function Sentinel(){ return null; } Component = () => { try { return <Child />; } catch { return null; } };`, Tsx: true},
	// ---- Dimension 4: memo callback alone is skipped by upstream's file heuristic. ----
	{Code: `const C = React.memo(function(){ try { return <Child />; } catch { return null; } });`, Tsx: true},
	// ---- Dimension 4: optional-chain wrappers are not compiler lint functions. ----
	{Code: `function Sentinel(){ return null; } React.memo?.(function(){ try { return <Child />; } catch { return null; } });`, Tsx: true},
	{Code: `function Sentinel(){ return null; } React?.memo(function(){ try { return <Child />; } catch { return null; } });`, Tsx: true},
	// ---- Dimension 4: element-access wrapper names are not accepted. ----
	{Code: `function Sentinel(){ return null; } React["memo"](function(){ try { return <Child />; } catch { return null; } });`, Tsx: true},
	// ---- Dimension 4: generator functions are skipped by the compiler lint target selection. ----
	{Code: `function* Component(){ try { yield <Child />; } catch { return null; } }`, Tsx: true},
	{Code: `const Component = function*(){ try { yield <Child />; } catch { return null; } };`, Tsx: true},
	{Code: `function Sentinel(){ return null; } React.memo(function*(){ try { yield <Child />; } catch { return null; } });`, Tsx: true},
	// ---- Dimension 4: JSX in finally is ignored, even under an outer try. ----
	{Code: `function Component(){ try { foo(); } finally { return <Child />; } }`, Tsx: true},
	{Code: `function Component(){ try { try { foo(); } finally { return <Child />; } } catch { return null; } }`, Tsx: true},
	{Code: `function Component(){ try { foo(); } finally { try { return <Child />; } catch { return null; } } }`, Tsx: true},
	// ---- Real-user: facebook/react#31237 local bare `use` is not a custom Hook target. ----
	{Code: `function use(){ try { return <Child />; } catch { return null; } }`, Tsx: true},
	// ---- Real-user: facebook/react#31446 component names follow upstream's ASCII PascalCase predicate. ----
	{Code: `function ÄndraVärde(){ try { return <Child />; } catch { return null; } }`, Tsx: true},
	// ---- Dimension 4: empty graceful-degradation forms. ----
	{Code: `function Component(){ try {} catch {} return null; }`, Tsx: true},
	{Code: `function Component(){ try { return null; } finally {} }`, Tsx: true},
	// ---- Real-user: facebook/react#34131 try/finally shape does not emit error-boundaries. ----
	{Code: `function Component(){ let el; try { el = null; } finally { console.log(el); } return el; }`, Tsx: true},
}

var errorBoundariesExtrasInvalid = []rule_tester.InvalidTestCase{
	// ---- Dimension 4: parenthesized JSX expression in a try block. ----
	{
		Code: `function Component(){ try { return (<Child />); } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 37, EndLine: 1, EndColumn: 46},
		},
	},
	// ---- Dimension 4: TS type-expression wrapper around JSX in a try block. ----
	{
		Code: `function Component(){ try { return <Child /> as React.ReactNode; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 36, EndLine: 1, EndColumn: 45},
		},
	},
	// ---- Dimension 4: TS satisfies wrapper around JSX in a try block. ----
	{
		Code: `function Component(){ try { return <Child /> satisfies React.ReactNode; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 36, EndLine: 1, EndColumn: 45},
		},
	},
	// ---- Dimension 4: top-level component variable declaration. ----
	{
		Code: `const Component = () => { try { return <Child />; } catch { return null; } };`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 40, EndLine: 1, EndColumn: 49},
		},
	},
	{
		Code: `export const Component = () => { try { return <Child />; } catch { return null; } };`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 47, EndLine: 1, EndColumn: 56},
		},
	},
	{
		Code: `const Component: React.FC = () => { try { return <Child />; } catch { return null; } };`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 50, EndLine: 1, EndColumn: 59},
		},
	},
	// ---- Dimension 4: async functions are still compiler lint functions. ----
	{
		Code: `async function Component(){ try { return <Child />; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 42, EndLine: 1, EndColumn: 51},
		},
	},
	{
		Code: `const Component = async () => { try { return <Child />; } catch { return null; } };`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 46, EndLine: 1, EndColumn: 55},
		},
	},
	// ---- Dimension 4: hook declarations are lint functions even when nested. ----
	{
		Code: `function Component(){ function useInner(){ try { return <Child />; } catch { return null; } } return null; }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 57, EndLine: 1, EndColumn: 66},
		},
	},
	{
		Code: `function Component(){ const useInner = () => { try { return <Child />; } catch { return null; } }; return null; }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 61, EndLine: 1, EndColumn: 70},
		},
	},
	{
		Code: `function use1(){ try { return <Child />; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 31, EndLine: 1, EndColumn: 40},
		},
	},
	// ---- Dimension 4: memo / forwardRef callbacks report once the file heuristic is enabled. ----
	{
		Code: `function Sentinel(){ return null; } const C = React.memo(function(){ try { return <Child />; } catch { return null; } });`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 83, EndLine: 1, EndColumn: 92},
		},
	},
	{
		Code: `function Sentinel(){ return null; } const C = React.forwardRef((props, ref) => { try { return <Child />; } catch { return null; } });`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 95, EndLine: 1, EndColumn: 104},
		},
	},
	{
		Code: `function Sentinel(){ return null; } React.memo(function(){ try { return <Child />; } catch { return null; } });`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 73, EndLine: 1, EndColumn: 82},
		},
	},
	{
		Code: `function Sentinel(){ return null; } (React.memo)(function(){ try { return <Child />; } catch { return null; } });`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 75, EndLine: 1, EndColumn: 84},
		},
	},
	{
		Code: `function Sentinel(){ return null; } foo(forwardRef((props, ref) => { try { return <Child />; } catch { return null; } }));`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 83, EndLine: 1, EndColumn: 92},
		},
	},
	{
		Code: `function Sentinel(){ return null; } const obj = { x: React.memo((function helper(){ try { return <Child />; } catch { return null; } })) };`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 98, EndLine: 1, EndColumn: 107},
		},
	},
	{
		Code: `function Sentinel(){ return null; } const C = React.memo(React.forwardRef((props, ref) => { try { return <Child />; } catch { return null; } }));`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 106, EndLine: 1, EndColumn: 115},
		},
	},
	{
		Code: `function Sentinel(){ return null; } function outer(){ const c = React.memo(function(){ try { return <Child />; } catch { return null; } }); }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 101, EndLine: 1, EndColumn: 110},
		},
	},
	// ---- Dimension 4: normal control-flow nesting under a try block still reports. ----
	{
		Code: `function Component(){ try { if (ok) { return <Child />; } return null; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 46, EndLine: 1, EndColumn: 55},
		},
	},
	{
		Code: `function Component(){ try { foo(); } catch { try { return <Child />; } catch { return null; } } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 59, EndLine: 1, EndColumn: 68},
		},
	},
	// ---- Dimension 4: multi-line diagnostics preserve upstream report ranges. ----
	{
		Code: `function Component() {
  try {
    if (ok) {
      return (
        <Child />
      );
    }
    return null;
  } catch {
    return null;
  }
}`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 5, Column: 9, EndLine: 5, EndColumn: 18},
		},
	},
	{
		Code: `function Component() {
  try {
    return (
      <A>
        <B />
      </A>
    );
  } catch {
    return null;
  }
}`,
		Tsx: true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 4, Column: 7, EndLine: 6, EndColumn: 11},
			{Line: 5, Column: 9, EndLine: 5, EndColumn: 14},
		},
	},
	// ---- Dimension 4: nested JSX reports both parent and child JSX constructions. ----
	{
		Code: `function Component(){ try { return <A><B /></A>; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 36, EndLine: 1, EndColumn: 48},
			{Line: 1, Column: 39, EndLine: 1, EndColumn: 44},
		},
	},
	// ---- Real-user: facebook/react#16026 component try/catch around hook result and JSX. ----
	{
		Code: `export function Things(){ try { const data = useThings(); return <Content>{!data ? <Loading /> : <ThingsList things={data} />}</Content>; } catch { return <Content><ThingsList roles={[]} /></Content>; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 66, EndLine: 1, EndColumn: 137},
			{Line: 1, Column: 84, EndLine: 1, EndColumn: 95},
			{Line: 1, Column: 98, EndLine: 1, EndColumn: 126},
		},
	},
	// Locks in upstream validateNoJSXInTryStatement arm 1: active try with JsxExpression.
	{
		Code: `function Component(){ try { const child = <Child />; return child; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 43, EndLine: 1, EndColumn: 52},
		},
	},
	// Locks in upstream validateNoJSXInTryStatement arm 2: active try with JsxFragment.
	{
		Code: `function Component(){ try { const child = <>x</>; return child; } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 43, EndLine: 1, EndColumn: 49},
		},
	},
	// Locks in upstream validateNoJSXInTryStatement arm 3: catch handler exits the inner try only.
	{
		Code: `function Component(){ try { try { foo(); } catch { return <Child />; } } catch { return null; } }`,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{
			{Line: 1, Column: 59, EndLine: 1, EndColumn: 68},
		},
	},
}

func TestErrorBoundariesExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &ErrorBoundariesRule,
		errorBoundariesExtrasValid,
		errorBoundariesExtrasInvalid,
	)
}
