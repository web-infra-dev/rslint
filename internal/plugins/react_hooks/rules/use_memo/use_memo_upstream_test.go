package use_memo

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestUseMemoUpstream migrates the react-hooks/use-memo cases from upstream
// compiler/packages/babel-plugin-react-compiler/src/Validation/ValidateUseMemo.ts
// and related React Compiler fixtures 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// use_memo_extras_test.go.
func TestUseMemoUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Upstream fixture: useMemo-simple.js. ----
		{Code: `function component(a) {
  let x = useMemo(() => [a], [a]);
  return <Foo x={x}></Foo>;
}
`, Tsx: true},
		// ---- Upstream fixture: aliased useMemo.js. ----
		{Code: `import {useMemo as myMemo} from 'react';

function Component({x}) {
  const v = myMemo(() => x * 2, [x]);
  return <div>{v}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-with-optional.js, missing deps list is allowed by use-memo. ----
		{Code: `import {useMemo} from 'react';
function Component(props) {
  return (
    useMemo(() => {
      return [props.value];
    }) || []
  );
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-arrow-implicit-return.js. ----
		{Code: `function Component() {
  const value = useMemo(() => computeValue(), []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-with-no-deps-list.ts, missing deps list is allowed by use-memo. ----
		{Code: `function Component(props) {
  const value = useMemo(() => props.value);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-empty-return.js belongs to void-use-memo, not use-memo. ----
		{Code: `function Component() {
  const value = useMemo(() => {
    return;
  }, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-explicit-null-return.js is valid for use-memo. ----
		{Code: `function Component() {
  const value = useMemo(() => {
    return null;
  }, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-multiple-returns.js is valid for use-memo. ----
		{Code: `function Component({cond, a, b}) {
  const value = useMemo(() => {
    if (cond) {
      return a;
    }
    return b;
  }, [cond, a, b]);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-switch-return.js is valid for use-memo. ----
		{Code: `function Component({kind, a, b}) {
  const value = useMemo(() => {
    switch (kind) {
      case 'a':
        return a;
      default:
        return b;
    }
  }, [kind, a, b]);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-named-function.ts inline named function expressions are valid. ----
		{Code: `function Component({a}) {
  const value = useMemo(function named() {
    return a;
  }, [a]);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Official react.dev use-memo docs: no-return callbacks belong to void-use-memo, not use-memo. ----
		{Code: `function Component({ data }) {
  const processed = useMemo(() => {
    data.forEach(item => console.log(item));
  }, [data]);
  return <div>{processed}</div>;
}
`, Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Upstream fixture: error.invalid-useMemo-callback-args.js. ----
		{
			Code: `function component(a, b) {
  let x = useMemo(c => a, []);
  return x;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noCallbackParameters", callbackParamsReason, callbackParamsDescription, "Callbacks with parameters are not supported", 2, 19, 2, 20),
			},
		},
		// ---- Upstream fixture: error.invalid-useMemo-async-callback.js. ----
		{
			Code: `function component(a, b) {
  let x = useMemo(async () => {
    await a;
  }, []);
  return x;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noAsyncOrGenerator", asyncGeneratorReason, asyncGeneratorDescription, "Async and generator functions are not supported", 2, 19, 4, 4),
			},
		},
		// ---- Upstream fixture: error.invalid-ReactUseMemo-async-callback.js. ----
		{
			Code: `function component(a, b) {
  let x = React.useMemo(async () => {
    await a;
  }, []);
  return x;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noAsyncOrGenerator", asyncGeneratorReason, asyncGeneratorDescription, "Async and generator functions are not supported", 2, 25, 4, 4),
			},
		},
		// ---- Upstream fixture: error.useMemo-callback-generator.js. ----
		{
			Code: `function component(a, b) {
  let x = useMemo(function* () {
    yield a;
  }, []);
  return x;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noAsyncOrGenerator", asyncGeneratorReason, asyncGeneratorDescription, "Async and generator functions are not supported", 2, 19, 4, 4),
			},
		},
		// ---- Upstream fixture: error.invalid-reassign-variable-in-useMemo.js. ----
		{
			Code: `function Component() {
  let x;
  const y = useMemo(() => {
    let z;
    x = [];
    z = true;
    return z;
  }, []);
  return [x, y];
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("noExternalReassignment", reassignOuterVariableReason, reassignOuterVariableDescription, "Cannot reassign variable", 5, 5, 5, 6),
			},
		},
		// ---- Upstream fixture: error.useMemo-non-literal-deps-list.ts. ----
		{
			Code: `import {useMemo} from 'react';

function App({text, hasDeps}) {
  const resolvedText = useMemo(
    () => {
      return text.toUpperCase();
    },
    hasDeps ? null : [text],
  );
  return resolvedText;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedArrayLiteral", expectedArrayLiteralReason, expectedArrayLiteralDescription, expectedArrayLiteralReason, 8, 5, 8, 28),
			},
		},
		// ---- Upstream fixture: preserve-memo-validation/error.validate-useMemo-named-function.js. ----
		{
			Code: `function Component(props) {
  const x = useMemo(someHelper, []);
  return x;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				useMemoError("expectedInlineFunction", expectedInlineFunctionReason, expectedInlineFunctionDescription, "Expected the first argument to be an inline function expression", 2, 21, 2, 31),
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UseMemoRule, valid, invalid)
}

func useMemoError(id, reason, description, detail string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: id,
		Message:   buildUseMemoMessage(id, reason, description, detail).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}
