// Dimension walk notes for prefer-object-has-own:
//   - Dimension 4 (declaration / container forms): N/A — the rule targets a
//     call expression, never a function or class declaration, so the form the
//     surrounding declaration takes cannot affect detection.
//   - Dimension 4 (graceful degradation: overload signatures / abstract /
//     declare members): N/A — a body-absent member holds no call expression
//     for the rule to see.
package prefer_object_has_own

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferObjectHasOwnExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the specific
// branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently
// regress them without breaking a named lock-in.
func TestPreferObjectHasOwnExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferObjectHasOwnRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TypeScript expression wrappers around the receiver ----
			// A wrapper is an expression of its own, so the receiver it wraps is no
			// longer `Object` / `Object.prototype` / an object literal.
			{Code: `Object!.prototype.hasOwnProperty.call(a, b);`},
			{Code: `(Object as any).prototype.hasOwnProperty.call(a, b);`},
			{Code: `(Object satisfies unknown).prototype.hasOwnProperty.call(a, b);`},
			{Code: `({} as Record<string, unknown>).hasOwnProperty.call(a, b);`},
			{Code: `Object.prototype!.hasOwnProperty.call(a, b);`},
			{Code: `Object.prototype.hasOwnProperty!.call(a, b);`},
			{Code: `(Object.prototype.hasOwnProperty as any).call(a, b);`},

			// ---- Dimension 4: computed keys that name nothing statically ----
			{Code: "Object[`${'prototype'}`].hasOwnProperty.call(a, b);"},
			{Code: `Object[('prototype' as any)].hasOwnProperty.call(a, b);`},
			{Code: `Object.prototype[('hasOwnProperty' as const)].call(a, b);`},
			{Code: `Object.prototype.hasOwnProperty[('call' as any)](a, b);`},

			// ---- Dimension 4: numeric keys are a separate equivalence class ----
			{Code: `Object[0].hasOwnProperty.call(a, b);`},
			{Code: `Object.prototype[0].call(a, b);`},
			{Code: `Object.prototype.hasOwnProperty[0](a, b);`},

			// ---- Dimension 4: object literals that are not empty ----
			{Code: `({ ...spread }).hasOwnProperty.call(a, b);`},
			{Code: `({ foo: 1 }).hasOwnProperty.call(a, b);`},
			{Code: `({ [key]: 1 }).hasOwnProperty.call(a, b);`},
			{Code: `({ method() {} }).hasOwnProperty.call(a, b);`},

			// ---- Dimension 4: the callee must be a call, not a construction or a tag ----
			{Code: `new ({}.hasOwnProperty.call)(a, b);`},
			{Code: `new Object.prototype.hasOwnProperty.call(a, b);`},

			// ---- Dimension 4: nesting boundary — an inner scope shadows only itself ----
			{Code: `function f() { function g() { return {}.hasOwnProperty.call(a, b); } const Object = 1; return [g, Object]; }`},
			{Code: `{ const Object = 1; ({}).hasOwnProperty.call(a, b); }`},
			{Code: `for (const Object of list) { ({}).hasOwnProperty.call(a, b); }`},
			{Code: `try { foo(); } catch (Object) { ({}).hasOwnProperty.call(a, b); }`},
			{Code: `class C { m(Object: unknown) { return {}.hasOwnProperty.call(a, b); } }`},

			// ---- Dimension 4: value declarations of `Object` in every form ----
			{Code: `const Object = globalThis.Object;
({}).hasOwnProperty.call(a, b);`},
			{Code: `class Object {}
({}).hasOwnProperty.call(a, b);`},
			{Code: `declare const Object: any;
({}).hasOwnProperty.call(a, b);`},
			{Code: `import Object from "mod";
({}).hasOwnProperty.call(a, b);`},
			{Code: `enum Object { A }
({}).hasOwnProperty.call(a, b);`},
			{Code: `function Object() {}
({}).hasOwnProperty.call(a, b);`},

			// ---- Dimension 4: `Object` un-declared through config globals ----
			{
				Code:    `({}).hasOwnProperty.call(a, b);`,
				Globals: map[string]any{"Object": "off"},
			},

			// Locks in upstream CallExpression() arm 1: the callee is not a member access.
			{Code: `call(a, b);`},
			{Code: `(function () {})(a, b);`},

			// Locks in upstream CallExpression() arm 2: the callee's own receiver is not a
			// member access, so there is no `Object` / `Object.prototype` in front of it.
			{Code: `Object.call(a, b);`},
			{Code: `hasOwnProperty.call(a, b);`},

			// Locks in upstream hasLeftHandObject() arm 2 false-branch: a receiver that is
			// a member access but does not name `prototype` stays as it is, and a member
			// access is never the `Object` identifier.
			{Code: `Object.foo.hasOwnProperty.call(a, b);`},
			{Code: `Object['proto'].hasOwnProperty.call(a, b);`},

			// Locks in upstream hasLeftHandObject() arm 3: an identifier receiver that is
			// not named `Object`.
			{Code: `Reflect.prototype.hasOwnProperty.call(a, b);`},
			{Code: `Object2.hasOwnProperty.call(a, b);`},

			// ---- Real-user: rule-change request to also report `obj.hasOwnProperty(key)`,
			// declined — only the `Object.prototype.hasOwnProperty.call()` shape is reported ----
			{Code: `obj.hasOwnProperty('a');`},
			{Code: `Object.prototype.hasOwnProperty(obj, 'a');`},

			// ---- Real-user: rule-change requests to also report `Object.keys().includes()`
			// and the `in` operator, both declined ----
			{Code: `Object.keys(obj).includes(key);`},
			{Code: `'a' in obj;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: optional chain — tsgo flags the access instead of wrapping
			// it, and every link of the chain still names the same call ----
			{
				Code:   `Object?.hasOwnProperty.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:   `Object?.prototype.hasOwnProperty.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:   `Object.prototype?.hasOwnProperty.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:   `Object.prototype.hasOwnProperty?.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:   `Object.prototype.hasOwnProperty.call?.(a, b);`,
				Output: []string{`Object.hasOwn?.(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:   `({})?.hasOwnProperty.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},

			// ---- Dimension 4: multi-level parenthesized receiver ----
			{
				Code:   `(((Object))).prototype.hasOwnProperty.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code:   `((({}))).hasOwnProperty.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			// A parenthesized computed key still names the property it spells.
			{
				Code:   `Object[('prototype')].hasOwnProperty.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
				},
			},

			// ---- Dimension 4: dotted and computed keys mix freely ----
			{
				Code:   `Object.prototype['hasOwnProperty'].call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code:   `Object['prototype'].hasOwnProperty['call'](a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 49},
				},
			},

			// ---- Dimension 4: TypeScript call syntax around a reported call ----
			{
				Code:   `Object.prototype.hasOwnProperty.call<any>(a, b);`,
				Output: []string{`Object.hasOwn<any>(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
				},
			},

			// ---- Dimension 4: graceful degradation on the argument list ----
			{
				Code:   `Object.prototype.hasOwnProperty.call();`,
				Output: []string{`Object.hasOwn();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:   `Object.prototype.hasOwnProperty.call(...args);`,
				Output: []string{`Object.hasOwn(...args);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
				},
			},

			// ---- Dimension 4: nesting boundary — an outer and an inner call both report ----
			{
				Code:   `({}).hasOwnProperty.call({}.hasOwnProperty.call(a, b), c);`,
				Output: []string{`Object.hasOwn(Object.hasOwn(a, b), c);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 58},
					{MessageId: "useHasOwn", Line: 1, Column: 26, EndLine: 1, EndColumn: 54},
				},
			},

			// ---- Dimension 4: a block-scoped `Object` does not reach the outer call ----
			{
				Code:   `{ let Object; } ({}).hasOwnProperty.call(a, b);`,
				Output: []string{`{ let Object; } Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 17, EndLine: 1, EndColumn: 47},
				},
			},

			// Locks in the message text for the rule's only messageId.
			{
				Code:   `Object.prototype.hasOwnProperty.call(a, b);`,
				Output: []string{`Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useHasOwn",
						Message:   "Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 43,
					},
				},
			},

			// Locks in the report range on a call spread over several lines.
			{
				Code: `const has = Object.prototype.hasOwnProperty.call(
	object,
	property,
);`,
				Output: []string{`const has = Object.hasOwn(
	object,
	property,
);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 13, EndLine: 4, EndColumn: 2},
				},
			},

			// Locks in upstream fix() arm 1: a comment inside the callee suppresses the fix
			// but not the diagnostic. A line comment ends the line the callee continues on.
			{
				Code: `Object.prototype.hasOwnProperty // why
	.call(a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 2, EndColumn: 13},
				},
			},
			{
				Code: `({} /* empty */).hasOwnProperty.call(a, b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
				},
			},
			// A comment outside the callee leaves the fix alone.
			{
				Code:   `Object.prototype.hasOwnProperty.call(/* key */ a, b);`,
				Output: []string{`Object.hasOwn(/* key */ a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 1, EndLine: 1, EndColumn: 53},
				},
			},

			// Locks in upstream fix() arm 2: the replacement starts an identifier, so it
			// takes a space when the token in front of it ends one.
			{
				Code:   `const has = typeof{}.hasOwnProperty.call(a, b);`,
				Output: []string{`const has = typeof Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 19, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code:   `const has = void{}.hasOwnProperty.call(a, b);`,
				Output: []string{`const has = void Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 17, EndLine: 1, EndColumn: 45},
				},
			},
			// Locks in upstream fix() arm 3: a punctuator in front of the callee needs no
			// space, and neither does a block comment.
			{
				Code:   `const has = 1+{}.hasOwnProperty.call(a, b);`,
				Output: []string{`const has = 1+Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 15, EndLine: 1, EndColumn: 43},
				},
			},

			// ---- Real-user: the shape the rule was reported not to fix cleanly —
			// a returned call with no space in front of it ----
			{
				Code:   `const f = () => {return{}.hasOwnProperty.call(object, property)};`,
				Output: []string{`const f = () => {return Object.hasOwn(object, property)};`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 24, EndLine: 1, EndColumn: 64},
				},
			},

			// ---- Dimension 4: TypeScript type-only declarations do not shadow a value ----
			// ESLint's scope manager gives `type Object` a variable of its own and stays
			// silent here; rslint reads only value declarations, so the call still reports.
			{
				Code: `type Object = string;
({}).hasOwnProperty.call(a, b);`,
				Output: []string{`type Object = string;
Object.hasOwn(a, b);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 2, Column: 1, EndLine: 2, EndColumn: 31},
				},
			},
			{
				Code:   `function f<Object>() { return {}.hasOwnProperty.call(a, b); }`,
				Output: []string{`function f<Object>() { return Object.hasOwn(a, b); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHasOwn", Line: 1, Column: 31, EndLine: 1, EndColumn: 59},
				},
			},
		},
	)
}

// TestPreferObjectHasOwnEditDemand exercises Dimension 3 (autofix boundaries):
// diagnostic count, message, and range must stay identical across every edit
// demand, and the fix must materialize only when autofix is requested.
func TestPreferObjectHasOwnEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"const has = Object.prototype.hasOwnProperty.call(object, property);\nconst had = ({}).hasOwnProperty.call(object, property);",
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      lintprogram.NewFromCompiler(program),
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     PreferObjectHasOwnRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferObjectHasOwnRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index := range allEdits {
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got, want := withoutEdits(diagnostics[index]), withoutEdits(allEdits[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d changed diagnostic %d:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
		if diagnosticsOnly[index].FixesPtr != nil || suggestionOnly[index].FixesPtr != nil {
			t.Fatalf("diagnostic %d: non-autofix demand materialized fixes", index)
		}
		if autofixOnly[index].FixesPtr == nil ||
			!reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Fatalf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
		}
		if fixes := *allEdits[index].FixesPtr; len(fixes) != 1 || fixes[0].Text != hasOwnReplacement {
			t.Fatalf("diagnostic %d: unexpected autofix %#v", index, fixes)
		}
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{
		diagnosticsOnly,
		autofixOnly,
		suggestionOnly,
		allEdits,
	} {
		for index, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Fatalf("diagnostic %d: autofix-only rule materialized suggestions", index)
			}
		}
	}
}
