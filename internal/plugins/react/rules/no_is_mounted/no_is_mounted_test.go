package no_is_mounted

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoIsMountedRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoIsMountedRule, []rule_tester.ValidTestCase{
		// ---- Upstream: function declaration without isMounted call ----
		{Code: `
        var Hello = function() {
        };
      `, Tsx: true},

		// ---- Upstream: createReactClass render without isMounted call ----
		{Code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: this.isMounted referenced but NOT called ----
		{Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            someNonMemberFunction(arg);
            this.someFunc = this.isMounted;
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `, Tsx: true},

		// ---- Upstream: class with similarly-named method (notIsMounted) ----
		{Code: `
        class Hello extends React.Component {
          notIsMounted() {}
          render() {
            this.notIsMounted();
            return <div>Hello</div>;
          }
        };
      `, Tsx: true},

		// ---- Edge: top-level, outside any property / method ----
		{Code: `this.isMounted();`, Tsx: true},

		// ---- Edge: plain function at top level, no class / object context ----
		{Code: `
        function foo() {
          this.isMounted();
        }
      `, Tsx: true},

		// ---- Edge: bracket notation is NOT matched (mirrors ESLint's `'name' in property` guard) ----
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            if (!this['isMounted']()) { return; }
          }
        };
      `, Tsx: true},

		// ---- Edge: property name mismatch ----
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            this.isMountedNot();
          }
        };
      `, Tsx: true},

		// ---- Edge: receiver is not `this` ----
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            other.isMounted();
          }
        };
      `, Tsx: true},

		// ---- Edge: private field name mismatch ----
		{Code: `
        class Hello extends React.Component {
          #isMountedNot() { return true; }
          someMethod() {
            this.#isMountedNot();
          }
        };
      `, Tsx: true},

		// ---- Edge: class field initializer (PropertyDeclaration is NOT MethodDefinition) ----
		{Code: `
        class Hello extends React.Component {
          foo = this.isMounted();
        };
      `, Tsx: true},

		// ---- Edge: class static block (StaticBlock is NOT MethodDefinition) ----
		{Code: `
        class Hello extends React.Component {
          static { this.isMounted(); }
        };
      `, Tsx: true},

		// ---- Edge: object-literal spread — SpreadElement is NOT Property ----
		{Code: `
        var obj = { ...this.isMounted() };
      `, Tsx: true},

		// ---- Edge: method that merely references isMounted (without invoking) ----
		{Code: `
        class Hello extends React.Component {
          someMethod() {
            const fn = this.isMounted;
            return fn;
          }
        };
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: createReactClass componentDidUpdate ----
		{
			Code: `
        var Hello = createReactClass({
          componentDidUpdate: function() {
            if (!this.isMounted()) {
              return;
            }
          },
          render: function() {
            return <div>Hello</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIsMounted",
					Message:   "Do not use isMounted",
					Line:      4, Column: 18,
				},
			},
		},

		// ---- Upstream: createReactClass custom method ----
		{
			Code: `
        var Hello = createReactClass({
          someMethod: function() {
            if (!this.isMounted()) {
              return;
            }
          },
          render: function() {
            return <div onClick={this.someMethod.bind(this)}>Hello</div>;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 18},
			},
		},

		// ---- Upstream: ES6 class component ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            if (!this.isMounted()) {
              return;
            }
          }
          render() {
            return <div onClick={this.someMethod.bind(this)}>Hello</div>;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 18},
			},
		},

		// ---- Edge: class getter ----
		{
			Code: `
        class Hello extends React.Component {
          get mounted() {
            return this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 20},
			},
		},

		// ---- Edge: class setter ----
		{
			Code: `
        class Hello extends React.Component {
          set mounted(_v) {
            this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 13},
			},
		},

		// ---- Edge: class constructor ----
		{
			Code: `
        class Hello extends React.Component {
          constructor() {
            super();
            this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 5, Column: 13},
			},
		},

		// ---- Edge: static class method ----
		{
			Code: `
        class Hello extends React.Component {
          static foo() {
            this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 13},
			},
		},

		// ---- Edge: async method ----
		{
			Code: `
        class Hello extends React.Component {
          async foo() {
            await Promise.resolve();
            this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 5, Column: 13},
			},
		},

		// ---- Edge: generator method ----
		{
			Code: `
        class Hello extends React.Component {
          *foo() {
            yield this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 19},
			},
		},

		// ---- Edge: computed-name class method (method name is irrelevant) ----
		{
			Code: `
        class Hello extends React.Component {
          ['foo']() {
            this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 13},
			},
		},

		// ---- Edge: object-literal shorthand method ----
		{
			Code: `
        var Hello = {
          someMethod() {
            this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 13},
			},
		},

		// ---- Edge: object-literal getter ----
		{
			Code: `
        var Hello = {
          get mounted() {
            return this.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 20},
			},
		},

		// ---- Edge: optional-call on receiver — `this.isMounted?.()` ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this.isMounted?.();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 13},
			},
		},

		// ---- Edge: optional-chain on `this` — `this?.isMounted()` ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            this?.isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 13},
			},
		},

		// ---- Edge: parenthesized receiver ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            (this).isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 13},
			},
		},

		// ---- Edge: parenthesized callee ----
		{
			Code: `
        class Hello extends React.Component {
          someMethod() {
            (this.isMounted)();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 14},
			},
		},

		// ---- Edge: deeply nested inside an arrow function chain inside a method ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            const f = () => () => () => this.isMounted();
            f();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 41},
			},
		},

		// ---- Edge: inside a nested function expression, still flagged (ESLint walks all ancestors) ----
		{
			Code: `
        class Hello extends React.Component {
          componentDidMount() {
            setTimeout(function() {
              if (!this.isMounted()) { return; }
            });
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 5, Column: 20},
			},
		},

		// ---- Edge: nested object literal inside a class method (ancestor chain hits PropertyAssignment first) ----
		{
			Code: `
        class Hello extends React.Component {
          foo() {
            const obj = { bar: () => this.isMounted() };
            return obj;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 38},
			},
		},

		// ---- Edge: private method `this.#isMounted()` — ESLint parity ----
		{
			Code: `
        class Hello extends React.Component {
          #isMounted() { return true; }
          someMethod() {
            this.#isMounted();
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 5, Column: 13},
			},
		},

		// ---- Edge: multiple violations in the same method ----
		{
			Code: `
        class Hello extends React.Component {
          foo() {
            this.isMounted();
            if (this.isMounted()) {}
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noIsMounted", Line: 4, Column: 13},
				{MessageId: "noIsMounted", Line: 5, Column: 17},
			},
		},
	})
}
