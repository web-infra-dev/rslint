package require_optimization

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestRequireOptimizationRule mirrors the upstream eslint-plugin-react test
// suite for `react/require-optimization` 1:1. Each case here corresponds
// to a valid/invalid entry from
// https://github.com/jsx-eslint/eslint-plugin-react/blob/master/tests/lib/rules/require-optimization.js.
//
// Additional rslint-specific edge cases (TypeScript-only syntax, tsgo AST
// quirks, walk-up boundary lock-ins, real-codebase regressions, nil-
// TypeChecker safety) live in `require_optimization_extra_test.go`.
func TestRequireOptimizationRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireOptimizationRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid: plain class without React extension is not a component ----
		{Code: `
        class A {}
      `, Tsx: true},

		// ---- Upstream valid: React.Component with shouldComponentUpdate ----
		{Code: `
        import React from "react";
        class YourComponent extends React.Component {
          shouldComponentUpdate () {}
        }
      `, Tsx: true},

		// ---- Upstream valid: bare Component with shouldComponentUpdate ----
		{Code: `
        import React, {Component} from "react";
        class YourComponent extends Component {
          shouldComponentUpdate () {}
        }
      `, Tsx: true},

		// ---- Upstream valid: PureRender decorator on bare Component ----
		{Code: `
        import React, {Component} from "react";
        @reactMixin.decorate(PureRenderMixin)
        class YourComponent extends Component {
          componentDidMount () {}
          render() {}
        }
      `, Tsx: true},

		// ---- Upstream valid: createReactClass with shouldComponentUpdate property ----
		{Code: `
        import React from "react";
        createReactClass({
          shouldComponentUpdate: function () {}
        })
      `, Tsx: true},

		// ---- Upstream valid: createReactClass with PureRenderMixin ----
		{Code: `
        import React from "react";
        createReactClass({
          mixins: [PureRenderMixin]
        })
      `, Tsx: true},

		// ---- Upstream valid: PureRender decorator alone ----
		{Code: `
        @reactMixin.decorate(PureRenderMixin)
        class DecoratedComponent extends Component {}
      `, Tsx: true},

		// ---- Upstream valid: stateless functional component as FunctionExpression ----
		{Code: `
        const FunctionalComponent = function (props) {
          return <div />;
        }
      `, Tsx: true},

		// ---- Upstream valid: stateless functional component as FunctionDeclaration ----
		{Code: `
        function FunctionalComponent(props) {
          return <div />;
        }
      `, Tsx: true},

		// ---- Upstream valid: stateless functional component as ArrowFunction ----
		{Code: `
        const FunctionalComponent = (props) => {
          return <div />;
        }
      `, Tsx: true},

		// ---- Upstream valid: custom decorator allow-listed ----
		{Code: `
        @bar
        @pureRender
        @foo
        class DecoratedComponent extends Component {}
      `, Tsx: true, Options: map[string]interface{}{"allowDecorators": []interface{}{"renderPure", "pureRender"}}},

		// ---- Upstream valid: React.PureComponent (allow option still applies) ----
		{Code: `
        import React from "react";
        class YourComponent extends React.PureComponent {}
      `, Tsx: true, Options: map[string]interface{}{"allowDecorators": []interface{}{"renderPure", "pureRender"}}},

		// ---- Upstream valid: bare PureComponent ----
		{Code: `
        import React, {PureComponent} from "react";
        class YourComponent extends PureComponent {}
      `, Tsx: true, Options: map[string]interface{}{"allowDecorators": []interface{}{"renderPure", "pureRender"}}},

		// ---- Upstream valid: object with elision-padded array — not a createReactClass call ----
		{Code: `
        const obj = { prop: [,,,,,] }
      `, Tsx: true},

		// ---- Upstream valid: class with class-field handler + shouldComponentUpdate method ----
		{Code: `
        import React from "react";
        class YourComponent extends React.Component {
          handleClick = () => {}
          shouldComponentUpdate(){
            return true;
          }
          render() {
            return <div onClick={this.handleClick}>123</div>
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid: empty React.Component class ----
		{
			Code: `
        import React from "react";
        class YourComponent extends React.Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 9},
			},
		},

		// ---- Upstream invalid: methods without shouldComponentUpdate ----
		{
			Code: `
        import React from "react";
        class YourComponent extends React.Component {
          handleClick() {}
          render() {
            return <div onClick={this.handleClick}>123</div>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 9},
			},
		},

		// ---- Upstream invalid: class field arrow handler, no SCU ----
		{
			Code: `
        import React from "react";
        class YourComponent extends React.Component {
          handleClick = () => {}
          render() {
            return <div onClick={this.handleClick}>123</div>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 9},
			},
		},

		// ---- Upstream invalid: bare Component, empty body ----
		{
			Code: `
        import React, {Component} from "react";
        class YourComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 9},
			},
		},

		// ---- Upstream invalid: empty createReactClass ----
		{
			Code: `
        import React from "react";
        createReactClass({})
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 26},
			},
		},

		// ---- Upstream invalid: createReactClass with non-PureRender mixin ----
		{
			Code: `
        import React from "react";
        createReactClass({
          mixins: [RandomMixin]
        })
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 3, Column: 26},
			},
		},

		// ---- Upstream invalid: decorator with non-PureRender mixin argument ----
		{
			Code: `
        @reactMixin.decorate(SomeOtherMixin)
        class DecoratedComponent extends Component {}
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},

		// ---- Upstream invalid: decorators not in allow-list ----
		{
			Code: `
        @bar
        @pure
        @foo
        class DecoratedComponent extends Component {}
      `,
			Tsx:     true,
			Options: map[string]interface{}{"allowDecorators": []interface{}{"renderPure", "pureRender"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noShouldComponentUpdate", Line: 2, Column: 9},
			},
		},
	})
}
