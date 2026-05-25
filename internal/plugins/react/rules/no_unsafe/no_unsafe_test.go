package no_unsafe

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// msg builds the diagnostic string upstream emits — kept inline so each
// invalid case stays grep-able to the upstream test file.
func msg(method, newMethod string) string {
	return method + " is unsafe for use in async rendering. Update the component to use " + newMethod + " instead. See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html."
}

// optsCheckAliases is the array-wrapped Options shape — exercises the JSON
// path the CLI / rule_tester takes (matches the multi-element config shape).
var optsCheckAliases = []interface{}{map[string]interface{}{"checkAliases": true}}

// settingsReact is the helper that builds `settings: { react: { version: ... } }`
// in the rule_tester shape.
func settingsReact(version string) map[string]interface{} {
	return map[string]interface{}{"react": map[string]interface{}{"version": version}}
}

func TestNoUnsafeRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Upstream valid cases (1:1 migration of `tests/lib/rules/no-unsafe.js`)
		// ============================================================

		// ---- Upstream #1: React.Component with safe lifecycle methods ----
		{Code: `
        class Foo extends React.Component {
          componentDidUpdate() {}
          render() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Upstream #2: createReactClass with safe lifecycle methods ----
		{Code: `
        const Foo = createReactClass({
          componentDidUpdate: function() {},
          render: function() {}
        });
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Upstream #3: non-React class with deprecated names allowed ----
		{Code: `
        class Foo extends Bar {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Upstream #4: non-React class with UNSAFE_ names allowed ----
		{Code: `
        class Foo extends Bar {
          UNSAFE_componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
          UNSAFE_componentWillUpdate() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Upstream #5: non-createReactClass call with deprecated names ----
		{Code: `
        const Foo = bar({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {},
        });
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Upstream #6: non-createReactClass call with UNSAFE_ names ----
		{Code: `
        const Foo = bar({
          UNSAFE_componentWillMount: function() {},
          UNSAFE_componentWillReceiveProps: function() {},
          UNSAFE_componentWillUpdate: function() {},
        });
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Upstream #7: React.Component with deprecated names — checkAliases default = false ----
		{Code: `
        class Foo extends React.Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Upstream #8: React 16.2.0 < 16.3.0 — entire rule disabled ----
		{Code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
          UNSAFE_componentWillUpdate() {}
        }
      `, Tsx: true, Settings: settingsReact("16.2.0")},

		// ---- Upstream #9: createReactClass with deprecated names — checkAliases default = false ----
		{Code: `
        const Foo = createReactClass({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {},
        });
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Upstream #10: createReactClass with UNSAFE_ names at React 16.2.0 — rule disabled ----
		{Code: `
        const Foo = createReactClass({
          UNSAFE_componentWillMount: function() {},
          UNSAFE_componentWillReceiveProps: function() {},
          UNSAFE_componentWillUpdate: function() {},
        });
      `, Tsx: true, Settings: settingsReact("16.2.0")},

		// ============================================================
		// Edge cases beyond upstream — Dimensions 1–4 + universal shapes
		// ============================================================

		// ---- React.PureComponent — same flag set ----
		// Locks in `ExtendsReactComponent`'s `(Pure)?Component` matcher.
		{Code: `
        class Foo extends React.PureComponent {
          componentDidMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Bare `Component` extends (e.g. `import { Component }`) ----
		{Code: `
        class Foo extends Component {
          componentDidMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Bare `PureComponent` extends ----
		{Code: `
        class Foo extends PureComponent {
          componentDidMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Class expression extending React.Component — no fire on safe lifecycle ----
		{Code: `
        const Foo = class extends React.Component {
          componentDidMount() {}
        };
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Computed key matching unsafe name — upstream's `getPropertyName`
		// returns `nameNode.name` which is undefined for ComputedPropertyName,
		// so `[X]() {}` never matches regardless of what X resolves to. Locks
		// in the divergence from `utils.GetStaticPropertyName` (which would
		// statically resolve `['UNSAFE_componentWillMount']` to its value). ----
		{Code: `
        const k = 'UNSAFE_componentWillMount';
        class Foo extends React.Component {
          [k]() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- StringLiteral key — upstream's `getPropertyName` returns
		// `nameNode.name`, and `Literal` nodes don't have `.name` (they have
		// `.value`), so `'UNSAFE_componentWillMount': fn` returns undefined
		// upstream and is NOT flagged. Lock the contract here so a future
		// refactor reaching for `utils.GetStaticPropertyName` doesn't silently
		// flip this. ----
		{Code: `
        const Foo = createReactClass({
          'UNSAFE_componentWillMount': function() {},
          'UNSAFE_componentWillReceiveProps': function() {},
          'UNSAFE_componentWillUpdate': function() {},
          render: function() {},
        });
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- StringLiteral key on a class — same contract, ES6 shape ----
		{Code: `
        class Foo extends React.Component {
          'UNSAFE_componentWillMount'() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- NumericLiteral key — upstream returns undefined; we don't flag.
		// (Pathological code; included as a graceful-degradation lock.) ----
		{Code: `
        class Foo extends React.Component {
          1() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Empty class body — graceful degradation ----
		{Code: `
        class Foo extends React.Component {}
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- createReactClass with empty object — graceful degradation ----
		{Code: `
        const Foo = createReactClass({});
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- createReactClass with spread — has no key, must not crash and
		// must not flag the spread itself ----
		{Code: `
        const Foo = createReactClass({
          ...mixins,
          render: function() {}
        });
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Default-version (no setting) treats project as latest → rule active.
		// A safe class still produces no diagnostic. ----
		{Code: `
        class Foo extends React.Component {
          componentDidMount() {}
        }
      `, Tsx: true},

		// ---- checkAliases at React 16.2.0 — version gate beats option ----
		// Even with `checkAliases: true`, rule is disabled below 16.3.0.
		{Code: `
        class Foo extends React.Component {
          componentWillMount() {}
        }
      `, Tsx: true, Options: optsCheckAliases, Settings: settingsReact("16.2.0")},

		// ---- Custom pragma — `React.Component` is not the configured pragma's
		// Component, so it's not detected as a React class component. ----
		{Code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact", "version": "16.4.0"}}},

		// ---- Object literal NOT in a `createReactClass(...)` call — generic
		// object-literal traversal must not over-match. ----
		{Code: `
        const config = {
          UNSAFE_componentWillMount: function() {},
        };
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Plain class without extends — not a React component. ----
		{Code: `
        class Foo {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- ObjectLiteralExpression nested inside a property — must not
		// match `createReactClass`. ----
		{Code: `
        const wrapper = {
          inner: {
            UNSAFE_componentWillMount: function() {},
          }
        };
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- TypeScript generic class extending React.Component — generics
		// don't affect the heritage check. ----
		{Code: `
        class Foo<P> extends React.Component<P> {
          render() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Custom createClass via settings — `myCreate(...)` is the
		// component factory; valid case has only safe methods. ----
		{Code: `
        const Foo = myCreate({
          render: function() {}
        });
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreate", "version": "16.4.0"}}},

		// ---- Default `createReactClass` with custom `createClass` setting —
		// the default no longer matches. ----
		{Code: `
        const Foo = createReactClass({
          UNSAFE_componentWillMount: function() {},
        });
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreate", "version": "16.4.0"}}},

		// ---- TS NonNullExpression in extends — `extends React.Component!`
		// is `TSNonNullExpression` upstream and `KindNonNullExpression` in
		// tsgo. Neither walks through it, so it does NOT flag. Lock the
		// alignment in. ----
		{Code: `
        class Foo extends React.Component! {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- TS AsExpression in extends — `extends (X as any)` doesn't match
		// either the Identifier or PropertyAccessExpression arm. ----
		{Code: `
        class Foo extends (React.Component as any) {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- ElementAccessExpression in extends — `React['Component']` is
		// not a PropertyAccessExpression, so doesn't match. (ESLint's
		// MemberExpression branch checks `.property.name` which is undefined
		// for computed access — so upstream also doesn't match.) ----
		{Code: `
        class Foo extends React['Component'] {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Class extends a HOC return value — not a literal Component reference. ----
		{Code: `
        class Foo extends withRouter(Base) {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Class with implements but no extends — not a React class. ----
		{Code: `
        class Foo implements IFoo {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Constructor + non-unsafe members in a React class — graceful. ----
		{Code: `
        class Foo extends React.Component {
          constructor(props) { super(props); }
          render() { return null; }
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- TS abstract class with no unsafe methods — graceful. ----
		{Code: `
        abstract class Foo extends React.Component {
          abstract render(): JSX.Element;
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- TS index signature — no `name`, must not crash, must not flag. ----
		{Code: `
        class Foo extends React.Component {
          [key: string]: any;
          render() { return null; }
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Static block — no name, must not crash. ----
		{Code: `
        class Foo extends React.Component {
          static {
            console.log('init');
          }
          render() { return null; }
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Private identifier method (`#foo`) named like an unsafe lifecycle ----
		// Private fields strip the leading `#` (matches upstream's
		// `nameNode.name`). But React doesn't call private methods as
		// lifecycles — upstream still flags structurally; we match. Locked
		// in the INVALID side; here we lock that an unrelated private name
		// (e.g. `#privateHelper`) does NOT flag.
		{Code: `
        class Foo extends React.Component {
          #privateHelper() {}
          render() { return null; }
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- TypeScript optional method (declaration only, body required) —
		// The `?` is on the parameter, not the method itself, when present.
		// Verify a typical signature stays graceful. ----
		{Code: `
        class Foo extends React.Component {
          render(p?: any) { return null; }
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Decorator on a non-unsafe method — graceful. ----
		{Code: `
        class Foo extends React.Component {
          @bind
          handleClick() {}
          render() { return null; }
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Computed string-literal key — `['UNSAFE_componentWillMount']: fn`
		// `utils.GetStaticPropertyName` would resolve this, but upstream's
		// `getPropertyName` returns `nameNode.name` and ComputedPropertyName
		// has no `.name`, so it does NOT flag. Locked in. ----
		{Code: `
        const Foo = createReactClass({
          ['UNSAFE_componentWillMount']: function() {},
          render: function() {},
        });
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- React.createClass via deprecated React-namespace pattern is
		// only matched when pragma matches — default pragma "React" + bare
		// createClass keyword "React.createClass" is NOT the same as
		// "createReactClass". Locked in. ----
		{Code: `
        const Foo = React.createClass({
          UNSAFE_componentWillMount: function() {},
        });
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Object literal as RHS of property — must not be detected as a
		// createReactClass arg. ----
		{Code: `
        const cfg = {
          payload: {
            UNSAFE_componentWillMount: 1,
          },
        };
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Object literal as JSX prop value — same; must not match. ----
		{Code: `
        const Foo = () => <div data-cfg={{ UNSAFE_componentWillMount: 1 }} />;
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Optional-chain in extends (`extends React?.Component`) —
		// upstream parses this as `ChainExpression` which is NOT
		// `MemberExpression`, so `componentUtil.isES6Component` returns
		// false. Verified empirically against eslint-plugin-react@latest.
		// We mirror by explicitly rejecting OptionalChain in
		// `ExtendsReactComponent`. ----
		{Code: `
        class Foo extends React?.Component {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Extends a sequence/conditional/IIFE — heritage is not
		// Identifier or PropertyAccessExpression, so no match. ----
		{Code: `
        class Foo extends (cond ? React.Component : Bar) {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		{Code: `
        class Foo extends (a, React.Component) {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Class extending an alias — only literal Component / PureComponent
		// matches the regex; aliases don't. ----
		{Code: `
        class Foo extends ComponentAlias {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Class extending deep namespace — only single-level pragma
		// qualification matches. ----
		{Code: `
        class Foo extends m.n.Component {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Class extends X (non-Component name) — bare Identifier must
		// be exactly `Component` or `PureComponent`. ----
		{Code: `
        class Foo extends MyBase {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- this.UNSAFE_componentWillMount = ... assignment — not a
		// declaration; never flag. ----
		{Code: `
        class Foo extends React.Component {
          render() {
            this.UNSAFE_componentWillMount = () => {};
            return null;
          }
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},

		// ---- Generic + paren: upstream's typescript-eslint parser
		// produces a `TSInstantiationExpression` for `React.Component<P>`;
		// wrapping it in parens makes the heritage's superClass not a
		// MemberExpression so upstream does NOT flag. tsgo's heritage
		// parsing of `extends (X<P>)` likewise doesn't expose
		// `React.Component` as a clean PropertyAccessExpression at the
		// SkipParentheses position, so rslint also does NOT flag —
		// alignment is automatic in this case. ----
		{Code: `
        class Foo<P> extends (React.Component<P>) {
          UNSAFE_componentWillMount() {}
        }
      `, Tsx: true, Settings: settingsReact("16.4.0")},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Upstream invalid cases (1:1 migration — line/column preserved
		// against the upstream `errors[*]` assertions).
		// ============================================================

		// ---- Upstream #1: React.Component + checkAliases: true (3 unprefixed reports) ----
		{
			Code: `
        class Foo extends React.Component {
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
			Tsx:      true,
			Options:  optsCheckAliases,
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("componentWillMount", "componentDidMount"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("componentWillReceiveProps", "getDerivedStateFromProps"), Line: 4, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("componentWillUpdate", "componentDidUpdate"), Line: 5, Column: 11},
			},
		},

		// ---- Upstream #2: React.Component + UNSAFE_ names at 16.3.0 (no option) ----
		{
			Code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
          UNSAFE_componentWillUpdate() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillReceiveProps", "getDerivedStateFromProps"), Line: 4, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillUpdate", "componentDidUpdate"), Line: 5, Column: 11},
			},
		},

		// ---- Upstream #3: createReactClass + checkAliases: true ----
		{
			Code: `
        const Foo = createReactClass({
          componentWillMount: function() {},
          componentWillReceiveProps: function() {},
          componentWillUpdate: function() {},
        });
      `,
			Tsx:      true,
			Options:  optsCheckAliases,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("componentWillMount", "componentDidMount"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("componentWillReceiveProps", "getDerivedStateFromProps"), Line: 4, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("componentWillUpdate", "componentDidUpdate"), Line: 5, Column: 11},
			},
		},

		// ---- Upstream #4: createReactClass + UNSAFE_ names at 16.3.0 ----
		{
			Code: `
        const Foo = createReactClass({
          UNSAFE_componentWillMount: function() {},
          UNSAFE_componentWillReceiveProps: function() {},
          UNSAFE_componentWillUpdate: function() {},
        });
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillReceiveProps", "getDerivedStateFromProps"), Line: 4, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillUpdate", "componentDidUpdate"), Line: 5, Column: 11},
			},
		},

		// ============================================================
		// Edge cases beyond upstream — lock in branch coverage and
		// nested / wrapper scenarios that real codebases produce.
		// ============================================================

		// ---- ClassExpression — same listener as ClassDeclaration ----
		{
			Code: `
        const Foo = class extends React.Component {
          UNSAFE_componentWillMount() {}
        };
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Bare `extends Component` (no pragma qualifier) ----
		{
			Code: `
        class Foo extends Component {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- React.PureComponent extends — same flag set as Component ----
		{
			Code: `
        class Foo extends React.PureComponent {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Source-order traversal — upstream's in-rule
		// `methods.sort(...)` is dead code (ESLint's reporter re-sorts by
		// (line, column)). Diagnostics arrive in source order, even when
		// the names are not alphabetical. ----
		{
			Code: `
        class Foo extends React.Component {
          UNSAFE_componentWillUpdate() {}
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillUpdate", "componentDidUpdate"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 4, Column: 11},
			},
		},

		// ---- Class field with arrow function — `astUtil.getComponentProperties`
		// returns all members, so `UNSAFE_componentWillMount = () => {}` flags
		// the same as a method-shorthand. Upstream behavior: same. ----
		{
			Code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount = () => {};
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Static method named like an unsafe lifecycle — upstream's
		// `getComponentProperties` returns all members with no static filter.
		// Locks in upstream behavior: a static method with this name still
		// flags (even though React doesn't call statics as lifecycles). ----
		{
			Code: `
        class Foo extends React.Component {
          static UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Getter named like an unsafe lifecycle — same membership rule. ----
		{
			Code: `
        class Foo extends React.Component {
          get UNSAFE_componentWillMount() { return null; }
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Default version (no `settings.react.version`) → "latest" → active ----
		{
			Code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Options-shape coverage: bare-map (single-option CLI shape) ----
		// Equivalent to the array-wrapped form upstream test #1 uses, but
		// passed via `map[string]interface{}` rather than
		// `[]interface{}{map[string]interface{}{...}}`. Catches a missing
		// `GetOptionsMap` call in `parseOptions`.
		{
			Code: `
        class Foo extends React.Component {
          componentWillMount() {}
        }
      `,
			Tsx:      true,
			Options:  map[string]interface{}{"checkAliases": true},
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Nested classes — only the inner class is a React component;
		// outer is plain. Both classes are visited; only Inner reports. ----
		{
			Code: `
        class Outer {
          run() {
            return class Inner extends React.Component {
              UNSAFE_componentWillMount() {}
            };
          }
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 5, Column: 15},
			},
		},

		// ---- Both nested classes are React components — each gets its own
		// independent listener invocation, so both flag. ----
		{
			Code: `
        class Outer extends React.Component {
          UNSAFE_componentWillMount() {}
          render() {
            return class Inner extends React.Component {
              UNSAFE_componentWillMount() {}
            };
          }
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 6, Column: 15},
			},
		},

		// ---- Wrapped class expression — `connect()(class extends ...)` ----
		// HOC wrapping doesn't move the class out of the listener; the
		// inner class is still detected via heritage clause.
		{
			Code: `
        const Foo = connect(state => state)(class extends React.Component {
          UNSAFE_componentWillMount() {}
        });
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- createReactClass with parenthesized object argument — rslint's
		// `IsCreateReactClassObjectArg` skips parens upward, matching tsgo's
		// preserved-paren AST. (Upstream ESTree flattens parens at parse
		// time, so this is a language-natural divergence: rslint flags,
		// upstream flags too because parens vanish before the rule sees
		// them.) Lock the observable behavior. ----
		{
			Code: `
        const Foo = createReactClass(({
          UNSAFE_componentWillMount: function() {},
        }));
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Pragma-qualified createClass — `Preact.createClass(...)` ----
		{
			Code: `
        const Foo = Preact.createReactClass({
          UNSAFE_componentWillMount: function() {},
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact", "version": "16.3.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Multiple components in one file — each independently checked. ----
		{
			Code: `
        class A extends React.Component {
          UNSAFE_componentWillMount() {}
        }
        class B extends React.Component {
          UNSAFE_componentWillUpdate() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillUpdate", "componentDidUpdate"), Line: 6, Column: 11},
			},
		},

		// ---- TypeScript generic class — type parameters don't affect the
		// extends-clause matcher. ----
		{
			Code: `
        class Foo<P> extends React.Component<P> {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Class with implements + extends React.Component — extends drives
		// the detection; implements is irrelevant. ----
		{
			Code: `
        class Foo extends React.Component implements IFoo {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Default export anonymous class ----
		{
			Code: `
        export default class extends React.Component {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- React.memo wrapping a class expression ----
		{
			Code: `
        const Foo = React.memo(class extends React.Component {
          UNSAFE_componentWillMount() {}
        });
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- React.forwardRef wrapping a class expression ----
		{
			Code: `
        const Foo = React.forwardRef((p, ref) =>
          new (class extends React.Component {
            UNSAFE_componentWillMount() {}
          })()
        );
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 4, Column: 13},
			},
		},

		// ---- Multiple HOC layers wrapping the class ----
		{
			Code: `
        const Foo = withRouter(connect(s => s)(class extends React.Component {
          UNSAFE_componentWillMount() {}
        }));
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- async unsafe method — modifier doesn't affect name detection. ----
		{
			Code: `
        class Foo extends React.Component {
          async UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Setter named like an unsafe lifecycle ----
		{
			Code: `
        class Foo extends React.Component {
          set UNSAFE_componentWillMount(v) {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Decorator on the unsafe method ----
		{
			Code: `
        class Foo extends React.Component {
          @bind
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Decorator on the class itself ----
		{
			Code: `
        @observer
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 4, Column: 11},
			},
		},

		// ---- TS readonly class field with arrow ----
		{
			Code: `
        class Foo extends React.Component {
          readonly UNSAFE_componentWillMount = () => {};
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- TS abstract method — body absent, name still extractable ----
		{
			Code: `
        abstract class Foo extends React.Component {
          abstract UNSAFE_componentWillMount(): void;
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- TS declare class — body absent, signature only ----
		{
			Code: `
        declare class Foo extends React.Component {
          UNSAFE_componentWillMount(): void;
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- createReactClass with method shorthand (ES6 syntax in obj literal) ----
		{
			Code: `
        const Foo = createReactClass({
          UNSAFE_componentWillMount() {},
          render() { return null; }
        });
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- createReactClass with arrow function value ----
		{
			Code: `
        const Foo = createReactClass({
          UNSAFE_componentWillMount: () => {},
          render: function() { return null; }
        });
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- createReactClass returned from a function ----
		{
			Code: `
        function makeFoo() {
          return createReactClass({
            UNSAFE_componentWillMount: function() {},
            render: function() { return null; }
          });
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 4, Column: 13},
			},
		},

		// ---- createReactClass nested in IIFE ----
		{
			Code: `
        const Foo = (() => createReactClass({
          UNSAFE_componentWillMount: function() {},
          render: function() { return null; }
        }))();
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Mixed UNSAFE_ + unprefixed names with checkAliases:true — both
		// flag, in source order (matches ESLint's reporter-layer sort). ----
		{
			Code: `
        class Foo extends React.Component {
          componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
        }
      `,
			Tsx:      true,
			Options:  optsCheckAliases,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("componentWillMount", "componentDidMount"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillReceiveProps", "getDerivedStateFromProps"), Line: 4, Column: 11},
			},
		},

		// ---- Class with parens around extends — `extends (React.Component)`
		// (tsgo preserves parens, ExtendsReactComponent skips them; upstream
		// flattens at parse time). Both flag — observable behavior aligned. ----
		{
			Code: `
        class Foo extends (React.Component) {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Class with parens around the pragma object — `(React).Component`
		// `ExtendsReactComponent` skips outer parens AND the parens around the
		// pragma identifier. Locked in. ----
		{
			Code: `
        class Foo extends (React).Component {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Class with constructor + unsafe lifecycle — constructor doesn't
		// affect detection; lifecycle still flags. ----
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) { super(props); }
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 4, Column: 11},
			},
		},

		// ---- All three UNSAFE_ + all three unprefixed with checkAliases:true ----
		// Diagnostics arrive in source order (line 3 → 8). Confirms ESLint's
		// reporter-layer line/column sort matches our source-order traversal.
		{
			Code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount() {}
          UNSAFE_componentWillReceiveProps() {}
          UNSAFE_componentWillUpdate() {}
          componentWillMount() {}
          componentWillReceiveProps() {}
          componentWillUpdate() {}
        }
      `,
			Tsx:      true,
			Options:  optsCheckAliases,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillReceiveProps", "getDerivedStateFromProps"), Line: 4, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillUpdate", "componentDidUpdate"), Line: 5, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("componentWillMount", "componentDidMount"), Line: 6, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("componentWillReceiveProps", "getDerivedStateFromProps"), Line: 7, Column: 11},
				{MessageId: "unsafeMethod", Message: msg("componentWillUpdate", "componentDidUpdate"), Line: 8, Column: 11},
			},
		},

		// ---- ES5 component nested inside an ES6 component (both detected
		// independently — only the ES5 inner one has unsafe). ----
		{
			Code: `
        class Outer extends React.Component {
          render() {
            return createReactClass({
              UNSAFE_componentWillMount: function() {},
              render: function() { return null; }
            });
          }
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 5, Column: 15},
			},
		},

		// ---- TypeScript class extending namespaced pragma form — `React.PureComponent` ----
		{
			Code: `
        class Foo<T> extends React.PureComponent<T> {
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- Custom createClass via settings — `myCreate(...)` is now the factory ----
		{
			Code: `
        const Foo = myCreate({
          UNSAFE_componentWillMount: function() {},
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "myCreate", "version": "16.3.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- ObjectExpression at NON-first arg position — upstream's
		// `componentUtil.isES5Component` only checks `node.parent.callee`
		// and accepts any argument position. Verified empirically against
		// eslint-plugin-react@latest: `createReactClass(other, {...})` IS
		// flagged. Lock the alignment in (rslint's `IsCreateReactClassObjectArg`
		// is more restrictive — we deliberately bypass it via a local
		// `isES5Component` to mirror upstream exactly). ----
		{
			Code: `
        const Foo = createReactClass(other, {
          UNSAFE_componentWillMount: function() {},
        });
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ---- `new createReactClass({...})` — upstream's `node.parent.callee`
		// check accepts NewExpression too (NewExpression also exposes
		// `.callee` in ESTree). Verified empirically: ESLint flags this.
		// Semantically nonsensical (createReactClass is not a constructor)
		// but we mirror the structural check. ----
		{
			Code: `
        const Foo = new createReactClass({
          UNSAFE_componentWillMount: function() {},
        });
      `,
			Tsx:      true,
			Settings: settingsReact("16.3.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Message: msg("UNSAFE_componentWillMount", "componentDidMount"), Line: 3, Column: 11},
			},
		},

		// ============================================================
		// EndLine / EndColumn assertions — verified against
		// eslint-plugin-react@latest. Upstream reports the diagnostic on
		// the entire MethodDefinition / Property node, so the range covers
		// modifiers, key, and (when present) body.
		// ============================================================

		// ---- Method shorthand: range = `UNSAFE_componentWillMount() {}` ----
		{
			Code:     `class Foo extends React.Component { UNSAFE_componentWillMount() {} }`,
			Tsx:      true,
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Line: 1, Column: 37, EndLine: 1, EndColumn: 67},
			},
		},

		// ---- Class field arrow: range covers `UNSAFE_componentWillMount = () => {};`
		// — verified empirically against ESLint 9.x: PropertyDefinition's
		// range INCLUDES the terminating semicolon (when present). Method
		// declarations don't. tsgo's `PropertyDeclaration.End()` already
		// matches this byte-for-byte, so no range adjustment is needed. ----
		{
			Code:     `class Foo extends React.Component { UNSAFE_componentWillMount = () => {}; }`,
			Tsx:      true,
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Line: 1, Column: 37, EndLine: 1, EndColumn: 74},
			},
		},

		// ---- Multi-line method: range spans multiple lines ----
		{
			Code: `class Foo extends React.Component {
  UNSAFE_componentWillMount() {
    return null;
  }
}`,
			Tsx:      true,
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Line: 2, Column: 3, EndLine: 4, EndColumn: 4},
			},
		},

		// ---- createReactClass property: range covers `KEY: function() {}` ----
		{
			Code: `const Foo = createReactClass({
  UNSAFE_componentWillMount: function() {},
  render: function() {}
});`,
			Tsx:      true,
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Line: 2, Column: 3, EndLine: 2, EndColumn: 43},
			},
		},

		// ---- Static method: includes the `static` modifier in the range ----
		{
			Code:     `class Foo extends React.Component { static UNSAFE_componentWillMount() {} }`,
			Tsx:      true,
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Line: 1, Column: 37, EndLine: 1, EndColumn: 74},
			},
		},

		// ---- Method overloads: upstream reports the diagnostic anchored
		// to the first matching declaration N times (because its
		// `methods.sort().forEach` iterates each occurrence of the name,
		// and each iteration calls `find` which returns the first match).
		// rslint mirrors this exactly: we collect (name, nodes[]) and
		// report N times all on `nodes[0]` to match upstream byte-for-byte.
		// Verified empirically against eslint-plugin-react@latest. ----
		{
			Code: `
        class Foo extends React.Component {
          UNSAFE_componentWillMount(): void;
          UNSAFE_componentWillMount() {}
        }
      `,
			Tsx:      true,
			Settings: settingsReact("16.4.0"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMethod", Line: 3, Column: 11},
				{MessageId: "unsafeMethod", Line: 3, Column: 11},
			},
		},
	})
}
