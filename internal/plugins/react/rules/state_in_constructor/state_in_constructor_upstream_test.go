package state_in_constructor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All 19 valid and 8 invalid cases from eslint-plugin-react v7.37.5.
// https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/state-in-constructor.js
// The four documentation examples are already covered by valid cases 3 and 14
// and invalid cases 1 and 3. Parser variants share these semantic cases.
func TestStateInConstructorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StateInConstructorRule, []rule_tester.ValidTestCase{
		// Upstream valid 1
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
		},
		// Upstream valid 2
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Upstream valid 3
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
		},
		// Upstream valid 4
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"always"},
		},
		// Upstream valid 5
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
		},
		// Upstream valid 6
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
		},
		// Upstream valid 7
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Upstream valid 8
		{
			Code: `
        class Foo extends React.Component {
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
		},
		// Upstream valid 9
		{
			Code: `
        class Foo extends React.Component {
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Upstream valid 10
		{
			Code: `
        const Foo = () => <div>Foo</div>
      `,
			Tsx: true,
		},
		// Upstream valid 11
		{
			Code: `
        const Foo = () => <div>Foo</div>
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Upstream valid 12
		{
			Code: `
        function Foo () {
          return <div>Foo</div>
        }
      `,
			Tsx: true,
		},
		// Upstream valid 13
		{
			Code: `
        function Foo () {
          return <div>Foo</div>
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Upstream valid 14
		{
			Code: `
        class Foo extends React.Component {
          state = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Upstream valid 15
		{
			Code: `
        class Foo extends React.Component {
          state = { bar: 0 }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Upstream valid 16
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Upstream valid 17
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            if (foobar) {
              this.state = { bar: 0 }
            }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
		},
		// Upstream valid 18
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            foobar = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
		},
		// Upstream valid 19
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            foobar = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
		},
	}, []rule_tester.InvalidTestCase{
		// Upstream invalid 1
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 5, Column: 13, EndLine: 5, EndColumn: 36},
			},
		},
		// Upstream invalid 2
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 5, Column: 13, EndLine: 5, EndColumn: 36},
			},
		},
		// Upstream invalid 3
		{
			Code: `
        class Foo extends React.Component {
          state = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 3, Column: 11, EndLine: 3, EndColumn: 29},
			},
		},
		// Upstream invalid 4
		{
			Code: `
        class Foo extends React.Component {
          state = { bar: 0 }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 3, Column: 11, EndLine: 3, EndColumn: 29},
			},
		},
		// Upstream invalid 5
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 7, Column: 11, EndLine: 7, EndColumn: 29},
			},
		},
		// Upstream invalid 6
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 7, Column: 11, EndLine: 7, EndColumn: 29},
			},
		},
		// Upstream invalid 7
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 5, Column: 13, EndLine: 5, EndColumn: 36},
			},
		},
		// Upstream invalid 8
		{
			Code: `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            if (foobar) {
              this.state = { bar: 0 }
            }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 6, Column: 15, EndLine: 6, EndColumn: 38},
			},
		},
	})
}
