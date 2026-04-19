package radix

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRadixRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RadixRule,
		[]rule_tester.ValidTestCase{
			// ---- Valid: integer radix literals ----
			{Code: `parseInt("10", 10);`},
			{Code: `parseInt("10", 2);`},
			{Code: `parseInt("10", 36);`},
			{Code: `parseInt("10", 0x10);`},
			{Code: `parseInt("10", 1.6e1);`},
			{Code: `parseInt("10", 10.0);`},

			// ---- Valid: variable radix (cannot determine statically) ----
			{Code: `parseInt("10", foo);`},
			{Code: `Number.parseInt("10", foo);`},

			// ---- Valid: non-literal radix expressions (ESLint only rejects Literal / bare undefined) ----
			{Code: `parseInt("10", -1);`},             // UnaryExpression
			{Code: `parseInt("10", x + 1);`},          // BinaryExpression
			{Code: `parseInt("10", x ? 10 : 16);`},    // ConditionalExpression
			{Code: `parseInt("10", Math.floor(x));`},  // CallExpression
			{Code: "parseInt(\"10\", `10`);"},         // NoSubstitutionTemplateLiteral
			{Code: `parseInt("10", NaN);`},            // Identifier other than `undefined`
			{Code: `parseInt("10", Infinity);`},       // Identifier
			{Code: `parseInt("10", (10));`},           // Parenthesized numeric literal
			{Code: `parseInt("10", ((10)));`},         // Deeply parenthesized

			// ---- Valid: numeric edge cases inside [2, 36] ----
			{Code: `parseInt("10", 2.0);`},
			{Code: `parseInt("10", 0b10);`},  // 2
			{Code: `parseInt("10", 0o10);`},  // 8
			{Code: `parseInt("10", 0x24);`},  // 36

			// ---- Valid: more than 2 arguments (rule only inspects args[1]) ----
			{Code: `parseInt("10", 10, extra);`},
			{Code: `Number.parseInt("10", 10, extra);`},

			// ---- Valid: not a call of the tracked functions ----
			{Code: `parseInt`},
			{Code: `Number.foo();`},
			{Code: `Number[parseInt]();`},

			// ---- Valid: private identifier Number.#parseInt (not the global) ----
			{Code: `class C { #parseInt; foo() { Number.#parseInt(); } }`},
			{Code: `class C { #parseInt; foo() { Number.#parseInt(foo); } }`},
			{Code: `class C { #parseInt; foo() { Number.#parseInt(foo, 'bar'); } }`},

			// ---- Valid: shadowing (various declaration kinds) ----
			{Code: `var parseInt; parseInt();`},
			{Code: `var Number; Number.parseInt();`},
			{Code: `let parseInt = foo; parseInt();`},
			{Code: `const parseInt = foo; parseInt();`},
			{Code: `function parseInt() {} parseInt();`},
			{Code: `function f(parseInt) { parseInt(); }`},
			{Code: `function f(Number) { Number.parseInt('x', 1); }`},
			{Code: `import parseInt from 'x'; parseInt();`},
			{Code: `import { parseInt } from 'x'; parseInt();`},
			{Code: `function f() { var parseInt; parseInt(); }`},
			{Code: `function f() { var Number = {}; Number.parseInt('x', 1); }`},
			// Shadow at intermediate scope blocks outer calls too
			{Code: `function f() { var parseInt = g; function h() { parseInt(); } }`},

			// ---- Valid: parseInt used in non-call positions ----
			{Code: `const obj = { parseInt: 1 };`},
			{Code: `const { parseInt } = obj;`},
			{Code: `obj.parseInt();`},
			{Code: `foo.parseInt('10');`},

			// ---- Valid: NewExpression / TaggedTemplate are NOT call expressions ----
			{Code: `new parseInt("10");`},
			{Code: "parseInt`10`;"},

			// ---- Valid: additional shadowing scopes ----
			{Code: `try {} catch (parseInt) { parseInt(); }`},
			{Code: `for (let parseInt = 0; parseInt < 1; parseInt++) {}`},
			{Code: `for (const parseInt of []) { parseInt(); }`},
			{Code: `for (const parseInt in {}) { parseInt(); }`},
			{Code: `{ var parseInt = foo; parseInt(); }`}, // var hoists to source file

			// ---- Valid: named function expression / class expression binding ----
			{Code: `var fn = function parseInt() { parseInt(); };`},
			{Code: `var fn = function Number() { Number.parseInt("x", 1); };`},
			{Code: `const C = class parseInt { static foo() { parseInt(); } };`},
			{Code: `const C = class Number { static foo() { Number.parseInt("x", 1); } };`},

			// ---- Valid: destructuring binds parseInt ----
			{Code: `const [parseInt] = arr; parseInt();`},
			{Code: `const [...parseInt] = arr; parseInt();`},
			{Code: `const { parseInt = foo } = obj; parseInt();`},
			{Code: `const { x: parseInt } = obj; parseInt();`},

			// ---- Valid: TS declare / export / namespace declarations ----
			{Code: `declare const parseInt: any; parseInt();`},
			{Code: `export const parseInt = foo; parseInt();`},

			// ---- Valid: Unicode-escaped identifier still resolves to parseInt ----
			// tsgo normalizes the Identifier.Text to "parseInt", so a local
			// declaration using the escaped form shadows the global.
			{Code: `var \u0070arseInt = foo; parseInt();`},

			// ---- Valid: TS-only wrappers on callee — ESLint doesn't treat
			//      these as an Identifier callee, so it never flags them.
			//      We mirror that to stay 1:1 with ESLint. ----
			{Code: `parseInt!("10");`},
			{Code: `(parseInt as Function)("10");`},
			{Code: `Number!.parseInt("10");`},
			{Code: `(Number as any).parseInt("10");`},

			// ---- Valid: labeled statement does NOT create a binding ----
			// `parseInt:` here is a label, not a variable, so the rule still
			// treats the inner `parseInt()` as a real call — but the outer
			// labeled statement does not shadow.
			// (Placed in valid because inside the loop we pass a radix.)
			{Code: `parseInt: for (;;) { parseInt("10", 10); break parseInt; }`},

			// ---- Valid: Number used in non-method-access positions ----
			{Code: `const x = Number;`},
			{Code: `const x = Number(foo);`},

			// ---- Valid: deprecated options "always" / "as-needed" (no behavior change) ----
			{Code: `parseInt("10", 10);`, Options: []any{"always"}},
			{Code: `parseInt("10", 10);`, Options: []any{"as-needed"}},
			{Code: `parseInt("10", 8);`, Options: []any{"always"}},
			{Code: `parseInt("10", 8);`, Options: []any{"as-needed"}},
			{Code: `parseInt("10", foo);`, Options: []any{"always"}},
			{Code: `parseInt("10", foo);`, Options: []any{"as-needed"}},

			// SKIP: rslint does not support ESLint's `/*global ... */` directive
			// SKIP: rslint does not support ESLint's `languageOptions.globals` override
			{Code: `/* globals parseInt:off */ parseInt(foo);`, Skip: true},
			{Code: `Number.parseInt(foo);`, Skip: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Missing parameters (+ exact message + full range) ----
			{
				Code: `parseInt();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingParameters",
						Message:   "Missing parameters.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 11,
					},
				},
			},

			// ---- Missing radix with suggestion (+ exact message + full range) ----
			{
				Code: `parseInt("10");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Message:   "Missing radix parameter.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `parseInt("10", 10);`,
							},
						},
					},
				},
			},
			{
				// Trailing comma: suggestion should preserve it.
				Code: `parseInt("10",);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `parseInt("10", 10,);`,
							},
						},
					},
				},
			},
			{
				// Sequence expression (no trailing comma).
				Code: `parseInt((0, "10"));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `parseInt((0, "10"), 10);`,
							},
						},
					},
				},
			},
			{
				// Sequence expression with trailing comma.
				Code: `parseInt((0, "10"),);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `parseInt((0, "10"), 10,);`,
							},
						},
					},
				},
			},

			// ---- Invalid radix literals (+ exact message + full range) ----
			{
				Code: `parseInt("10", null);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidRadix",
						Message:   "Invalid radix parameter, must be an integer between 2 and 36.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			{
				Code: `parseInt("10", undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code: `parseInt("10", true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code: `parseInt("10", "foo");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code: `parseInt("10", "123");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code: `parseInt("10", 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code: `parseInt("10", 37);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code: `parseInt("10", 10.5);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},

			// ---- Number.parseInt ----
			{
				Code: `Number.parseInt();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingParameters", Line: 1, Column: 1},
				},
			},
			{
				Code: `Number.parseInt("10");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `Number.parseInt("10", 10);`,
							},
						},
					},
				},
			},
			{
				Code: `Number.parseInt("10", 1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code: `Number.parseInt("10", 37);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code: `Number.parseInt("10", 10.5);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},

			// ---- Optional chaining ----
			{
				Code: `parseInt?.("10");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `parseInt?.("10", 10);`,
							},
						},
					},
				},
			},
			{
				Code: `Number.parseInt?.("10");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `Number.parseInt?.("10", 10);`,
							},
						},
					},
				},
			},
			{
				Code: `Number?.parseInt("10");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `Number?.parseInt("10", 10);`,
							},
						},
					},
				},
			},
			{
				Code: `(Number?.parseInt)("10");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    `(Number?.parseInt)("10", 10);`,
							},
						},
					},
				},
			},

			// ---- Deprecated options still trigger (no behavior change) ----
			{
				Code:    `parseInt();`,
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingParameters", Line: 1, Column: 1},
				},
			},
			{
				Code:    `parseInt();`,
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingParameters", Line: 1, Column: 1},
				},
			},
			{
				Code:    `parseInt("10");`,
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `parseInt("10", 10);`},
						},
					},
				},
			},
			{
				Code:    `parseInt("10");`,
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `parseInt("10", 10);`},
						},
					},
				},
			},
			{
				Code:    `parseInt("10", 1);`,
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code:    `parseInt("10", 1);`,
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				Code:    `Number.parseInt();`,
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingParameters", Line: 1, Column: 1},
				},
			},
			{
				Code:    `Number.parseInt();`,
				Options: []any{"as-needed"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingParameters", Line: 1, Column: 1},
				},
			},

			// ---- Extra edge cases: position assertions with multi-line and various containers ----
			{
				Code: "foo(\n  parseInt()\n);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingParameters", Line: 2, Column: 3},
				},
			},
			{
				Code: "function f() { return parseInt('10'); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addRadixParameter10",
								Output:    "function f() { return parseInt('10', 10); }",
							},
						},
					},
				},
			},
			// Parenthesized callee
			{
				Code: `(parseInt)("10");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `(parseInt)("10", 10);`},
						},
					},
				},
			},
			// BigInt literal as radix — not a valid integer-in-range literal
			{
				Code: `parseInt("10", 10n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},

			// ---- Numeric literal edge cases ----
			{
				// 0 is below range.
				Code: `parseInt("10", 0);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},
			{
				// 0x25 = 37, above range.
				Code: `parseInt("10", 0x25);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidRadix", Line: 1, Column: 1},
				},
			},

			// ---- Multiple errors in one file ----
			{
				Code: "parseInt();\nparseInt('10');\nparseInt('10', 1);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingParameters", Line: 1, Column: 1},
					{
						MessageId: "missingRadix",
						Line:      2,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: "parseInt();\nparseInt('10', 10);\nparseInt('10', 1);"},
						},
					},
					{MessageId: "invalidRadix", Line: 3, Column: 1},
				},
			},

			// ---- Nested parseInt calls — both should be flagged ----
			{
				Code: `parseInt(parseInt("10"));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `parseInt(parseInt("10"), 10);`},
						},
					},
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `parseInt(parseInt("10", 10));`},
						},
					},
				},
			},

			// ---- Class / arrow / IIFE contexts ----
			{
				Code: `class C { foo() { parseInt('x'); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `class C { foo() { parseInt('x', 10); } }`},
						},
					},
				},
			},
			{
				Code: `class C { x = parseInt('x'); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `class C { x = parseInt('x', 10); }`},
						},
					},
				},
			},
			{
				Code: `class C { static { parseInt('x'); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `class C { static { parseInt('x', 10); } }`},
						},
					},
				},
			},
			{
				Code: `const f = () => parseInt('x');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `const f = () => parseInt('x', 10);`},
						},
					},
				},
			},
			{
				Code: `(function () { parseInt('x'); })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `(function () { parseInt('x', 10); })();`},
						},
					},
				},
			},

			// ---- Shadow scope does not reach: inner call still flagged ----
			{
				Code: `function f() { parseInt(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingParameters", Line: 1, Column: 16},
				},
			},
			{
				// parseInt shadowed in inner function only; outer call still flagged.
				Code: "parseInt('x');\nfunction f() { var parseInt; parseInt(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: "parseInt('x', 10);\nfunction f() { var parseInt; parseInt(); }"},
						},
					},
				},
			},

			// ---- Suggestion with comments between tokens ----
			{
				// Comment inside the argument list should be preserved by the fix.
				Code: `parseInt("10" /* hi */);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `parseInt("10" /* hi */, 10);`},
						},
					},
				},
			},

			// ---- Spread argument (kept for ESLint parity even though the
			//      suggestion is semantically questionable — ESLint emits the
			//      same suggestion) ----
			{
				Code: `parseInt(...args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `parseInt(...args, 10);`},
						},
					},
				},
			},
			// SpreadElement as args[1] is not a Literal / Identifier, so it is
			// treated as valid (falls into the default branch).

			// ---- Real-world containers: async / generator / object method /
			//      throw / template literal arg / computed key / labeled stmt ----
			{
				Code: `async function f() { return parseInt("x"); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `async function f() { return parseInt("x", 10); }`},
						},
					},
				},
			},
			{
				Code: `function* g() { yield parseInt("x"); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `function* g() { yield parseInt("x", 10); }`},
						},
					},
				},
			},
			{
				Code: `const obj = { m() { parseInt("x"); } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `const obj = { m() { parseInt("x", 10); } };`},
						},
					},
				},
			},
			{
				Code: `throw parseInt("x");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `throw parseInt("x", 10);`},
						},
					},
				},
			},
			{
				Code: "const x = parseInt(`${foo}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: "const x = parseInt(`${foo}`, 10);"},
						},
					},
				},
			},
			{
				// parseInt inside a computed class member key.
				Code: `class C { [parseInt("x")]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `class C { [parseInt("x", 10)]() {} }`},
						},
					},
				},
			},
			{
				// Default parameter value expression.
				Code: `function f(x = parseInt("x")) { return x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `function f(x = parseInt("x", 10)) { return x; }`},
						},
					},
				},
			},
			{
				// Export default.
				Code: `export default parseInt("x");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingRadix",
						Line:      1,
						Column:    16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addRadixParameter10", Output: `export default parseInt("x", 10);`},
						},
					},
				},
			},

			// ---- Multi-line case with EndLine/EndColumn assertion ----
			{
				Code: "parseInt(\n  \"10\",\n  37\n);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalidRadix",
						Line:      1,
						Column:    1,
						EndLine:   4,
						EndColumn: 2,
					},
				},
			},
		},
	)
}
