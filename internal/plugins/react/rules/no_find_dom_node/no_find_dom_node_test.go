package no_find_dom_node

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoFindDomNodeRule(t *testing.T) {
	const msg = "Do not use findDOMNode. It doesn’t work with function components and is deprecated in StrictMode. See https://reactjs.org/docs/react-dom.html#finddomnode"

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoFindDomNodeRule, []rule_tester.ValidTestCase{
		// ---- Upstream: plain function declaration ----
		{Code: `
        var Hello = function() {};
      `, Tsx: true},

		// ---- Upstream: createReactClass with no findDOMNode reference ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: findDOMNode referenced as a value but not called ----
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            someNonMemberFunction(arg);
            this.someFunc = React.findDOMNode;
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: unrelated method on React ----
		{Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            React.someFunc(this);
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},

		// ---- Edge: bracket notation is NOT matched (ESLint's `'name' in property` guard) ----
		{Code: `
        class Hello extends Component {
          componentDidMount() {
            React['findDOMNode'](this).scrollIntoView();
          }
        };
      `, Tsx: true},

		// ---- Edge: similar-but-not-equal identifier name ----
		{Code: `
        class Hello extends Component {
          componentDidMount() {
            findDOMNodes(this);
          }
        };
      `, Tsx: true},

		// ---- Edge: property with a different name ----
		{Code: `
        class Hello extends Component {
          componentDidMount() {
            React.findNode(this);
          }
        };
      `, Tsx: true},

		// ---- Edge: `new findDOMNode(...)` is a NewExpression, not a CallExpression ----
		{Code: `
        class Hello extends Component {
          componentDidMount() {
            new findDOMNode(this);
          }
        };
      `, Tsx: true},

		// ---- Edge: tagged template is not a CallExpression ----
		{Code: `
        const r = findDOMNode` + "`hello`" + `;
      `, Tsx: true},

		// ---- Edge: findDOMNode passed as a value, never invoked ----
		{Code: `
        class Hello extends Component {
          componentDidMount() {
            foo(this.findDOMNode);
            bar(React.findDOMNode);
          }
        };
      `, Tsx: true},

		// ---- Edge: object-literal property key happens to be `findDOMNode` ----
		{Code: `
        var api = { findDOMNode: function() {} };
      `, Tsx: true},

		// ---- Edge: assignment target, no call ----
		{Code: `
        class Hello extends Component {
          componentDidMount() {
            this.findDOMNode = null;
          }
        };
      `, Tsx: true},

		// ---- Edge: `new findDOMNode(this)` is a NewExpression, not a CallExpression ----
		{Code: `
        class Hello extends Component {
          componentDidMount() {
            new React.findDOMNode(this);
          }
        };
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: React.findDOMNode inside createReactClass ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            React.findDOMNode(this).scrollIntoView();
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Message: msg, Line: 4, Column: 13},
			},
		},

		// ---- Upstream: ReactDOM.findDOMNode inside createReactClass ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            ReactDOM.findDOMNode(this).scrollIntoView();
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: bare findDOMNode(this) in class method ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            findDOMNode(this).scrollIntoView();
          }
          render() {
            return <div>Hello</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: bare findDOMNode stored on a field ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            this.node = findDOMNode(this);
          }
          render() {
            return <div>Hello</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 25},
			},
		},

		// ---- Edge: parenthesized callee ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            (findDOMNode)(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 14},
			},
		},

		// ---- Edge: parenthesized member-access callee ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            (React.findDOMNode)(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 14},
			},
		},

		// ---- Edge: optional-chain member access — `React?.findDOMNode(this)` ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            React?.findDOMNode(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
			},
		},

		// ---- Edge: optional-call — `React.findDOMNode?.(this)` ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            React.findDOMNode?.(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
			},
		},

		// ---- Edge: chained access — findDOMNode is the last property ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            pkg.react.findDOMNode(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
			},
		},

		// ---- Edge: private-identifier property (ESLint parity — `.name === 'findDOMNode'`) ----
		{
			Code: `
        class Hello extends Component {
          #findDOMNode() { return null; }
          componentDidMount() {
            this.#findDOMNode();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 5, Column: 13},
			},
		},

		// ---- Edge: callee is a member of a call result — `getDom().findDOMNode(this)` ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            getDom().findDOMNode(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
			},
		},

		// ---- Edge: multiple violations in the same method ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            findDOMNode(this);
            React.findDOMNode(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
				{MessageId: "noFindDOMNode", Line: 5, Column: 13},
			},
		},

		// ---- Edge: nested — outer call flags, inner call also flags ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            foo(findDOMNode(this));
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 17},
			},
		},

		// ---- Edge: TypeScript type arguments on the call ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            React.findDOMNode<HTMLDivElement>(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
			},
		},

		// ---- Edge: non-null assertion on the receiver — `React!.findDOMNode(this)` ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            React!.findDOMNode(this);
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 13},
			},
		},

		// ---- Edge: inside a JSX attribute arrow callback (nested function expr) ----
		{
			Code: `
        class Hello extends Component {
          render() {
            return <div onClick={() => findDOMNode(this)} />;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 40},
			},
		},

		// ---- Edge: inside a SpreadElement of an array literal ----
		{
			Code: `
        class Hello extends Component {
          componentDidMount() {
            const list = [...findDOMNode(this)];
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 30},
			},
		},

		// ---- Edge: inside a template-literal expression slot ----
		{
			Code: "\n" +
				"        class Hello extends Component {\n" +
				"          componentDidMount() {\n" +
				"            const s = `id=${findDOMNode(this).id}`;\n" +
				"          }\n" +
				"        };\n" +
				"      ",
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noFindDOMNode", Line: 4, Column: 29},
			},
		},
	})
}
