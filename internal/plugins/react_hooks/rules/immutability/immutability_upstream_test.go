package immutability

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestImmutabilityUpstream migrates the react-hooks/immutability cases from
// upstream packages/eslint-plugin-react-hooks/__tests__/ReactCompilerRuleTypescript-test.ts
// and ReactCompilerRuleFlow-test.ts 1:1. Position assertions cover line/column
// for every invalid case. rslint-specific lock-in cases live in the
// immutability_extras_test.go file.

func upstreamImmutabilityError(info immutableInfo, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "immutableMutation",
		Message:   buildImmutabilityMessage(info).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestImmutabilityUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- ReactCompilerRuleTypescript: Basic example. ----
		{Code: `
function Button(props) {
  return null;
}
		`, FileName: "test.tsx", Tsx: true},
		// ---- ReactCompilerRuleTypescript: Repro for hooks as normal values. ----
		{Code: `
function Button(props) {
  const scrollView = React.useRef<ScrollView>(null);
  return <Button thing={scrollView} />;
}
		`, FileName: "test.tsx", Tsx: true},
		// ---- ReactCompilerRuleTypescript: [Heuristic] skips files with only lowercase utility functions. ----
		{Code: `
function helper(obj) {
  obj.key = 'value';
  return obj;
}
		`, FileName: "utils.ts"},
		// ---- ReactCompilerRuleTypescript: [Heuristic] skips lowercase arrow functions even with mutations. ----
		{Code: `
const processData = (input) => {
  input.modified = true;
  return input;
};
		`, FileName: "helpers.ts"},
		// ---- ReactCompilerRuleFlow: lowercase utility functions are skipped. ----
		{Code: `
function helper(obj) {
  obj.key = 'value';
  return obj;
}
		`, Skip: true}, // SKIP: rslint does not parse Flow-only React Compiler test files with hermes-eslint.
		// ---- ReactCompilerRuleFlow: lowercase arrow functions are skipped. ----
		{Code: `
const processData = (input) => {
  input.modified = true;
  return input;
};
		`, Skip: true}, // SKIP: rslint does not parse Flow-only React Compiler test files with hermes-eslint.
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- ReactCompilerRuleTypescript: Mutating useState value. ----
		{
			Code: `
import { useState } from 'react';
function Component(props) {
  const x: ` + "`foo${1}`" + ` = 'foo1';
  const [state, setState] = useState({a: 0});
  state.a = 1;
  return <div>{props.foo}</div>;
}
			`,
			FileName: "test.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamImmutabilityError(immutableInfo{kind: immutableUseState, hookName: "useState"}, 6, 3, 6, 8),
			},
		},
		// ---- ReactCompilerRuleTypescript: PascalCase function declaration detects prop mutation. ----
		{
			Code: `
function MyComponent({a}) {
  a.key = 'value';
  return <div />;
}
			`,
			FileName: "component.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamImmutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 4),
			},
		},
		// ---- ReactCompilerRuleTypescript: PascalCase arrow function detects prop mutation. ----
		{
			Code: `
const MyComponent = ({a}) => {
  a.key = 'value';
  return <div />;
};
			`,
			FileName: "component.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamImmutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 4),
			},
		},
		// ---- ReactCompilerRuleTypescript: PascalCase function expression detects prop mutation. ----
		{
			Code: `
const MyComponent = function({a}) {
  a.key = 'value';
  return <div />;
};
			`,
			FileName: "component.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamImmutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 4),
			},
		},
		// ---- ReactCompilerRuleTypescript: exported function declaration detects prop mutation. ----
		{
			Code: `
export function MyComponent({a}) {
  a.key = 'value';
  return <div />;
}
			`,
			FileName: "component.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamImmutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 4),
			},
		},
		// ---- ReactCompilerRuleTypescript: exported arrow function detects prop mutation. ----
		{
			Code: `
export const MyComponent = ({a}) => {
  a.key = 'value';
  return <div />;
};
			`,
			FileName: "component.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamImmutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 4),
			},
		},
		// ---- ReactCompilerRuleTypescript: default exported function detects prop mutation. ----
		{
			Code: `
export default function MyComponent({a}) {
  a.key = 'value';
  return <div />;
}
			`,
			FileName: "component.tsx",
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamImmutabilityError(immutableInfo{kind: immutablePropsOrHookArgs}, 3, 3, 3, 4),
			},
		},
		// ---- ReactCompilerRuleFlow: Flow component declaration detects prop mutation. ----
		{
			Code: `
component MyComponent(a: {key: string}) {
  a.key = 'value';
  return <div />;
}
			`,
			Skip: true, // SKIP: rslint does not parse Flow component declarations.
		},
		// ---- ReactCompilerRuleFlow: exported Flow component declaration detects prop mutation. ----
		{
			Code: `
export component MyComponent(a: {key: string}) {
  a.key = 'value';
  return <div />;
}
			`,
			Skip: true, // SKIP: rslint does not parse Flow component declarations.
		},
		// ---- ReactCompilerRuleFlow: default exported Flow component declaration detects prop mutation. ----
		{
			Code: `
export default component MyComponent(a: {key: string}) {
  a.key = 'value';
  return <div />;
}
			`,
			Skip: true, // SKIP: rslint does not parse Flow component declarations.
		},
		// ---- ReactCompilerRuleFlow: Flow hook declaration detects argument mutation. ----
		{
			Code: `
hook useMyHook(a: {key: string}) {
  a.key = 'value';
  return a;
}
			`,
			Skip: true, // SKIP: rslint does not parse Flow hook declarations.
		},
		// ---- ReactCompilerRuleFlow: exported Flow hook declaration detects argument mutation. ----
		{
			Code: `
export hook useMyHook(a: {key: string}) {
  a.key = 'value';
  return a;
}
			`,
			Skip: true, // SKIP: rslint does not parse Flow hook declarations.
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ImmutabilityRule, valid, invalid)
}
