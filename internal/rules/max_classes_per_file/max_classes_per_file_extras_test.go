package max_classes_per_file

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestMaxClassesPerFileExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
func TestMaxClassesPerFileExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxClassesPerFileRule,
		[]rule_tester.ValidTestCase{
			// N/A: rule counts ClassDeclaration/ClassExpression node kinds
			// directly; it never reads a property key or walks a member/
			// element-access chain, so the "Access / key forms" and
			// "Optional chain" Dimension 4 rows do not apply.

			// N/A: the rule never inspects function bodies, destructuring
			// patterns, spreads, or class members, so "Graceful degradation"
			// (SpreadAssignment / RestElement / empty bodies / overload
			// signatures) does not apply.

			// ---- Dimension 4: empty Program body ----
			{Code: ""},

			// ---- Dimension 4: receiver wrappers around a class expression ----
			// A single wrapped class expression still counts as exactly one
			// class, so it stays under the default max of 1.
			{Code: "var x = (class {});"},
			{Code: "var x = (class {})!;"},
			{Code: "var x = (class {}) satisfies object;"},

			// ---- Branch: classCount == max is not > max (boundary), object
			// option form with both `max` and `ignoreExpressions` set ----
			{Code: "class Foo {}\nclass Bar {}", Options: option(map[string]interface{}{"ignoreExpressions": true, "max": 2})},

			// ---- Options contract: explicit `ignoreExpressions: false` behaves
			// the same as the (implicit-default) unset case -- expressions
			// still count towards max. ----
			{Code: "var x = class {};\nvar y = class {};", Options: option(map[string]interface{}{"ignoreExpressions": false, "max": 2})},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: `as`-wrapped class expressions still count ----
			{
				Code: "var x = (class {}) as any;\nvar y = (class {}) as any;",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 27,
				}},
			},

			// ---- Dimension 4: declaration/container forms ----
			// `export default class {}` parses as an (anonymous)
			// ClassDeclaration, not a ClassExpression, so it counts the same
			// as a named declaration.
			{
				Code: "export default class {}\nclass Named {}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 15,
				}},
			},
			// A named class expression (`class Named {}` as an expression)
			// still counts as a ClassExpression, not a ClassDeclaration.
			{
				Code: "var x = class Named {};\nvar y = class Named2 {};",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 25,
				}},
			},

			// ---- Dimension 1: TypeScript-specific class syntax ----
			{
				Code: "abstract class Foo {}\nabstract class Bar {}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 22,
				}},
			},
			{
				Code: "declare class Foo {}\ndeclare class Bar {}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 21,
				}},
			},
			{
				Code: "class Foo<T> {}\nclass Bar<T> {}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 16,
				}},
			},

			// ---- Dimension 2: nesting boundaries ----
			// A class declared inside a method body of another class still
			// counts; the rule does not stop at the enclosing class boundary.
			{
				Code: "class Outer {\n  method() {\n    class Inner {}\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 5, EndColumn: 2,
				}},
			},
			// A class declared inside a function body counts too, and the
			// reported range still spans the Program's own top-level
			// statements, not the classes themselves.
			{
				Code: "function factory() {\n  class Local {}\n  return Local;\n}\nclass Top {}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 5, EndColumn: 13,
				}},
			},
			// Deeply nested (3+ levels): a ClassExpression returned from a
			// function, whose own method contains a nested function
			// declaring a further ClassDeclaration. Both are still counted
			// even though there is a single top-level statement.
			{
				Code: "function outer() {\n  return class Level1 {\n    method() {\n      function inner() {\n        class Level2 {}\n        return Level2;\n      }\n      return inner;\n    }\n  };\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 11, EndColumn: 2,
				}},
			},

			// ---- Locks in upstream parseOptions() arm: object option with
			// only `max` (no `ignoreExpressions` key) leaves ignoreExpressions
			// undefined, which is falsy -- class expressions still count. ----
			{
				Code:    "var x = class {};\nvar y = class {};",
				Options: option(map[string]interface{}{"max": 1}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 18,
				}},
			},
			// ---- Locks in upstream parseOptions() arm: object option with
			// only `ignoreExpressions` (no `max` key) falls back to the
			// `option.max || 1` default. ----
			{
				Code:    "class Foo {}\nclass Bar {}",
				Options: option(map[string]interface{}{"ignoreExpressions": true}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 13,
				}},
			},

			// ---- Options contract: an empty options object `{}` behaves
			// identically to omitting options entirely -- both fall back to
			// the same defaults (max: 1, ignoreExpressions: false). ----
			{
				Code:    "class Foo {}\nclass Bar {}",
				Options: option(map[string]interface{}{}),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 2, EndColumn: 13,
				}},
			},

			// ---- Real-user: barrel-style module exporting several model
			// classes from one file (a common source of real violations). ----
			{
				Code:    "export class UserModel {}\nexport class PostModel {}\nexport class CommentModel {}",
				Options: option(2),
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(3, 2),
					Line:      1, Column: 1, EndLine: 3, EndColumn: 29,
				}},
			},
			// ---- Real-user: TypeScript mixin factories, each returning an
			// anonymous class expression -- a common false-positive
			// complaint against this rule with default options. ----
			{
				Code: "function Serializable(Base) {\n  return class extends Base {\n    serialize() {}\n  };\n}\nfunction Comparable(Base) {\n  return class extends Base {\n    compare() {}\n  };\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "maximumExceeded",
					Message:   maximumExceededMessage(2, 1),
					Line:      1, Column: 1, EndLine: 10, EndColumn: 2,
				}},
			},
		},
	)
}
