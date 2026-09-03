package function_component_definition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestFunctionComponentDefinitionUpstream migrates the full valid/invalid
// suite from upstream tests/lib/rules/function-component-definition.js 1:1.
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in the
// function_component_definition_extras_test.go file.
func TestFunctionComponentDefinitionUpstream(t *testing.T) {
	named := func(value interface{}) []interface{} {
		return []interface{}{map[string]interface{}{"namedComponents": value}}
	}
	unnamed := func(value interface{}) []interface{} {
		return []interface{}{map[string]interface{}{"unnamedComponents": value}}
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &FunctionComponentDefinitionRule, []rule_tester.ValidTestCase{
		// ---- class components are never reported ----
		{
			Code: `
        class Hello extends React.Component {
          render() { return <div>Hello {this.props.name}</div> }
        }
      `,
			Tsx:     true,
			Options: named("arrow-function"),
		},
		{
			Code: `
        class Hello extends React.Component {
          render() { return <div>Hello {this.props.name}</div> }
        }
      `,
			Tsx:     true,
			Options: named("function-declaration"),
		},
		{
			Code: `
        class Hello extends React.Component {
          render() { return <div>Hello {this.props.name}</div> }
        }
      `,
			Tsx:     true,
			Options: named("function-expression"),
		},

		// ---- named components matching the configured form ----
		{Code: `var Hello = (props) => { return <div/> }`, Tsx: true, Options: named("arrow-function")},
		{Code: `const Hello = (props) => { return <div/> }`, Tsx: true, Options: named("arrow-function")},
		{Code: `function Hello(props) { return <div/> }`, Tsx: true, Options: named("function-declaration")},
		{Code: `var Hello = function(props) { return <div/> }`, Tsx: true, Options: named("function-expression")},
		{Code: `const Hello = function(props) { return <div/> }`, Tsx: true, Options: named("function-expression")},

		// ---- unnamed components matching the configured form ----
		{Code: `function Hello() { return function() { return <div/> } }`, Tsx: true, Options: unnamed("function-expression")},
		{Code: `function Hello() { return () => { return <div/> }}`, Tsx: true, Options: unnamed("arrow-function")},

		// ---- pragma-wrapped components are registered on the wrapper call ----
		{Code: `var Foo = React.memo(function Foo() { return <p/> })`, Tsx: true, Options: named("function-declaration")},
		{Code: `const Foo = React.memo(function Foo() { return <p/> })`, Tsx: true, Options: named("function-declaration")},

		// ---- functions starting with a lowercase letter are not components ----
		{
			Code: `
        const selectAvatarByUserId = (state, id) => {
          const user = selectUserById(state, id)
          return null
        }
      `,
			Tsx:     true,
			Options: named("function-declaration"),
		},
		{
			Code: `
        function ensureValidSourceType(sourceType: string) {
          switch (sourceType) {
            case 'ALBUM':
            case 'PLAYLIST':
              return sourceType;
            default:
              return null;
          }
        }
      `,
			Tsx:     true,
			Options: named("arrow-function"),
		},

		// ---- TypeScript-annotated components matching the configured form ----
		{Code: `function Hello(props: Test) { return <p/> }`, Tsx: true, Options: named("function-declaration")},
		{Code: `var Hello = function(props: Test) { return <p/> }`, Tsx: true, Options: named("function-expression")},
		{Code: `var Hello = (props: Test) => { return <p/> }`, Tsx: true, Options: named("arrow-function")},
		{Code: `var Hello: React.FC<Test> = function(props) { return <p/> }`, Tsx: true, Options: named("function-expression")},
		{Code: `var Hello: React.FC<Test> = (props) => { return <p/> }`, Tsx: true, Options: named("arrow-function")},
		{Code: `function Hello<Test>(props: Props<Test>) { return <p/> }`, Tsx: true, Options: named("function-declaration")},
		{Code: `function Hello<Test extends {}>(props: Props<Test>) { return <p/> }`, Tsx: true, Options: named("function-declaration")},
		{Code: `var Hello = function<Test>(props: Props<Test>) { return <p/> }`, Tsx: true, Options: named("function-expression")},
		{Code: `var Hello = function<Test extends {}>(props: Props<Test>) { return <p/> }`, Tsx: true, Options: named("function-expression")},
		{Code: `var Hello = <Test extends {}>(props: Props<Test>) => { return <p/> }`, Tsx: true, Options: named("arrow-function")},
		{Code: `function wrapper() { return function<Test>(props: Props<Test>) { return <p/> } } `, Tsx: true, Options: unnamed("function-expression")},
		{Code: `function wrapper() { return function<Test extends {}>(props: Props<Test>) { return <p/> } } `, Tsx: true, Options: unnamed("function-expression")},
		{Code: `function wrapper() { return<Test extends {}>(props: Props<Test>) => { return <p/> } } `, Tsx: true, Options: unnamed("arrow-function")},
		{Code: `var Hello = function(props): ReactNode { return <p/> }`, Tsx: true, Options: named("function-expression")},
		{Code: `var Hello = (props): ReactNode => { return <p/> }`, Tsx: true, Options: named("arrow-function")},
		{Code: `function wrapper() { return function(props): ReactNode { return <p/> } }`, Tsx: true, Options: unnamed("function-expression")},
		{Code: `function wrapper() { return (props): ReactNode => { return <p/> } }`, Tsx: true, Options: unnamed("arrow-function")},
		{Code: `function Hello(props): ReactNode { return <p/> }`, Tsx: true, Options: named("function-declaration")},

		// ---- object properties are never reported (issue #2765) ----
		{
			Code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: named("function-declaration"),
		},
		{
			Code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        }
      `,
			Tsx:     true,
			Options: named("arrow-function"),
		},
		{
			Code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        }
      `,
			Tsx:     true,
			Options: named("function-expression"),
		},
		{
			Code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        }
      `,
			Tsx:     true,
			Options: named("function-declaration"),
		},
		{
			Code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: named("arrow-function"),
		},
		{
			Code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: named("function-expression"),
		},
		{
			Code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: named("function-declaration"),
		},
		{
			Code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: named("arrow-function"),
		},
		{
			Code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: named("function-expression"),
		},
		{
			Code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: unnamed("arrow-function"),
		},
		{
			Code: `
        const obj = {
          serialize(el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: unnamed("function-expression"),
		},
		{
			Code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: unnamed("arrow-function"),
		},
		{
			Code: `
        const obj = {
          serialize: (el) => {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: unnamed("function-expression"),
		},
		{
			Code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: unnamed("arrow-function"),
		},
		{
			Code: `
        const obj = {
          serialize: function (el) {
            return <p/>
          }
        };
      `,
			Tsx:     true,
			Options: unnamed("function-expression"),
		},

		// ---- array-valued options accept any listed form ----
		{Code: `function Hello(props) { return <div/> }`, Tsx: true, Options: named([]interface{}{"function-declaration", "function-expression"})},
		{Code: `var Hello = function(props) { return <div/> }`, Tsx: true, Options: named([]interface{}{"function-declaration", "function-expression"})},
		{Code: `var Foo = React.memo(function Foo() { return <p/> })`, Tsx: true, Options: named([]interface{}{"function-declaration", "function-expression"})},
		{Code: `function Hello(props: Test) { return <p/> }`, Tsx: true, Options: named([]interface{}{"function-declaration", "function-expression"})},
		{Code: `var Hello = function(props: Test) { return <p/> }`, Tsx: true, Options: named([]interface{}{"function-expression", "function-expression"})},
		{Code: `var Hello = (props: Test) => { return <p/> }`, Tsx: true, Options: named([]interface{}{"arrow-function", "function-expression"})},
		{
			Code: `
        function wrap(Component) {
          return function(props) {
            return <div><Component {...props}/></div>;
          };
        }
      `,
			Tsx:     true,
			Options: unnamed([]interface{}{"arrow-function", "function-expression"}),
		},
		{
			Code: `
        function wrap(Component) {
          return (props) => {
            return <div><Component {...props}/></div>;
          };
        }
      `,
			Tsx:     true,
			Options: unnamed([]interface{}{"arrow-function", "function-expression"}),
		},

		// ---- non-JSX functions are not components ----
		{
			Code: `
        export default (key, subTree = {}) => {
          return (state) => {
            const dataInStore = getFromDataModel(key)(state);
            const fullPaths = dataInStore.map((item, index) => {
              return [key, index];
            });

            return {
              key,
              paths: fullPaths.map((p) => [p[1]]),
              fullPaths,
              subTree: Object.keys(subTree).length ? subTree : null,
            }
          };
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        function mapStateToProps() {
          const internItems = makeInternArray();
          const internClassList = makeInternArray();

          return (state, props) => {
            const { store, bucket, singleCharacter } = props;

            return {
              store: null,
              destinyVersion: store.destinyVersion,
              storeId: store.id,
            }
          }
        }
      `,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- named components: JavaScript forms ----
		{
			Code: `
        function Hello(props) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        const Hello = (props) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Message: "Function component is not an arrow function", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = function(props) {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        var Hello = (props) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = (props) => {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        function Hello(props) {
          return <div/>;
        }
      `},
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Message: "Function component is not a function declaration", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = function(props) {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        function Hello(props) {
          return <div/>;
        }
      `},
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = (props) => {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        var Hello = function(props) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Message: "Function component is not a function expression", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        let Hello = (props) => {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        let Hello = function(props) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        let Hello;
        Hello = (props) => {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        let Hello;
        Hello = function(props) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 3, Column: 17, EndLine: 5, EndColumn: 10}},
		},
		{
			Code: `
        let Hello = (props) => {
          return <div/>;
        }
        Hello = function(props) {
          return <span/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        let Hello = function(props) {
          return <div/>;
        }
        Hello = function(props) {
          return <span/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        function Hello(props) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        const Hello = function(props) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},

		// ---- unnamed components ----
		{
			Code: `
        function wrap(Component) {
          return function(props) {
            return <div><Component {...props}/></div>;
          };
        }
      `,
			Tsx: true,
			Output: []string{`
        function wrap(Component) {
          return (props) => {
            return <div><Component {...props}/></div>;
          };
        }
      `},
			Options: unnamed("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 3, Column: 18, EndLine: 5, EndColumn: 12}},
		},
		{
			Code: `
        function wrap(Component) {
          return (props) => {
            return <div><Component {...props}/></div>;
          };
        }
      `,
			Tsx: true,
			Output: []string{`
        function wrap(Component) {
          return function(props) {
            return <div><Component {...props}/></div>;
          };
        }
      `},
			Options: unnamed("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 3, Column: 18, EndLine: 5, EndColumn: 12}},
		},

		// ---- TypeScript parameter annotations ----
		{
			Code: `
        var Hello = (props: Test) => {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        function Hello(props: Test) {
          return <div/>;
        }
      `},
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = function(props: Test) {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        function Hello(props: Test) {
          return <div/>;
        }
      `},
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        function Hello(props: Test) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        const Hello = (props: Test) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = function(props: Test) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        var Hello = (props: Test) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        function Hello(props: Test) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        const Hello = function(props: Test) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},

		// ---- fileVarType: `var` unless the file uses ES module syntax or JSX ----
		{
			Code: `
        function Hello(props: Test) {
          return React.createElement('div');
        }
      `,
			Tsx: true,
			Output: []string{`
        var Hello = function(props: Test) {
          return React.createElement('div');
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        import * as React from 'react';
        function Hello(props: Test) {
          return React.createElement('div');
        }
      `,
			Tsx: true,
			Output: []string{`
        import * as React from 'react';
        const Hello = function(props: Test) {
          return React.createElement('div');
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 3, Column: 9, EndLine: 5, EndColumn: 10}},
		},
		{
			Code: `
        export function Hello(props: Test) {
          return React.createElement('div');
        }
      `,
			Tsx: true,
			Output: []string{`
        export const Hello = function(props: Test) {
          return React.createElement('div');
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 16, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        function Hello(props) {
          return React.createElement('div');
        }
        export default Hello;
      `,
			Tsx: true,
			Output: []string{`
        const Hello = function(props) {
          return React.createElement('div');
        }
        export default Hello;
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = (props: Test) => {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        var Hello = function(props: Test) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},

		// ---- variable type annotations ----
		{
			Code: `
        var Hello: React.FC<Test> = (props) => {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        var Hello: React.FC<Test> = function(props) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 37, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello: React.FC<Test> = function(props) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        var Hello: React.FC<Test> = (props) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 37, EndLine: 4, EndColumn: 10}},
		},
		{
			// A type-annotated variable has no function-declaration form to move
			// the annotation to, so the diagnostic carries no fix.
			Code: `
        var Hello: React.FC<Test> = function(props) {
          return <div/>;
        }
      `,
			Tsx:     true,
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 37, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello: React.FC<Test> = (props) => {
          return <div/>;
        };
      `,
			Tsx:     true,
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 37, EndLine: 4, EndColumn: 10}},
		},

		// ---- type parameters ----
		{
			Code: `
        function Hello<Test extends {}>(props: Test) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        const Hello = <Test extends {}>(props: Test) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},
		{
			// One unconstrained type parameter would produce `<Test>(props) =>`,
			// which is ambiguous with JSX, so no fix is offered.
			Code: `
        function Hello<Test>(props: Test) {
          return <div/>;
        }
      `,
			Tsx:     true,
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        function Hello<Test extends {}>(props: Test) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        const Hello = function<Test extends {}>(props: Test) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = function<Test extends {}>(props: Test) {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        function Hello<Test extends {}>(props: Test) {
          return <div/>;
        }
      `},
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = <Test extends {}>(props: Test) => {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        var Hello = function<Test extends {}>(props: Test) {
          return <div/>;
        }
      `},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = function<Test extends {}>(props: Test) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        var Hello = <Test extends {}>(props: Test) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = function<Test extends {}>(props: Test) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        function Hello<Test extends {}>(props: Test) {
          return <div/>;
        }
      `},
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        function wrap(Component) {
          return function<Test extends {}>(props) {
            return <div><Component {...props}/></div>
          }
        }
      `,
			Tsx: true,
			Output: []string{`
        function wrap(Component) {
          return <Test extends {}>(props) => {
            return <div><Component {...props}/></div>
          }
        }
      `},
			Options: unnamed("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 3, Column: 18, EndLine: 5, EndColumn: 12}},
		},
		{
			Code: `
        function wrap(Component) {
          return function<Test>(props) {
            return <div><Component {...props}/></div>
          }
        }
      `,
			Tsx:     true,
			Options: unnamed("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 3, Column: 18, EndLine: 5, EndColumn: 12}},
		},
		{
			Code: `
        function wrap(Component) {
          return <Test extends {}>(props) => {
            return <div><Component {...props}/></div>
          }
        }
      `,
			Tsx: true,
			Output: []string{`
        function wrap(Component) {
          return function<Test extends {}>(props) {
            return <div><Component {...props}/></div>
          }
        }
      `},
			Options: unnamed("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 3, Column: 18, EndLine: 5, EndColumn: 12}},
		},

		// ---- return type annotations ----
		{
			Code: `
        function wrap(Component) {
          return function(props): ReactNode {
            return <div><Component {...props}/></div>
          }
        }
      `,
			Tsx: true,
			Output: []string{`
        function wrap(Component) {
          return (props): ReactNode => {
            return <div><Component {...props}/></div>
          }
        }
      `},
			Options: unnamed("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 3, Column: 18, EndLine: 5, EndColumn: 12}},
		},
		{
			Code: `
        function wrap(Component) {
          return (props): ReactNode => {
            return <div><Component {...props}/></div>
          }
        }
      `,
			Tsx: true,
			Output: []string{`
        function wrap(Component) {
          return function(props): ReactNode {
            return <div><Component {...props}/></div>
          }
        }
      `},
			Options: unnamed("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 3, Column: 18, EndLine: 5, EndColumn: 12}},
		},

		// ---- exported components ----
		{
			Code: `
        export function Hello(props) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        export const Hello = (props) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 16, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        export var Hello = function(props) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        export var Hello = (props) => {
          return <div/>;
        }
      `},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 28, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        export var Hello = (props) => {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        export function Hello(props) {
          return <div/>;
        }
      `},
			Options: named("function-declaration"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 28, EndLine: 4, EndColumn: 10}},
		},
		{
			// `export default var Hello = …` is not valid syntax, so this shape
			// has to be fixed by hand.
			Code: `
        export default function Hello(props) {
          return <div/>;
        }
      `,
			Tsx:     true,
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 24, EndLine: 4, EndColumn: 10}},
		},
		{
			// A named function expression keeps its own binding, so rewriting it
			// into an arrow would drop the name.
			Code: `
        module.exports = function Hello(props) {
          return <div/>;
        }
      `,
			Tsx:     true,
			Options: unnamed("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 26, EndLine: 4, EndColumn: 10}},
		},

		// ---- array-valued options report the first listed form ----
		{
			Code: `
        function Hello(props) {
          return <div/>;
        }
      `,
			Tsx: true,
			Output: []string{`
        const Hello = (props) => {
          return <div/>;
        }
      `},
			Options: named([]interface{}{"arrow-function", "function-expression"}),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 2, Column: 9, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = (props) => {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        function Hello(props) {
          return <div/>;
        }
      `},
			Options: named([]interface{}{"function-declaration", "function-expression"}),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        var Hello = (props) => {
          return <div/>;
        };
      `,
			Tsx: true,
			Output: []string{`
        var Hello = function(props) {
          return <div/>;
        }
      `},
			Options: named([]interface{}{"function-expression", "function-declaration"}),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 21, EndLine: 4, EndColumn: 10}},
		},
		{
			Code:    "\n        const genX = (symbol) => `the symbol is ${symbol}`;\n\n        const IndexPage = () => {\n          return (\n            <div>\n              Hello World.{genX('$')}\n            </div>\n          )\n        }\n\n        export default IndexPage;\n      ",
			Tsx:     true,
			Output:  []string{"\n        const genX = (symbol) => `the symbol is ${symbol}`;\n\n        function IndexPage() {\n          return (\n            <div>\n              Hello World.{genX('$')}\n            </div>\n          )\n        }\n\n        export default IndexPage;\n      "},
			Options: named([]interface{}{"function-declaration"}),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 4, Column: 27, EndLine: 10, EndColumn: 10}},
		},
	})
}
