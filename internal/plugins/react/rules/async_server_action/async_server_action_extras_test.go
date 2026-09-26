package async_server_action

// cspell:ignore actionasync ction

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// These AST adaptations were checked against the pinned upstream rule using
// ESLint 10.9.0 and @typescript-eslint/parser 8.65.0.
func TestAsyncServerActionExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AsyncServerActionRule,
		[]rule_tester.ValidTestCase{
			// An angle-bracket TypeScript assertion is not a string literal.
			{Code: `function action() { (<string>'use server'); }`,
				FileName: `review-assertion.ts`},

			// A file-level directive does not make every function a match.
			{Code: `'use server'; export function action() { return save(); }`,
				Tsx: true},
			// Only a block directly owned by a function is inspected.
			{Code: `function action() { { 'use server'; } }
if (ready) { 'use server'; }
class C { static { 'use server'; } }`,
				Tsx: true},
			// An empty statement still counts as the first statement.
			{Code: `function action() { ; 'use server'; }`,
				Tsx: true},
			{Code: `function action() { return 'use server'; }`,
				Tsx: true},
			{Code: `function action() { 'use ' + 'server'; }`,
				Tsx: true},
			// Authored TypeScript wrappers retain their ESTree node kind.
			{Code: `function action() { 'use server' as string; }`,
				Tsx: true},
			{Code: `function action() { 'use server'!; }`,
				Tsx: true},
			{Code: `function action() { 'use server' satisfies string; }`,
				Tsx: true},
			// Declarations without bodies are not function blocks.
			{Code: `declare function action(): void;
abstract class C { abstract action(): void; }
interface I { action(): void; }`,
				Tsx: true},
			// Method generators are excluded, just like function generators.
			{Code: `const obj = { *action() { 'use server'; }, async *other() { 'use server'; } };
class C { *action() { 'use server'; } async *other() { 'use server'; } }`,
				Tsx: true},
			{Code: `class C { action = async () => { 'use server'; }; }`,
				Tsx: true},
			{Code: `async function action() { 'use server'; }`,
				Tsx:     true,
				Options: []any{}}},
		[]rule_tester.InvalidTestCase{
			// A method named async is still synchronous.
			{Code: `const obj = { async() { 'use server'; } }; class C { static async() { 'use server'; } }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 20, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { async async() { 'use server'; } }; class C { static async() { 'use server'; } }`}}}, {MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 66, EndLine: 1, EndColumn: 86, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { async() { 'use server'; } }; class C { static async async() { 'use server'; } }`}}}}},

			// The async identifier can also be an ordinary arrow parameter.
			{Code: `const action = async => { 'use server'; return async; };`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 16, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const action = async async => { 'use server'; return async; };`}}}}},

			// A line break separates an async expression from the following declaration.
			{Code: `async
function action() { 'use server'; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 2, Column: 1, EndLine: 2, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `async
async function action() { 'use server'; }`}}}}},

			// An action in a default parameter has its own function body.
			{Code: `function action(callback = () => { 'use server'; }) {}`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 28, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `function action(callback = async () => { 'use server'; }) {}`}}}}},

			// Parenthesized default exports report and edit only the inner function.
			{Code: `export default (function action() { 'use server'; });`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 17, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `export default (async function action() { 'use server'; });`}}}}},

			// An optional method marker precedes the function range.
			{Code: `class C { action?(): void { 'use server'; } }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 18, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { async action?(): void { 'use server'; } }`}}}}},

			// A typed computed key retains its assertion and has no identifier name.
			{Code: `const obj = { [(action as any)]() { 'use server'; } };`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 32, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { [(async action as any)]() { 'use server'; } };`}}}}},

			// Escapes in quoted constructor keys still select the anonymous suggestion.
			{Code: `class C { 'constr\u0075ctor'() { 'use server'; } }`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 29, EndLine: 1, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { async 'constr\u0075ctor'() { 'use server'; } }`}}}}},

			// BOM, CRLF, and a non-BMP comment preserve UTF-16 diagnostic ranges.
			{Code: "\ufeff// 😀\r\nexport default function action() {\r\n  'use server';\r\n}",
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 2, Column: 16, EndLine: 4, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: "\ufeff// 😀\r\nexport default async function action() {\r\n  'use server';\r\n}"}}}}},

			// A string line continuation is decoded before matching the directive.
			{Code: `function action() { 'use ser\
ver'; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 1, EndLine: 2, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `async function action() { 'use ser\
ver'; }`}}}}},

			// JavaScript JSDoc satisfies casts are transparent, unlike authored TypeScript.
			{Code: `function action() { (/** @satisfies {string} */ ('use server')); }`,
				FileName: `review-jsdoc.js`,
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 1, EndLine: 1, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `async function action() { (/** @satisfies {string} */ ('use server')); }`}}}}},

			// A function in a computed key and its enclosing method both report.
			{Code: `const obj = { [function action() { 'use server'; }]() { 'use server'; } };`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 16, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { [async function action() { 'use server'; }]() { 'use server'; } };`}}}, {MessageId: "asyncServerAction", Message: "Server Actions must be async", Line: 1, Column: 52, EndLine: 1, EndColumn: 72, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { [async function action() { 'use server'; }]() { 'use server'; } };`}}}}},

			// tsgo folds quoted constructor keys into Constructor declarations.
			{
				Code: `class C { 'constructor'() { 'use server'; } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "asyncServerAction", Message: "Server Actions must be async",
					Line: 1, Column: 24, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { async 'constructor'() { 'use server'; } }`}},
				}},
			},
			// Parentheses do not hide the first expression, and comments are not statements.
			{Code: `function action() { /* comment */ ('use server'); }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 52,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `async function action() { /* comment */ ('use server'); }`}},
				}}},
			// Compare the decoded string value.
			{Code: `function action() { 'use\u0020server'; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `async function action() { 'use\u0020server'; }`}},
				}}},
			// The arrow range excludes outer parentheses but keeps a bare parameter.
			{Code: `const action = ((value => { 'use server'; return value; }));`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 18, EndLine: 1, EndColumn: 58,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const action = ((async value => { 'use server'; return value; }));`}},
				}}},
			{Code: `const action = (function named() { 'use server'; });`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 17, EndLine: 1, EndColumn: 51,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const action = (async function named() { 'use server'; });`}},
				}}},
			// Both nested functions report independently.
			{Code: `function outer() { 'use server'; function inner() { 'use server'; } }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 70,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `async function outer() { 'use server'; function inner() { 'use server'; } }`}},
				}, {
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 34, EndLine: 1, EndColumn: 68,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `function outer() { 'use server'; async function inner() { 'use server'; } }`}},
				}}},
			// An ordinary property value is a function expression, not a method.
			{Code: `const obj = { action: function named() { 'use server'; } };`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 23, EndLine: 1, EndColumn: 57,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { action: async function named() { 'use server'; } };`}},
				}}},
			// Private keys do not supply an identifier name; field values stay anonymous.
			{Code: `class C { #action() { 'use server'; } field = () => { 'use server'; }; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 18, EndLine: 1, EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { async #action() { 'use server'; } field = () => { 'use server'; }; }`}},
				}, {
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 47, EndLine: 1, EndColumn: 70,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { #action() { 'use server'; } field = async () => { 'use server'; }; }`}},
				}}},
			// Literal method keys use the anonymous suggestion.
			{Code: `const obj = { 'action'() { 'use server'; }, 1() { 'use server'; } };`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 23, EndLine: 1, EndColumn: 43,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { async 'action'() { 'use server'; }, 1() { 'use server'; } };`}},
				}, {
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 46, EndLine: 1, EndColumn: 66,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { 'action'() { 'use server'; }, async 1() { 'use server'; } };`}},
				}}},
			// Preserve upstream insertion before the ESTree key, even inside computed brackets.
			{Code: `class C { static [action]() { 'use server'; } [('other')]() { 'use server'; } }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 26, EndLine: 1, EndColumn: 46,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { static [async action]() { 'use server'; } [('other')]() { 'use server'; } }`}},
				}, {
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 58, EndLine: 1, EndColumn: 78,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { static [action]() { 'use server'; } [(async 'other')]() { 'use server'; } }`}},
				}}},
			{Code: `const obj = { [getKey()]() { 'use server'; } };`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 25, EndLine: 1, EndColumn: 45,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { [async getKey()]() { 'use server'; } };`}},
				}}},
			// Upstream offers suggestions for constructors and accessors, even though async is invalid here.
			{Code: `class C { constructor() { 'use server'; } get action() { 'use server'; } set action(value) { 'use server'; } }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 22, EndLine: 1, EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { async constructor() { 'use server'; } get action() { 'use server'; } set action(value) { 'use server'; } }`}},
				}, {
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 53, EndLine: 1, EndColumn: 73,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { constructor() { 'use server'; } get async action() { 'use server'; } set action(value) { 'use server'; } }`}},
				}, {
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 84, EndLine: 1, EndColumn: 109,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { constructor() { 'use server'; } get action() { 'use server'; } set async action(value) { 'use server'; } }`}},
				}}},
			// Object accessors have Property.method=false upstream and use the function-value insertion point.
			{Code: `const obj = { get action() { 'use server'; }, set action(value) { 'use server'; } };`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 25, EndLine: 1, EndColumn: 45,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { get actionasync () { 'use server'; }, set action(value) { 'use server'; } };`}},
				}, {
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 57, EndLine: 1, EndColumn: 82,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { get action() { 'use server'; }, set actionasync (value) { 'use server'; } };`}},
				}}},
			// Export modifiers and trivia stay outside the function range.
			{Code: `// leading comment
export /* keep */ default function action() {
  'use server';
}`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 2, Column: 27, EndLine: 4, EndColumn: 2,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `// leading comment
export /* keep */ default async function action() {
  'use server';
}`}},
				}}},
			// Columns use UTF-16 while the name retains its decoded Unicode text.
			{Code: `const emoji = '😀'; function 𐐀() { 'use server'; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 21, EndLine: 1, EndColumn: 52,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const emoji = '😀'; async function 𐐀() { 'use server'; }`}},
				}}},
			{Code: `function \u0061ction() { 'use server'; }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `async function \u0061ction() { 'use server'; }`}},
				}}},
			// Generic arrow heads belong to the function range.
			{Code: `const action = <T,>(value: T): T => { 'use server'; return value; };`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 16, EndLine: 1, EndColumn: 68,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const action = async <T,>(value: T): T => { 'use server'; return value; };`}},
				}}},
			// Method keys and decorators stay outside the ESTree function range.
			{Code: `class C { @decorator public action<T>(value: T): T { 'use server'; return value; } }`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 35, EndLine: 1, EndColumn: 83,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `class C { @decorator public async action<T>(value: T): T { 'use server'; return value; } }`}},
				}}},
			{Code: `const obj = { action /* keep */ <T>(value: T) { 'use server'; return value; } };`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 33, EndLine: 1, EndColumn: 78,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `const obj = { async action /* keep */ <T>(value: T) { 'use server'; return value; } };`}},
				}}},
			// JavaScript JSDoc cast wrappers are transparent.
			{Code: `function action() { /** @type {string} */ ('use server'); }`,
				FileName: `action.js`,
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
					Line: 1, Column: 1, EndLine: 1, EndColumn: 60,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{Output: `async function action() { /** @type {string} */ ('use server'); }`}},
				}}}})
}

