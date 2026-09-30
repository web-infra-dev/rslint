// TestPreferExactPropsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case identifies the source branch,
// universal edge shape, or real-user report that it covers.
package prefer_exact_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferExactPropsExtras(t *testing.T) {
	objectExactSettings := map[string]interface{}{
		"propWrapperFunctions": []interface{}{
			map[string]interface{}{"object": "Object", "property": "freeze", "exact": true},
		},
	}
	plainWrapperSettings := map[string]interface{}{
		"propWrapperFunctions": []interface{}{
			"exact",
			map[string]interface{}{"property": "other", "exact": false},
		},
	}
	errorAtStart := []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 1, Column: 1}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferExactPropsRule, []rule_tester.ValidTestCase{
		// Locks in getExactPropWrappers: string entries and exact:false entries are not exact wrappers.
		{Code: `Component.propTypes = { foo: PropTypes.string };`, Settings: plainWrapperSettings, Tsx: true},
		// Locks in ClassProperty branch 1: an empty object is already exact enough.
		{Code: `class Component { static propTypes = {}; }`, Settings: exactSettings, Tsx: true},
		// Locks in MemberExpression identifier branch: an empty initializer is allowed.
		{Code: `const props = {}; Component.propTypes = props;`, Settings: exactSettings, Tsx: true},
		// Locks in MemberExpression identifier fallback: unresolved and imported values are opaque.
		{Code: `import props from './props'; A.propTypes = props; B.propTypes = unresolved;`, Settings: exactSettings, Tsx: true},
		// Locks in ClassProperty branch 3: identifier initializers are not resolved in this listener.
		{Code: `const props = { foo: P }; class Component { static propTypes = props; }`, Settings: exactSettings, Tsx: true},
		// Locks in isExactPropWrapperFunction's property-only match.
		{Code: `Component.propTypes = exact({ foo: PropTypes.string });`, Settings: exactSettings, Tsx: true},
		// Locks in isExactPropWrapperFunction's object-and-property match and property fallback.
		{Code: `A.propTypes = Object.freeze({ foo: P }); B.propTypes = freeze({ foo: P });`, Settings: objectExactSettings, Tsx: true},
		// ---- Dimension 4: parentheses around exact wrapper callees are ESTree-transparent ----
		{Code: `Component.propTypes = ((exact))({ foo: P });`, Settings: exactSettings, Tsx: true},
		// ---- Dimension 4: optional exact calls retain their direct identifier callee ----
		{Code: `Component.propTypes = exact?.({ foo: P });`, Settings: exactSettings, Tsx: true},
		// ---- Dimension 4: optional non-exact calls are ChainExpressions upstream, not calls ----
		{Code: `A.propTypes = other?.({ foo: P }); B.propTypes = Object?.freeze({ bar: P });`, Settings: objectExactSettings, Tsx: true},
		// ---- Dimension 4: optional member reads degrade gracefully instead of following a declaration path ----
		{Code: `Component?.propTypes;`, Settings: exactSettings, Tsx: true},
		// ---- Intentional divergence: non-assignment uses never define propTypes ----
		{Code: `const fallback = Component.propTypes || { foo: P }; Component.propTypes + { bar: P }; consume(Component.propTypes); delete Component.propTypes; condition ? Component.propTypes : fallback; Component.propTypes.foo;`, Settings: exactSettings, Tsx: true},
		// ---- Intentional divergence: for-in/of targets do not define propTypes from the iterable ----
		{Code: `for (Component.propTypes in { foo: P }) {} for (Other.propTypes of other()) {}`, Settings: exactSettings, Tsx: true},
		// ---- Dimension 4: dynamic and non-string element keys remain unknown ----
		{Code: `const propTypes = 'notPropTypes'; const key = 'propTypes'; A[propTypes] = { foo: P }; B[key] = { foo: P }; C[0] = { foo: P };`, Settings: exactSettings, Tsx: true},
		// ---- Intentional divergence: only public runtime static propTypes fields are checked ----
		{Code: `class Component { propTypes = { a: P }; static #propTypes = { b: P }; static [propTypes] = { c: P }; props: { d: string } = { d: 'x' }; } const Other = class { propTypes = { e: P }; };`, Settings: exactSettings, Tsx: true},
		// ---- Dimension 4: authored TS wrappers around the value remain visible upstream ----
		{Code: `A.propTypes = ({ foo: P } as object); B.propTypes = ({ foo: P })!; C.propTypes = ({ foo: P } satisfies object);`, Settings: exactSettings, Tsx: true},
		// ---- Dimension 4: empty declarations and body-absent members are ignored ----
		{Code: `declare class Component { static propTypes: object; } class Empty {}`, Settings: exactSettings, Tsx: true},
		// ---- Dimension 4: TypeScript props types are not Flow ObjectTypeAnnotation nodes ----
		{Code: `type Props = { foo: string }; function Component(props: Props) { return <div />; } class C { props: { foo: string }; }`, Settings: exactSettings, Tsx: true},
		// ---- Dimension 4: the closest lexical definition wins for an identifier RHS ----
		{Code: `const props = { foo: P }; function assign() { const props = {}; Component.propTypes = props; }`, Settings: exactSettings, Tsx: true},
		// ---- Intentional divergence: mutable bindings are unknown instead of using a stale initializer ----
		{Code: `let before = {}; before = { foo: P }; A.propTypes = before; let after = { bar: P }; after = exact(after); B.propTypes = after;`, Settings: exactSettings, Tsx: true},
		// ---- Intentional divergence: mutations and escapes make const-backed objects unknown ----
		{Code: `const added = {}; added.foo = P; A.propTypes = added; const removed = { foo: P }; delete removed.foo; B.propTypes = removed; const escaped = { bar: P }; consume(escaped); C.propTypes = escaped;`, Settings: exactSettings, Tsx: true},
		// ---- Intentional divergence: lexical resolution cannot select an inaccessible child binding ----
		{Code: `/* global props */ function hidden() { const props = { foo: P }; } Component.propTypes = props;`, Settings: exactSettings, Tsx: true},
		// ---- Real-user: PR #3190 local type annotations inside a component are not component props ----
		{Code: `function Component(props: {}) { const local: { foo: string } = { foo: 'x' }; return <div />; }`, Settings: exactSettings, Tsx: true},
		// N/A Dimension 3: the upstream rule provides neither autofixes nor suggestions.
		// N/A Dimension 4 function variants: runtime propTypes checks target declarations, not function kinds.
		// N/A Dimension 4 ancestor walks: the runtime path does not walk component ancestors.
		// N/A Dimension 4 RestElement: binding patterns are outside the runtime propTypes value shapes.
	}, []rule_tester.InvalidTestCase{
		// Locks in ClassProperty branch 2: every call not matching an exact wrapper reports.
		{Code: `class Component { static propTypes = other(); }`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 1, Column: 19}}},
		// Locks in MemberExpression branch 1: any non-empty object reports, including spread-only objects.
		{Code: `Component.propTypes = { ...shared };`, Settings: exactSettings, Tsx: true, Errors: errorAtStart},
		// Locks in MemberExpression branch 2: a non-exact zero-argument call still reports.
		{Code: `Component.propTypes = other();`, Settings: exactSettings, Tsx: true, Errors: errorAtStart},
		// ---- Intentional divergence: assignment operators remain declaration writes ----
		{Code: `Component.propTypes ||= { foo: P };`, Settings: exactSettings, Tsx: true, Errors: errorAtStart},
		// Locks in MemberExpression branch 3: forward declarations are visible through scope lookup.
		{Code: `Component.propTypes = props; const props = { foo: P };`, Settings: exactSettings, Tsx: true, Errors: errorAtStart},
		// Locks in MemberExpression branch 3b: identifier initializers that are non-exact calls report.
		{Code: `const props = other(); Component.propTypes = props;`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 1, Column: 24}}},
		// ---- Intentional divergence: a type-space merge does not hide the runtime const declaration ----
		{Code: `const Props = { foo: P }; type Props = { foo: string }; Component.propTypes = Props;`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 1, Column: 57}}},
		// ---- Cached alias classification permits repeated safe propTypes assignments ----
		{Code: `const props = { foo: P }; A.propTypes = props; B.propTypes = props;`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Line: 1, Column: 27}, {MessageId: "propTypes", Line: 1, Column: 48}}},
		// Locks in the absence of component-name validation on runtime declarations.
		{Code: `NotAComponent.propTypes = { foo: P };`, Settings: exactSettings, Tsx: true, Errors: errorAtStart},
		// ---- Dimension 4: single and multi-level parenthesized receivers remain visible ----
		{Code: `(Component).propTypes = { foo: P }; ((Other)).propTypes = { bar: P };`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Line: 1, Column: 1}, {MessageId: "propTypes", Line: 1, Column: 37}}},
		// ---- Dimension 4: TS non-null, as, and satisfies receiver wrappers do not hide propTypes ----
		{Code: `Component!.propTypes = { foo: P }; (Other as any).propTypes = { bar: P }; (Third satisfies object).propTypes = { baz: P };`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Line: 1, Column: 1}, {MessageId: "propTypes", Line: 1, Column: 36}, {MessageId: "propTypes", Line: 1, Column: 75}}},
		// ---- Dimension 4: string-literal and template-literal element keys name runtime propTypes ----
		{Code: `Component['propTypes'] = { foo: P };`, Settings: exactSettings, Tsx: true, Errors: errorAtStart},
		{Code: "Component[`propTypes`] = { foo: P };", Settings: exactSettings, Tsx: true, Errors: errorAtStart},
		// ---- Dimension 4: a static string class key names the public runtime field ----
		{Code: `class Component { static ['propTypes'] = { foo: P }; }`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Line: 1, Column: 19}}},
		// ---- Dimension 4: class declarations and class expressions share static-field behavior ----
		{Code: `class A { static propTypes = { a: P }; } const B = class { static propTypes = { b: P }; };`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Line: 1, Column: 11}, {MessageId: "propTypes", Line: 1, Column: 60}}},
		// ---- Dimension 4: nested classes retain independent declaration reports ----
		{Code: `class Outer { static propTypes = { outer: P }; method() { return class Inner { static propTypes = { inner: P }; }; } }`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Line: 1, Column: 15}, {MessageId: "propTypes", Line: 1, Column: 80}}},
		// ---- Dimension 4: authored TS wrappers around a callee prevent exact-wrapper matching ----
		{Code: `Component.propTypes = (exact as any)({ foo: P });`, Settings: exactSettings, Tsx: true, Errors: errorAtStart},
		// ---- Dimension 4: parentheses inside a dotted wrapper name remain observable source text ----
		{Code: `Component.propTypes = (Object).freeze({ foo: P });`, Settings: objectExactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: "Component propTypes should be exact by using 'Object.freeze'.", Line: 1, Column: 1}}},
		// ---- Real-user: issue #1455 runtime PropTypes definitions require an exact wrapper ----
		{Code: `const Card = ({ title }) => <h1>{title}</h1>; Card.propTypes = { title: PropTypes.string };`, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 1, Column: 47}}},
	})
}
