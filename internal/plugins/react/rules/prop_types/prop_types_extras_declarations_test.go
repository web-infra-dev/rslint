package prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPropTypesRuleExtrasDeclarations locks in runtime declarations, validator
// aliases, wrapper syntax, and TypeScript inference beyond the upstream suite.
// Component ownership and parameter bindings live in prop_types_extras_bindings_test.go.
func TestPropTypesRuleExtrasDeclarations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PropTypesRule,
		[]rule_tester.ValidTestCase{
			// ---- Validator aliases and opaque runtime declarations ----
			{Code: `const base = PropTypes.string; const alias = base; function Hello(props) { return <div>{props.a}{props.b}</div>; } Hello.propTypes = { a: alias, b: alias };`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = condition ? typesA : typesB;`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = typesA || typesB;`, Tsx: true},

			// ---- Qualified types, wrapper keys, and dynamic accesses ----
			{Code: `type Props = { local: string }; function Hello(props: Models.Props) { return <div>{props.missing}</div>; }`, Tsx: true},
			{Code: `type Props = { name: string }; const Hello = React.forwardRef<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.name}</div>);`, Tsx: true},
			{Code: `type Props = { name: string }; const Hello = React['forwardRef']<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.missing}</div>);`, Tsx: true},
			{Code: `type Props = { name: string }; const Hello = React[(forwardRef as unknown)]<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.missing}</div>);`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello[propTypes] = { name: PropTypes.string };`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello[propTypes].name = PropTypes.string;`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes[name] = PropTypes.string;`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props[key].child}</div>; }`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.user[key].child}</div>; } Hello.propTypes = { user: PropTypes.shape({}) };`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = null;`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = new Types();`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = () => types;`, Tsx: true},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = [types];`, Tsx: true},
			{Code: `function Same(props) { return <div>{props.outer}</div>; } { function Same(props) { return <div>{props.inner}</div>; } Same.propTypes = { inner: PropTypes.string }; } { Same.propTypes = { outer: PropTypes.string }; }`, Tsx: true},

			// ---- Type/runtime declaration composition ----
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = ({ name: PropTypes.string } as const);`, Tsx: true},
			{Code: `function Outer(props) { function Inner(other) { return <span>{props.name}</span>; } Inner.propTypes = {}; return <Inner />; } Outer.propTypes = { name: PropTypes.string }; Outer.propTypes = condition ? typesA : typesB;`, Tsx: true},
			{Code: `function Outer(props) { function Inner(other) { return <span>{props.name}</span>; } Inner.propTypes = {}; return <Inner />; } Outer.propTypes = { name: PropTypes.string, ...shared };`, Tsx: true},
			{Code: `type Props = { typed: string }; class Hello extends React.Component<Props> { static propTypes = { runtime: PropTypes.string }; render() { return <div>{this.props.typed}{this.props.runtime}</div>; } }`, Tsx: true},
			{Code: `type Props = { typed: string }; class Hello extends React.Component<Props> { static propTypes: {}; render() { return <div>{this.props.typed}</div>; } }`, Tsx: true},
			{Code: `type Props = ReturnType<typeof missingSelector>; function Hello(props: Props) { return <div>{props.name}</div>; }`, Tsx: true},
			{Code: `class Hello extends React.Component { static get propTypes() {} render() { return <div>{this.props.name}</div>; } }`, Options: map[string]any{"skipUndeclared": true}, Tsx: true},
			{Code: `type Props = { name: string }; class Hello extends React.Component<Props> { render() { return <>{this.props.name}{this.props.missing}</>; } } Hello.propTypes = condition ? typesA : typesB;`, Tsx: true},

			// ---- forwardRef type precedence and partial assignments ----
			{Code: `type Param = { param: string }; type Generic = { generic: string }; const Hello = React.forwardRef<HTMLDivElement, Generic>((props: Param, ref) => <div ref={ref}>{props.generic}</div>);`, Tsx: true},
			{Code: `type Variable = { variable: string }; const Hello: React.FC<Variable> = React.forwardRef((props, ref) => <div ref={ref}>{props.variable}{props.missing}</div>);`, Options: map[string]any{"skipUndeclared": true}, Tsx: true},
			{Code: `function Hello(props) { return <>{props.a.b}{props.a.c}{props.a.d}</>; } Hello.propTypes.a.b = PropTypes.string; Hello.propTypes = { a: PropTypes.shape({ c: PropTypes.string }) };`, Tsx: true},
			{Code: `function Hello(props) { return <>{props.a.b.c}{props.other}</>; } Hello.propTypes = { a: PropTypes.shape({}) }; Hello.propTypes.a.b.c = PropTypes.string;`, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			// ---- Unresolved declarations and ReturnType inference ----
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = makePropTypes();`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `const makeProps = () => { return { name: 'x' }; }; type Props = ReturnType<typeof makeProps>; function Hello(props: Props) { return <div>{props.name}{props.missing}</div>; }`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `const makeProps = () => { return { first: 'x' }; return { last: 'x' }; }; type Props = ReturnType<typeof makeProps>; function Hello(props: Props) { return <div>{props.first}{props.last}</div>; }`, Tsx: true, Errors: missingPropTypes("'first' is missing in props validation")},
			{Code: `type Props = { name: string }; const makeProps = () => factory<Props>(); type Made = ReturnType<typeof makeProps>; function Hello(props: Made) { return <div>{props.name}{props.missing}</div>; }`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `type Props = { name: string }; type Made = ReturnType<() => Props>; function Hello(props: Made) { return <div>{props.name}{props.missing}</div>; }`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `import React from 'react'; type Props = { name: string }; const makeProps = () => factory<React.FC<Props>>(); type Made = ReturnType<typeof makeProps>; function Hello(props: Made) { return <div>{props.name}{props.missing}</div>; }`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `import React from 'react'; type Props = { name: string }; type Made = ReturnType<() => React.FC<Props>>; function Hello(props: Made) { return <div>{props.name}{props.missing}</div>; }`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `declare const makeProps: () => ({ name: string }); type Props = ReturnType<typeof makeProps>; function Hello(props: Props) { return <div>{props.name}{props.missing}</div>; }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation", "'missing' is missing in props validation")},
			{Code: `const makeProps = {}; type Props = ReturnType<typeof makeProps>; function Hello(props: Props) { return <div>{props.name}</div>; }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `Hello.propTypes = condition ? typesA : typesB; type Props = { name: string }; function Hello(props: Props) { return <>{props.name}{props.missing}</>; }`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `type Props = { name: string }; const Hello = class extends React.Component<Props> { render() { return <>{this.props.name}{this.props.missing}</>; } }; Hello.propTypes = condition ? typesA : typesB;`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},

			// ---- Ordered partial assignments and computed propTypes keys ----
			{Code: `function Hello(props) { return <>{props.a.b}{props.a.c}{props.a.d}</>; } Hello.propTypes = { a: PropTypes.shape({ c: PropTypes.string }) }; Hello.propTypes.a.b = PropTypes.string;`, Tsx: true, Errors: missingPropTypes("'a.d' is missing in props validation")},
			{Code: `function Hello(props) { return <>{props.a.b}{props.a.d}</>; } Hello.propTypes.a = PropTypes.shape({ b: PropTypes.string });`, Tsx: true, Errors: missingPropTypes("'a.d' is missing in props validation")},
			{Code: `function Hello(props) { return <>{props.a.b}{props.other}</>; } Hello.propTypes = { a: PropTypes.object }; Hello.propTypes.a.b = PropTypes.string;`, Tsx: true, Errors: missingPropTypes("'other' is missing in props validation")},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello['propTypes'] = { name: PropTypes.string };`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello['propTypes'].name = PropTypes.string;`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},

			// ---- forwardRef, imported React generics, and key syntax ----
			{Code: `type Props = { name: string }; const Hello = React.forwardRef<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.missing}</div>);`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `type Param = { param: string }; type Generic = { generic: string }; type Variable = { variable: string }; const Hello: React.FC<Variable> = React.forwardRef<HTMLDivElement, Generic>((props: Param, ref) => <>{props.param}{props.generic}{props.variable}</>);`, Tsx: true, Errors: missingPropTypes("'param' is missing in props validation", "'variable' is missing in props validation")},
			{Code: `type Param = { param: string }; type Variable = { variable: string }; const Hello: React.FC<Variable> = React.forwardRef<HTMLDivElement>((props: Param, ref) => <>{props.param}{props.variable}</>);`, Tsx: true, Errors: missingPropTypes("'param' is missing in props validation", "'variable' is missing in props validation")},
			{Code: `type Variable = { variable: string }; const Hello: React.FC<Variable> = React.forwardRef((props, ref) => <>{props.variable}{props.missing}</>);`, Tsx: true, Errors: missingPropTypes("'variable' is missing in props validation", "'missing' is missing in props validation")},
			{Code: `type Props = { name: string }; const Hello = React[forwardRef]<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.name}{props.missing}</div>);`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `type Props = { name: string }; const Hello = React[forwardRef]<HTMLDivElement>((props: Props, ref) => <div ref={ref}>{props.name}{props.missing}</div>);`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation", "'missing' is missing in props validation")},
			{Code: `type Props = { name: string }; const Hello = React[forwardRef]((props: Props, ref) => <div ref={ref}>{props.name}{props.missing}</div>);`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `type Props = { name: string }; const Hello = (React.forwardRef)?.<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.name}{props.missing}</div>);`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `type Props = { name: string }; const Hello = (React[(forwardRef)])?.<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.name}{props.missing}</div>);`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `const Hello = (React.forwardRef)?.((props, ref) => <div ref={ref}>{props.name}</div>);`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `import { forwardRef } from 'react'; type Props = { name: string }; const Hello = forwardRef?.<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.name}{props.missing}</div>);`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `function Hello(props) { return <div>{props.name}</div>; } Hello[(propTypes as unknown)] = { name: PropTypes.string };`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `import React from 'react'; type Props = { name: string }; const Hello: React.FC<Props> = (props) => <>{props.name}{props.children}{props.missing}</>;`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `import React from 'react'; type Props = { name: string }; const Hello: React.ForwardRefRenderFunction<HTMLDivElement, Props> = (props, ref) => <div ref={ref}>{props.name}{props.children}{props.missing}</div>;`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},

			// ---- No-initializer fields, computed accesses, and declaration replay ----
			{Code: `class Hello extends React.Component { static propTypes: { name: string }; render() { return <div>{this.props.name}</div>; } }`, Options: map[string]any{"skipUndeclared": true}, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { static propTypes!: { name: string }; render() { return <div>{this.props.name}</div>; } }`, Options: map[string]any{"skipUndeclared": true}, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `function Hello(props) { return <div>{props.user[key].child}</div>; }`, Tsx: true, Errors: missingPropTypes("'user' is missing in props validation", "'user[].child' is missing in props validation")},
			{Code: `function Hello(props) { return <div>{props[123].child}</div>; }`, Tsx: true, Errors: missingPropTypes("'123' is missing in props validation", "'123.child' is missing in props validation")},
			{Code: `function Hello(props) { return <div>{props[123].child}</div>; } Hello.propTypes = { 123: PropTypes.shape({}) };`, Tsx: true, Errors: missingPropTypes("'123.child' is missing in props validation")},
			{Code: `function Outer(props) { function Inner(other) { return <span>{props.name}</span>; } Inner.propTypes = {}; return <Inner />; } Outer.propTypes = condition ? typesA : typesB;`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `function Outer(props) { function Inner(other) { return <span>{props.missing}</span>; } Inner.propTypes = {}; return <Inner />; } Outer.propTypes = { known: PropTypes.string, ...shared };`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `const makeProps = () => ({ known: '', ...shared }); type Props = ReturnType<typeof makeProps>; function Outer(props: Props) { function Inner(other) { return <span>{props.missing}</span>; } Inner.propTypes = {}; return <Inner />; }`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `let types; function Hello(props) { return <div>{props.name}</div>; } Hello.propTypes = types;`, Options: map[string]any{"skipUndeclared": true}, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `type Props = { typed: string }; class Hello extends React.Component<Props> { static propTypes = { runtime: PropTypes.string }; render() { return <div>{this.props.typed}{this.props.runtime}{this.props.missing}</div>; } }`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `function Hello(props) { return <>{props.user.name}{props.user.age}</>; } Hello.propTypes = { user: PropTypes.shape({ name: PropTypes.string }) }; Hello.propTypes = { user: PropTypes.shape({ age: PropTypes.string }) };`, Tsx: true, Errors: missingPropTypes("'user.name' is missing in props validation")},
			{Code: `Hello.propTypes = { user: PropTypes.shape({ before: PropTypes.string }) }; class Hello extends React.Component { static propTypes = { user: PropTypes.shape({ inside: PropTypes.string }) }; render() { return <>{this.props.user.before}{this.props.user.inside}{this.props.user.after}</>; } }`, Tsx: true, Errors: missingPropTypes("'user.before' is missing in props validation", "'user.after' is missing in props validation")},
			{Code: `class Hello extends React.Component { static get propTypes() { return; } render() { return <div>{this.props.name}</div>; } }`, Options: map[string]any{"skipUndeclared": true}, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
		},
	)
}

func missingPropTypes(messages ...string) []rule_tester.InvalidTestCaseError {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(messages))
	for _, message := range messages {
		errors = append(errors, rule_tester.InvalidTestCaseError{MessageId: "missingPropType", Message: message})
	}
	return errors
}
