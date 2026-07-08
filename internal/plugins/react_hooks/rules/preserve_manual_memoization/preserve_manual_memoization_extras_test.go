package preserve_manual_memoization

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreserveManualMemoizationExtras locks in branches and edge shapes that
// the upstream fixture suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / issue shape it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.

func preserveManualMemoizationMutationError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "preserveManualMemoization",
		Message:   buildPreservedManualMemoizationMutationMessage().Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func TestPreserveManualMemoizationExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized receiver and callback body are transparent. ----
		{Code: `
function Component({ data, filter }) {
  const filtered = (useMemo)((() => ((data)).filter(filter)), [data, filter]);
  return <List items={filtered} />;
}
		`, Tsx: true},
		// ---- Dimension 4: TS assertion wrappers in dependency expressions are transparent. ----
		{Code: `
type Props = { data: string[], filter: (value: string) => boolean };
function Component({ data, filter }: Props) {
  const filtered = useMemo(() => (data as string[]).filter(filter), [data as string[], filter satisfies unknown]);
  return <List items={filtered} />;
}
		`, Tsx: true},
		// ---- Dimension 4: static element access dependencies are equivalent to property dependencies. ----
		{Code: `
function Component({ data }) {
  const label = useMemo(() => data['label'], [data.label]);
  return <span>{label}</span>;
}
		`, Tsx: true},
		// ---- Dimension 4: dynamic element access collects the dynamic key. ----
		{Code: `
function Component({ data, keyName }) {
  const label = useMemo(() => data[keyName], [data, keyName]);
  return <span>{label}</span>;
}
		`, Tsx: true},
		// ---- Dimension 4: nested callbacks are traversal boundaries. ----
		{Code: `
function Component({ value }) {
  const fn = useCallback(() => {
    return () => value;
  }, []);
  return <button onClick={fn} />;
}
		`, Tsx: true},
		// ---- Dimension 4: state setters are stable and do not need source deps. ----
		{Code: `
function Component() {
  const [count, setCount] = useState(0);
  const onClick = useCallback(() => {
    setCount(count => count + 1);
  }, []);
  return <button onClick={onClick}>{count}</button>;
}
		`, Tsx: true},
		// ---- Dimension 4: React import aliases and namespace imports are recognized. ----
		{Code: `
import R, { useCallback as callback } from 'react';
function Component({ value }) {
  const first = callback(() => value, [value]);
  const second = R.useMemo(() => value + 1, [value]);
  return <button onClick={first}>{second}</button>;
}
		`, Tsx: true},
		// ---- Dimension 4: class methods are not Compiler render functions for this rule. ----
		{Code: `
class Component {
  render(data, filter) {
    return useMemo(() => data.filter(filter), [data]);
  }
}
		`, Tsx: true},
		// ---- Dimension 4: nested function bodies after the memo call are traversal boundaries. ----
		{Code: `
function Component({ items }) {
  const value = useMemo(() => items.length, [items]);
  function resetLater() {
    items.push(1);
  }
  return <span onClick={resetLater}>{value}</span>;
}
		`, Tsx: true},
		// ---- Dimension 4: expression-bodied callbacks that return functions do not inspect the returned function body. ----
		{Code: `
function Component({ value }) {
  const makeHandler = useMemo(() => () => value, []);
  return <button onClick={makeHandler()} />;
}
		`, Tsx: true},
		// ---- Dimension 4: empty dependency arrays degrade gracefully. ----
		{Code: `
function Component() {
  const label = useMemo(() => 'static', []);
  return <span>{label}</span>;
}
		`, Tsx: true},
		// ---- Dimension 4: unsupported dynamic dependency arrays are left to use-memo/exhaustive-deps. ----
		{Code: `
function Component({ data, deps }) {
  const label = useMemo(() => data.label, deps);
  return <span>{label}</span>;
}
		`, Tsx: true},
		// ---- Real-user: facebook/react#34957 state setter in useCallback remains stable. ----
		{Code: `
function Example({ things }) {
  const thing = things.find(thing => thing.something > 0);
  const [count, setCount] = React.useState(0);
  const label = thing.label.toLowerCase();
  const handleClick = React.useCallback(() => {
    setCount(count => count + 1);
  }, []);
  return <button onClick={handleClick}>{label}: {count}</button>;
}
		`, Tsx: true},
		// ---- Real-user: facebook/react#36384 state setter does not need a manual dep. ----
		{Code: `
function Gallery({ images }) {
  const [selected, setImages] = useState(images);
  const update = useCallback(() => {
    setImages(prev => prev.slice(1));
  }, []);
  return <button onClick={update}>{selected.length}</button>;
}
		`, Tsx: true},
		// N/A: private property names, class fields, overload signatures, and
		// React.memo comparison functions are not useMemo/useCallback dependency
		// arrays for this syntax-level preservation check.
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Dimension 4: optional chain source deps must preserve optional precision. ----
		{
			Code: `
function Component({ data }) {
  const label = useMemo(() => data?.label, [data.label]);
  return <span>{label}</span>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "data", path: []dependencyPathPart{{property: "label", optional: true}}},
					[]memoDependency{{root: "data", path: []dependencyPathPart{{property: "label"}}}},
					compareDependencyPathDifference,
					3, 25, 3, 42,
				),
			},
		},
		// ---- Dimension 4: ref.current access requires exact source precision for non-stable refs. ----
		{
			Code: `
function Component({ ref }) {
  const value = useMemo(() => ref.current, [ref]);
  return <span>{value}</span>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "ref", path: []dependencyPathPart{{property: "current"}}},
					[]memoDependency{{root: "ref"}},
					compareDependencyRefAccessDifference,
					3, 25, 3, 42,
				),
			},
		},
		// Locks in upstream compareDeps() arm: source path that is more specific than inferred is invalid.
		{
			Code: `
function Component({ data }) {
  const value = useMemo(() => data, [data.label]);
  return <span>{value.label}</span>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMessage(
					memoDependency{root: "data"},
					[]memoDependency{{root: "data", path: []dependencyPathPart{{property: "label"}}}},
					compareDependencySubpath,
					3, 25, 3, 35,
				),
			},
		},
		// Locks in upstream StartMemoize operand arm: source dependency mutated later.
		{
			Code: `
function Component({ items }) {
  const value = useMemo(() => items.length, [items]);
  items.push(1);
  return <span>{value}</span>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMutationError(3, 46, 3, 51),
			},
		},
		// ---- Dimension 4: destructuring writes are later mutations of source dependencies. ----
		{
			Code: `
function Component({ items }) {
  const value = useMemo(() => items.length, [items]);
  [items] = [[]];
  return <span>{value}</span>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMutationError(3, 46, 3, 51),
			},
		},
		// ---- Dimension 4: property writes through TS wrappers are later mutations. ----
		{
			Code: `
function Component({ item }) {
  const value = useMemo(() => item.label, [item]);
  (item as any).label = 'next';
  return <span>{value}</span>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationMutationError(3, 44, 3, 48),
			},
		},
		// Locks in upstream collectMaybeMemoDependencies() arm: dynamic element key is inferred.
		{
			Code: `
function Component({ data, keyName }) {
  const label = useMemo(() => data[keyName], [data]);
  return <span>{label}</span>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationDependencyError("keyName", []string{"data"}, compareDependencyRootDifference, 3, 25, 3, 44),
			},
		},
		// ---- Dimension 2: member-call receivers that are calls still traverse receiver arguments. ----
		{
			Code: `
function Component({ makeObj, value }) {
  const result = useMemo(() => makeObj(value).method(), [makeObj]);
  return <span>{result}</span>;
}
			`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				preserveManualMemoizationDependencyError("value", []string{"makeObj"}, compareDependencyRootDifference, 3, 26, 3, 55),
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreserveManualMemoizationRule, valid, invalid)
}
