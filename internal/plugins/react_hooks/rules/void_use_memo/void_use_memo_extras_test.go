package void_use_memo

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestVoidUseMemoExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
func TestVoidUseMemoExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized callee and callback wrappers are transparent. ----
		{Code: `function Component({value}) {
  const memo = ((useMemo))((() => value), [value]);
  return <div>{memo}</div>;
}
`, Tsx: true},
		// Locks in upstream validateUseMemo() arm: missing callback is handled by react-hooks/use-memo, not void-use-memo.
		{Code: `function Component() {
  useMemo();
  return <div />;
}
`, Tsx: true},
		// Locks in upstream validateUseMemo() arm: non-inline callback is handled by react-hooks/use-memo, not void-use-memo.
		{Code: `function Component({value}) {
  const compute = () => value;
  const memo = useMemo(compute, [value]);
  return <div>{memo}</div>;
}
`, Tsx: true},
		// Locks in upstream unusedUseMemos deletion arm: the void operator consumes the call result.
		{Code: `function Component() {
  void useMemo(() => {
    return [];
  }, []);
  return <div />;
}
`, Tsx: true},
		// Locks in upstream operand deletion arms: logical, conditional, call, array, assignment, and JSX positions consume the result.
		{Code: `function Component({ok}) {
  let value;
  ok && useMemo(() => {
    return [];
  }, []);
  ok ? useMemo(() => {
    return [];
  }, []) : value;
  consume(useMemo(() => {
    return [];
  }, []));
  [useMemo(() => {
    return [];
  }, [])];
  value = useMemo(() => {
    return [];
  }, []);
  return <div>{useMemo(() => {
    return value;
  }, [value])}</div>;
}
`, Tsx: true},
		// Locks in upstream comma-expression arm: the final operand's result is not reported.
		{Code: `function Component() {
  sideEffect(), useMemo(() => {
    return [];
  }, []);
  return <div />;
}
`, Tsx: true},
		// Locks in upstream comma-expression arm: a final useMemo operand remains consumed when the whole sequence is non-null asserted.
		{Code: `function Component() {
  ((sideEffect(), useMemo(() => {
    return [];
  }, []))!, last());
  return <div />;
}
`, Tsx: true},
		// ---- Dimension 4: TS assertion wrappers around a returned value consume the result. ----
		{Code: `function Component() {
  (useMemo(() => {
    return [];
  }, []) as unknown);
  (useMemo(() => {
    return [];
  }, []) satisfies unknown);
  return <div />;
}
`, Tsx: true},
		// Locks in upstream unusedUseMemos deletion arm: assigning to a variable consumes the call result even when the binding is never read.
		{Code: `function Component() {
  const _ = useMemo(() => {
    return [];
  }, []);
  return <div />;
}
`, Tsx: true},
		// Locks in upstream terminal-operand arm: returning the useMemo result consumes it.
		{Code: `function Component() {
  return useMemo(() => {
    return <div />;
  }, []);
}
`, Tsx: true},
		// Locks in upstream hasNonVoidReturn() arm: any explicit return terminal is enough.
		{Code: `function Component({cond}) {
  const value = useMemo(() => {
    if (cond) {
      return 1;
    }
  }, [cond]);
  return <div>{value}</div>;
}
`, Tsx: true},
		// Locks in upstream hasNonVoidReturn() arm: implicit void expression returns still count as a return.
		{Code: `function Component({effect}) {
  const value = useMemo(() => void effect(), [effect]);
  return <div>{value}</div>;
}
`, Tsx: true},
		// Locks in upstream HIR terminal scan: returns inside try/catch count when there is no finally.
		{Code: `function Component() {
  const value = useMemo(() => {
    try {
      return 1;
    } catch (error) {
      return 2;
    }
  }, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// Locks in upstream HIR terminal scan: a return after try/finally still counts.
		{Code: `function Component() {
  const value = useMemo(() => {
    try {
      work();
    } finally {
      cleanup();
    }
    return 1;
  }, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// Locks in upstream HIR terminal scan: catch returns count even with a finally block.
		{Code: `function Component() {
  const value = useMemo(() => {
    try {
      work();
    } catch (error) {
      return 1;
    } finally {
      cleanup();
    }
  }, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Options contract: upstream accepts an arbitrary object and void-use-memo ignores it. ----
		{
			Code: `function Component({value}) {
  const memo = useMemo(() => value, [value]);
  return <div>{memo}</div>;
}
`,
			Options: map[string]interface{}{"someFutureOption": true},
			Tsx:     true,
		},
		// ---- Dimension 4: local shadowing means the callee is not React's useMemo. ----
		{Code: `function Component() {
  const useMemo = (fn) => fn();
  useMemo(() => {
    console.log('local');
  }, []);
  return <div />;
}
`, Tsx: true},
		// ---- Dimension 4: element access and optional namespace calls are outside upstream's input surface. ----
		{Code: `function Component() {
  React["useMemo"](() => {
    console.log('ignored');
  }, []);
  React?.useMemo(() => {
    console.log('ignored');
  }, []);
  return <div />;
}
`, Tsx: true},
		// ---- Dimension 4: TS assertion wrappers around callback arguments are handled by react-hooks/use-memo. ----
		{Code: `type Fn = () => number;
function Component() {
  const value = useMemo((() => 1) as Fn, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// N/A: PrivateIdentifier and object/class property key forms are not inputs inspected by this rule.
		// N/A: Autofix boundaries do not apply because react-hooks/void-use-memo has no autofix.
		// N/A: Class declaration/class expression containers are not rule targets; the rule validates hook calls.
		// N/A: Empty class bodies and overload signatures do not affect useMemo call validation.
	}

	invalid := []rule_tester.InvalidTestCase{
		// Locks in upstream validateUseMemo() arm: no callback return reports missingReturn and does not also report unusedResult.
		{
			Code: `function Component() {
  useMemo(() => {
    console.log('effect');
  }, []);
  return <div />;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 2, 11, 4, 4),
			},
		},
		// ---- Dimension 4: parenthesized expression statements still discard the result. ----
		{
			Code: `function Component() {
  (useMemo(() => {
    return [];
  }, []));
  return <div />;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("unusedResult", unusedResultReason, unusedResultDescription, 2, 4, 2, 11),
			},
		},
		// ---- Dimension 4: TS non-null wrappers around the result are transparent for unused-result detection. ----
		{
			Code: `function Component() {
  (useMemo(() => {
    return [];
  }, []))!;
  return <div />;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("unusedResult", unusedResultReason, unusedResultDescription, 2, 4, 2, 11),
			},
		},
		// Locks in upstream comma-expression arm: non-final operands are discarded even when the overall expression is used.
		{
			Code: `function Component() {
  const x = (useMemo(() => {
    return [];
  }, []), sideEffect());
  return <div>{x}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("unusedResult", unusedResultReason, unusedResultDescription, 2, 14, 2, 21),
			},
		},
		// Locks in upstream comma-expression arm: nested non-final sequence operands are discarded inside logical expressions.
		{
			Code: `function Component({ok}) {
  ok && (useMemo(() => {
    return [];
  }, []), sideEffect());
  return <div />;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("unusedResult", unusedResultReason, unusedResultDescription, 2, 10, 2, 17),
			},
		},
		// ---- Dimension 4: React.useMemo expression statements report the full callee. ----
		{
			Code: `function Component() {
  React.useMemo(() => {
    return [];
  }, []);
  return <div />;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("unusedResult", unusedResultReason, unusedResultDescription, 2, 3, 2, 16),
			},
		},
		// ---- Dimension 4: nested function returns do not satisfy the outer useMemo callback. ----
		{
			Code: `function Component() {
  const value = useMemo(() => {
    function nested() {
      return 1;
    }
    nested();
  }, []);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 2, 25, 7, 4),
			},
		},
		// Locks in upstream HIR terminal scan: a return inside a try block with finally does not satisfy useMemo.
		{
			Code: `function Component() {
  const value = useMemo(() => {
    try {
      return 1;
    } finally {
      cleanup();
    }
  }, []);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 2, 25, 8, 4),
			},
		},
		// Locks in upstream HIR terminal scan: a return inside finally does not satisfy useMemo.
		{
			Code: `function Component() {
  const value = useMemo(() => {
    try {
      work();
    } finally {
      return 1;
    }
  }, []);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 2, 25, 8, 4),
			},
		},
		// ---- Options contract: arbitrary options do not change diagnostics. ----
		{
			Code: `function Component() {
  const value = useMemo(() => {
    console.log('effect');
  }, []);
  return <div>{value}</div>;
}
`,
			Options: []interface{}{map[string]interface{}{"someFutureOption": true}},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 2, 25, 4, 4),
			},
		},
		// ---- Real-user: facebook/react#25379 useMemo used for side effects and no returned value. ----
		{
			Code: `import {useMemo} from 'react';
function Component() {
  const value = useMemo(() => {
    console.log("Yippee!");
  }, []);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 3, 25, 5, 4),
			},
		},
		// ---- Real-user: facebook/react#17962 object-literal-looking callback body is a labeled statement, not a return. ----
		{
			Code: `function Component() {
  const ref2 = useMemo(() => {
    current: null;
  }, []);
  return <div>{ref2}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 2, 24, 4, 4),
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &VoidUseMemoRule, valid, invalid)
}
