package preserve_manual_memoization

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreserveManualMemoizationUpstream migrates the public docs and
// representative React Compiler preserve-memo fixtures for
// react-hooks/preserve-manual-memoization. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in the
// preserve_manual_memoization_extras_test.go file.

func preserveManualMemoizationDependencyError(dep string, sourceDeps []string, result compareDependencyResult, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	source := make([]memoDependency, 0, len(sourceDeps))
	for _, sourceDep := range sourceDeps {
		source = append(source, memoDependency{root: sourceDep})
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: "preserveManualMemoization",
		Message: buildPreservedManualMemoizationDependencyMessage(
			memoDependency{root: dep},
			source,
			result,
		).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func preserveManualMemoizationMessage(dep memoDependency, sourceDeps []memoDependency, result compareDependencyResult, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "preserveManualMemoization",
		Message:   buildPreservedManualMemoizationDependencyMessage(dep, sourceDeps, result).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestPreserveManualMemoizationUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- react.dev docs: complete useMemo dependencies. ----
		{Code: `
function Component({ data, filter }) {
  const filtered = useMemo(
    () => data.filter(filter),
    [data, filter]
  );
  return <List items={filtered} />;
}
		`, Tsx: true},
		// ---- react.dev docs: complete useCallback dependencies. ----
		{Code: `
function Component({ onUpdate, value }) {
  const handleClick = useCallback(() => {
    onUpdate(value);
  }, [onUpdate, value]);
  return <button onClick={handleClick} />;
}
		`, Tsx: true},
		// ---- react.dev docs: no manual memoization needed. ----
		{Code: `
function Component({ data, filter }) {
  const filtered = data.filter(filter);
  return <List items={filtered} />;
}
		`, Tsx: true},
		// ---- React Compiler fixture: preserve-use-callback-stable-built-ins.ts. ----
		{Code: `
function Component({ items }) {
  const onClick = useCallback(() => {
    console.log(items.length, Math.max(1, items.length));
  }, [items]);
  return <button onClick={onClick} />;
}
		`, Tsx: true},
		// ---- React Compiler fixture: preserve-use-memo-ref-missing-ok.ts. ----
		{Code: `
function Component() {
  const ref = useRef(null);
  const value = useMemo(() => ref.current, []);
  return <div>{value}</div>;
}
		`, Tsx: true},
		// ---- React Compiler fixture: preserve-use-memo-transition.ts. ----
		{Code: `
function Component() {
  const [pending, startTransition] = useTransition();
  const onClick = useCallback(() => {
    startTransition(() => {});
  }, []);
  return <button disabled={pending} onClick={onClick} />;
}
		`, Tsx: true},
		// ---- React Compiler fixture: useMemo-alias-property-load-dep.ts. ----
		{Code: `
import {useMemo} from 'react';
import {sum} from 'shared-runtime';

function Component({propA, propB}) {
  const x = propB.x.y;
  return useMemo(() => {
    return sum(propA.x, x);
  }, [propA.x, x]);
}
		`, Tsx: true},
		// ---- React Compiler fixture: useCallback-alias-property-load-dep.ts. ----
		{Code: `
import {useCallback} from 'react';
import {sum} from 'shared-runtime';

function Component({propA, propB}) {
  const x = propB.x.y;
  return useCallback(() => {
    return sum(propA.x, x);
  }, [propA.x, x]);
}
		`, Tsx: true},
		// ---- React Compiler fixture: useMemo-infer-more-specific.ts. ----
		{Code: `
import {useMemo} from 'react';

function useHook(x) {
  return useMemo(() => [x.y.z], [x]);
}
		`},
		// ---- React Compiler fixture: useCallback-infer-more-specific.ts. ----
		{Code: `
import {useCallback} from 'react';

function useHook(x) {
  return useCallback(() => [x.y.z], [x]);
}
		`},
		// ---- React Compiler fixture: useMemo-dep-array-literal-access.ts. ----
		{Code: `
import {useMemo} from 'react';

function Foo(props) {
  const x = makeArray(props);
  return useMemo(() => [x[0]], [x[0]]);
}
		`},
		// ---- React Compiler fixture: useMemo-inner-decl.ts. ----
		{Code: `
import {useMemo} from 'react';
import {identity} from 'shared-runtime';

function useFoo(data) {
  return useMemo(() => {
    const temp = identity(data.a);
    return {temp};
  }, [data.a]);
}
		`},
		// ---- React Compiler ESLint options schema: first option object is accepted. ----
		{Code: `
function Component({ data, filter }) {
  const filtered = useMemo(() => data.filter(filter), [data, filter]);
  return <List items={filtered} />;
}
		`, Tsx: true, Options: map[string]interface{}{"compilationMode": "annotation", "panicThreshold": "none"}},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- react.dev docs: missing useMemo dependency. ----
		{
			Code: `
function Component({ data, filter }) {
  const filtered = useMemo(
    () => data.filter(filter),
    [data]
  );
  return <List items={filtered} />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationDependencyError("filter", []string{"data"}, compareDependencyRootDifference, 4, 5, 4, 30),
			},
		},
		// ---- react.dev docs: missing useCallback dependency. ----
		{
			Code: `
function Component({ onUpdate, value }) {
  const handleClick = useCallback(() => {
    onUpdate(value);
  }, [onUpdate]);
  return <button onClick={handleClick} />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationDependencyError("value", []string{"onUpdate"}, compareDependencyRootDifference, 3, 35, 5, 4),
			},
		},
		// ---- React Compiler fixture: error.useMemo-property-call-dep.ts. ----
		{
			Code: `
import {useMemo} from 'react';

function Component({propA}) {
  return useMemo(() => {
    return propA.x();
  }, [propA.x]);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "propA"},
					[]memoDependency{{root: "propA", path: []dependencyPathPart{{property: "x"}}}},
					compareDependencySubpath,
					5, 18, 7, 4,
				),
			},
		},
		// ---- React Compiler fixture: error.useMemo-property-call-chained-object.ts. ----
		{
			Code: `
import {useMemo} from 'react';

function Component({propA}) {
  return useMemo(() => {
    return {
      value: propA.x().y,
    };
  }, [propA.x]);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "propA"},
					[]memoDependency{{root: "propA", path: []dependencyPathPart{{property: "x"}}}},
					compareDependencySubpath,
					5, 18, 9, 4,
				),
			},
		},
		// ---- React Compiler fixture: error.useCallback-property-call-dep.ts. ----
		{
			Code: `
import {useCallback} from 'react';

function Component({propA}) {
  return useCallback(() => {
    return propA.x();
  }, [propA.x]);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "propA"},
					[]memoDependency{{root: "propA", path: []dependencyPathPart{{property: "x"}}}},
					compareDependencySubpath,
					5, 22, 7, 4,
				),
			},
		},
		// ---- React Compiler fixture: error.preserve-use-memo-ref-missing-reactive.ts. ----
		{
			Code: `
function Component({ cond, ref1, ref2 }) {
  const ref = cond ? ref1 : ref2;
  const cb = useCallback(() => {
    ref.current();
  }, []);
  return <button onClick={cb} />;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "ref"},
					nil,
					compareDependencyRootDifference,
					4, 26, 6, 4,
				),
			},
		},
		// ---- React Compiler fixture: error.useMemo-aliased-var.ts. ----
		{
			Code: `
function useHook(x) {
  const aliasedX = x;
  const aliasedProp = x.y.z;

  return useMemo(() => [x, x.y.z], [aliasedX, aliasedProp]);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "x"},
					[]memoDependency{{root: "aliasedX"}, {root: "aliasedProp"}},
					compareDependencyRootDifference,
					6, 18, 6, 34,
				),
			},
		},
		// ---- React Compiler fixture: error.useCallback-aliased-var.ts. ----
		{
			Code: `
function useHook(x) {
  const aliasedX = x;
  const aliasedProp = x.y.z;

  return useCallback(() => [aliasedX, x.y.z], [x, aliasedProp]);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "aliasedX"},
					[]memoDependency{{root: "x"}, {root: "aliasedProp"}},
					compareDependencyRootDifference,
					6, 22, 6, 45,
				),
			},
		},
		// ---- React Compiler fixture: error.useCallback-conditional-access-noAlloc.ts. ----
		{
			Code: `
import {useCallback} from 'react';

function Component({propA, propB}) {
  return useCallback(() => {
    return {
      value: propB?.x.y,
      other: propA,
    };
  }, [propA, propB.x.y]);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "propB", path: []dependencyPathPart{{property: "x", optional: true}, {property: "y"}}},
					[]memoDependency{{root: "propA"}, {root: "propB", path: []dependencyPathPart{{property: "x"}, {property: "y"}}}},
					compareDependencyPathDifference,
					5, 22, 10, 4,
				),
			},
		},
		// SKIP: React Compiler fixture error.false-positive-useMemo-dropped-infer-always-invalidating.ts
		// depends on the compiler's pruned reactive-scope output, which this syntax-level rule cannot observe.
		{Skip: true, Code: `
function useFoo(props) {
  const x = [];
  useHook();
  x.push(props);
  return useMemo(() => [x], [x]);
}
		`},
		// SKIP: React Compiler fixture error.false-positive-useMemo-infer-mutate-deps.ts
		// reports a conservative HIR mutable-range diagnostic without a later source write.
		{Skip: true, Code: `
function useFoo() {
  const val = [1, 2, 3];
  return useMemo(() => identity(val), [val]);
}
		`},
		// SKIP: React Compiler fixture error.useMemo-infer-less-specific-conditional-access.ts
		// relies on conditional reactive-scope hoisting rather than a syntax-visible dependency mismatch.
		{Skip: true, Code: `
function Component({propA, propB}) {
  return useMemo(() => {
    const x = {};
    if (propA?.a) {
      mutate(x);
      return {value: propB.x.y};
    }
  }, [propA?.a, propB.x.y]);
}
		`},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreserveManualMemoizationRule, valid, invalid)
}
