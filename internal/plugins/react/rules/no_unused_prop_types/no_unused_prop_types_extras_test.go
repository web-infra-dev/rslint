// TestNoUnusedPropTypesExtras locks in tsgo edge shapes and branches that the
// upstream suite does not fully exercise. The upstream-mirror cases live in
// no_unused_prop_types_upstream_test.go.
package no_unused_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedPropTypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedPropTypesRule,
		[]rule_tester.ValidTestCase{
			{Code: `type Param = { param: string }; type Generic = { generic: string }; const Foo = React.forwardRef<HTMLDivElement, Generic>((props: Param, ref) => <div>{props.generic}</div>);`, Tsx: true},
			{Code: `function Foo(props) { return <div>{props.outer}</div>; } Foo.propTypes = { outer: PropTypes.arrayOf(Custom.shape({ unused: Custom.string }).isRequired) };`, Options: map[string]any{"customValidators": []any{"Custom"}, "skipShapeProps": false}, Tsx: true},
			{Code: `const Foo = React.memo(React.forwardRef((props, ref) => <div>{props.name}</div>)); Foo.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `type Props = { name: string }; const Foo = React.memo(React.forwardRef<HTMLDivElement, Props>(({ name }, ref) => <div>{name}</div>));`, Tsx: true},
			{Code: `const Foo = React.memo(React.forwardRef(function(props, ref) { return <div>{props.name}</div>; })); Foo.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `type Props = { name: string }; const Foo = React.memo(React.forwardRef<HTMLDivElement, Props>(function({ name }, ref) { return <div>{name}</div>; }));`, Tsx: true},
			{Code: `function Foo(props) { return <div>{props.outer}</div>; } Foo.propTypes = { outer: PropTypes.arrayOf(Custom.shape({ inner: Custom.string })) };`, Options: map[string]any{"customValidators": []any{"Custom"}, "skipShapeProps": false}, Tsx: true},
			{Code: `function Foo(props) { return <div>{props.outer}</div>; } Foo.propTypes = { outer: PropTypes.oneOfType([Custom.shape({ inner: Custom.string })]) };`, Options: map[string]any{"customValidators": []any{"Custom"}, "skipShapeProps": false}, Tsx: true},
			// Requiring a prop does not change a custom validator's opaque semantics.
			{Code: `function Foo(props) { return <div>{props.outer}</div>; } Foo.propTypes = { outer: Custom.shape({ inner: Custom.string }).isRequired };`, Options: map[string]any{"customValidators": []any{"Custom"}, "skipShapeProps": false}, Tsx: true},
			{Code: `type Props = { user: { used: string; unused: string } }; const Foo = (props: Props) => <div>{props.user.used}</div>;`, Tsx: true},
			{Code: `interface Props { user: { used: string; unused: string } } const Foo = (props: Props) => <div>{props.user.used}</div>;`, Options: map[string]any{"skipShapeProps": false}, Tsx: true},
			{Code: `type Extra = { phantom: string }; const Foo: Wrapper<Extra> = () => <div />;`, Tsx: true},
			{Code: `import { FC } from 'other'; type Props = { unused: string }; const Foo: FC<Props> = () => <div />;`, Tsx: true},
			{Code: `type FC<T> = () => unknown; type Props = { unused: string }; const Foo: FC<Props> = () => <div />;`, Tsx: true},
			{Code: `function Foo(props) { const x = props.a; function helper() { const x = props.b; return x.other; } return <div>{x.used}</div>; } Foo.propTypes = { a: PropTypes.shape({ used: PropTypes.string }), b: PropTypes.shape({ other: PropTypes.string }) };`, Options: map[string]any{"skipShapeProps": false}, Tsx: true},
			{Code: `function Foo(props) { const x = props.a; { const x = props.b; x.other; } return <div>{x.used}</div>; } Foo.propTypes = { a: PropTypes.shape({ used: PropTypes.string }), b: PropTypes.shape({ other: PropTypes.string }) };`, Options: map[string]any{"skipShapeProps": false}, Tsx: true},
			{Code: `const shared = { unused: PropTypes.string }; function factory(shared) { function Foo() { return <div />; } Foo.propTypes = shared; }`, Tsx: true},
			{Code: `const a = b; const b = a; function Foo() { return <div />; } Foo.propTypes = a;`, Tsx: true},
			{Code: `function Foo(props) { const key = unknown; const { [key]: value } = props; return <div>{value}</div>; } Foo.propTypes = { unused: PropTypes.string };`, Tsx: true},
			{Code: `function Foo(props) { return <div {...other} />; } Foo.propTypes = { unused: PropTypes.string };`, Tsx: true},
			{Code: `class Foo extends React.Component { static propTypes = { used: PropTypes.string }; render() { const { props } = this; return <div>{props.used}</div>; } }`, Tsx: true},
			{Code: `const makeProps = () => bindActionCreators({ unused: fn }, dispatch); type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div />;`, Tsx: true},
			{Code: `const makeProps = () => ({ ...bindActionCreators({ unused: fn }, dispatch) }); type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div />;`, Tsx: true},
			// Own prop declarations can legitimately shadow Object.prototype methods.
			{Code: `class Foo extends React.Component { static propTypes = { toString: PropTypes.func, constructor: PropTypes.func }; render() { return <div>{this.props.toString()}{String(this.props.constructor)}</div>; } }`, Tsx: true},
			{Code: `class Comp1 extends React.Component { render() { return <span>{this.props.prop1}</span>; } } Comp1.propTypes = { prop1: PropTypes.number }; class Comp2 extends React.Component { static propTypes = { prop2: PropTypes.arrayOf(Comp1.propTypes.prop1) }; render() { return <span>{this.props.prop2}</span>; } }`, Tsx: true},
			// Typed helpers are not components merely because their parameter has a type.
			{Code: `function helper(props: { used: string; unused: string }) { return props.used.length; } helper.propTypes = { used: PropTypes.string, unused: PropTypes.string };`, Tsx: true},
			// Same-name type aliases resolve from the nearest lexical scope.
			{Code: `{ type Props = { first: string }; const Foo = (props: Props) => <div>{props.first}</div>; } { type Props = { second: string }; const Foo = (props: Props) => <div>{props.second}</div>; }`, Tsx: true},
			// Configured custom validators remain opaque, including shape-like method names.
			{Code: `function Foo(props) { return <div>{props.outer}</div>; } Foo.propTypes = { outer: CustomValidator.shape({ inner: CustomValidator.string }) };`, Options: map[string]interface{}{"customValidators": []interface{}{"CustomValidator"}, "skipShapeProps": false}, Tsx: true},
			// ---- Dimension 4: parenthesized and TypeScript expression wrappers ----
			{Code: `class Hello extends React.Component { render() { return <div>{(this.props).name}</div>; } } Hello.propTypes = ({ name: PropTypes.string });`, Tsx: true},
			{Code: `const Hello = ((props: Props) => <div>{props.name}</div>) as any; type Props = { name: string }; Hello.propTypes = { name: PropTypes.string };`, Tsx: true},
			// ---- Dimension 4: computed/static property keys ----
			{Code: `class Hello extends React.Component { render() { return <div>{this.props["name"]}</div>; } } Hello.propTypes = { ["name"]: PropTypes.string };`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>{this.props[0]}</div>; } } Hello.propTypes = { 0: PropTypes.string };`, Tsx: true},
			// ---- Dimension 4: optional-chain links ----
			{Code: `class Hello extends React.Component { static propTypes = { name: PropTypes.string }; render() { return <div>{this.props?.name}</div>; } }`, Tsx: true},
			// ---- Real-user: nested helper reads ----
			{Code: `function Hello(props) { function renderLabel() { return <span>{props.label}</span>; } return <div>{renderLabel()}</div>; } Hello.propTypes = { label: PropTypes.string };`, Tsx: true},
			// Nested helpers retain eslint-plugin-react's conventional `props` alias.
			{Code: `function Foo() { function helper(props) { return props.unused; } return <div />; } Foo.propTypes = { unused: PropTypes.string };`, Tsx: true},
			// Callback parameters named `props` retain the same conventional alias.
			{Code: `function Foo() { return [1].map((props) => <div>{props.used}</div>); } Foo.propTypes = { used: PropTypes.string };`, Tsx: true},
			// Each member-path propTypes assignment resolves through its own root binding.
			{Code: `{ const obj = { Foo: function(props) { return <div>{props.first}</div>; } }; obj.Foo.propTypes = { first: PropTypes.string }; } { const obj = { Foo: function(props) { return <div>{props.second}</div>; } }; obj.Foo.propTypes = { second: PropTypes.string }; }`, Tsx: true},
			// Lifecycle class fields receive the same props aliases as methods.
			{Code: `class Foo extends React.Component { static propTypes = { used: PropTypes.string }; componentDidUpdate = (nextProps) => nextProps.used; render() { return <div />; } }`, Tsx: true},
			{Code: `class Foo extends React.Component { static propTypes = { used: PropTypes.string }; UNSAFE_componentWillReceiveProps = ({ used }) => used; render() { return <div />; } }`, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.3.0"}}, Tsx: true},
			// ---- Real-user: rest props make the accessed set open ----
			{Code: `function Hello({ name, ...rest }) { return <div>{rest.other}</div>; } Hello.propTypes = { name: PropTypes.string, other: PropTypes.string };`, Tsx: true},
			// Locks in upstream reportUnusedPropType()'s `props === true` arm.
			{Code: `class Hello extends React.Component { render() { return <div>{this.props.name}</div>; } } Hello.propTypes = { name: PropTypes.string, ...external };`, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			{Code: `class Foo extends React.Component { static get propTypes() { return shared; } render() { return <div />; } } const shared = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `import { forwardRef } from 'react'; type Props = { used: string; unused: string }; const Foo = forwardRef<HTMLDivElement, Props>(({ used }, ref) => <div>{used}</div>);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `type P = { unused: string }; const makeProps = () => { if (flag) return factory<P>(); return factory<P>(); }; type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `function Foo(props) { return <div>{props.outer}</div>; } Foo.propTypes = { outer: PropTypes.arrayOf(PropTypes.shape({ unused: PropTypes.string }).isRequired) };`, Options: map[string]any{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'outer.*.unused' PropType is defined but prop is never used"}}},
			{Code: `function Foo(props) { return <div>{props.outer}</div>; } Foo.propTypes = { outer: PropTypes.arrayOf(PropTypes.arrayOf(PropTypes.shape({ unused: PropTypes.string }))) };`, Options: map[string]any{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'outer.*.*.unused' PropType is defined but prop is never used"}}},
			{Code: `function Foo() { return <div />; } Foo.propTypes = { outer: PropTypes.shape({ unused: PropTypes.string }) };`, Options: map[string]any{"ignore": []any{"outer"}, "skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'outer.unused' PropType is defined but prop is never used"}}},
			{Code: `const Foo = React.memo(React.forwardRef((props, ref) => <div>{props.name}</div>)); Foo.propTypes = { name: PropTypes.string, unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `type Props = { used: string; unused: string }; const Foo = React.memo(React.forwardRef<HTMLDivElement, Props>(({ used }, ref) => <div>{used}</div>));`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `import React from 'react'; type Props = { used: string; unused: string }; const Foo = ({ used }: React.PropsWithChildren<Props>) => <div>{used}</div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `import { PropsWithChildren as WithChildren } from 'react'; type Props = { used: string; unused: string }; const Foo = ({ used }: WithChildren<Props>) => <div>{used}</div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `import * as R from 'react'; type Props = { used: string; unused: string }; const Foo = ({ used }: R.PropsWithChildren<Props>) => <div>{used}</div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `import React from 'react'; type Props = { used: string; unused: string }; const Foo = React.forwardRef<HTMLDivElement, React.PropsWithChildren<Props>>(({ used }, ref) => <div>{used}</div>);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// A React component annotation declares props even without a parameter.
			{Code: `import { FC as C } from 'react'; type Props = { unused: string }; const Foo: C<Props> = () => <div />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `import React from 'react'; type Props = { unused: string }; const Foo: React.ForwardRefRenderFunction<HTMLDivElement, Props> = () => <div />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `const makeProps = () => bindActionCreators<{ used: () => void; unused: () => void }>({}, dispatch); type Props = ReturnType<typeof makeProps>; const Foo = ({ used }: Props) => <div>{String(used)}</div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `const makeProps = () => { return factory<{ unused: string }>(); }; type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `const makeProps = () => ({ ...factory<{ unused: string }>() }); type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `const unused = ''; const makeProps = () => ({ unused }); type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `const makeProps = () => ({ unused() {} }); type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div />;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `const makeProps = (x: boolean) => { switch (x) { case true: return { used: '' }; default: return { unused: '' }; } }; type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div>{props.used}</div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Declaration-site lexical scope wins over an unrelated component-local name.
			{Code: `const shared = { first: PropTypes.string }; function Foo() { const shared = { second: PropTypes.string }; return <div />; } Foo.propTypes = shared;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'first' PropType is defined but prop is never used"}}},
			{Code: `var shared = { first: PropTypes.string }; var shared = { second: PropTypes.string }; function Foo() { return <div />; } Foo.propTypes = shared;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'second' PropType is defined but prop is never used"}}},
			{Code: `function Foo(props) { function helper() { const x = props.a; return null; } return <div>{x.used}</div>; } Foo.propTypes = { a: PropTypes.shape({ used: PropTypes.string }) };`, Options: map[string]any{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'a.used' PropType is defined but prop is never used"}}},
			// Unrelated local variables and parameters cannot consume an outer props alias.
			{Code: `function Foo(props) { const x = props.a; function helper(x) { return x.used; } return <div />; } Foo.propTypes = { a: PropTypes.shape({ used: PropTypes.string }) };`, Options: map[string]any{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'a.used' PropType is defined but prop is never used"}}},
			{Code: `function Foo(props) { function helper() { const props = { used: 1 }; return props.used; } return <div />; } Foo.propTypes = { used: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'used' PropType is defined but prop is never used"}}},
			{Code: `class Foo extends React.Component { static propTypes = { unused: PropTypes.string }; render() { this.props; return <div />; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `function Foo(props) { const copy = { ...props }; return <div />; } Foo.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `function Foo(props) { const copy = [...props]; return <div />; } Foo.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `function Foo(props) { fn(...props); return <div />; } Foo.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Static computed spellings denote the same component and propTypes property.
			{Code: `class Foo extends React.Component { static ['propTypes'] = { unused: PropTypes.string }; render() { return <div />; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			{Code: `const obj = { ['Foo']: function() { return <div />; } }; obj.Foo.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Shadowed components must receive their own propTypes assignment.
			{Code: `{ function Foo(props) { return <div>{props.second}</div>; } Foo.propTypes = { first: PropTypes.string, second: PropTypes.string }; } { function Foo(props) { return <div>{props.first}</div>; } Foo.propTypes = { first: PropTypes.string, second: PropTypes.string }; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'first' PropType is defined but prop is never used"}, {MessageId: "unusedPropType", Message: "'second' PropType is defined but prop is never used"}}},
			// Only callbacks with a props-derived receiver may destructure prop paths.
			{Code: `function Foo(props) { [1].map(({ used }) => used); return <div />; } Foo.propTypes = { used: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'used' PropType is defined but prop is never used"}}},
			// User-configured component wrappers participate in component discovery.
			{Code: `const Foo = observer((props) => <div />); Foo.propTypes = { unused: PropTypes.string };`, Settings: map[string]interface{}{"componentWrapperFunctions": []interface{}{"observer"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// UNSAFE lifecycle aliases are unavailable before React 16.3.
			{Code: `class Foo extends React.Component { static propTypes = { value: PropTypes.string }; UNSAFE_componentWillReceiveProps(nextProps) { return nextProps.value; } render() { return <div />; } }`, Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.2.0"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'value' PropType is defined but prop is never used"}}},
			// TypeScript method signatures are declared props too.
			{Code: `interface Props { used: string; onClick(): void } function Foo(props: Props) { return <div>{props.used}</div>; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'onClick' PropType is defined but prop is never used"}}},
			// Synthetic array children report at their validator argument, not the outer key.
			{Code: "class Comp1 extends React.Component {\n  render() { return <span>{this.props.prop1}</span>; }\n}\nComp1.propTypes = { prop1: PropTypes.number };\nclass Comp2 extends React.Component {\n  static propTypes = {\n    prop2: PropTypes.arrayOf(\n    Comp1.propTypes.prop1\n    )\n  };\n  render() { return <span />; }\n}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'prop2' PropType is defined but prop is never used", Line: 7, Column: 5, EndLine: 7, EndColumn: 10}, {MessageId: "unusedPropType", Message: "'prop2.*' PropType is defined but prop is never used", Line: 8, Column: 5, EndLine: 8, EndColumn: 26}}},
			// Nested assignment diagnostics cover the assignment, not the terminal key.
			{Code: "class Foo extends React.Component {\n  render() { return <div />; }\n}\nFoo.propTypes = { a: PropTypes.shape({ b: PropTypes.shape({}) }) };\nFoo.propTypes.a.b.c = PropTypes.number;", Options: map[string]interface{}{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'a' PropType is defined but prop is never used", Line: 4, Column: 19, EndLine: 4, EndColumn: 20}, {MessageId: "unusedPropType", Message: "'a.b' PropType is defined but prop is never used", Line: 4, Column: 40, EndLine: 4, EndColumn: 41}, {MessageId: "unusedPropType", Message: "'a.b.c' PropType is defined but prop is never used", Line: 5, Column: 1, EndLine: 5, EndColumn: 39}}},
			// Locks in upstream isPropUsed()'s nested-name comparison arm.
			{Code: `class Hello extends React.Component { render() { return <div>{this.props.user.name}</div>; } } Hello.propTypes = { user: PropTypes.shape({ first: PropTypes.string, last: PropTypes.string }) };`, Options: map[string]interface{}{"skipShapeProps": false}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'user.first' PropType is defined but prop is never used"}, {MessageId: "unusedPropType", Message: "'user.last' PropType is defined but prop is never used"}}},
			// Locks in the custom validator option branch.
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = { unused: Custom.string };`, Options: map[string]interface{}{"customValidators": []interface{}{"Custom"}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Locks in the direct propTypes member-assignment branch.
			{Code: `class Hello extends React.Component { render() { return <div />; } } Hello.propTypes = {}; Hello.propTypes.unused = PropTypes.string;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Only a function component's first parameter is its props object.
			{Code: `function Foo(props, options) { return <div>{options.used}</div>; } Foo.propTypes = { used: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'used' PropType is defined but prop is never used"}}},
			// Lifecycle-looking methods on unrelated objects do not receive component props.
			{Code: `function Foo() { const helper = { componentDidUpdate(nextProps) { return nextProps.used; } }; return <div />; } Foo.propTypes = { used: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'used' PropType is defined but prop is never used"}}},
			// Only the updater (first) callback of setState receives props as its second argument.
			{Code: `class Foo extends React.Component { static propTypes = { used: PropTypes.string }; render() { this.setState({}, (state, props) => props.used); return <div />; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'used' PropType is defined but prop is never used"}}},
			// Validator callback identifiers are not aliases for the component props object.
			{Code: `class Foo extends React.Component { static propTypes = { a: PropTypes.string, b: (props) => props.b }; render() { return <div>{this.props.a}</div>; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'b' PropType is defined but prop is never used"}}},
			// Object-literal methods are method-valued propTypes declarations.
			{Code: `function Foo() { return <div />; } Foo.propTypes = { unused() {} };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// An arbitrary first parameter name is not a component props alias.
			{Code: `function Foo(p) { return <div>{p.used}</div>; } Foo.propTypes = { used: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'used' PropType is defined but prop is never used"}}},
			// ReturnType derives props from block-bodied function producers.
			{Code: `const makeProps = function() { return { used: '', unused: '' }; }; type Props = ReturnType<typeof makeProps>; const Foo = (props: Props) => <div>{props.used}</div>;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Non-static computed declarations retain their identifier key.
			{Code: `const key = 'unused'; function Foo() { return <div />; } Foo.propTypes = { [key]: PropTypes.string, other: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'key' PropType is defined but prop is never used"}, {MessageId: "unusedPropType", Message: "'other' PropType is defined but prop is never used"}}},
			// Components assigned to object members retain their full assignment path.
			{Code: `const obj = { Foo: function(props) { return <div />; } }; obj.Foo.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Object-literal methods and nested paths are component targets too.
			{Code: `const root = { nested: { Foo() { return <div />; } } }; root.nested.Foo.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// Member assignments also retain the complete component path.
			{Code: `const obj = {}; obj.Foo = function() { return <div />; }; obj.Foo.propTypes = { unused: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'unused' PropType is defined but prop is never used"}}},
			// bindActionCreators derives ReturnType props from explicit type arguments.
			{Code: `type DispatchProps = ReturnType<typeof mapDispatchToProps>; const Component = ({ runtimeOnly }: DispatchProps) => <div>{runtimeOnly}</div>; const mapDispatchToProps = () => ({ ...bindActionCreators<{ typedOnly: () => void }>({ runtimeOnly: fn }, dispatch) });`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedPropType", Message: "'typedOnly' PropType is defined but prop is never used"}}},
		},
	)
}
