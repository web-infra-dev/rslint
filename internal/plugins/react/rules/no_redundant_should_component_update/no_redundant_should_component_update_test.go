package no_redundant_should_component_update

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRedundantShouldComponentUpdateRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRedundantShouldComponentUpdateRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid: shouldComponentUpdate on plain React.Component ----
		{Code: `
        class Foo extends React.Component {
          shouldComponentUpdate() {
            return true;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: class field arrow shouldComponentUpdate on Component ----
		{Code: `
        class Foo extends React.Component {
          shouldComponentUpdate = () => {
            return true;
          }
        }
      `, Tsx: true},

		// ---- Upstream valid: nested class expression on plain Component ----
		{Code: `
        function Foo() {
          return class Bar extends React.Component {
            shouldComponentUpdate() {
              return true;
            }
          };
        }
      `, Tsx: true},

		// ---- Edge: PureComponent without shouldComponentUpdate — clean ----
		{Code: `
        class Foo extends React.PureComponent {
          render() { return null; }
        }
      `, Tsx: true},

		// ---- Edge: bare PureComponent without shouldComponentUpdate ----
		{Code: `
        class Foo extends PureComponent {
          render() { return null; }
        }
      `, Tsx: true},

		// ---- Edge: extends some other namespace's PureComponent — strict regex ----
		{Code: `
        class Foo extends Other.PureComponent {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Edge: extends PureComponent.Inner — full-text regex requires terminal `PureComponent` ----
		{Code: `
        class Foo extends PureComponent.Inner {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Edge: extends a CallExpression returning PureComponent — only Identifier / qualified `<pragma>.PureComponent` matches ----
		{Code: `
        class Foo extends getBase() {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Edge: shouldComponentUpdate as STRING-LITERAL key — upstream getPropertyName returns undefined for non-Identifier keys, so no match ----
		{Code: `
        class Foo extends React.PureComponent {
          "shouldComponentUpdate"() { return true; }
        }
      `, Tsx: true},

		// ---- Edge: shouldComponentUpdate as COMPUTED key — non-Identifier key never matches ----
		{Code: "\n        class Foo extends React.PureComponent {\n          [`shouldComponentUpdate`]() { return true; }\n        }\n      ", Tsx: true},

		// ---- Edge: settings.react.pragma="Preact" — `React.PureComponent` no longer matches ----
		{Code: `
        class Foo extends React.PureComponent {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true, Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}}},

		// ---- Edge: empty class body extending PureComponent — clean (no scu) ----
		{Code: `
        class Foo extends React.PureComponent {}
      `, Tsx: true},

		// ---- Edge: extends without superClass — graceful ----
		{Code: `
        class Foo {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Edge: a sibling NON-Pure class is fine even when same file has a Pure one ----
		{Code: `
        class A extends React.Component {
          shouldComponentUpdate() { return true; }
        }
        class B extends React.PureComponent {
          render() { return null; }
        }
      `, Tsx: true},

		// ---- Edge (TS): TypeAssertion `as` wrapper on extends — ESLint's text-regex
		// won't match either, so we mirror that miss. ----
		{Code: `
        class Foo extends (React.PureComponent as any) {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Edge (TS): ElementAccess form `React['PureComponent']` — neither side matches ----
		{Code: `
        class Foo extends React['PureComponent'] {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Edge: SubClass off PureComponent — strict regex requires terminal `PureComponent` ----
		{Code: `
        class Foo extends React.PureComponent.SubClass {
          shouldComponentUpdate() { return true; }
        }
      `, Tsx: true},

		// ---- Edge: doubly-nested classes where ONLY outer extends PureComponent
		// and inner non-Pure class hosts shouldComponentUpdate — outer must NOT
		// report (inner's scu doesn't pollute outer's Members) ----
		{Code: `
        class Outer extends React.PureComponent {
          render() {
            return class Inner extends React.Component {
              shouldComponentUpdate() { return true; }
            };
          }
        }
      `, Tsx: true},

		// ---- Edge: shouldComponentUpdate on a UNRELATED inner class extending Component
		// nested inside an outer PureComponent class with NO scu ----
		{Code: `
        class Outer extends React.PureComponent {
          method() {
            class Local {
              shouldComponentUpdate() { return true; }
            }
          }
        }
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid: ClassDeclaration extends React.PureComponent ----
		{
			Code: `
        class Foo extends React.PureComponent {
          shouldComponentUpdate() {
            return true;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Upstream invalid: bare PureComponent ----
		{
			Code: `
        class Foo extends PureComponent {
          shouldComponentUpdate() {
            return true;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Upstream invalid: class field arrow shouldComponentUpdate ----
		{
			Code: `
        class Foo extends React.PureComponent {
          shouldComponentUpdate = () => {
            return true;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Upstream invalid: nested ClassExpression with own name (Bar) ----
		{
			Code: `
        function Foo() {
          return class Bar extends React.PureComponent {
            shouldComponentUpdate() {
              return true;
            }
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Bar does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      3,
					Column:    18,
				},
			},
		},

		// ---- Upstream invalid: nested ClassExpression bare PureComponent ----
		{
			Code: `
        function Foo() {
          return class Bar extends PureComponent {
            shouldComponentUpdate() {
              return true;
            }
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Bar does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      3,
					Column:    18,
				},
			},
		},

		// ---- Upstream invalid: anonymous ClassExpression assigned to var (name from binding) ----
		{
			Code: `
        var Foo = class extends PureComponent {
          shouldComponentUpdate() {
            return true;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    19,
				},
			},
		},

		// ---- Edge: ClassExpression IIFE with no enclosing binding — name is empty ----
		{
			Code: `
        (class extends React.PureComponent {
          shouldComponentUpdate() { return true; }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					// Empty class name — upstream's `{{component}}` placeholder
					// substitutes to "" producing a leading-space message.
					Message: " does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:    2,
					Column:  10,
				},
			},
		},

		// ---- Edge: parens around extends expression — paren-transparent on our side, matching ESTree ----
		{
			Code: `
        class Foo extends (React.PureComponent) {
          shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge: settings.react.pragma="Preact" + extends Preact.PureComponent ----
		{
			Code: `
        class Foo extends Preact.PureComponent {
          shouldComponentUpdate() { return true; }
        }
      `,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge: outer Component class encloses inner PureComponent — only inner reports (listener fires per node) ----
		{
			Code: `
        class Outer extends React.Component {
          render() {
            class Inner extends React.PureComponent {
              shouldComponentUpdate() { return true; }
            }
            return null;
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Inner does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      4,
					Column:    13,
				},
			},
		},

		// ---- Edge (TS): generic class with type params on extends ----
		{
			Code: `
        class Foo extends React.PureComponent<Props, State> {
          shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge (TS): bare PureComponent with type args ----
		{
			Code: `
        class Foo extends PureComponent<Props> {
          shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge: PrivateIdentifier `#shouldComponentUpdate` — upstream's
		// `getPropertyName` strips the leading `#` and matches; we mirror that. ----
		{
			Code: `
        class Foo extends React.PureComponent {
          #shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge: static shouldComponentUpdate — upstream doesn't filter by
		// `static`, so we report (mirroring upstream's behavior even though
		// runtime React doesn't actually use a static lifecycle hook). ----
		{
			Code: `
        class Foo extends React.PureComponent {
          static shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge: getter `get shouldComponentUpdate()` — same upstream behavior ----
		{
			Code: `
        class Foo extends React.PureComponent {
          get shouldComponentUpdate() { return () => true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge (TS): abstract class — same shape, must still report ----
		{
			Code: `
        abstract class Foo extends React.PureComponent {
          shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge (TS): decorated class — decorators are siblings of the
		// class node, must not affect detection ----
		{
			Code: `
        @decorator
        class Foo extends React.PureComponent {
          shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge (TS): overload signatures + implementation — Members()
		// includes both the bodyless signature and the implementation; we
		// only need ONE Identifier-keyed match so we still report. ----
		{
			Code: `
        class Foo extends React.PureComponent {
          shouldComponentUpdate(): boolean;
          shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge: multi-level parens around extends — SkipParentheses must
		// peel every level, matching ESLint's identifier-text-only regex match. ----
		{
			Code: `
        class Foo extends ((React.PureComponent)) {
          shouldComponentUpdate() { return true; }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Foo does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    9,
				},
			},
		},

		// ---- Edge: deeply nested ClassDeclaration inside object method body —
		// listener must still fire on the inner class. ----
		{
			Code: `
        const utils = {
          make() {
            class Inner extends React.PureComponent {
              shouldComponentUpdate() { return true; }
            }
            return Inner;
          }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   "Inner does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      4,
					Column:    13,
				},
			},
		},

		// ---- Edge: anonymous class inside `module.exports = class extends PureComponent {...}`
		// — parent is BinaryExpression, not VariableDeclaration → empty name. ----
		{
			Code: `
        module.exports = class extends React.PureComponent {
          shouldComponentUpdate() { return true; }
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noShouldCompUpdate",
					Message:   " does not need shouldComponentUpdate when extending React.PureComponent.",
					Line:      2,
					Column:    26,
				},
			},
		},
	})
}
