package globals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestGlobalsUpstream migrates the react-hooks/globals cases from upstream
// React Compiler fixtures and official ESLint behavior 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in globals_extras_test.go.

func upstreamGlobalsError(name string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "globalReassignment",
		Message:   buildGlobalReassignmentMessage(name).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestGlobalsUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Upstream fixture: allow-reassignment-to-global-function-jsx-prop.js. ----
		{Code: `
function Component() {
  const onClick = () => {
    someUnknownGlobal = true;
    moduleLocal = true;
  };
  return <div onClick={onClick} />;
}
		`, Tsx: true},
		// ---- Upstream fixture: globals-Boolean.js. ----
		{Code: `
function Component(props) {
  const x = {};
  const y = Boolean(x);
  return [x, y];
}
		`, Tsx: true},
		// ---- Upstream fixture: globals-Number.js. ----
		{Code: `
function Component(props) {
  const x = {};
  const y = Number(x);
  return [x, y];
}
		`, Tsx: true},
		// ---- Upstream fixture: globals-String.js. ----
		{Code: `
function Component(props) {
  const x = {};
  const y = String(x);
  return [x, y];
}
		`, Tsx: true},
		// ---- Upstream fixture: globals-dont-resolve-local-useState.js. ----
		{Code: `
import {useState as _useState, useCallback, useEffect} from 'react';

function useState(value) {
  const [state, setState] = _useState(value);
  return [state, setState];
}

function Component() {
  const [state, setState] = useState('hello');

  return <div onClick={() => setState('goodbye')}>{state}</div>;
}
		`, Tsx: true},
		// ---- Official ESLint: PascalCase without JSX or hook call is skipped. ----
		{Code: `
function Component() {
  someUnknownGlobal = true;
  return null;
}
		`, Tsx: true},
		// ---- Official ESLint: local bindings can be reassigned. ----
		{Code: `
function Component() {
  let moduleLocal;
  moduleLocal = true;
  return <div />;
}
		`, Tsx: true},
		// ---- Official ESLint: parameters are local to the component. ----
		{Code: `
function Component(moduleLocal) {
  moduleLocal = true;
  return <div />;
}
		`, Tsx: true},
		// ---- Official ESLint: property writes are handled by react-hooks/immutability, not globals. ----
		{Code: `
function Component() {
  window.location.href = "/";
  return <div />;
}
		`, Tsx: true},
		// ---- Official ESLint: event callbacks do not run during render. ----
		{Code: `
function Component() {
  const onClick = () => {
    someGlobal = true;
  };
  return <button onClick={onClick} />;
}
		`, Tsx: true},
		// ---- Official ESLint: effect callbacks do not run during render. ----
		{Code: `
function Component() {
  useEffect(() => {
    someGlobal = true;
  });
  return <div />;
}
		`, Tsx: true},
		// ---- Official ESLint: useCallback callbacks do not run during render. ----
		{Code: `
function Component() {
  useCallback(() => {
    someGlobal = true;
  });
  return <div />;
}
		`, Tsx: true},
		// ---- Official ESLint: update expressions are not reported by globals. ----
		{Code: `
function Component() {
  someGlobal++;
  ++otherGlobal;
  return <div />;
}
		`, Tsx: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Upstream fixture: error.reassignment-to-global.js, adapted with JSX render marker. ----
		{
			Code: `
function Component() {
  someUnknownGlobal = true;
  moduleLocal = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamGlobalsError("someUnknownGlobal", 3, 3, 3, 20),
				upstreamGlobalsError("moduleLocal", 4, 3, 4, 14),
			},
		},
		// ---- Upstream fixture: new-mutability/error.reassignment-to-global.js, adapted with JSX render marker. ----
		{
			Code: `
// @enableNewMutationAliasingModel
function Component() {
  someUnknownGlobal = true;
  moduleLocal = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamGlobalsError("someUnknownGlobal", 4, 3, 4, 20),
				upstreamGlobalsError("moduleLocal", 5, 3, 5, 14),
			},
		},
		// ---- Official ESLint: top-level module bindings are still outside the component. ----
		{
			Code: `
let moduleLocal;
function Component() {
  moduleLocal = true;
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamGlobalsError("moduleLocal", 4, 3, 4, 14),
			},
		},
		// ---- Upstream fixture: error.reassignment-to-global-indirect.js, adapted with JSX render marker. ----
		{
			Code: `
function Component() {
  const foo = () => {
    someUnknownGlobal = true;
    moduleLocal = true;
  };
  foo();
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamGlobalsError("someUnknownGlobal", 4, 5, 4, 22),
				upstreamGlobalsError("moduleLocal", 5, 5, 5, 16),
			},
		},
		// ---- Upstream fixture: new-mutability/error.reassignment-to-global-indirect.js, adapted with JSX render marker. ----
		{
			Code: `
// @enableNewMutationAliasingModel
function Component() {
  const foo = () => {
    someUnknownGlobal = true;
    moduleLocal = true;
  };
  foo();
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamGlobalsError("someUnknownGlobal", 5, 5, 5, 22),
				upstreamGlobalsError("moduleLocal", 6, 5, 6, 16),
			},
		},
		// ---- Upstream fixture: error.assign-global-in-jsx-children.js. ----
		{
			Code: `
function Component() {
  const foo = () => {
    someGlobal = true;
  };
  return <Foo>{foo}</Foo>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamGlobalsError("someGlobal", 4, 5, 4, 15),
			},
		},
		// ---- Official ESLint: useMemo callbacks run during render. ----
		{
			Code: `
function Component() {
  useMemo(() => {
    someGlobal = true;
    return 1;
  });
  return <div />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamGlobalsError("someGlobal", 4, 5, 4, 15),
			},
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &GlobalsRule,
		valid,
		invalid,
	)
}
