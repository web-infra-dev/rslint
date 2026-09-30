// TestPreferExactPropsUpstream migrates the full valid/invalid suite from
// eslint-plugin-react v7.37.5's tests/lib/rules/prefer-exact-props.js 1:1.
// Position assertions cover every executable invalid case. rslint-specific
// lock-in cases live in prefer_exact_props_extras_test.go.
package prefer_exact_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var exactSettings = map[string]interface{}{
	"propWrapperFunctions": []interface{}{
		map[string]interface{}{"property": "exact", "exact": true},
	},
}

var multipleExactSettings = map[string]interface{}{
	"propWrapperFunctions": []interface{}{
		map[string]interface{}{"property": "exact", "exact": true},
		map[string]interface{}{"property": "forbidExtraProps", "exact": true},
	},
}

const (
	propTypesError = "Component propTypes should be exact by using 'exact'."
	flowError      = "Component flow props should be set with exact objects."
)

func TestPreferExactPropsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferExactPropsRule, []rule_tester.ValidTestCase{
		{Code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {};
      `, Settings: exactSettings, Tsx: true},
		{Code: `
        class Component extends React.Component {
          static propTypes = {};
          render() {
            return <div />;
          }
        }
      `, Settings: exactSettings, Tsx: true},
		{Code: `
        class Component extends React.Component {
          props: {};
          render() {
            return <div />;
          }
        }
      `, Settings: exactSettings, Tsx: true},
		{Code: `
        function Component(props) {
          return <div />;
        }
        Component.propTypes = {};
      `, Settings: exactSettings, Tsx: true},
		{Code: `
        function Component(props: {}) {
          return <div />;
        }
      `, Tsx: true},
		// SKIP: rslint does not parse Flow exact-object syntax.
		{Code: `
        type Props = {|
          foo: string
        |}
        function Component(props: Props) {
          return <div />;
        }
      `, Skip: true, Tsx: true},
		// SKIP: rslint does not parse Flow exact-object syntax.
		{Code: `
        type Props = {|
          foo: string
        |}
        function Component(props: Props) {
          let someVar: { foo: string };
          return <div />;
        }
      `, Skip: true, Tsx: true},
		// SKIP: rslint does not parse Flow exact-object syntax.
		{Code: `
        function Component(props: {| foo : string |}) {
          return <div />;
        }
      `, Skip: true, Tsx: true},
		{Code: `
        type Props = {}
        function Component(props: Props) {
          return <div />;
        }
      `, Tsx: true},
		{Code: `
        import type Props from 'foo';
        function Component(props: Props) {
          return <div />;
        }
      `, Tsx: true},
		{Code: `
        const props = {};
        function Component(props) {
          return <div />;
        }
        Component.propTypes = props;
      `, Settings: exactSettings, Tsx: true},
		{Code: `
        const props = {};
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `, Settings: exactSettings, Tsx: true},
		{Code: `
        import props from 'foo';
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `, Settings: exactSettings, Tsx: true},
		{Code: `
        class Component extends React.Component {
          state = {hi: 'hi'}
          render() {
            return <div>{this.state.hi}</div>;
          }
        }
      `, Tsx: true},
		{Code: `
        import exact from "prop-types-exact";
        function Component({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        Component.propTypes = exact({
          foo: PropTypes.string,
          bar: PropTypes.string,
        });
      `, Settings: exactSettings, Tsx: true},
		{Code: `
        function Component({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        Component.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `, Tsx: true},
		{Code: `
        class Component extends React.Component {
          render() {
            const { foo, bar } = this.props;
            return <div>{foo}{bar}</div>;
          }
        }
        Component.propTypes = {
          foo: PropTypes.string,
          bar: PropTypes.string,
        };
      `, Tsx: true},
		{Code: `
        import somethingElse from "something-else";
        const props = {
          foo: PropTypes.string,
          bar: PropTypes.shape({
            baz: PropTypes.string
          })
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = somethingElse(props);
      `, Tsx: true},
		{Code: `
        import somethingElse from "something-else";
        const props =
        class Component extends React.Component {
          static propTypes = somethingElse({
            foo: PropTypes.string,
            bar: PropTypes.shape({
              baz: PropTypes.string
            })
          });
          render() {
            return <div />;
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{Code: `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = {
          foo: PropTypes.string
        };
      `, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 7, Column: 9, EndLine: 7, EndColumn: 28}}},
		{Code: `
        class Component extends React.Component {
          static propTypes = {
            foo: PropTypes.string
          }
          render() {
            return <div />;
          }
        }
      `, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 3, Column: 11, EndLine: 5, EndColumn: 12}}},
		// SKIP: rslint does not parse Flow object-type syntax.
		{Code: `class Component extends React.Component { props: { foo: string } }`, Skip: true, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "flow", Message: flowError, Line: 1, Column: 43}}},
		// SKIP: rslint does not parse Flow object-type syntax.
		{Code: `function Component(props: { foo: string }) { return <div />; }`, Skip: true, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "flow", Message: flowError, Line: 1, Column: 20}}},
		// SKIP: rslint does not parse Flow object-type syntax.
		{Code: `type Props = { foo: string }; function Component(props: Props) { return <div />; }`, Skip: true, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "flow", Message: flowError, Line: 1, Column: 49}}},
		{Code: `
        const props = {
          foo: PropTypes.string
        };
        function Component(props) {
          return <div />;
        }
        Component.propTypes = props;
      `, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 8, Column: 9, EndLine: 8, EndColumn: 28}}},
		{Code: `
        const props = {
          foo: PropTypes.string
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 10, Column: 9, EndLine: 10, EndColumn: 28}}},
		{Code: `
        const props = {
          foo: PropTypes.string
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `, Settings: multipleExactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: "Component propTypes should be exact by using one of 'exact', 'forbidExtraProps'.", Line: 10, Column: 9, EndLine: 10, EndColumn: 28}}},
		{Code: `
        const props = {
          foo: PropTypes.string,
          bar: PropTypes.shape({
            baz: PropTypes.string
          })
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = props;
      `, Settings: multipleExactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: "Component propTypes should be exact by using one of 'exact', 'forbidExtraProps'.", Line: 13, Column: 9, EndLine: 13, EndColumn: 28}}},
		{Code: `
        import somethingElse from "something-else";
        function Component({ foo, bar }) {
          return <div>{foo}{bar}</div>;
        }
        Component.propTypes = somethingElse({
          foo: PropTypes.string,
          bar: PropTypes.string,
        });
      `, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 6, Column: 9, EndLine: 6, EndColumn: 28}}},
		{Code: `
        import somethingElse from "something-else";
        const props = {
          foo: PropTypes.string,
          bar: PropTypes.shape({
            baz: PropTypes.string
          })
        };
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.propTypes = somethingElse(props);
      `, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 14, Column: 9, EndLine: 14, EndColumn: 28}}},
		{Code: `
        import somethingElse from "something-else";
        const props =
        class Component extends React.Component {
          static propTypes = somethingElse({
            foo: PropTypes.string,
            bar: PropTypes.shape({
              baz: PropTypes.string
            })
          });
          render() {
            return <div />;
          }
        }
      `, Settings: exactSettings, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propTypes", Message: propTypesError, Line: 5, Column: 11, EndLine: 10, EndColumn: 14}}},
	})
}
