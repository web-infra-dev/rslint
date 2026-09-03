package prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPropTypesRuleExtrasBindings locks in component ownership, lifecycle
// parameters, and alias propagation beyond the upstream suite. Declaration and
// type-inference edge shapes live in prop_types_extras_declarations_test.go.
func TestPropTypesRuleExtrasBindings(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PropTypesRule,
		[]rule_tester.ValidTestCase{
			// ---- Nested component ownership ----
			{Code: `function Outer(props) { function Inner(props) { return <span>{props.inner}</span>; } Inner.propTypes = { inner: PropTypes.string }; return <Inner />; }`, Tsx: true},
			{Code: `class Outer extends React.Component { render() { class Inner extends React.Component { static propTypes = { inner: PropTypes.string }; render() { return <span>{this.props.inner}</span>; } } return <Inner />; } }`, Tsx: true},

			// ---- Lifecycle and destructuring negative controls ----
			{Code: `class Hello extends React.Component { componentDidUpdate(old) { void old.name; } render() { return <div />; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { 'componentDidUpdate'(prevProps) { void prevProps.name; } render() { return <div />; } }`, Tsx: true},
			{Code: `function Hello(props) { const { [key]: dynamic, name } = props; return <div>{name}</div>; }`, Tsx: true},

			// ---- Receiver and ESTree-key negative controls ----
			{Code: `class Outer extends React.Component { render() { class Helper { read() { return this.props.name; } } return <div>{new Helper().read()}</div>; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>{this['props'].name}</div>; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { render() { return <div>{this[(props as unknown)].name}</div>; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { componentDidUpdate() { [1].map(({ name }) => name); } render() { return <div />; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { constructor() { super(); [1].map(({ name }) => name); } render() { return <div />; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { update() { store['setState']((state, { name }) => name); } render() { return <div />; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { update() { store[setState!]((state, { name }) => name); } render() { return <div />; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { [(componentDidUpdate satisfies unknown)]({ name }) { return name; } render() { return <div />; } }`, Tsx: true},
			{Code: `class Hello extends React.Component { read(props) { void props.name; } render() { return <div />; } }`, Tsx: true},

			// ---- Captured lifecycle aliases and createReactClass controls ----
			{Code: `class Outer extends React.Component { componentDidUpdate(prevProps) { class Helper { read() { void prevProps.name; const value = prevProps; return value.name; } } } render() { return <div />; } }`, Tsx: true},
			{Code: `class Outer extends React.Component { componentDidUpdate(prevProps) { class Helper { componentDidUpdate() { const { name } = prevProps.user; return name; } } } render() { return <div />; } }`, Tsx: true},
			{Code: `const Hello = createReactClass({ constructor(props) { void props.name; }, render() { return <div />; } });`, Tsx: true},
			{Code: `const Hello = createReactClass({ componentDidUpdate: (function(prevProps) { void prevProps.name; } as Handler), render() { return <div />; } });`, Tsx: true},
			{Code: `class Outer extends React.Component { componentDidUpdate(prevProps) { const value = prevProps; class Helper { read() { return value.name; } } } render() { return <div />; } }`, Tsx: true},
			{Code: `class Outer extends React.Component { render() { const { props: value } = this; class Helper { read() { return value.name; } } return <div />; } }`, Tsx: true},
			{Code: `class Outer extends React.Component { render() { const direct = this.props; const copied = direct; class Helper { read() { return copied.name; } } return <div />; } }`, Tsx: true},

			// ---- Property-owned and named components ----
			{Code: `const O = { C: function Inner(props) { return <div>{props.name}</div>; } }; O.C.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `const O = { C: class Inner extends React.Component { render() { return <div>{this.props.name}</div>; } } }; O.C.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `const O = { C(props) { return <div>{props.name}</div>; } }; O.C.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `const Hello = class Inner extends React.Component { render() { return <div>{this.props.name}</div>; } }; Inner.propTypes = { name: PropTypes.string };`, Tsx: true},
			{Code: `const Hello = function Inner(props) { return <div>{props.name}</div>; }; Inner.propTypes = { name: PropTypes.string };`, Tsx: true},
		}, []rule_tester.InvalidTestCase{
			// ---- Nested owners and special props parameters ----
			{Code: `function Outer(props) { function Inner(props) { return <span>{props.inner}</span>; } return <Inner />; }`, Tsx: true, Errors: missingPropTypes("'inner' is missing in props validation")},
			{Code: `class Outer extends React.Component { render() { class Inner extends React.Component { render() { return <span>{this.props.inner}</span>; } } return <Inner />; } }`, Tsx: true, Errors: missingPropTypes("'inner' is missing in props validation")},
			{Code: `class Outer extends React.Component { render() { class Helper { read(props) { void props.name; } } return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Outer extends React.Component { render() { class Helper { read(props) { const user = props.user; return user.name; } } return <div />; } }`, Tsx: true, Errors: missingPropTypes("'user' is missing in props validation", "'user.name' is missing in props validation")},
			{Code: `class Outer extends React.Component { render() { class Helper { read(props) { const { name } = props.user; return name; } } return <div />; } }`, Tsx: true, Errors: missingPropTypes("'user.name' is missing in props validation", "'user' is missing in props validation")},
			{Code: `class Outer extends React.Component { componentDidUpdate(prevProps) { class Helper { read() { const { name } = prevProps; return name; } } } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Outer extends React.Component { render() { class Helper { componentDidUpdate() { const { name } = props.user; return name; } } return <div />; } }`, Tsx: true, Errors: missingPropTypes("'user.name' is missing in props validation", "'user' is missing in props validation")},
			{Code: `class Outer extends React.Component { render() { class Helper { componentDidUpdate({ name }) { return name; } } return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Outer extends React.Component { render() { class Helper { update() { store.setState((state, { name }) => name); } } return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { update() { store[setState]((state, { name }) => name); } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { update() { store[(setState)]((state, { name }) => name); } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},

			// ---- createReactClass lifecycle forms and captured parameters ----
			{Code: `const Hello = createReactClass({ componentDidUpdate(prevProps) { void prevProps.name; }, render() { return <div />; } });`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `const Hello = createReactClass({ componentDidUpdate: function(prevProps) { void prevProps.name; }, render() { return <div />; } });`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `const Hello = createReactClass({ componentDidUpdate: (prevProps) => { void prevProps.name; }, render() { return <div />; } });`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `const Hello = createReactClass({ componentDidUpdate: (function(prevProps) { void prevProps.name; }), render() { return <div />; } });`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `const Hello = createReactClass({ componentDidUpdate: (({ name }) => { void name; }), render() { return <div />; } });`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { get componentDidUpdate() { void props.name; return null; } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { set componentDidUpdate(prevProps) { void prevProps.name; } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `const Hello = createReactClass({ get componentDidUpdate() { void props.name; return null; }, render() { return <div />; } });`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `const Hello = createReactClass({ set componentDidUpdate(prevProps) { void prevProps.name; }, render() { return <div />; } });`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { componentDidUpdate(prevProps) { setTimeout(() => prevProps.name); } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { componentDidUpdate = (prevProps) => { setTimeout(() => prevProps.name); }; render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},

			// ---- Scoped, named, and property-owned components ----
			{Code: `function Same(props) { return <div>{props.outer}</div>; } { function Same(props) { return <div>{props.inner}</div>; } } { Same.propTypes = { outer: PropTypes.string }; }`, Tsx: true, Errors: missingPropTypes("'inner' is missing in props validation")},
			{Code: `const Hello = function Inner(props) { return <>{props.name}{props.missing}</>; }; Hello.propTypes = { name: PropTypes.string };`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `const Hello = class Inner extends React.Component { render() { return <>{this.props.name}{this.props.missing}</>; } }; Hello.propTypes = { name: PropTypes.string };`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `const Hello = function Inner(props) { Inner.propTypes = { name: PropTypes.string }; return <>{props.name}{props.missing}</>; };`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `const O = { C: function Inner(props) { return <>{props.name}{props.missing}</>; } }; O.C.propTypes = { name: PropTypes.string };`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `let O; O = { C: class Inner extends React.Component { render() { return <>{this.props.name}{this.props.missing}</>; } } }; O.C.propTypes = { name: PropTypes.string };`, Tsx: true, Errors: missingPropTypes("'missing' is missing in props validation")},
			{Code: `const O = { C: (props) => <>{props.name}{props.missing}</> }; O.propTypes = { name: PropTypes.string, missing: PropTypes.string };`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation", "'missing' is missing in props validation")},
			{Code: `const O = { C(props) { return <>{props.name}{props.missing}</>; } }; O.propTypes = { name: PropTypes.string, missing: PropTypes.string };`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation", "'missing' is missing in props validation")},

			// ---- Constructor, lifecycle, and setState binding provenance ----
			{Code: `class Hello extends React.Component { constructor({ name }) { super(); this.name = name; } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Outer extends React.Component { render() { const value = this.props; class Helper { read() { return value.name; } } return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { render() { return <div>{this[props].name}</div>; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { render() { return <div>{this[(props)].name}</div>; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { [(componentDidUpdate)]({ name }) { return name; } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { componentDidUpdate({ user: { name } }) { void name; } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'user' is missing in props validation", "'user.name' is missing in props validation")},
			{Code: `class Hello extends React.Component { update() { this.setState((state, { user: { name } }) => ({ name })); } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'user' is missing in props validation", "'user.name' is missing in props validation")},
			{Code: `class Hello extends React.Component { update() { store.setState((state, props) => ({ name: props.name })); } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
			{Code: `class Hello extends React.Component { componentDidUpdate(old) { const prevProps = {}; void prevProps.name; } render() { return <div />; } }`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},

			// ---- Nested non-React classes inside createReactClass ----
			{Code: `const Hello = createReactClass({ render() { class Helper { read() { return this.props.name; } } return <div>{new Helper().read()}</div>; } });`, Tsx: true, Errors: missingPropTypes("'name' is missing in props validation")},
		},
	)
}