func TestAsyncServerActionEditDemand(t *testing.T) {
	for _, test := range []struct {
		name, code, message, output string
	}{
		{"named", `function action() { 'use server'; }`, "Make `action` an `async` function", `async function action() { 'use server'; }`},
		{"anonymous", `const action = () => { 'use server'; };`, "Make this function `async`", `const action = async () => { 'use server'; };`},
		{"method", `class C { static action() { 'use server'; } }`, "Make `action` an `async` function", `class C { static async action() { 'use server'; } }`},
		{"constructor", `class C { public constructor() { 'use server'; } }`, "Make `constructor` an `async` function", `class C { public async constructor() { 'use server'; } }`},
		{"quoted constructor", `class C { 'constructor'() { 'use server'; } }`, "Make this function `async`", `class C { async 'constructor'() { 'use server'; } }`},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/action.ts", Path: "/action.ts",
			}, test.code, core.ScriptKindTS)
			run := func(demand rule.EditDemand) rule.RuleDiagnostic {
				t.Helper()
				var diagnostics []rule.RuleDiagnostic
				ctx := rule.RuleContext{
					SourceFile:     sourceFile,
					DisableManager: rule.NewDisableManager(sourceFile, rule.NewCommentStore(sourceFile)),
				}.WithDiagnosticConsumer(AsyncServerActionRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Demand: demand,
					Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
				})
				listeners := AsyncServerActionRule.Run(ctx, nil)
				var visit func(*ast.Node) bool
				visit = func(node *ast.Node) bool {
					if listener := listeners[node.Kind]; listener != nil {
						listener(node)
					}
					return node.ForEachChild(visit)
				}
				sourceFile.AsNode().ForEachChild(visit)
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
				}
				return diagnostics[0]
			}
			all := run(rule.EditDemandAll)
			if all.Suggestions == nil || len(*all.Suggestions) != 1 {
				t.Fatal("expected one suggestion")
			}
			suggestion := (*all.Suggestions)[0]
			if suggestion.Message.Id != "" || suggestion.Message.Description != test.message {
				t.Fatalf("unexpected suggestion message: %#v", suggestion.Message)
			}
			output, _, _ := linter.ApplyRuleFixes(test.code, []rule.RuleSuggestion{suggestion})
			if output != test.output {
				t.Fatalf("suggestion output = %q, want %q", output, test.output)
			}
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				got := run(demand)
				if got.FixesPtr != nil {
					t.Fatalf("demand %d produced an automatic fix", demand)
				}
				if demand&rule.EditDemandSuggestion != 0 {
					if !reflect.DeepEqual(got.Suggestions, all.Suggestions) {
						t.Fatalf("demand %d changed suggestions", demand)
					}
				} else if got.Suggestions != nil {
					t.Fatalf("demand %d produced unexpected suggestions", demand)
				}
				want := all
				want.Suggestions, got.Suggestions = nil, nil
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("demand %d changed diagnostic metadata", demand)
				}
			}
		})
	}
}
