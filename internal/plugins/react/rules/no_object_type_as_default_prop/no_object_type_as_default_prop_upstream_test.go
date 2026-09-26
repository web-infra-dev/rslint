package no_object_type_as_default_prop

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Ported from eslint-plugin-react v7.37.5:
// https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/no-object-type-as-default-prop.js
func TestNoObjectTypeAsDefaultPropUpstream(t *testing.T) {
	expectedViolations := []rule_tester.InvalidTestCaseError{
		{MessageId: "forbiddenTypeDefaultParam", Message: "a has a/an object literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of object literal.", Line: 3, Column: 11, EndLine: 3, EndColumn: 17},
		{MessageId: "forbiddenTypeDefaultParam", Message: "b has a/an array literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of array literal.", Line: 4, Column: 11, EndLine: 4, EndColumn: 29},
		{MessageId: "forbiddenTypeDefaultParam", Message: "c has a/an regex literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of regex literal.", Line: 5, Column: 11, EndLine: 5, EndColumn: 23},
		{MessageId: "forbiddenTypeDefaultParam", Message: "d has a/an arrow function as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of arrow function.", Line: 6, Column: 11, EndLine: 6, EndColumn: 23},
		{MessageId: "forbiddenTypeDefaultParam", Message: "e has a/an function expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of function expression.", Line: 7, Column: 11, EndLine: 7, EndColumn: 28},
		{MessageId: "forbiddenTypeDefaultParam", Message: "f has a/an class expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of class expression.", Line: 8, Column: 11, EndLine: 8, EndColumn: 23},
		{MessageId: "forbiddenTypeDefaultParam", Message: "g has a/an construction expression as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of construction expression.", Line: 9, Column: 11, EndLine: 9, EndColumn: 26},
		{MessageId: "forbiddenTypeDefaultParam", Message: "h has a/an JSX element as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of JSX element.", Line: 10, Column: 11, EndLine: 10, EndColumn: 24},
		{MessageId: "forbiddenTypeDefaultParam", Message: "i has a/an Symbol literal as default prop. This could lead to potential infinite render loop in React. Use a variable reference instead of Symbol literal.", Line: 11, Column: 11, EndLine: 11, EndColumn: 28},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoObjectTypeAsDefaultPropRule, []rule_tester.ValidTestCase{
		{Code: `
      function Foo({
        bar = emptyFunction,
      }) {
        return null;
      }
    `, Tsx: true},
		{Code: `
      function Foo({
        bar = emptyFunction,
        ...rest
      }) {
        return null;
      }
    `, Tsx: true},
		{Code: `
      function Foo({
        bar = 1,
        baz = 'hello',
      }) {
        return null;
      }
    `, Tsx: true},
		{Code: `
      function Foo(props) {
        return null;
      }
    `, Tsx: true},
		{Code: `
      function Foo(props) {
        return null;
      }

      Foo.defaultProps = {
        bar: () => {}
      }
    `, Tsx: true},
		{Code: `
      const Foo = () => {
        return null;
      };
    `, Tsx: true},
		{Code: `
      const Foo = ({bar = 1}) => {
        return null;
      };
    `, Tsx: true},
		{Code: `
      const Foo = ({bar = 1}, context) => {
        return null;
      };
    `, Tsx: true},
		{Code: `
      export default function NotAComponent({foo = {}}) {}
    `, Tsx: true},
		// Exact documentation examples omit returns and are not detected as components.
		{Code: `const emptyArray = [];

function Component({
  items = emptyArray,
}) {}`, Tsx: true},
		{Code: `function Component({
  items = [],
}) {}`, Tsx: true},
		{Code: `const Component = ({
  items = {},
}) => {}`, Tsx: true},
		{Code: `const Component = ({
  items = () => {},
}) => {}`, Tsx: true},
		{Code: `const emptyArray = [];

function Component({
  items = emptyArray,
}) {}`, Tsx: true},
		{Code: `const emptyObject = {};
const Component = ({
  items = emptyObject,
}) => {}`, Tsx: true},
		{Code: `const noopFunc = () => {};
const Component = ({
  items = noopFunc,
}) => {}`, Tsx: true},
		{Code: `// primitives are all compared by value, so are safe to be inlined
function Component({
  num = 3,
  str = 'foo',
  bool = true,
}) {}`, Tsx: true},
	},
		[]rule_tester.InvalidTestCase{
			{Code: `
        function Foo({
          a = {},
          b = ['one', 'two'],
          c = /regex/i,
          d = () => {},
          e = function() {},
          f = class {},
          g = new Thing(),
          h = <Thing />,
          i = Symbol('foo')
        }) {
          return null;
        }
      `, Tsx: true, Errors: expectedViolations},
			{Code: `
        const Foo = ({
          a = {},
          b = ['one', 'two'],
          c = /regex/i,
          d = () => {},
          e = function() {},
          f = class {},
          g = new Thing(),
          h = <Thing />,
          i = Symbol('foo')
        }) => {
          return null;
        }
      `, Tsx: true, Errors: expectedViolations},
			{Code: `
        const Foo = ({
          a = {},
          b = ['one', 'two'],
          c = /regex/i,
          d = () => {},
          e = function() {},
          f = class {},
          g = new Thing(),
          h = <Thing />,
          i = Symbol('foo')
        }, context) => {
          return null;
        }
      `, Tsx: true, Errors: expectedViolations},
		})
}
