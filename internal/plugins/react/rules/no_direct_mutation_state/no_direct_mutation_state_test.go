package no_direct_mutation_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDirectMutationStateRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDirectMutationStateRule, []rule_tester.ValidTestCase{
		// ---- Upstream: createReactClass with no mutation ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: local object shaped like { state: {} } is not this.state ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            var obj = {state: {}};
            obj.state.name = "foo";
            return <div>Hello {obj.state.name}</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: not a component at all ----
		{Code: `
        var Hello = "foo";
        module.exports = {};
      `, Tsx: true},

		// ---- Upstream: non-component class (no extends) — mutation allowed ----
		{Code: `
        class Hello {
          getFoo() {
            this.state.foo = 'bar'
            return this.state.foo;
          }
        }
      `, Tsx: true},

		// ---- Upstream: mutation inside constructor is exempt ----
		{Code: `
        class Hello extends React.Component {
          constructor() {
            this.state.foo = "bar"
          }
        }
      `, Tsx: true},

		// ---- Upstream: numeric literal assignment in constructor is still exempt ----
		{Code: `
        class Hello extends React.Component {
          constructor() {
            this.state.foo = 1;
          }
        }
      `, Tsx: true},

		// ---- Upstream: nested non-mutating class inside a component constructor ----
		{Code: `
        class OneComponent extends Component {
          constructor() {
            super();
            class AnotherComponent extends Component {
              constructor() {
                super();
              }
            }
            this.state = {};
          }
        }
      `, Tsx: true},

		// ---- Edge: bracket access `this['state']` is not matched (ESLint checks Identifier .name only) ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this['state'].foo = 'bar';
          }
        }
      `, Tsx: true},

		// ---- Edge: receiver is not `this` ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            that.state.foo = 'bar';
          }
        }
      `, Tsx: true},

		// ---- Edge: a class that doesn't extend Component/PureComponent ----
		{Code: `
        class Hello extends SomethingElse {
          componentDidMount() {
            this.state.foo = 'bar';
          }
        }
      `, Tsx: true},

		// ---- Edge: destructuring assignment whose LHS is a pattern, not a member ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            const {x} = this.state;
            ({x} = this.state);
          }
        }
      `, Tsx: true},

		// ---- Edge: inside constructor, outside any call: compound assignment is still exempt ----
		{Code: `
        class Hello extends React.Component {
          constructor() {
            this.state.count += 1;
          }
        }
      `, Tsx: true},

		// ---- Edge: mutation sits in a non-function position of createReactClass — no component ----
		{Code: `
        var Hello = createReactClass({
          foo: (this.state.bar = 1)
        });
      `, Tsx: true},

		// ---- Edge: `super.state.foo = ...` — not `this`, not reported (matches ESLint `object.type === 'ThisExpression'`) ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            super.state.foo = "bar";
          }
        }
      `, Tsx: true},

		// ---- Edge: receiver is an identifier coincidentally named `state` ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            const state = { foo: 0 };
            state.foo = 1;
          }
        }
      `, Tsx: true},

		// ---- Edge: `this['state'].foo = x` — computed outer, not flagged (matches ESLint) ----
		{Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this['state'].nested.foo = 1;
          }
        }
      `, Tsx: true},

		// ---- Edge: React.PureComponent subclass is still a component ----
		{Code: `
        class Hello extends React.PureComponent {
          constructor() {
            super();
            this.state = { foo: 'bar' };
          }
        }
      `, Tsx: true},

		// ---- Edge: pragma-qualified createClass via React settings ----
		{
			Code: `
        var Hello = React.createClass({
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass"}},
		},

		// ---- Edge: lowercase-named function is NOT a component (no capital letter) ----
		{Code: `
        function hello() {
          this.state.x = 1;
          return <div/>;
        }
      `, Tsx: true},

		// ---- Edge: capital-named function that doesn't return JSX/null is NOT a component ----
		{Code: `
        function Hello() {
          this.state.x = 1;
          return "hi";
        }
      `, Tsx: true},

		// ---- Edge: arrow with lowercase VariableDeclarator name is NOT a component ----
		{Code: `
        const hello = () => {
          this.state.x = 1;
          return <div/>;
        };
      `, Tsx: true},

		// ---- Edge: bare function body (no enclosing component) ----
		{Code: `
        function foo() {
          this.state.x = 1;
        }
      `, Tsx: true},

		// ---- Edge: lowercase variable beats named FE — `const outer = function Inner() {...}`
		// matches ESLint's getStatelessComponent, which early-returns `undefined`
		// when VariableDeclarator.id is lowercase; the FE's own capitalized name
		// MUST NOT override that. ----
		{Code: `
        const outer = function Inner() {
          this.state.x = 1;
          return <div/>;
        };
      `, Tsx: true},

		// ---- Edge: non-wrapper CallExpression parent (e.g. plain helper) — not a component ----
		{Code: `
        helper(function () {
          this.state.x = 1;
          return <div/>;
        });
      `, Tsx: true},

		// ---- Edge: lowercase property name in object literal — not a component ----
		{Code: `
        const obj = {
          render: () => {
            this.state.x = 1;
            return <div/>;
          },
        };
      `, Tsx: true},

		// ---- Edge: IIFE (any name) is NOT in an allowed position for component
		// per ESLint's isInAllowedPositionForComponent — CallExpression parent is
		// rejected, so these must NOT be reported. ----
		{Code: `
        (function Hello() {
          this.state.x = 1;
          return <div/>;
        })();
      `, Tsx: true},
		{Code: `
        (() => {
          this.state.x = 1;
          return <div/>;
        })();
      `, Tsx: true},

		// ---- Edge: obj.lowercase = fn — lowercase property name is NOT a component ----
		{Code: `
        obj.foo = function () {
          this.state.x = 1;
          return <div/>;
        };
      `, Tsx: true},

		// ---- Edge: plain helper callback (not memo/forwardRef) — not a component ----
		{Code: `
        setTimeout(() => {
          this.state.x = 1;
          return <div/>;
        });
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: createReactClass — simple mutation ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            this.state.foo = "bar"
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: createReactClass — update expression ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            this.state.foo++;
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: createReactClass — nested property assignment ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            this.state.person.name= "bar"
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: createReactClass — deeper nested assignment ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            this.state.person.name.first = "bar"
            return <div>Hello</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: two mutations in the same render body ----
		{
			Code: `
        var Hello = createReactClass({
          render: function() {
            this.state.person.name.first = "bar"
            this.state.person.name.last = "baz"
            return <div>Hello</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noDirectMutation",
					Message:   "Do not mutate state directly. Use setState().",
					Line:      4, Column: 13,
				},
				{
					MessageId: "noDirectMutation",
					Message:   "Do not mutate state directly. Use setState().",
					Line:      5, Column: 13,
				},
			},
		},

		// ---- Upstream: mutation in a class method called from the constructor ----
		{
			Code: `
        class Hello extends React.Component {
          constructor() {
            someFn()
          }
          someFn() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 7, Column: 13},
			},
		},

		// ---- Upstream: constructor body but wrapped in a nested CallExpression (arrow in async helper) ----
		{
			Code: `
        class Hello extends React.Component {
          constructor(props) {
            super(props)
            doSomethingAsync(() => {
              this.state = "bad";
            });
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 6, Column: 15},
			},
		},

		// ---- Upstream: componentWillMount ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillMount() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: componentDidMount ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: componentWillReceiveProps ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillReceiveProps() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: shouldComponentUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          shouldComponentUpdate() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: componentWillUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUpdate() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: componentDidUpdate ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidUpdate() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: componentWillUnmount ----
		{
			Code: `
        class Hello extends React.Component {
          componentWillUnmount() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: bare `class Hello extends Component` (unqualified) ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            this.state.foo = "bar"
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: compound assignment is still a mutation ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.state.count += 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: prefix decrement ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            --this.state.count;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 15},
			},
		},

		// ---- Edge: mutation inside an ArrowFunction callback inside a class method is NOT in the constructor ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            setTimeout(() => { this.state.x = 1; });
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 32},
			},
		},

		// ---- Edge: parenthesized LHS is unwrapped ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            (this.state.foo) = "bar";
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 14},
			},
		},

		// ---- Edge: bracket access within the chain still traverses inward ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.state['foo'] = "bar";
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: static method — ESLint flags (does not special-case `this` binding) ----
		{
			Code: `
        class Hello extends React.Component {
          static foo() {
            this.state.x = 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: class getter body ----
		{
			Code: `
        class Hello extends React.Component {
          get foo() {
            this.state.x = 1;
            return 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: class setter body ----
		{
			Code: `
        class Hello extends React.Component {
          set foo(_v) {
            this.state.x = 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: arrow-function class property body (fires outside the constructor) ----
		{
			Code: `
        class Hello extends React.Component {
          handleClick = () => {
            this.state.x = 1;
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: class field initializer (PropertyDeclaration, not a Constructor) ----
		// The mutation runs during construction semantically, but syntactically
		// it is NOT inside a constructor MethodDefinition, so ESLint flags it
		// (`inConstructor` never becomes true).
		{
			Code: `
        class Hello extends React.Component {
          foo = (this.state.x = 1);
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 18},
			},
		},

		// ---- Edge: anonymous class expression component ----
		{
			Code: `
        var Hello = class extends React.Component {
          componentDidMount() {
            this.state.x = 1;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: IIFE inside constructor — nested CallExpression voids the exemption ----
		{
			Code: `
        class Hello extends React.Component {
          constructor() {
            super();
            (() => { this.state.x = 1; })();
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 5, Column: 22},
			},
		},

		// ---- Edge: object-literal method named `constructor` inside createReactClass
		// is NOT the ES6 constructor — mutation is flagged. ----
		{
			Code: `
        var Hello = createReactClass({
          constructor: function() {
            this.state.x = 1;
          },
          render: function() {
            return <div/>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: mutation deep inside nested CallExpressions inside a method ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            foo(bar(baz(() => { this.state.x = 1; })));
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 33},
			},
		},

		// ---- Edge: logical assignment ||= on this.state ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.state.x ||= 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: mutation inside a generator method ----
		{
			Code: `
        class Hello extends React.Component {
          *step() {
            this.state.x = 1;
            yield 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: mutation inside async method ----
		{
			Code: `
        class Hello extends React.Component {
          async load() {
            await Promise.resolve();
            this.state.x = 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 5, Column: 13},
			},
		},

		// ---- Edge: computed-name class method ----
		{
			Code: `
        class Hello extends React.Component {
          ['foo']() {
            this.state.x = 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: private method ----
		{
			Code: `
        class Hello extends React.Component {
          #foo() {
            this.state.x = 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: pragma-qualified createReactClass via settings: `React.createClass({...})` ----
		{
			Code: `
        var Hello = React.createClass({
          render: function() {
            this.state.x = 1;
            return <div/>;
          }
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: React.PureComponent — a component, should be flagged ----
		{
			Code: `
        class Hello extends React.PureComponent {
          componentDidMount() {
            this.state.x = 1;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
			},
		},

		// ---- Edge: nested React component inside a non-component class method ----
		{
			Code: `
        class Outer {
          method() {
            class Inner extends React.Component {
              componentDidMount() {
                this.state.x = 1;
              }
            }
            return Inner;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 6, Column: 17},
			},
		},

		// ---- Edge: two mutations across different lifecycle methods on same class ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            this.state.a = 1;
          }
          componentWillUnmount() {
            this.state.b = 2;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 4, Column: 13},
				{MessageId: "noDirectMutation", Line: 7, Column: 13},
			},
		},

		// ---- Edge: mutation inside constructor wrapped in a PostfixUnary call — still flagged ----
		{
			Code: `
        class Hello extends React.Component {
          constructor() {
            super();
            someHelper(() => { this.state.x++; });
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 5, Column: 32},
			},
		},

		// ---- Edge: stateless FunctionDeclaration component (capital name + JSX return) ----
		{
			Code: `
        function Hello() {
          this.state.x = 1;
          return <div/>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless arrow assigned to capital-cased VariableDeclarator ----
		{
			Code: `
        const Hello = () => {
          this.state.x = 1;
          return <div/>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless FunctionExpression assigned to capital-cased VariableDeclarator ----
		{
			Code: `
        const Hello = function() {
          this.state.x = 1;
          return <div/>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless arrow with implicit JSX return body ----
		{
			Code: `
        const Hello = () => (this.state.x = 1, <div/>);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 2, Column: 30},
			},
		},

		// ---- Edge: stateless component returning `null` on some branch ----
		{
			Code: `
        function Hello() {
          this.state.x = 1;
          if (cond) return null;
          return <div/>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless component with ternary returning JSX-or-null ----
		{
			Code: `
        function Hello(props) {
          this.state.x = 1;
          return props.show ? <div/> : null;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless component's inner arrow does NOT shadow — outer stateless wins ----
		{
			Code: `
        function Hello() {
          const onClick = () => { this.state.x = 1; };
          onClick();
          return <div onClick={onClick}/>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 35},
			},
		},

		// ---- Edge: capital variable + lowercase named FE — `const Outer = function inner() {...}`
		// ESLint's VariableDeclarator branch decides by parent.id.name ("Outer"),
		// NOT by the FE's own inner name — so this IS detected as a component. ----
		{
			Code: `
        const Outer = function inner() {
          this.state.x = 1;
          return <div/>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless component with `cond && <div/>` (common conditional render) ----
		{
			Code: `
        function Hello(props) {
          this.state.x = 1;
          return props.visible && <div/>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless component with `cond || <div/>` fallback pattern ----
		{
			Code: `
        function Hello(props) {
          this.state.x = 1;
          return props.fallback || <div/>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: stateless component with nullish coalescing `x ?? <div/>` ----
		{
			Code: `
        function Hello(props) {
          this.state.x = 1;
          return props.cached ?? <div/>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: React.memo wrapping an anonymous arrow ----
		{
			Code: `
        const Hello = React.memo(() => {
          this.state.x = 1;
          return <div/>;
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: bare `memo(...)` (destructured from React) ----
		{
			Code: `
        export default memo(() => {
          this.state.x = 1;
          return <div/>;
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: React.forwardRef wrapping an anonymous arrow ----
		{
			Code: `
        const Hello = React.forwardRef((props, ref) => {
          this.state.x = 1;
          return <div ref={ref}/>;
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: bare `forwardRef(...)` ----
		{
			Code: `
        const Hello = forwardRef((props, ref) => {
          this.state.x = 1;
          return <div ref={ref}/>;
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: pragma-qualified memo under `settings.react.pragma = "Preact"` ----
		{
			Code: `
        const Hello = Preact.memo(() => {
          this.state.x = 1;
          return <div/>;
        });
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: `export default function() {...}` (anonymous default-export FD) ----
		{
			Code: `
        export default function () {
          this.state.x = 1;
          return <div/>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: `export default () => ...` (arrow default export) ----
		{
			Code: `
        export default () => {
          this.state.x = 1;
          return <div/>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: `module.exports = function() {...}` — blanket true per ESLint's
		// isModuleExportsAssignment carve-out. ----
		{
			Code: `
        module.exports = function () {
          this.state.x = 1;
          return <div/>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},

		// ---- Edge: `obj.Hello = function() {...}` — capitalized property name
		// assignment on an arbitrary MemberExpression LHS is a component. ----
		{
			Code: `
        obj.Hello = function () {
          this.state.x = 1;
          return <div/>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noDirectMutation", Line: 3, Column: 11},
			},
		},
	})
}
