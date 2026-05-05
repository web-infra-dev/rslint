package boolean_prop_naming

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// patternIs is the most-common rule pattern across upstream tests.
const patternIs = "^is[A-Z]([A-Za-z0-9]?)+"

// patternIsHas is the documented default pattern; upstream uses it whenever
// both prefixes are accepted.
const patternIsHas = "^(is|has)[A-Z]([A-Za-z0-9]?)+"

// optsIs / optsIsHas are array-wrapped Options shapes — they exercise the
// JSON-array path that CLI / multi-element configs take. The bare-map shape
// is exercised separately under "Options-shape coverage".
var (
	optsIs    = []interface{}{map[string]interface{}{"rule": patternIs}}
	optsIsHas = []interface{}{map[string]interface{}{"rule": patternIsHas}}
)

func TestBooleanPropNamingRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &BooleanPropNamingRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Upstream valid cases (all migrated; Flow-only cases use the
		// equivalent TypeScript shape since rslint parses TS, not Flow).
		// ============================================================

		// ---- Default behavior: missing `rule` regex → rule is a no-op ----
		// Upstream sets `rule = config.rule ? new RegExp(config.rule) : null`
		// and bails on every listener; both `is*` and `has*` names pass not
		// because they match the documented default but because nothing is
		// checked.
		{Code: `
          var Hello = createReactClass({
            propTypes: {isSomething: PropTypes.bool, hasValue: PropTypes.bool},
            render: function() { return <div />; }
          });
        `, Tsx: true},

		// ---- createReactClass + PropTypes.bool / React.PropTypes.bool ----
		{Code: `
          var Hello = createReactClass({
            propTypes: {isSomething: PropTypes.bool},
            render: function() { return <div />; }
          });
        `, Tsx: true, Options: optsIs},
		{Code: `
          var Hello = createReactClass({
            propTypes: {isSomething: React.PropTypes.bool},
            render: function() { return <div />; }
          });
        `, Tsx: true, Options: optsIs},
		{
			Code: `
              var Hello = React.createClass({
                propTypes: {isSomething: PropTypes.bool},
                render: function() { return <div />; }
              });
            `,
			Tsx:      true,
			Options:  optsIs,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass"}},
		},
		{
			// Non-boolean PropTypes.any is ignored entirely.
			Code: `
              var Hello = React.createClass({
                propTypes: {something: PropTypes.any},
                render: function() { return <div />; }
              });
            `,
			Tsx:      true,
			Options:  optsIs,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass"}},
		},

		// ---- ES6 class extends React.Component / Component ----
		{Code: `
          class Hello extends React.Component {
            render () { return <div />; }
          }
          Hello.propTypes = {isSomething: PropTypes.bool}
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends React.Component {
            render () { return <div />; }
          }
          Hello.propTypes = wrap({ a: PropTypes.bool })
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends React.Component {
            render () { return <div />; }
          }
          Hello.propTypes = {something: PropTypes.any}
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends Component {
            render () { return <div />; }
          }
          Hello.propTypes = {isSomething: PropTypes.bool}
        `, Tsx: true, Options: optsIs},

		// ---- ES6 components with static class properties ----
		{Code: `
          class Hello extends React.Component {
            static propTypes = {isSomething: PropTypes.bool};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          const spreadProps = { aSpreadProp: PropTypes.string };
          class Hello extends React.Component {
            static propTypes = {isSomething: PropTypes.bool, ...spreadProps};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          const spreadProps = { aSpreadProp: PropTypes.string };
          class Hello extends Component {
            render () { return <div />; }
          }
          Hello.propTypes = {isSomething: PropTypes.bool, ...spreadProps}
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends React.Component {
            static propTypes = {isSomething: React.PropTypes.bool};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends React.Component {
            static propTypes = {something: PropTypes.any};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},

		// ---- TS-equivalent of the upstream Flow `props: {x: boolean}` ----
		// Upstream marks this `features: ['flow']`. In tsgo the same syntax
		// parses as a TS class field with a TypeLiteral annotation, and the
		// rule path validates it.
		{Code: `
          class Hello extends React.Component {
            props: {isSomething: boolean};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends React.Component {
            props: {something: any};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},

		// ---- Stateless components with PropTypes ----
		{Code: `
          var Hello = ({isSomething}) => { return <div /> }
          Hello.propTypes = {isSomething: PropTypes.bool};
        `, Tsx: true, Options: optsIs},
		{Code: `
          type Props = { isSomething: boolean };
          function Hello(props: Props): React.Element { return <div /> }
        `, Tsx: true, Options: optsIs},

		// ---- Custom propTypeNames ----
		{Code: `
          class Hello extends React.Component {
            static propTypes = {
              isSomething: PropTypes.mutuallyExclusiveTrueProps,
              something: PropTypes.bool
            };
            render () { return <div />; }
          }
        `, Tsx: true, Options: []interface{}{map[string]interface{}{
			"propTypeNames": []interface{}{"mutuallyExclusiveTrueProps"},
			"rule":          patternIs,
		}}},
		{Code: `
          class Hello extends React.Component {
            static propTypes = {
              isSomething: mutuallyExclusiveTrueProps,
              isSomethingElse: bool
            };
            render () { return <div />; }
          }
        `, Tsx: true, Options: []interface{}{map[string]interface{}{
			"propTypeNames": []interface{}{"bool", "mutuallyExclusiveTrueProps"},
			"rule":          patternIs,
		}}},

		// ---- Misc shapes that must not crash ----
		{Code: `
          var x = {a: 1}
          var y = {...x}
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends PureComponent {
            props: PropsType;
            render () { return <div /> }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          function Card(props) {
            return <div>{props.showScore ? 'yeh' : 'no'}</div>;
          }
          Card.propTypes = merge({}, Card.propTypes, {
              showScore: PropTypes.bool
          });`,
			Tsx: true, Options: optsIsHas,
		},

		// ---- isRequired chains ----
		{Code: `
          var Hello = createReactClass({
            propTypes: {isSomething: PropTypes.bool.isRequired, hasValue: PropTypes.bool.isRequired},
            render: function() { return <div />; }
          });
        `, Tsx: true},
		{Code: `
          class Hello extends React.Component {
            static propTypes = {
              isSomething: PropTypes.bool.isRequired,
              hasValue: PropTypes.bool.isRequired
            };
            render() { return ( <div /> ); }
          }
        `, Tsx: true},
		{Code: `
          class Hello extends React.Component {
            render() { return ( <div /> ); }
          }
          Hello.propTypes = {
            isSomething: PropTypes.bool.isRequired,
            hasValue: PropTypes.bool.isRequired
          }
        `, Tsx: true},
		{Code: `
          var Hello = createReactClass({
            propTypes: {something: PropTypes.shape({}).isRequired},
            render: function() { return <div />; }
          });
        `, Tsx: true, Options: optsIs},

		// ---- validateNested ----
		{Code: `
          class Hello extends React.Component {
            render() { return ( <div /> ); }
          }
          Hello.propTypes = {
            isSomething: PropTypes.bool.isRequired,
            nested: PropTypes.shape({
              isWorking: PropTypes.bool
            })
          };
        `, Tsx: true},
		{Code: `
          class Hello extends React.Component {
            render() { return ( <div /> ); }
          }
          Hello.propTypes = {
            isSomething: PropTypes.bool.isRequired,
            nested: PropTypes.shape({
              nested: PropTypes.shape({
                isWorking: PropTypes.bool
              })
            })
          };
        `, Tsx: true, Options: []interface{}{map[string]interface{}{
			"rule":           patternIs,
			"validateNested": true,
		}}},

		// ---- TypeScript type annotations on functional components ----
		{Code: `
          type TestFNType = {
            isEnabled: boolean
          }
          const HelloNew = (props: TestFNType) => { return <div /> };
        `, Tsx: true, Options: optsIs},
		{Code: `
          type Props = {
            isEnabled: boolean
          } & OtherProps
          const HelloNew = (props: Props) => { return <div /> };
        `, Tsx: true, Options: optsIs},
		{Code: `
          type Props = {
            isEnabled: boolean
          } & {
            hasLOL: boolean
          } & OtherProps
          const HelloNew = (props: Props) => { return <div /> };
        `, Tsx: true, Options: []interface{}{map[string]interface{}{"rule": "(is|has)[A-Z]([A-Za-z0-9]?)+"}}},
		{Code: `
          type Props = {
            isEnabled: boolean
          }
          const HelloNew: React.FC<Props> = (props) => { return <div /> };
        `, Tsx: true, Options: optsIs},
		{Code: `
          type Props = {
            isEnabled: boolean
          } & {
            hasLOL: boolean
          }
          const HelloNew: React.FC<Props> = (props) => { return <div /> };
        `, Tsx: true, Options: optsIsHas},
		{Code: `
          type Props = {
            isEnabled: boolean
          } | {
            hasLOL: boolean
          }
          const HelloNew = (props: Props) => { return <div /> };
        `, Tsx: true, Options: optsIsHas},
		{Code: `
          type Props = {
            isEnabled: boolean
          } & ({
            hasLOL: boolean
          } | {
            isLOL: boolean
          })
          const HelloNew = (props: Props) => { return <div /> };
        `, Tsx: true, Options: optsIsHas},
		{Code: `
          export const DataRow = (props: { label: string; value: string; } & React.HTMLAttributes<HTMLDivElement>) => {
              const { label, value, ...otherProps } = props;
              return (
                  <div {...otherProps}>
                      <span>{label}</span>
                      <span>{value}</span>
                  </div>
              );
          };
        `, Tsx: true, Options: []interface{}{map[string]interface{}{"rule": "(^(is|has|should|without)[A-Z]([A-Za-z0-9]?)+|disabled|required|checked|defaultChecked)"}}},
		{Code: `
          // Non-component code — must not crash.
          const resultCode = result.code
            .replace('/** @jsxRuntime automatic */', '')
            .replace('/** @jsxImportSource @fluentui/react-jsx-runtime */', '');
        `, Tsx: true},

		// ============================================================
		// Additional valid cases (Phase 1 Dimensions 1–4 + real-world).
		// ============================================================

		// ---- Dimension 1 / 4: Empty containers ----
		{Code: `
          class Hello extends React.Component {
            static propTypes = {};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `const Hello = (props: {}) => <div />;`, Tsx: true, Options: optsIs},
		{Code: `
          interface Empty {}
          const Hello = (props: Empty) => <div />;
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 4: Non-component class / non-component identifier ----
		// Class without `extends Component` — propTypes field is irrelevant.
		{Code: `
          class NotAComponent {
            propTypes = {something: PropTypes.bool};
          }
        `, Tsx: true, Options: optsIs},
		// Non-component identifier `.propTypes` assignment must not report.
		{Code: `
          var notAComponent = 1;
          notAComponent.propTypes = {something: PropTypes.bool};
        `, Tsx: true, Options: optsIs},
		// Non-component receiver via member chain.
		{Code: `
          var ns = {};
          ns.notAComponent.propTypes = {something: PropTypes.bool};
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 4: Key-form variants on PropTypes object ----
		// String literal key — same equivalence class as identifier, valid name.
		{Code: `
          class Hello extends React.Component {
            static propTypes = { "isSomething": PropTypes.bool };
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 3: TS expression wrappers on prop value ----
		// `(PropTypes.bool as any)` / `PropTypes.bool!` / `(X satisfies Y)`
		// must unwrap to the same chain — matching name → no report.
		{Code: `
          class Hello extends React.Component {
            static propTypes = {isSomething: (PropTypes.bool as any)};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends React.Component {
            static propTypes = {isSomething: PropTypes.bool!};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          class Hello extends React.Component {
            static propTypes = {isSomething: (PropTypes.bool satisfies any)};
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},
		// Parens around the whole object literal initializer.
		{Code: `
          class Hello extends React.Component {
            static propTypes = ({isSomething: PropTypes.bool});
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 4: ElementAccess assignment LHS ----
		// `Hello['propTypes'] = {...}` — must report identical to the dotted form.
		{Code: `
          class Hello extends React.Component {
            render () { return <div />; }
          }
          Hello['propTypes'] = {isSomething: PropTypes.bool};
        `, Tsx: true, Options: optsIs},
		// ElementAccess with a non-string-literal arg → not recognized.
		{Code: `
          var key = 'propTypes';
          class Hello extends React.Component {
            render () { return <div />; }
          }
          Hello[key] = {something: PropTypes.bool};
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 4: Multi-segment receiver `outer.inner.Hello.propTypes` ----
		// Rightmost ID is the component name — accepted (conservative match).
		{Code: `
          class Hello extends React.Component {
            render () { return <div />; }
          }
          ns.Hello.propTypes = {isSomething: PropTypes.bool};
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 1: TypeReference to QualifiedName (`A.B`) ----
		// Rightmost segment is the lookup key. With the alias `Inner` defined,
		// `A.Inner` resolves and validates.
		{Code: `
          type Inner = { isFoo: boolean };
          const Hello = (props: ns.Inner) => <div />;
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 1: Interface declaration merging ----
		{Code: `
          interface Props { isFoo: boolean }
          interface Props { isBar: boolean }
          const Hello = (props: Props) => <div />;
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 1: Type alias indirection ----
		{Code: `
          type A = B;
          type B = { isEnabled: boolean };
          const Hello = (props: A) => <div />;
        `, Tsx: true, Options: optsIs},

		// ---- Dimension 1: Three-level type composition ----
		{Code: `
          type A = { isFoo: boolean };
          type B = { hasBar: boolean };
          type C = { isBaz: boolean };
          const Hello = (props: A & (B | C)) => <div />;
        `, Tsx: true, Options: optsIsHas},

		// ---- Both static propTypes and props: TypeLiteral (dual-path) ----
		// Both names match → no report from either path.
		{Code: `
          class Hello extends React.Component {
            static propTypes = { isFoo: PropTypes.bool };
            props: { isBar: boolean };
            render () { return <div />; }
          }
        `, Tsx: true, Options: optsIs},

		// ---- HOC wrapping — name is still a known component ----
		// Bare `memo` requires destructure-import from React (matches
		// upstream's `isDestructuredFromPragmaImport` gate).
		{Code: `
          import { memo } from 'react';
          const Hello = memo((props) => <div/>);
          Hello.propTypes = {isSomething: PropTypes.bool};
        `, Tsx: true, Options: optsIs},
		// ---- HOC wrapping via React.memo (no import gate) ----
		{Code: `
          const Hello = React.memo((props: { isFoo: boolean }) => <div/>);
        `, Tsx: true, Options: optsIs},
		// ---- Nested HOC wrappers ----
		{Code: `
          const Hello = React.memo(React.forwardRef((props: { isFoo: boolean }, ref) => <div ref={ref}/>));
        `, Tsx: true, Options: optsIs},

		// ---- propWrapperFunctions: `wrap` not configured → silently skip ----
		{Code: `
          function Card(props) { return <div /> }
          Card.propTypes = wrap({ showScore: PropTypes.bool });
        `, Tsx: true, Options: optsIsHas},

		// ---- Locks in: invalid regex degrades to no-op ----
		// Rather than throw (upstream behavior), the Go implementation
		// silently disables the rule. This is a documented divergence; the
		// test ensures the chosen behavior cannot regress.
		{Code: `
          class Hello extends React.Component {
            static propTypes = {something: PropTypes.bool};
            render () { return <div />; }
          }
        `, Tsx: true, Options: []interface{}{map[string]interface{}{"rule": "[unclosed"}}},

		// ---- Locks in: empty `propTypeNames` array → falls back to default ['bool'] ----
		// Upstream's schema requires `minItems: 1`; an empty array is
		// malformed config. We coerce to the default rather than throw.
		{Code: `
          class Hello extends React.Component {
            static propTypes = {isSomething: PropTypes.bool};
            render () { return <div />; }
          }
        `, Tsx: true, Options: []interface{}{map[string]interface{}{
			"rule":          patternIs,
			"propTypeNames": []interface{}{},
		}}},

		// ---- Locks in: empty `rule` string → no-op ----
		{Code: `
          class Hello extends React.Component {
            static propTypes = {something: PropTypes.bool};
            render () { return <div />; }
          }
        `, Tsx: true, Options: []interface{}{map[string]interface{}{"rule": ""}}},

		// ---- Bare-map Options shape (single-option CLI shape) ----
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = {isSomething: PropTypes.bool};
                render () { return <div />; }
              }
            `,
			Tsx:     true,
			Options: map[string]interface{}{"rule": patternIs},
		},

		// ---- nil Options ----
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = {something: PropTypes.bool};
                render () { return <div />; }
              }
            `,
			Tsx: true,
		},

		// ---- Options array with a non-map first element (malformed) → defaults ----
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = {something: PropTypes.bool};
                render () { return <div />; }
              }
            `,
			Tsx:     true,
			Options: []interface{}{"not-a-map"},
		},

		// ---- Interface heritage: `interface A extends B` ----
		// Both A's and B's bool members must match.
		{Code: `
          interface Base { isVisible: boolean }
          interface Props extends Base { isReady: boolean }
          const Hello = (props: Props) => <div />;
        `, Tsx: true, Options: optsIs},

		// ---- 3-segment QualifiedName: `a.b.C` ----
		{Code: `
          type C = { isFoo: boolean };
          const Hello = (props: ns.sub.C) => <div />;
        `, Tsx: true, Options: optsIs},

		// ---- Namespace-scoped component (recursive descent) ----
		// pre-walk should descend into ModuleDeclaration bodies so the
		// inner const HelloIn is registered as a component.
		{Code: `
          namespace UI {
            export const HelloIn = (props: { isFoo: boolean }) => <div/>;
          }
        `, Tsx: true, Options: optsIs},

		// ---- React.PropsWithChildren<Props> ----
		// firstTypeArgumentOfType returns Props; validateTypeNode resolves.
		{Code: `
          type Props = { isFoo: boolean };
          const Hello: React.PropsWithChildren<Props> = (props) => <div/>;
        `, Tsx: true, Options: optsIs},

		// ---- Optional / readonly modifiers on PropertySignature ----
		// `?` modifier and `readonly` modifier both pass through;
		// validateMembers should still recognize the bool annotation.
		{Code: `
          type Props = { isFoo?: boolean; readonly isBar: boolean };
          const Hello = (props: Props) => <div/>;
        `, Tsx: true, Options: optsIs},

		// ---- Generic component ----
		{Code: `
          function Hello<T>(props: { isReady: boolean; data: T }) { return <div/>; }
        `, Tsx: true, Options: optsIs},

		// ---- React.FC with composed generic ----
		{Code: `
          type Props = { isFoo: boolean };
          const Hello: React.FC<Props & { isExtra?: boolean }> = (props) => <div/>;
        `, Tsx: true, Options: optsIs},

		// ---- prop-types via various import shapes (no behavioral diff) ----
		{Code: `
          import PropTypes from 'prop-types';
          class Hello extends React.Component {
            static propTypes = { isFoo: PropTypes.bool };
            render() { return <div/>; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          import * as PropTypes from 'prop-types';
          class Hello extends React.Component {
            static propTypes = { isFoo: PropTypes.bool };
            render() { return <div/>; }
          }
        `, Tsx: true, Options: optsIs},
		{Code: `
          import { bool, func, string } from 'prop-types';
          class Hello extends React.Component {
            static propTypes = { isFoo: bool, name: string };
            render() { return <div/>; }
          }
        `, Tsx: true, Options: optsIs},

		// ---- Mapped types — opaque (upstream too) ----
		// A mapped type produces no PropertySignature members in the AST,
		// so nothing is checked; the rule must not crash.
		{Code: `
          type Props = { [K in 'a' | 'b']: boolean };
          const Hello = (props: Props) => <div/>;
        `, Tsx: true, Options: optsIs},

		// ---- Utility-type generics — opaque (upstream too) ----
		// Pick / Omit / Partial / Required don't expand without a
		// TypeChecker; rslint treats them as opaque, mirroring upstream's
		// "objectTypeAnnotations.get(name) returns undefined" behavior.
		{Code: `
          type Base = { something: boolean };
          const Hello = (props: Pick<Base, 'something'>) => <div/>;
        `, Tsx: true, Options: optsIs},

		// ---- as const suffix on object literal ----
		{Code: `
          class Hello extends React.Component {
            static propTypes = ({isFoo: PropTypes.bool} as const);
            render() { return <div/>; }
          }
        `, Tsx: true, Options: optsIs},

		// ---- Numeric / unicode regex pattern ----
		// rule pattern that accepts 中文 prefix; matching name passes.
		{Code: `
          class Hello extends React.Component {
            static propTypes = { 是否启用: PropTypes.bool };
            render() { return <div/>; }
          }
        `, Tsx: true, Options: []interface{}{map[string]interface{}{"rule": "^[\\p{L}]+$"}}},

		// ---- Empty propTypeNames → user-cleared list, no PropTypes match ----
		{Code: `
          class Hello extends React.Component {
            static propTypes = { something: PropTypes.bool };
            render () { return <div />; }
          }
        `, Tsx: true, Options: []interface{}{map[string]interface{}{
			"rule":          patternIs,
			"propTypeNames": []interface{}{},
		}}},

		// ---- Empty propTypeNames doesn't disable TS `: boolean` path ----
		// (Documented difference from upstream; lock-in.)
		// Although this is `valid` because the name matches the pattern,
		// the test verifies that the TS path runs even without
		// propTypeNames entries.
		{Code: `
          const Hello = (props: { isFoo: boolean }) => <div/>;
        `, Tsx: true, Options: []interface{}{map[string]interface{}{
			"rule":          patternIs,
			"propTypeNames": []interface{}{},
		}}},

		// ---- 5-level validateNested (depth) ----
		{Code: `
          class Hello extends React.Component {
            render() { return <div/>; }
          }
          Hello.propTypes = {
            isA: PropTypes.shape({
              isB: PropTypes.shape({
                isC: PropTypes.shape({
                  isD: PropTypes.shape({
                    isE: PropTypes.bool
                  })
                })
              })
            })
          };
        `, Tsx: true, Options: []interface{}{map[string]interface{}{
			"rule":           patternIs,
			"validateNested": true,
		}}},

		// ---- Long propTypes object (50 keys) — performance / no crash ----
		// Generated inline so the test reads as one string; all keys
		// match the pattern.
		{Code: `
          class Hello extends React.Component {
            static propTypes = {
              isProp00: PropTypes.bool, isProp01: PropTypes.bool, isProp02: PropTypes.bool, isProp03: PropTypes.bool,
              isProp04: PropTypes.bool, isProp05: PropTypes.bool, isProp06: PropTypes.bool, isProp07: PropTypes.bool,
              isProp08: PropTypes.bool, isProp09: PropTypes.bool, isProp10: PropTypes.bool, isProp11: PropTypes.bool,
              isProp12: PropTypes.bool, isProp13: PropTypes.bool, isProp14: PropTypes.bool, isProp15: PropTypes.bool,
              isProp16: PropTypes.bool, isProp17: PropTypes.bool, isProp18: PropTypes.bool, isProp19: PropTypes.bool,
              isProp20: PropTypes.bool, isProp21: PropTypes.bool, isProp22: PropTypes.bool, isProp23: PropTypes.bool,
              isProp24: PropTypes.bool, isProp25: PropTypes.bool, isProp26: PropTypes.bool, isProp27: PropTypes.bool,
              isProp28: PropTypes.bool, isProp29: PropTypes.bool, isProp30: PropTypes.bool, isProp31: PropTypes.bool,
              isProp32: PropTypes.bool, isProp33: PropTypes.bool, isProp34: PropTypes.bool, isProp35: PropTypes.bool,
              isProp36: PropTypes.bool, isProp37: PropTypes.bool, isProp38: PropTypes.bool, isProp39: PropTypes.bool,
              isProp40: PropTypes.bool, isProp41: PropTypes.bool, isProp42: PropTypes.bool, isProp43: PropTypes.bool,
              isProp44: PropTypes.bool, isProp45: PropTypes.bool, isProp46: PropTypes.bool, isProp47: PropTypes.bool,
              isProp48: PropTypes.bool, isProp49: PropTypes.bool
            };
            render() { return <div/>; }
          }
        `, Tsx: true, Options: optsIs},

		// ---- regex with lookahead (Go RE2 unsupported) → silent no-op ----
		// `^(?=is)` is a JS-only lookahead; Go's regexp.Compile rejects
		// it; the rule degrades silently.
		{Code: `
          class Hello extends React.Component {
            static propTypes = { something: PropTypes.bool };
            render() { return <div/>; }
          }
        `, Tsx: true, Options: []interface{}{map[string]interface{}{"rule": "^(?=is)"}}},

		// ---- Empty file / imports-only file ----
		{Code: ``, Tsx: true, Options: optsIs},
		{Code: `import * as React from 'react';`, Tsx: true, Options: optsIs},

		// ---- .ts file (no JSX) with lower-cased non-component fn ----
		// Lowercase function name fails upstream's capitalized-id check
		// in `getStatelessComponent`; the function is not a component, so
		// nothing is validated.
		{Code: `
          type Props = { something: boolean };
          function helper(props: Props) { return null; }
        `, Options: optsIs},

		// ---- Real-world smoke: large mixed file ----
		// A single file containing several components in different shapes
		// shouldn't false-positive on any unrelated identifier and
		// shouldn't crash on missing pieces.
		{Code: `
          import * as React from 'react';
          import PropTypes from 'prop-types';

          // Class component (matches).
          class Card extends React.Component {
            static propTypes = { isOpen: PropTypes.bool };
            render() { return <div />; }
          }

          // Stateless component (matches).
          const Avatar = ({ isOnline }: { isOnline: boolean }) => <div />;

          // Plain helper class — no propTypes consideration.
          class Repo {
            propTypes = { something: PropTypes.bool };
            run() {}
          }

          // Pure data shape — no React linkage.
          const config = { propTypes: { something: 1 } };

          // Function that isn't a component (no JSX, lower-cased name).
          function helper(props: { isReady: boolean }) { return null; }

          export { Card, Avatar };
        `, Tsx: true, Options: optsIs},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Upstream invalid cases (all migrated, 1:1 with optional Flow→TS).
		// ============================================================

		// ---- createReactClass ----
		{
			Code: `
              var Hello = createReactClass({
                propTypes: {something: PropTypes.bool},
                render: function() { return <div />; }
              });
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Message:   "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
				Line:      3,
				Column:    29,
				EndLine:   3,
				EndColumn: 54,
			}},
		},
		{
			Code: `
              var Hello = createReactClass({
                propTypes: {something: React.PropTypes.bool},
                render: function() { return <div />; }
              });
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              var Hello = React.createClass({
                propTypes: {something: PropTypes.bool},
                render: function() { return <div />; }
              });
            `,
			Tsx:      true,
			Options:  optsIs,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- ES6 class extends React.Component / Component ----
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello.propTypes = {something: PropTypes.bool}
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Message:   "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
				Line:      5,
				Column:    34,
				EndLine:   5,
				EndColumn: 59,
			}},
		},
		{
			Code: `
              class Hello extends Component {
                render () { return <div />; }
              }
              Hello.propTypes = {something: PropTypes.bool}
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = {something: PropTypes.bool};
                render () { return <div />; }
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			// Spread in static propTypes.
			Code: `
              const spreadProps = { aSpreadProp: PropTypes.string };
              class Hello extends Component {
                render () { return <div />; }
              }
              Hello.propTypes = {something: PropTypes.bool, ...spreadProps}
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              const spreadProps = { aSpreadProp: PropTypes.string };
              class Hello extends React.Component {
                static propTypes = {something: PropTypes.bool, ...spreadProps};
                render () { return <div />; }
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			// Class with TS-style `props: {something: boolean}` field
			// (upstream's Flow-equivalent test).
			Code: `
              class Hello extends React.Component {
                props: {something: boolean};
                render () { return <div />; }
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Message:   "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
				Line:      3,
				Column:    25,
				EndLine:   3,
				EndColumn: 43,
			}},
		},

		// ---- Stateless components ----
		{
			Code: `
              var Hello = ({something}) => { return <div /> }
              Hello.propTypes = {something: PropTypes.bool};
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		// Function component with TS type alias on first param (TS form
		// of upstream's Flow `function Hello(props: Props): React.Element`).
		{
			Code: `
              type Props = {
                something: boolean;
              };
              function Hello(props: Props): React.Element { return <div /> }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Custom propTypeNames ----
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = {something: PropTypes.mutuallyExclusiveTrueProps};
                render () { return <div />; }
              }
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"propTypeNames": []interface{}{"bool", "mutuallyExclusiveTrueProps"},
				"rule":          patternIs,
			}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = {
                  something: PropTypes.mutuallyExclusiveTrueProps,
                  somethingElse: PropTypes.bool
                };
                render () { return <div />; }
              }
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"propTypeNames": []interface{}{"bool", "mutuallyExclusiveTrueProps"},
				"rule":          patternIs,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},
		{
			// Bare-identifier values (`mutuallyExclusiveTrueProps`, `bool`).
			Code: `
              class Hello extends React.Component {
                static propTypes = {
                  something: mutuallyExclusiveTrueProps,
                  somethingElse: bool
                };
                render () { return <div />; }
              }
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"propTypeNames": []interface{}{"bool", "mutuallyExclusiveTrueProps"},
				"rule":          patternIs,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ---- propWrapperFunctions matrix ----
		{
			Code: `
              function Card(props) {
                return <div>{props.showScore ? 'yeh' : 'no'}</div>;
              }
              Card.propTypes = merge({}, Card.propTypes, {
                  showScore: PropTypes.bool
              });`,
			Tsx:      true,
			Settings: map[string]interface{}{"propWrapperFunctions": []interface{}{"merge"}},
			Options:  optsIsHas,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			// `Object.assign` configured as `{object, property}` pair.
			Code: `
              function Card(props) {
                return <div>{props.showScore ? 'yeh' : 'no'}</div>;
              }
              Card.propTypes = Object.assign({}, Card.propTypes, {
                  showScore: PropTypes.bool
              });`,
			Tsx: true,
			Settings: map[string]interface{}{"propWrapperFunctions": []interface{}{
				map[string]interface{}{"object": "Object", "property": "assign"},
			}},
			Options: optsIsHas,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			// `_.assign` via `{object, property}`.
			Code: `
              function Card(props) {
                return <div>{props.showScore ? 'yeh' : 'no'}</div>;
              }
              Card.propTypes = _.assign({}, Card.propTypes, {
                  showScore: PropTypes.bool
              });`,
			Tsx: true,
			Settings: map[string]interface{}{"propWrapperFunctions": []interface{}{
				map[string]interface{}{"object": "_", "property": "assign"},
			}},
			Options: optsIsHas,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			// Legacy `"Object.assign"` string shape — split on first dot.
			Code: `
              function Card(props) { return <div /> }
              Card.propTypes = Object.assign({}, Card.propTypes, {
                  showScore: PropTypes.bool
              });`,
			Tsx:      true,
			Settings: map[string]interface{}{"propWrapperFunctions": []interface{}{"Object.assign"}},
			Options:  optsIsHas,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              function Card(props) {
                return <div>{props.showScore ? 'yeh' : 'no'}</div>;
              }
              Card.propTypes = forbidExtraProps({
                  showScore: PropTypes.bool
              });`,
			Tsx:      true,
			Settings: map[string]interface{}{"propWrapperFunctions": []interface{}{"forbidExtraProps"}},
			Options:  optsIsHas,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              class Card extends React.Component {
                render() { return <div>{this.props.showScore ? 'yeh' : 'no'}</div>; }
              }
              Card.propTypes = forbidExtraProps({
                  showScore: PropTypes.bool
              });`,
			Tsx:      true,
			Settings: map[string]interface{}{"propWrapperFunctions": []interface{}{"forbidExtraProps"}},
			Options:  optsIsHas,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              class Card extends React.Component {
                static propTypes = forbidExtraProps({
                  showScore: PropTypes.bool
                });
                render() { return <div /> }
              }`,
			Tsx:      true,
			Settings: map[string]interface{}{"propWrapperFunctions": []interface{}{"forbidExtraProps"}},
			Options:  optsIsHas,
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Custom message ----
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello.propTypes = {something: PropTypes.bool}
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"rule":    patternIs,
				"message": "Boolean prop names must begin with either 'is' or 'has'",
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Message:   "Boolean prop names must begin with either 'is' or 'has'",
			}},
		},
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello.propTypes = {something: PropTypes.bool}
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"rule":    patternIs,
				"message": "It is better if your prop ({{ propName }}) matches this pattern: ({{ pattern }})",
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Message:   "It is better if your prop (something) matches this pattern: (^is[A-Z]([A-Za-z0-9]?)+)",
			}},
		},

		// ---- isRequired chains report on the inner type ----
		{
			Code: `
              var Hello = createReactClass({
                propTypes: {something: PropTypes.bool.isRequired},
                render: function() { return <div />; }
              });
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = { something: PropTypes.bool.isRequired };
                render() { return <div />; }
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              class Hello extends React.Component {
                render() { return <div />; }
              }
              Hello.propTypes = { something: PropTypes.bool.isRequired }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Inline TypeScript parameter type literal ----
		{
			Code: `
              function SomeComponent({ something }: { something: boolean }) {
                  return ( <span>{something}</span> );
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Message:   "Prop name `something` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
				Line:      2,
				Column:    55,
				EndLine:   2,
				EndColumn: 73,
			}},
		},

		// ---- validateNested ----
		{
			Code: `
              class Hello extends React.Component {
                render() { return <div />; }
              }
              Hello.propTypes = {
                isSomething: PropTypes.bool.isRequired,
                nested: PropTypes.shape({ failingItIs: PropTypes.bool })
              };
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"rule":           patternIs,
				"validateNested": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			// Two-level nesting.
			Code: `
              class Hello extends React.Component {
                render() { return <div />; }
              }
              Hello.propTypes = {
                isSomething: PropTypes.bool.isRequired,
                nested: PropTypes.shape({
                  nested: PropTypes.shape({ failingItIs: PropTypes.bool })
                })
              };
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"rule":           patternIs,
				"validateNested": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Bare identifier (imported `bool`) ----
		{
			Code: `
              import { bool } from 'prop-types';
              var Hello = createReactClass({
                propTypes: {something: bool},
                render: function() { return <div />; }
              });
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"rule":           patternIs,
				"validateNested": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- TypeScript type annotations on functional components ----
		{
			Code: `
              type TestConstType = {
                enabled: boolean
              }
              const HelloNew = (props: TestConstType) => { return <div /> };
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              type Props = {
                enabled: boolean
              } & OtherProps
              const HelloNew = (props: Props) => { return <div /> };
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              type Props = {
                enabled: boolean
              } & {
                hasLOL: boolean
              } & OtherProps
              const HelloNew = (props: Props) => { return <div /> };
            `,
			Tsx:     true,
			Options: optsIsHas,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              type Props = {
                enabled: boolean
              }
              const HelloNew: React.FC<Props> = (props) => { return <div /> };
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Message:   "Prop name `enabled` doesn't match rule `^is[A-Z]([A-Za-z0-9]?)+`",
				Line:      3,
				Column:    17,
				EndLine:   3,
				EndColumn: 33,
			}},
		},
		{
			Code: `
              type Props = {
                enabled: boolean
              } & {
                hasLOL: boolean
              }
              const HelloNew: React.FC<Props> = (props) => { return <div /> };
            `,
			Tsx:     true,
			Options: optsIsHas,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              type Props = {
                enabled: boolean
              } | {
                hasLOL: boolean
              }
              const HelloNew = (props: Props) => { return <div /> };
            `,
			Tsx:     true,
			Options: optsIsHas,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              type Props = {
                enabled: boolean
              } & ({
                hasLOL: boolean
              } | {
                lol: boolean
              })
              const HelloNew = (props: Props) => { return <div /> };
            `,
			Tsx:     true,
			Options: optsIsHas,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},
		{
			Code: `
              interface TestFNType {
                enabled: boolean
              }
              const HelloNew = (props: TestFNType) => { return <div /> };
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code:    `const Hello = (props: {enabled:boolean}) => <div />;`,
			Tsx:     true,
			Options: optsIsHas,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              type Props = {
                enabled: boolean
              };
              type BaseProps = {
                semi: boolean
              };
              const Hello = (props: Props & BaseProps) => <div />;
            `,
			Tsx:     true,
			Options: optsIsHas,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},
		{
			Code: `
              type Props = {
                enabled: boolean
              };
              const Hello = (props: Props & {
                semi: boolean
              }) => <div />;
            `,
			Tsx:     true,
			Options: optsIsHas,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ============================================================
		// Additional invalid cases: real-world / lock-in tests.
		// ============================================================

		// ---- ElementAccess on assignment LHS ----
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello['propTypes'] = {something: PropTypes.bool};
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Multi-segment receiver `ns.Hello.propTypes = {...}` ----
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              ns.Hello.propTypes = {something: PropTypes.bool};
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- TS expression wrappers on prop value (mismatch reports) ----
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello.propTypes = {something: (PropTypes.bool as any)}
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello.propTypes = {something: PropTypes.bool!}
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- TypeReference using QualifiedName (`A.B`) ----
		{
			Code: `
              type Inner = { enabled: boolean };
              const Hello = (props: ns.Inner) => <div />;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Interface declaration merging both report ----
		{
			Code: `
              interface Props { enabled: boolean }
              interface Props { switched: boolean }
              const Hello = (props: Props) => <div />;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ---- Type alias indirection (one hop) ----
		{
			Code: `
              type A = B;
              type B = { enabled: boolean };
              const Hello = (props: A) => <div />;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Three-level type composition ----
		{
			Code: `
              type A = { enabled: boolean };
              type B = { hasBar: boolean };
              type C = { lol: boolean };
              const Hello = (props: A & (B | C)) => <div />;
            `,
			Tsx:     true,
			Options: optsIsHas,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ---- Both static propTypes + props: TypeLiteral mismatch ----
		// Both paths must report; reportedSet keeps each diagnostic distinct
		// (different source node) but doesn't deduplicate across paths.
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = { something: PropTypes.bool };
                props: { fooBar: boolean };
                render () { return <div />; }
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ---- HOC wrapping doesn't strip the component name ----
		// Bare `memo` matches the wrapper list only when the binding is
		// destructure-imported from the React pragma; without the
		// import, upstream's `Components.detect` likewise wouldn't
		// classify (matches its `isDestructuredFromPragmaImport` gate).
		{
			Code: `
              import { memo } from 'react';
              const Hello = memo((props) => <div/>);
              Hello.propTypes = {something: PropTypes.bool};
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		// ---- HOC wrapping via React.memo / nested HOC ----
		{
			Code: `
              import * as React from 'react';
              const Hello = React.memo(React.forwardRef((props: { enabled: boolean }, ref) => <div ref={ref}/>));
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
		// ---- HOC wrapping via user-configured `componentWrapperFunctions` ----
		{
			Code: `
              const Hello = myMemo((props: { enabled: boolean }) => <div/>);
            `,
			Tsx:      true,
			Options:  optsIs,
			Settings: map[string]interface{}{"componentWrapperFunctions": []interface{}{"myMemo"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Multi-line TypeLiteral + position assertions ----
		{
			Code: `
              const Hello = (props: {
                enabled: boolean,
                isReady: boolean,
              }) => <div />;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Line:      3,
				Column:    17,
				EndLine:   3,
				EndColumn: 34,
			}},
		},

		// ---- `propTypeNames` only allowing a custom name → bool no longer counts ----
		// Bare identifier `bool` is no longer in propTypeNames, so it isn't
		// even considered; only `mutuallyExclusiveTrueProps` triggers.
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = {
                  something: bool,
                  somethingElse: mutuallyExclusiveTrueProps
                };
                render () { return <div />; }
              }
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"rule":          patternIs,
				"propTypeNames": []interface{}{"mutuallyExclusiveTrueProps"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Interface heritage: `interface Props extends Base` ----
		// Both members reported.
		{
			Code: `
              interface Base { hidden: boolean }
              interface Props extends Base { ready: boolean }
              const Hello = (props: Props) => <div />;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ---- 3-segment QualifiedName: `a.b.C` ----
		{
			Code: `
              type C = { enabled: boolean };
              const Hello = (props: ns.sub.C) => <div />;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Namespace-scoped component ----
		{
			Code: `
              namespace UI {
                export const HelloIn = (props: { enabled: boolean }) => <div/>;
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- React.PropsWithChildren<Props> mismatch ----
		{
			Code: `
              type Props = { enabled: boolean };
              const Hello: React.PropsWithChildren<Props> = (props) => <div/>;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Optional / readonly modifiers don't gate validation ----
		{
			Code: `
              type Props = { enabled?: boolean; readonly visible: boolean };
              const Hello = (props: Props) => <div/>;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ---- Generic component mismatch ----
		{
			Code: `
              function Hello<T>(props: { enabled: boolean; data: T }) { return <div/>; }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- React.FC composed generic mismatch ----
		{
			Code: `
              type Props = { enabled: boolean };
              const Hello: React.FC<Props & { extra: boolean }> = (props) => <div/>;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ---- as const suffix on object literal ----
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = ({something: PropTypes.bool} as const);
                render() { return <div/>; }
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Default-export named class component ----
		{
			Code: `
              export default class Hello extends React.Component {
                static propTypes = { something: PropTypes.bool };
                render() { return <div/>; }
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Anonymous default-export function component ----
		{
			Code: `
              export default function (props: { enabled: boolean }) { return <div/>; }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Named export const arrow component ----
		{
			Code: `
              export const Hello = (props: { enabled: boolean }) => <div/>;
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- HOC + import + .propTypes assignment ----
		{
			Code: `
              import { memo } from 'react';
              const Hello = memo((props: { enabled: boolean }) => <div/>);
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Multi-component file (5 components, mixed shapes) ----
		// Three are invalid (something / enabled / disabled), two are
		// valid (isFoo / hasBar). All five must be detected and only the
		// three mismatches reported, in source order.
		{
			Code: `
              import * as React from 'react';
              import PropTypes from 'prop-types';

              class Card extends React.Component {
                static propTypes = { something: PropTypes.bool };
                render() { return <div/>; }
              }

              const Avatar = (props: { isFoo: boolean }) => <div/>;

              type RowProps = { hasBar: boolean; enabled: boolean };
              const Row: React.FC<RowProps> = (props) => <div/>;

              function Header(props: { disabled: boolean }) { return <div/>; }

              const Footer = React.memo((props: { isOK: boolean }) => <div/>);
            `,
			Tsx:     true,
			Options: optsIsHas,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
				{MessageId: "patternMismatch"},
			},
		},

		// ---- 5-level validateNested mismatch at deepest level ----
		{
			Code: `
              class Hello extends React.Component {
                render() { return <div/>; }
              }
              Hello.propTypes = {
                isA: PropTypes.shape({
                  isB: PropTypes.shape({
                    isC: PropTypes.shape({
                      isD: PropTypes.shape({
                        leaf: PropTypes.bool
                      })
                    })
                  })
                })
              };
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"rule":           patternIs,
				"validateNested": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- TS expression wrapper around the whole assignment RHS ----
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello.propTypes = ({something: PropTypes.bool} as any);
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- ElementAccess with TS-wrapped RHS ----
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello['propTypes'] = ({something: PropTypes.bool}!);
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Settings: custom pragma changes class detection ----
		// `extends Preact.Component` only counts when pragma=Preact.
		{
			Code: `
              class Hello extends Preact.Component {
                static propTypes = { something: PropTypes.bool };
                render() { return <div/>; }
              }
            `,
			Tsx:      true,
			Options:  optsIs,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},

		// ---- Custom message with unknown placeholder kept literal ----
		// Mirrors ESLint's `interpolate`: keys not in `data` stay as
		// `{{key}}` in the rendered output rather than being deleted or
		// throwing.
		{
			Code: `
              class Hello extends React.Component {
                render () { return <div />; }
              }
              Hello.propTypes = {something: PropTypes.bool}
            `,
			Tsx: true,
			Options: []interface{}{map[string]interface{}{
				"rule":    patternIs,
				"message": "{{ propName }} fails {{ pattern }} (unknown: {{ unknown }})",
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "patternMismatch",
				Message:   "something fails ^is[A-Z]([A-Za-z0-9]?)+ (unknown: {{ unknown }})",
			}},
		},

		// ---- Double TS expression wrappers on prop value ----
		// `((PropTypes.bool as any) satisfies any)` — peeled correctly
		// by reactutil.SkipExpressionWrappers, name still reported.
		{
			Code: `
              class Hello extends React.Component {
                static propTypes = {something: ((PropTypes.bool as any) satisfies any)};
                render () { return <div />; }
              }
            `,
			Tsx:     true,
			Options: optsIs,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "patternMismatch"}},
		},
	})
}
