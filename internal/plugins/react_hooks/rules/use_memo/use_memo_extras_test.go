package use_memo

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestUseMemoExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
func TestUseMemoExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized receiver and callee wrappers. ----
		{Code: `function Component(props) {
  const a = (React).useMemo(() => props.value, [props.value]);
  const b = ((useMemo))(() => props.value, [props.value]);
  return <div>{a}{b}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: TS non-null wrappers are transparent in dependency expressions. ----
		{Code: `type Props = { value: number };
function Component(props: Props) {
  const value = useMemo(() => props.value, [props!.value]);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: static element access forms are simple dependency expressions. ----
		{Code: `function Component(props) {
  const a = useMemo(() => props.value, [props["value"]]);
  const b = useMemo(() => props.items[0], [props.items[0]]);
  return <div>{a}{b}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: dynamic element access does not masquerade as React.useMemo. ----
		{Code: `function Component(props) {
  const hook = "useMemo";
  return React[hook](async () => props.value, []);
}
`, Tsx: true},
		// ---- Dimension 4: static element access and optional namespace calls are outside upstream's input surface. ----
		{Code: `function Component(props) {
  const a = React["useMemo"](async () => props.value, []);
  const b = React?.useMemo(async () => props.value, []);
  return <div>{a}{b}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: TS assertion wrappers around the callee are not transparent. ----
		{Code: `function Component(props) {
  const a = React!.useMemo(async () => props.value, []);
  const b = (React as any).useMemo(async () => props.value, []);
  const c = (useMemo as any)(async () => props.value, []);
  return <div>{a}{b}{c}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: import aliases are outside upstream's input surface. ----
		{Code: `import {useMemo as memo} from 'react';
import R from 'react';
import * as Namespace from 'react';
function Component() {
  const a = memo(async () => 1, []);
  const b = R.useMemo(async () => 1, []);
  const c = Namespace.useMemo(async () => 1, []);
  return <div>{a}{b}{c}</div>;
}
`, Tsx: true},
		// Locks in upstream extractManualMemoizationArgs() arm 2: spread first argument is not modeled as a callback.
		{Code: `function Component(args) {
  const value = useMemo(...args);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: local callback assignments are allowed. ----
		{Code: `function Component() {
  const value = useMemo(() => {
    let x;
    x = [];
    return x;
  }, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: nested function body does not bleed into the useMemo callback scan. ----
		{Code: `function Component() {
  let x;
  const value = useMemo(() => {
    function nested() {
      x = [];
    }
    return nested;
  }, []);
  return <div>{value}{x}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: shadowed local useMemo is not treated as React's hook. ----
		{Code: `function Component() {
  const useMemo = (value) => value;
  const value = useMemo(1, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: shadowed imported alias is resolved as local, not React's useMemo. ----
		{Code: `import {useMemo as memo} from 'react';
function Component() {
  const memo = (value) => value;
  const value = memo(1, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: shadowed React namespace import is resolved as local. ----
		{Code: `import * as R from 'react';
function Component() {
  const R = {useMemo: (value) => value};
  const value = R.useMemo(1, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Dimension 4: property writes are not variable reassignments for this rule. ----
		{Code: `function Component() {
  const ref = {current: 0};
  const value = useMemo(() => {
    ref.current = 1;
    return ref.current;
  }, [ref]);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Real-user: facebook/react#25379 no-return useMemo belongs to void-use-memo, not use-memo. ----
		{Code: `function Component({value}) {
  const memo = useMemo(() => {
    console.log(value);
  }, [value]);
  return <div>{memo}</div>;
}
`, Tsx: true},
		// ---- Options contract: upstream accepts an arbitrary object and use-memo ignores it. ----
		{
			Code: `function Component({value}) {
  const memo = useMemo(() => value, [value]);
  return <div>{memo}</div>;
}
`,
			Options: map[string]interface{}{"someFutureOption": true},
			Tsx:     true,
		},
		// N/A: PrivateIdentifier and object/class property key forms are not inputs inspected by this rule.
		// N/A: Autofix boundaries do not apply because react-hooks/use-memo has no autofix.
		// N/A: Class declaration/class expression containers are not rule targets; the rule validates hook calls.
		// N/A: Empty class bodies and overload signatures do not affect useMemo call validation.
	}

	invalid := []rule_tester.InvalidTestCase{
		// Locks in upstream extractManualMemoizationArgs() arm 1: missing first argument.
		{
			Code: `function Component() {
  useMemo();
  return <div />;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedMemoizationFunction", expectedMemoizationFunctionReason, expectedMemoizationFunctionDescription, "Expected a memoization function", 2, 3, 2, 12),
			},
		},
		// Locks in upstream extractManualMemoizationArgs() arm 1: first argument is not a function.
		{
			Code: `function Component() {
  const value = useMemo(123, []);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedInlineFunction", expectedInlineFunctionReason, expectedInlineFunctionDescription, "Expected the first argument to be an inline function expression", 2, 25, 2, 28),
			},
		},
		// Locks in upstream extractManualMemoizationArgs() arm 3: dependency element is not simple.
		{
			Code: `function Component({x}) {
  const value = useMemo(() => x, [x + 1]);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedSimpleDependencies", expectedSimpleDepsReason, expectedSimpleDepsDescription, expectedSimpleDepsReason, 2, 35, 2, 40),
			},
		},
		// ---- Dimension 4: spread in dependency array is not modeled as an array dependency list upstream. ----
		{
			Code: `function Component({x, deps}) {
  const value = useMemo(() => x, [...deps]);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedArrayLiteral", expectedArrayLiteralReason, expectedArrayLiteralDescription, expectedArrayLiteralReason, 2, 34, 2, 43),
			},
		},
		// ---- Dimension 4: sparse dependency arrays are not modeled as array dependency lists upstream. ----
		{
			Code: `function Component({x}) {
  const value = useMemo(() => x, [, x]);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedArrayLiteral", expectedArrayLiteralReason, expectedArrayLiteralDescription, expectedArrayLiteralReason, 2, 34, 2, 39),
			},
		},
		// ---- Dimension 4: TS as/satisfies wrappers around callback arguments are not transparent upstream. ----
		{
			Code: `type Fn = () => number;
function Component() {
  const a = useMemo((() => 1) as Fn, []);
  const b = useMemo((() => 2) satisfies Fn, []);
  return <div>{a}{b}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedInlineFunction", expectedInlineFunctionReason, expectedInlineFunctionDescription, "Expected the first argument to be an inline function expression", 3, 21, 3, 36),
				useMemoError("expectedInlineFunction", expectedInlineFunctionReason, expectedInlineFunctionDescription, "Expected the first argument to be an inline function expression", 4, 21, 4, 43),
			},
		},
		// ---- Dimension 4: TS as/satisfies wrappers in dependency expressions are not simple upstream. ----
		{
			Code: `type Props = { value: number };
function Component(props: Props) {
  const a = useMemo(() => props.value, [(props as Props).value]);
  const b = useMemo(() => props.value, [(props satisfies Props).value]);
  return <div>{a}{b}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedSimpleDependencies", expectedSimpleDepsReason, expectedSimpleDepsDescription, expectedSimpleDepsReason, 3, 41, 3, 63),
				useMemoError("expectedSimpleDependencies", expectedSimpleDepsReason, expectedSimpleDepsDescription, expectedSimpleDepsReason, 4, 41, 4, 70),
			},
		},
		// ---- Dimension 4: no-substitution template element keys are not simple dependency paths upstream. ----
		{
			Code: "function Component({props}) {\n  const value = useMemo(() => props.value, [props[`value`]]);\n  return <div>{value}</div>;\n}\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedSimpleDependencies", expectedSimpleDepsReason, expectedSimpleDepsDescription, expectedSimpleDepsReason, 2, 45, 2, 59),
			},
		},
		// ---- Real-user: facebook/react#34986 TS wrappers around inline callbacks must still report. ----
		{
			Code: `type Noop = () => void;
function Component() {
  const handleClick = useMemo(
    (() => {
      return 1;
    }) satisfies Noop,
    [],
  );
  return <button>{handleClick}</button>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedInlineFunction", expectedInlineFunctionReason, expectedInlineFunctionDescription, "Expected the first argument to be an inline function expression", 4, 5, 6, 22),
			},
		},
		// ---- Dimension 4: same-name imports remain inside upstream's input surface. ----
		{
			Code: `import {foo as useMemo} from 'not-react';
import React from 'not-react';
function Component() {
  const a = useMemo(async () => 1, []);
  const b = React.useMemo(async () => 2, []);
  return <div>{a}{b}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noAsyncOrGenerator", asyncGeneratorReason, asyncGeneratorDescription, "Async and generator functions are not supported", 4, 21, 4, 34),
				useMemoError("noAsyncOrGenerator", asyncGeneratorReason, asyncGeneratorDescription, "Async and generator functions are not supported", 5, 27, 5, 40),
			},
		},
		// ---- Options contract: arbitrary options do not change diagnostics. ----
		{
			Code: `function Component({value}) {
  const memo = useMemo(param => value, []);
  return <div>{memo}</div>;
}
`,
			Options: []interface{}{map[string]interface{}{"someFutureOption": true}},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noCallbackParameters", callbackParamsReason, callbackParamsDescription, "Callbacks with parameters are not supported", 2, 24, 2, 29),
			},
		},
		// ---- Dimension 4: assignment pattern reassigns a captured variable. ----
		{
			Code: `function Component({next}) {
  let x;
  const value = useMemo(() => {
    ({x} = next);
    return x;
  }, [next]);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noExternalReassignment", reassignOuterVariableReason, reassignOuterVariableDescription, "Cannot reassign variable", 4, 7, 4, 8),
			},
		},
		// ---- Dimension 4: nested array/object assignment patterns find every captured identifier. ----
		{
			Code: `function Component({next}) {
  let x;
  let y;
  const value = useMemo(() => {
    [x, {value: y}] = next;
    return x + y;
  }, [next]);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noExternalReassignment", reassignOuterVariableReason, reassignOuterVariableDescription, "Cannot reassign variable", 5, 6, 5, 7),
				useMemoError("noExternalReassignment", reassignOuterVariableReason, reassignOuterVariableDescription, "Cannot reassign variable", 5, 17, 5, 18),
			},
		},
		// ---- Dimension 4: update expression reassigns a captured variable. ----
		{
			Code: `function Component() {
  let x = 0;
  const value = useMemo(() => {
    x++;
    return x;
  }, []);
  return <div>{value}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noExternalReassignment", reassignOuterVariableReason, reassignOuterVariableDescription, "Cannot reassign variable", 4, 5, 4, 6),
			},
		},
		// ---- Real-user: facebook/react#16096 conditional dependency lists are not array literals. ----
		{
			Code: `function Component({position}) {
  const memoPosition = React.useMemo(
    () => position && gpsToLatLng(position),
    position ? [position.latitude, position.longitude] : [position],
  );
  return <div>{memoPosition}</div>;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedArrayLiteral", expectedArrayLiteralReason, expectedArrayLiteralDescription, expectedArrayLiteralReason, 4, 5, 4, 68),
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UseMemoRule, valid, invalid)
}
