// TestSortDefaultPropsUpstream migrates all active cases from eslint-plugin-react
// v7.37.5 tests/lib/rules/sort-default-props.js. Parser duplicates share the
// native parser; every diagnostic asserts its complete range. Additional
// branches and edge shapes live in sort_default_props_extras_test.go.
package sort_default_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortDefaultPropsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortDefaultPropsRule, []rule_tester.ValidTestCase{
		// Upstream valid 1.
		{Code: `
        var First = createReactClass({
          render: function() {
            return <div />;
          }
        });
      `, Tsx: true},
		// Upstream valid 2.
		{Code: `
        var First = createReactClass({
          propTypes: {
            A: PropTypes.any,
            Z: PropTypes.string,
            a: PropTypes.any,
            z: PropTypes.string
          },
          getDefaultProps: function() {
            return {
              A: "A",
              Z: "Z",
              a: "a",
              z: "z"
            };
          },
          render: function() {
            return <div />;
          }
        });
      `, Tsx: true},
		// Upstream valid 3.
		{Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            A: PropTypes.any,
            z: PropTypes.string,
            Z: PropTypes.string
          },
          getDefaultProps: function() {
            return {
              a: "a",
              A: "A",
              z: "z",
              Z: "Z"
            };
          },
          render: function() {
            return <div />;
          }
        });
      `, Tsx: true, Options: map[string]any{"ignoreCase": true}},
		// Upstream valid 4.
		{Code: `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any,
            z: PropTypes.string
          },
          getDefaultProps: function() {
            return {
              a: "a",
              z: "z"
            };
          },
          render: function() {
            return <div />;
          }
        });
        var Second = createReactClass({
          propTypes: {
            AA: PropTypes.any,
            ZZ: PropTypes.string
          },
          getDefaultProps: function() {
            return {
              AA: "AA",
              ZZ: "ZZ"
            };
          },
          render: function() {
            return <div />;
          }
        });
      `, Tsx: true},
		// Upstream valid 5.
		{Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          a: PropTypes.string,
          z: PropTypes.string
        };
        First.propTypes.justforcheck = PropTypes.string;
        First.defaultProps = {
          a: a,
          z: z
        };
        First.defaultProps.justforcheck = "justforcheck";
      `, Tsx: true},
		// Upstream valid 6.
		{Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          a: PropTypes.any,
          A: PropTypes.any,
          z: PropTypes.string,
          Z: PropTypes.string
        };
        First.defaultProps = {
          a: "a",
          A: "A",
          z: "z",
          Z: "Z"
        };
      `, Tsx: true, Options: map[string]any{"ignoreCase": true}},
		// Upstream valid 7.
		{Code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.any,
            b: PropTypes.any,
            c: PropTypes.any
          };
          static defaultProps = {
            a: "a",
            b: "b",
            c: "c"
          };
          render() {
            return <div />;
          }
        }
      `, Tsx: true},
		// Upstream valid 8.
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "aria-controls": PropTypes.string
        };
        Hello.defaultProps = {
          "aria-controls": "aria-controls"
        };
      `, Tsx: true, Options: map[string]any{"ignoreCase": true}},
		// Upstream valid 9.
		{Code: `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `, Tsx: true},
		// Upstream valid 10.
		{Code: `
        var First = createReactClass({
          propTypes: {
            barRequired: PropTypes.func.isRequired,
            onBar: PropTypes.func,
            z: PropTypes.any
          },
          getDefaultProps: function() {
            return {
              barRequired: "barRequired",
              onBar: "onBar",
              z: "z"
            };
          },
          render: function() {
            return <div />;
          }
        });
      `, Tsx: true},
		// Upstream valid 11.
		{Code: `
        export default class ClassWithSpreadInPropTypes extends BaseClass {
          static propTypes = {
            b: PropTypes.string,
            ...c.propTypes,
            a: PropTypes.string
          }
          static defaultProps = {
            b: "b",
            ...c.defaultProps,
            a: "a"
          }
        }
      `, Tsx: true},
		// Upstream valid 12.
		{Code: `
        export default class ClassWithSpreadInPropTypes extends BaseClass {
          static propTypes = {
            a: PropTypes.string,
            b: PropTypes.string,
            c: PropTypes.string,
            d: PropTypes.string,
            e: PropTypes.string,
            f: PropTypes.string
          }
          static defaultProps = {
            a: "a",
            b: "b",
            ...c.defaultProps,
            e: "e",
            f: "f",
            ...d.defaultProps
          }
        }
      `, Tsx: true},
		// Upstream valid 13.
		{Code: `
        const defaults = {
          b: "b"
        };
        const types = {
          a: PropTypes.string,
          b: PropTypes.string,
          c: PropTypes.string
        };
        function StatelessComponentWithSpreadInPropTypes({ a, b, c }) {
          return <div>{a}{b}{c}</div>;
        }
        StatelessComponentWithSpreadInPropTypes.propTypes = types;
        StatelessComponentWithSpreadInPropTypes.defaultProps = {
          c: "c",
          ...defaults,
          a: "a"
        };
      `, Tsx: true},
		// Upstream valid 14.
		{Code: `
        const propTypes = require('./externalPropTypes')
        const defaultProps = require('./externalDefaultProps')
        const TextFieldLabel = (props) => {
          return <div />;
        };
        TextFieldLabel.propTypes = propTypes;
        TextFieldLabel.defaultProps = defaultProps;
      `, Tsx: true},
		// Upstream valid 15.
		{Code: `
        const First = (props) => <div />;
        export const propTypes = {
            a: PropTypes.any,
            z: PropTypes.string,
        };
        export const defaultProps = {
            a: "a",
            z: "z",
        };
        First.propTypes = propTypes;
        First.defaultProps = defaultProps;
      `, Tsx: true},
		// Upstream valid 16.
		{Code: `
        const defaults = {
          b: "b"
        };
        const First = (props) => <div />;
        export const propTypes = {
            a: PropTypes.string,
            b: PropTypes.string,
            z: PropTypes.string,
        };
        export const defaultProps = {
            ...defaults,
            a: "a",
            z: "z",
        };
        First.propTypes = propTypes;
        First.defaultProps = defaultProps;
      `, Tsx: true},
		// Upstream valid 17.
		{Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }

        First.defaultProps = {
            a: PropTypes.any,
            onBar: PropTypes.func,
            onFoo: PropTypes.func,
            z: PropTypes.string,
        };
      `, Tsx: true},
		// Legacy Babel <9 case: these are type annotations, not initializers.
		{Code: `
        class Component extends React.Component {
          propTypes: {
            a: PropTypes.any,
            c: PropTypes.any,
            b: PropTypes.any
          };
          defaultProps: {
            a: "a",
            c: "c",
            b: "b"
          };
          render() {
            return <div />;
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// Upstream invalid 1.
		{Code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.any,
            b: PropTypes.any,
            c: PropTypes.any
          };
          static defaultProps = {
            a: "a",
            c: "c",
            b: "b"
          };
          render() {
            return <div />;
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 11, Column: 13, EndLine: 11, EndColumn: 19},
		}},
		// Upstream invalid 2.
		{Code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.any,
            b: PropTypes.any,
            c: PropTypes.any
          };
          static defaultProps = {
            c: "c",
            b: "b",
            a: "a"
          };
          render() {
            return <div />;
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 10, Column: 13, EndLine: 10, EndColumn: 19},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 11, Column: 13, EndLine: 11, EndColumn: 19},
		}},
		// Upstream invalid 3.
		{Code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.any,
            b: PropTypes.any
          };
          static defaultProps = {
            Z: "Z",
            a: "a",
          };
          render() {
            return <div />;
          }
        }
      `, Tsx: true, Options: map[string]any{"ignoreCase": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 9, Column: 13, EndLine: 9, EndColumn: 19},
		}},
		// Upstream invalid 4.
		{Code: `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.any,
            z: PropTypes.any
          };
          static defaultProps = {
            a: "a",
            Z: "Z",
          };
          render() {
            return <div />;
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 9, Column: 13, EndLine: 9, EndColumn: 19},
		}},
		// Upstream invalid 5.
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "a": PropTypes.string,
          "b": PropTypes.string
        };
        Hello.defaultProps = {
          "b": "b",
          "a": "a"
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 13, Column: 11, EndLine: 13, EndColumn: 19},
		}},
		// Upstream invalid 6.
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "a": PropTypes.string,
          "b": PropTypes.string,
          "c": PropTypes.string
        };
        Hello.defaultProps = {
          "c": "c",
          "b": "b",
          "a": "a"
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 14, Column: 11, EndLine: 14, EndColumn: 19},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 15, Column: 11, EndLine: 15, EndColumn: 19},
		}},
		// Upstream invalid 7.
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "a": PropTypes.string,
          "B": PropTypes.string,
        };
        Hello.defaultProps = {
          "a": "a",
          "B": "B",
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 13, Column: 11, EndLine: 13, EndColumn: 19},
		}},
		// Upstream invalid 8.
		{Code: `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "a": PropTypes.string,
          "B": PropTypes.string,
        };
        Hello.defaultProps = {
          "B": "B",
          "a": "a",
        };
      `, Tsx: true, Options: map[string]any{"ignoreCase": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 13, Column: 11, EndLine: 13, EndColumn: 19},
		}},
		// Upstream invalid 9.
		{Code: `
        const First = (props) => <div />;
        const propTypes = {
          z: PropTypes.string,
          a: PropTypes.any,
        };
        const defaultProps = {
          z: "z",
          a: "a",
        };
        First.propTypes = propTypes;
        First.defaultProps = defaultProps;
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 9, Column: 11, EndLine: 9, EndColumn: 17},
		}},
		// Upstream invalid 10.
		{Code: `
        export default class ClassWithSpreadInPropTypes extends BaseClass {
          static propTypes = {
            b: PropTypes.string,
            ...c.propTypes,
            a: PropTypes.string
          }
          static defaultProps = {
            b: "b",
            a: "a",
            ...c.defaultProps
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 10, Column: 13, EndLine: 10, EndColumn: 19},
		}},
		// Upstream invalid 11.
		{Code: `
        export default class ClassWithSpreadInPropTypes extends BaseClass {
          static propTypes = {
            a: PropTypes.string,
            b: PropTypes.string,
            c: PropTypes.string,
            d: PropTypes.string,
            e: PropTypes.string,
            f: PropTypes.string
          }
          static defaultProps = {
            b: "b",
            a: "a",
            ...c.defaultProps,
            f: "f",
            e: "e",
            ...d.defaultProps
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 13, Column: 13, EndLine: 13, EndColumn: 19},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 16, Column: 13, EndLine: 16, EndColumn: 19},
		}},
		// Upstream invalid 12.
		{Code: `
        const defaults = {
          b: "b"
        };
        const types = {
          a: PropTypes.string,
          b: PropTypes.string,
          c: PropTypes.string
        };
        function StatelessComponentWithSpreadInPropTypes({ a, b, c }) {
          return <div>{a}{b}{c}</div>;
        }
        StatelessComponentWithSpreadInPropTypes.propTypes = types;
        StatelessComponentWithSpreadInPropTypes.defaultProps = {
          c: "c",
          a: "a",
          ...defaults,
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 16, Column: 11, EndLine: 16, EndColumn: 17},
		}},
		// Upstream invalid 13.
		{Code: `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }

        First.defaultProps = {
            a: PropTypes.any,
            z: PropTypes.string,
            onFoo: PropTypes.func,
            onBar: PropTypes.func,
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 11, Column: 13, EndLine: 11, EndColumn: 34},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 12, Column: 13, EndLine: 12, EndColumn: 34},
		}},
	})
}
