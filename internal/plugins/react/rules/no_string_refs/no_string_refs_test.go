package no_string_refs

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoStringRefsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoStringRefsRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{
			Code: `
var Hello = createReactClass({
  componentDidMount: function() {
    var component = this.hello;
  },
  render: function() {
    return <div ref={c => this.hello = c}>Hello {this.props.name}</div>;
  }
});
`,
			Tsx: true,
		},
		{
			Code: "\nvar Hello = createReactClass({\n  render: function() {\n    return <div ref={`hello`}>Hello {this.props.name}</div>;\n  }\n});\n",
			Tsx:  true,
		},
		{
			Code: "\nvar Hello = createReactClass({\n  render: function() {\n    return <div ref={`hello${index}`}>Hello {this.props.name}</div>;\n  }\n});\n",
			Tsx:  true,
		},
		{
			Code: `
var Hello = createReactClass({
  componentDidMount: function() {
    var component = this.refs.hello;
  },
  render: function() {
    return <div>Hello {this.props.name}</div>;
  }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.3.0"}},
		},

		// ---- Additional edge cases ----
		// `this.refs` outside any React component is not flagged.
		{
			Code: `
function f() {
  return this.refs.hello;
}
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
		},
		// ES6 component that does NOT extend Component/PureComponent: this.refs is not flagged.
		{
			Code: `
class Hello extends SomethingElse {
  componentDidMount() {
    var c = this.refs.hello;
  }
}
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
		},
		// No version setting → defaults to "latest" (999.999.999) → this.refs is not checked.
		{
			Code: `
var Hello = createReactClass({
  componentDidMount: function() {
    var c = this.refs.hello;
  }
});
`,
			Tsx: true,
		},
		// String assigned via callback ref is fine (it's a function, not a string literal).
		{
			Code: `
var Hello = createReactClass({
  render: function() {
    return <div ref={setRef}>hi</div>;
  }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
		},
		// `ref` prop on a JSX element with no initializer (boolean shorthand) is not flagged.
		{
			Code: `
var Hello = createReactClass({
  render: function() { return <div ref />; }
});
`,
			Tsx: true,
		},
		// Pragma gate: with default pragma=React, extending `Preact.Component`
		// is NOT classified as a React component, so this.refs is allowed.
		// Pair with the invalid case that sets pragma=Preact below.
		{
			Code: `
class Hello extends Preact.Component {
  componentDidMount() { var c = this.refs.foo; }
}
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
		},
		// createClass pragma gate: with default (`createReactClass`), a call
		// to `React.createClass({...})` without `settings.react.createClass =
		// "createClass"` is NOT an ES5 component.
		{
			Code: `
var Hello = React.createClass({
  componentDidMount: function() { var c = this.refs.foo; }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
		},
		// `ref` attribute on a non-JSX (e.g. object property) is unaffected.
		{Code: `var o = { ref: 'hello' };`, Tsx: true},
		// JSX spread attribute containing a string-valued `ref` is not flagged
		// (the AST-level spread is opaque to this rule).
		{Code: `const props = { ref: 'hello' as const }; <div {...props} />;`, Tsx: true},
		// Non-string template expression with identifier — not flagged without
		// noTemplateLiterals.
		{Code: "<div ref={`hello${x}`} />;", Tsx: true},
		// Nearest-class gate: a non-React inner class nested inside a React
		// component should NOT make its methods "inside a component". The
		// ES6 path is decided by the nearest enclosing class only.
		{
			Code: `
class Hello extends React.Component {
  method() {
    class Inner {
      foo() { var c = this.refs.x; }
    }
  }
}
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
		},
		// ES5 path requires this.refs to be inside some function. A bare
		// object-property `this.refs` reference (no enclosing function) must
		// not be treated as inside an ES5 component.
		{
			Code: `
var bag = createReactClass({ x: this.refs && 1 });
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `
var Hello = createReactClass({
  componentDidMount: function() {
    var component = this.refs.hello;
  },
  render: function() {
    return <div>Hello {this.props.name}</div>;
  }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Message: "Using this.refs is deprecated.", Line: 4, Column: 21},
			},
		},
		{
			Code: `
var Hello = createReactClass({
  render: function() {
    return <div ref="hello">Hello {this.props.name}</div>;
  }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stringInRefDeprecated", Message: "Using string literals in ref attributes is deprecated.", Line: 4, Column: 17},
			},
		},
		{
			Code: `
var Hello = createReactClass({
  render: function() {
    return <div ref={'hello'}>Hello {this.props.name}</div>;
  }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stringInRefDeprecated", Line: 4, Column: 17},
			},
		},
		{
			Code: `
var Hello = createReactClass({
  componentDidMount: function() {
    var component = this.refs.hello;
  },
  render: function() {
    return <div ref="hello">Hello {this.props.name}</div>;
  }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 4, Column: 21},
				{MessageId: "stringInRefDeprecated", Line: 7, Column: 17},
			},
		},
		{
			Code: "\nvar Hello = createReactClass({\n  componentDidMount: function() {\n    var component = this.refs.hello;\n  },\n  render: function() {\n    return <div ref={`hello`}>Hello {this.props.name}</div>;\n  }\n});\n",
			Options: map[string]interface{}{
				"noTemplateLiterals": true,
			},
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 4, Column: 21},
				{MessageId: "stringInRefDeprecated", Line: 7, Column: 17},
			},
		},
		{
			Code: "\nvar Hello = createReactClass({\n  componentDidMount: function() {\n    var component = this.refs.hello;\n  },\n  render: function() {\n    return <div ref={`hello${index}`}>Hello {this.props.name}</div>;\n  }\n});\n",
			Options: map[string]interface{}{
				"noTemplateLiterals": true,
			},
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 4, Column: 21},
				{MessageId: "stringInRefDeprecated", Line: 7, Column: 17},
			},
		},
		{
			Code: "\nvar Hello = createReactClass({\n  componentDidMount: function() {\n    var component = this.refs.hello;\n  },\n  render: function() {\n    return <div ref={`hello${index}`}>Hello {this.props.name}</div>;\n  }\n});\n",
			Options: map[string]interface{}{
				"noTemplateLiterals": true,
			},
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.3.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stringInRefDeprecated", Line: 7, Column: 17},
			},
		},

		// ---- Additional edge cases ----
		// ES6 class component extending React.Component with this.refs.
		{
			Code: `
class Hello extends React.Component {
  componentDidMount() {
    var c = this.refs.hello;
  }
  render() { return <div />; }
}
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 4, Column: 13},
			},
		},
		// ES6 class component extending PureComponent (bare identifier) with this.refs.
		{
			Code: `
class Hello extends PureComponent {
  componentDidMount() {
    var c = this.refs.hello;
  }
}
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 4, Column: 13},
			},
		},
		// `React.createClass` (legacy pragma form) also counts as an ES5 component.
		{
			Code: `
var Hello = React.createClass({
  componentDidMount: function() {
    var c = this.refs.hello;
  }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0", "createClass": "createClass"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 4, Column: 13},
			},
		},
		// Class expression assigned to a variable.
		{
			Code: `
var Hello = class extends React.Component {
  componentDidMount() { var c = this.refs.foo; }
};
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 3, Column: 33},
			},
		},
		// Arrow property on an ES6 component — class field, still inside the class.
		{
			Code: `
class Hello extends React.Component {
  onClick = () => { var c = this.refs.foo; };
}
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 3, Column: 29},
			},
		},
		// Parenthesized `this` inside the refs access.
		{
			Code: `
var Hello = createReactClass({
  componentDidMount: function() { var c = (this).refs.hello; }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 3, Column: 43},
			},
		},
		// Nested createReactClass — inner this.refs is inside a component too.
		{
			Code: `
var Outer = createReactClass({
  render: function() {
    var Inner = createReactClass({
      componentDidMount: function() { var c = this.refs.x; }
    });
    return null;
  }
});
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 5, Column: 47},
			},
		},
		// Pragma setting — `Preact.Component` is recognized when pragma=Preact.
		{
			Code: `
class Hello extends Preact.Component {
  componentDidMount() { var c = this.refs.foo; }
}
`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.2.0", "pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "thisRefsDeprecated", Line: 3, Column: 33},
			},
		},
		// `ref` with parenthesized string literal.
		{
			Code: `
var Hello = createReactClass({
  render: function() { return <div ref={('hello')} />; }
});
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stringInRefDeprecated", Line: 3, Column: 36},
			},
		},
		// `ref` with double-parenthesized string literal.
		{
			Code: `
var Hello = createReactClass({
  render: function() { return <div ref={(('hello'))} />; }
});
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stringInRefDeprecated", Line: 3, Column: 36},
			},
		},
		// `ref={'x' as string}` — the TS `as` cast must not hide the literal.
		{
			Code: `
var Hello = createReactClass({
  render: function() { return <div ref={'hello' as string} />; }
});
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stringInRefDeprecated", Line: 3, Column: 36},
			},
		},
		// Pragma is not `React` by default → `Preact.Component` is not a
		// component, so `this.refs` inside it should NOT fire. (Locks the
		// pragma gate — compare with the matching invalid case above.)
		// The symmetric invalid version uses `pragma=Preact`.
	})
}
