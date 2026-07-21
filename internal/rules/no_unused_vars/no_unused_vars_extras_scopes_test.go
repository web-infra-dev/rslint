// TestNoUnusedVarsExtrasScopes covers scope ownership and execution boundaries
// that are easy to misclassify with ancestor-only checks. It includes nested
// bindings, static scopes, inline globals, JSX names, and discarded self-updates.
package no_unused_vars

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnusedVarsExtrasScopes(t *testing.T) {
	signaturesCode := "type Fn = (callback: string) => void;\n" +
		"interface I { method(methodArg: number): void; }\n" +
		"abstract class C { abstract run(abstractArg: boolean): void; }\n" +
		"consume(null as Fn); consume(null as I); consume(C);"
	overloadsCode := "function overloaded(overloadArg: string): void;\n" +
		"function overloaded(implementationArg: string) {}\n" +
		"overloaded(\"x\");\n" +
		"function withThis(this: unknown) {}\n" +
		"withThis();\n" +
		"class ParameterProperty { constructor(private propertyArg?: string) {} }\n" +
		"consume(ParameterProperty);"
	ambientCode := "declare namespace N { const value: string; function f(arg: number): void; }\nconsume(N);"
	rangesCode := "let definite!: string;\n" +
		"function f(\n" +
		"  value: {\n" +
		"    key: string;\n" +
		"  },\n" +
		"  ...rest: string[]\n" +
		") {}\n" +
		"f();"
	typeScopeShadowCode := `function outer(context: string, callback: (context: number) => void) { consume(context); consume(callback); } outer("x", () => {});`
	mappedAndInferCode := `type A<T> = { [K in keyof T]: T[K] }; type B<T> = { [P in keyof T]: never }; type C<T> = T extends infer U ? U : never; type D<T> = T extends infer V ? string : never; consume(null as A<{}>); consume(null as B<{}>); consume(null as C<string>); consume(null as D<string>);`
	enumCode := `export enum E { A = 1, B = A, C = 3 }`
	globalAugmentationCode := `declare global { const X: string; interface I {} } consume(X); consume(null as I);`

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// Only direct properties beside an object rest are ignored.
			{Code: `const { value: direct, ...rest } = source; console.log(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},
			{Code: `let direct; let rest; ({ value: direct, ...rest } = source); console.log(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},

			// A static block is its own variable scope, so an outer value read
			// there can be observed independently of the outer declaration.
			{Code: `let x = 0; class C { static { x = x + 1; } } new C();`},
			{Code: `let x = 0; namespace N { x = x + 1; } consume(N);`},

			// A callback passed through the RHS can execute later and stores its read.
			{Code: `let x; x = consume(() => x);`},

			// Local shadows do not consume an inline global, but a real global read does.
			{Code: `/*global foo*/ function f(foo) { return foo; } consume(foo); f(1);`},
			{Code: `/*global foo*/ consume(foo);`},
			{Code: `/*global Foo*/ type Alias = Foo; consume({} as Alias);`},

			// Every binding introduced by an exported destructuring declaration is exported.
			{Code: `export const { nested: { value }, list: [item] } = source;`},

			// ---- Real-user: rspack RsdoctorPlugin nested namespace type exports ----
			// A local export specifier can be nested arbitrarily deeply inside
			// namespaces. It still consumes the exact local symbol it resolves to.
			{Code: `import { type T } from "./foo"; export declare namespace Data { export type { T as Alias }; }`},
			{Code: `type A = {}; export namespace Outer { export namespace Inner { export type { A as B }; } }`},
			{Code: `type A = {}; namespace N { export { type A as B }; } consume(N);`},
			{Code: `type A = {}; type C = {}; export namespace N { export type { A as B, C as D }; }`},

			// Core ESLint deliberately ignores the syntactically required setter
			// parameter even when TypeScript gives it a type annotation.
			{Code: `const target = { set value(input: string) {} }; consume(target);`},
			// Index-signature keys are property placeholders, not scope variables.
			{Code: `interface Indexable { [key: string]: unknown; } consume(null as Indexable);`},
			// A type-predicate reference consumes its parameter inside the otherwise
			// bodyless signature scope.
			{Code: `type Guard = (value: unknown) => value is string; consume(null as Guard);`},
			{Code: `declare function guard(value: unknown): value is string; consume(guard);`},
			// Recursive type declarations are used from their nested type scope in
			// core ESLint; function/class runtime self-references remain different.
			{Code: `interface RecursiveInterface { next: RecursiveInterface }`},
			{Code: `type RecursiveAlias = { next: RecursiveAlias };`},
			{Code: `declare namespace Rspack { interface ExportInfo {} interface ExportsInfo { [key: string]: ExportInfo & ExportsInfo } } consume(Rspack);`},

			// Core no-unused-vars follows the TypeScript scope manager and accepts
			// a value/type declaration consumed from a type position.
			{Code: `class Foo {} let value: Foo; consume(value);`},

			// Capitalized JSX tags are component references.
			{Code: `const Component = () => null; const view = <Component />; consume(view);`, Tsx: true},

			// A write in another variable scope, a later loop iteration, or a
			// storable callback can make a self-read observable.
			{Code: `let x = 0; function update() { x = x + 1; } update();`},
			{Code: `let x = 0; for (let i = 0; i < 2; i++) { x = x + 1; }`},
			{Code: `let x; x = consume({ value: () => x });`},
			{Code: `let x; let stored; x = (stored = { value: () => x }); consume(stored);`},

			// Consuming the assignment result makes the read meaningful; logical
			// assignment is a conditional read rather than a discarded update.
			{Code: `let x = 0; consume(x = x + 1);`},
			{Code: `let x; x ||= 1;`},

			// TypeScript wrappers retain the parser's execution semantics.
			{Code: `const f = ((function () { return f(); }) as () => unknown);`},
			{Code: `let x: any = 0; (x as any) = x + 1;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint core with @typescript-eslint/parser: TS-only scopes ----
			// Unlike the TypeScript extension rule, core reports parameters that
			// exist only in type signatures and abstract declarations.
			{
				Code: signaturesCode,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("callback", false, 1, 12, 28, strings.Replace(signaturesCode, "callback: string", "", 1)),
					extraUnusedErrorWithSuggestion("methodArg", false, 2, 22, 39, strings.Replace(signaturesCode, "methodArg: number", "", 1)),
					extraUnusedErrorWithSuggestion("abstractArg", false, 3, 33, 53, strings.Replace(signaturesCode, "abstractArg: boolean", "", 1)),
				},
			},
			{
				Code: overloadsCode,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("overloadArg", false, 1, 21, 40, strings.Replace(overloadsCode, "overloadArg: string", "", 1)),
					extraUnusedErrorWithSuggestion("implementationArg", false, 2, 21, 46, strings.Replace(overloadsCode, "implementationArg: string", "", 1)),
					extraUnusedErrorWithSuggestion("this", false, 4, 19, 32, strings.Replace(overloadsCode, "this: unknown", "", 1)),
					extraUnusedErrorWithSuggestion("propertyArg", false, 6, 47, 67, strings.Replace(overloadsCode, "propertyArg?: string", "", 1)),
				},
			},
			{
				Code: ambientCode,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("value", false, 1, 29, 42, strings.Replace(ambientCode, "const value: string;", "", 1)),
					extraUnusedErrorWithSuggestion("f", false, 1, 53, 54, strings.Replace(ambientCode, "function f(", "function (", 1)),
					extraUnusedErrorWithSuggestion("arg", false, 1, 55, 66, strings.Replace(ambientCode, "arg: number", "", 1)),
				},
			},
			// Identifier ranges from the TS parser include a non-rest binding's
			// optional/type suffix, including a multiline annotation. Rest and
			// destructured bindings retain the bare identifier range.
			{
				Code: rangesCode,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("definite", false, 1, 5, 22, strings.Replace(rangesCode, "let definite!: string;", "", 1)),
					{
						MessageId: "unusedVar",
						Message:   "'value' is defined but never used.",
						Line:      3,
						Column:    3,
						EndLine:   5,
						EndColumn: 4,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeVar",
								Output: strings.Replace(
									rangesCode,
									"value: {\n    key: string;\n  },",
									"",
									1,
								),
							},
						},
					},
					extraUnusedErrorWithSuggestion("rest", false, 6, 6, 10, strings.Replace(rangesCode, ",\n  ...rest: string[]", "", 1)),
				},
			},
			// A parameter declared inside a function type owns a distinct scope,
			// even when it shadows a used runtime parameter with the same name.
			{
				Code: typeScopeShadowCode,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("context", false, 1, 44, 59, strings.Replace(typeScopeShadowCode, "context: number", "", 1)),
				},
			},
			// Mapped and infer bindings are ordinary scope variables to the core
			// rule: a reference in the produced type consumes K/U, while P/V stay unused.
			{
				Code: mappedAndInferCode,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("P", false, 1, 54, 55, strings.Replace(mappedAndInferCode, "[P in", "[ in", 1)),
					extraUnusedErrorWithSuggestion("V", false, 1, 149, 150, strings.Replace(mappedAndInferCode, "infer V", "infer ", 1)),
				},
			},
			// Enum members have their own bindings. A is consumed by B's
			// initializer; removing B/C follows core's identifier-only suggestion.
			{
				Code: enumCode,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("B", false, 1, 24, 25, strings.Replace(enumCode, ", B", "", 1)),
					extraUnusedErrorWithSuggestion("C", false, 1, 31, 32, strings.Replace(enumCode, ", C", "", 1)),
				},
			},
			// References outside a `declare global` block do not consume the
			// augmentation block's local definitions in ESLint core's scope graph.
			{
				Code: globalAugmentationCode,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("X", false, 1, 24, 33, strings.Replace(globalAugmentationCode, "const X: string;", "", 1)),
					extraUnusedErrorWithSuggestion("I", false, 1, 45, 46, strings.Replace(globalAugmentationCode, "interface I", "interface ", 1)),
				},
			},
			// The checker-resolved target, rather than the exported spelling,
			// decides which shadowed symbol is consumed by a nested export.
			{
				Code: `type A = {}; export namespace N { type A = {}; export type { A as B }; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("A", false, 1, 6, 7, `type  = {}; export namespace N { type A = {}; export type { A as B }; }`),
				},
			},
			// A source-bearing export is a re-export and must not consume a
			// same-named local declaration.
			{
				Code: `type A = {}; export type { A } from "./foo";`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("A", false, 1, 6, 7, `type  = {}; export type { A } from "./foo";`),
				},
			},
			{
				Code:    `function f(a = (() => { const nested = 1; return 0; })()) {} f();`,
				Options: map[string]interface{}{"args": "none"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 31, 37, `function f(a = (() => {  return 0; })()) {} f();`),
				},
			},
			{
				Code:    `try {} catch (error) { const nested = 1; console.log(error); }`,
				Options: map[string]interface{}{"caughtErrors": "none"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 30, 36, `try {} catch (error) {  console.log(error); }`),
				},
			},
			{
				Code:    `const [head, ...tail] = source; console.log(tail);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("head", true, 1, 8, 12, `const [, ...tail] = source; console.log(tail);`),
				},
			},
			{
				Code:    `const { value: [nested], ...rest } = source; console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 17, 23, `const {  ...rest } = source; console.log(rest);`),
				},
			},
			{
				Code:    `const { value: { nested }, ...rest } = source; console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 18, 24, `const {  ...rest } = source; console.log(rest);`),
				},
			},
			{
				Code:    `const { value = 1, ...rest } = source; console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("value", true, 1, 9, 14, `const {  ...rest } = source; console.log(rest);`),
				},
			},
			{
				Code:    `let value; let rest; ({ value = 1, ...rest } = source); console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("value", true, 1, 25, 30, ""),
				},
			},
			{
				Code:    `let nested; let rest; ({ value: [nested], ...rest } = source); console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 34, 40, ""),
				},
			},
			extraUnusedCase(`let x; class C { static { x = 1; } } new C();`, "x", true, 1, 5, 6, ""),
			extraUnusedCase(`let x; namespace N { x = 1; } consume(N);`, "x", true, 1, 5, 6, ""),
			{
				Code: `const f = (function () { return f(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("f", true, 1, 7, 8, ""),
				},
			},
			{
				Code: `const f = ((() => f()));`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("f", true, 1, 7, 8, ""),
				},
			},
			{
				Code:    `function f(a = f) {}`,
				Options: map[string]interface{}{"args": "none"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("f", false, 1, 10, 11, ""),
				},
			},
			{
				Code: `/*global foo*/ function f(foo) { return foo; } f(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo*/ { const foo = 1; consume(foo); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo:writable*/ foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo:writable*/ foo += 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo:writable*/ foo = foo + 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo:writable*/ foo++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `const div = 1; const view = <div />; consume(view);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("div", true, 1, 7, 10, ` const view = <div />; consume(view);`),
				},
			},
			{
				Code: "import React from \"react\";\nconst view = <div />;\nconsume(view);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("React", false, 1, 8, 13, "import \"react\";\nconst view = <div />;\nconsume(view);"),
				},
			},

			// Discarded self-updates stay unused through varied expression trees.
			extraUnusedCase(`let x = []; x = x["concat"](x);`, "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x = []; x = x?.["concat"](x);`, "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x = 0; x = true ? x : 1;`, "x", true, 1, 12, 13, ""),
			extraUnusedCase(`let x = 0; x = [x][0];`, "x", true, 1, 12, 13, ""),
			extraUnusedCase(`let x = 0; x = ({ value: x }).value;`, "x", true, 1, 12, 13, ""),
			extraUnusedCase("let x = ''; x = `${x}`;", "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x; x = new Box(x);`, "x", true, 1, 8, 9, ""),
			extraUnusedCase("let x = ''; x = tag`${x}`;", "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x; x = (() => x)();`, "x", true, 1, 8, 9, ""),
			extraUnusedCase(`let x; x = { value: () => x };`, "x", true, 1, 8, 9, ""),
			extraUnusedCase(`let x; x = [() => x];`, "x", true, 1, 8, 9, ""),
			extraUnusedCase(`let x = 0; (x) += 1;`, "x", true, 1, 13, 14, ""),
			extraUnusedCase(`let x = 0; (x)++;`, "x", true, 1, 13, 14, ""),

			// TypeScript expression wrappers inside the RHS must not hide the read.
			extraUnusedCase(`let x: any = []; x = (x as any)["concat"](x);`, "x", true, 1, 18, 19, ""),
			extraUnusedCase(`let x: any = []; x = x!["concat"](x);`, "x", true, 1, 18, 19, ""),
			extraUnusedCase(`let x: any = []; x = (x satisfies any)["concat"](x);`, "x", true, 1, 18, 19, ""),
		},
	)
}

func TestNoUnusedVarsGapJavaScriptJSDoc(t *testing.T) {
	t.Parallel()

	code := `/** @param {(msg?: string) => void} logFn @returns {(level: string, msg?: string) => void} */ function logGroup(logFn) { return logFn; } consume(logGroup);
/** @template Unused @returns {number} */ function answer() { return 42; } consume(answer);`
	tmpDir := t.TempDir()
	filePath := tspath.NormalizePath(filepath.Join(tmpDir, "file.js"))
	if err := os.WriteFile(filePath, []byte(code), 0o644); err != nil {
		t.Fatalf("write JavaScript fixture: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		AllowJs:      core.TSTrue,
		CheckJs:      core.TSTrue,
		Module:       core.ModuleKindESNext,
		SkipLibCheck: core.TSTrue,
		Target:       core.ScriptTargetESNext,
	}, []string{filePath}, host)
	if err != nil {
		t.Fatalf("create JavaScript program: %v", err)
	}

	ruleRan := false
	var diagnostics []rule.RuleDiagnostic
	linter.RunLinterInProgram(
		program,
		nil,
		nil,
		utils.ExcludePaths,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			if sourceFile.FileName() != filePath {
				return nil
			}
			return []linter.ConfiguredRule{{
				Name:                NoUnusedVarsRule.Name,
				RequiresBindingInfo: NoUnusedVarsRule.RequiresBindingInfo,
				Severity:            rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					ruleRan = true
					if ctx.TypeChecker != nil || ctx.BindingChecker == nil {
						t.Error("gap JavaScript should receive only the binding checker")
					}
					return NoUnusedVarsRule.Run(ctx, nil)
				},
			}}
		},
		false,
		func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
		map[string]struct{}{filepath.Join(tmpDir, "project-only.ts"): {}},
		nil,
	)

	if !ruleRan {
		t.Fatal("core no-unused-vars did not run for gap JavaScript")
	}
	if len(diagnostics) != 0 {
		t.Fatalf("JSDoc function-type names must not become variables: %v", diagnostics)
	}
}
