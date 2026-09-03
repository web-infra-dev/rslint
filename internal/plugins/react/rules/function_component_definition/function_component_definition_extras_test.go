package function_component_definition

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestFunctionComponentDefinitionExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
func TestFunctionComponentDefinitionExtras(t *testing.T) {
	named := func(value interface{}) []interface{} {
		return []interface{}{map[string]interface{}{"namedComponents": value}}
	}
	unnamed := func(value interface{}) []interface{} {
		return []interface{}{map[string]interface{}{"unnamedComponents": value}}
	}

	// Dimension 4 rows that do not apply to this rule:
	// N/A: optional chain / member-access receivers — the rule inspects
	// function nodes and their declarators, never a property access.
	// N/A: literal kinds — no literal value is ever compared.
	// N/A: computed member keys as an equivalence class — an object property of
	// any key shape hits the same `Property` early return.
	// N/A: class bodies — methods and accessors are MethodDeclaration /
	// GetAccessor / SetAccessor in tsgo and are never collected, matching
	// upstream's `Property` early return for the object-literal forms.

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &FunctionComponentDefinitionRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: TS type-expression wrapper on the initializer ----
		// TSESTree keeps `as` / `satisfies` / `!` as real nodes, so the function
		// is not in one of the positions component detection accepts and is not
		// a component at all.
		{
			Code:    `var Hello = ((props) => { return <div/>; }) as React.FC;`,
			Tsx:     true,
			Options: named("function-expression"),
		},

		// A function that returns no JSX is only a component upstream because a
		// later `Hello.propTypes = …` assignment registers it; rslint has no
		// equivalent of that late registration, so nothing is reported. See the
		// rule doc's "Differences from ESLint".
		{
			Code:    "var Hello = function(props) { return 1; }\nHello.propTypes = {};",
			Tsx:     true,
			Options: named("arrow-function"),
		},

		// ---- Dimension 4: body-absent function forms ----
		// An overload signature and an ambient declaration have no body, so
		// they can never return JSX and are never components.
		{
			Code:    "declare function Hello(props: Test): JSX.Element;",
			Tsx:     true,
			Options: named("arrow-function"),
		},
		// ---- Dimension 4: empty function body ----
		{
			Code:    `function Hello(props) {}`,
			Tsx:     true,
			Options: named("arrow-function"),
		},

		// Locks in upstream Components.detect FunctionDeclaration listener arm 1:
		// `node.async && node.generator` registers confidence 0, which bans the
		// node from `components.get` no matter what it returns.
		{
			Code:    `async function* Hello(props) { return <div/>; }`,
			Tsx:     true,
			Options: named("arrow-function"),
		},
		{
			Code:    `var Hello = async function*(props) { return <div/>; };`,
			Tsx:     true,
			Options: named("arrow-function"),
		},

		// Locks in upstream getStatelessComponent's pragma-wrapper arm: the
		// component is registered against the WRAPPER call, so
		// `components.get(fn)` is null and the inner function is never checked.
		{
			Code:    `const Foo = React.memo((props) => { return <div/>; });`,
			Tsx:     true,
			Options: unnamed("function-expression"),
		},
		{
			Code:    `const Foo = React.forwardRef(function(props, ref) { return <div/>; });`,
			Tsx:     true,
			Options: unnamed("arrow-function"),
		},
		// Nested wrappers redirect to the OUTER-MOST call, so neither the inner
		// function nor an intermediate wrapper is validated.
		{
			Code:    `const Foo = React.memo(React.forwardRef((props, ref) => { return <div/>; }));`,
			Tsx:     true,
			Options: unnamed("function-expression"),
		},

		// ---- Dimension 4: computed property key ----
		// Upstream's `node.parent.type === 'Property'` early return covers
		// computed keys too.
		{
			Code:    `const key = 'a'; const obj = { [key]: (props) => { return <div/>; } };`,
			Tsx:     true,
			Options: named("function-declaration"),
		},

		// Locks in the default options: named components default to
		// `function-declaration` and unnamed ones to `function-expression`.
		{
			Code: `function Hello(props) { return <div/>; }`,
			Tsx:  true,
		},
		{
			Code:    `function wrap() { return function(props) { return <div/>; }; }`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{}},
		},
	}, []rule_tester.InvalidTestCase{
		// Schema defaults: omitting the options entirely and passing an empty
		// object must produce byte-identical output.
		{
			Code:   `var Hello = (props) => { return <div/>; };`,
			Tsx:    true,
			Output: []string{`function Hello(props) { return <div/>; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 1, Column: 13, EndLine: 1, EndColumn: 42}},
		},
		{
			Code:    `var Hello = (props) => { return <div/>; };`,
			Tsx:     true,
			Output:  []string{`function Hello(props) { return <div/>; }`},
			Options: []interface{}{map[string]interface{}{}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-declaration", Line: 1, Column: 13, EndLine: 1, EndColumn: 42}},
		},

		// ---- Dimension 4: parenthesized initializer ----
		// ESTree has no ParenthesizedExpression, so upstream sees the variable
		// declarator as the arrow's direct parent and treats the component as
		// named. The parentheses are inside the replaced declaration, so the
		// fix drops them.
		{
			Code:    `var Hello = ((props) => { return <div/>; });`,
			Tsx:     true,
			Output:  []string{`var Hello = function(props) { return <div/>; }`},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 1, Column: 14, EndLine: 1, EndColumn: 43}},
		},

		// The same shape returning JSX is a component on its own merit, so the
		// `propTypes` assignment makes no difference here.
		{
			Code:    "function Hello(props) { return <div/>; }\nHello.propTypes = {};",
			Tsx:     true,
			Output:  []string{"const Hello = (props) => { return <div/>; }\nHello.propTypes = {};"},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 1, Column: 1, EndLine: 1, EndColumn: 41}},
		},

		// ---- Dimension 4: empty parameter list ----
		// Locks in upstream getParams()'s `params.length === 0 → null` arm,
		// which `buildFunction` turns into the empty string.
		{
			Code:    `function Hello() { return <div/>; }`,
			Tsx:     true,
			Output:  []string{`const Hello = () => { return <div/>; }`},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 1, Column: 1, EndLine: 1, EndColumn: 36}},
		},
		// ---- Dimension 4: arrow with an expression body ----
		// Locks in upstream getBody()'s non-BlockStatement arm, which wraps the
		// expression in a synthesized block with fixed two-space indentation.
		{
			Code:    `var Hello = (props) => <div/>;`,
			Tsx:     true,
			Output:  []string{"var Hello = function(props) {\n  return <div/>\n}"},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 1, Column: 13, EndLine: 1, EndColumn: 30}},
		},
		// ---- Dimension 4: parameter with a leading comment ----
		// ESTree parameter ranges exclude leading comments, and so does the
		// trimmed tsgo range, so the comment is dropped from the rewritten
		// parameter list either way.
		{
			Code:    `function Hello(/* props */ props) { return <div/>; }`,
			Tsx:     true,
			Output:  []string{`const Hello = (props) => { return <div/>; }`},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 1, Column: 1, EndLine: 1, EndColumn: 53}},
		},
		// ---- Dimension 4: trailing comma in the parameter list ----
		// The copied span ends at the last parameter, so the trailing comma is
		// dropped.
		{
			Code:    `function Hello(props,) { return <div/>; }`,
			Tsx:     true,
			Output:  []string{`const Hello = (props) => { return <div/>; }`},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 1, Column: 1, EndLine: 1, EndColumn: 42}},
		},
		// ---- Dimension 4: default and rest parameters ----
		{
			Code:    `function Hello(props = {}, ...rest) { return <div/>; }`,
			Tsx:     true,
			Output:  []string{`const Hello = (props = {}, ...rest) => { return <div/>; }`},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 1, Column: 1, EndLine: 1, EndColumn: 55}},
		},

		// ---- Dimension 4: async function component ----
		// The rewrite templates have no slot for `async`, so upstream's fix
		// drops the keyword. Mirrored verbatim rather than diverging.
		{
			Code:    `async function Hello(props) { return <div/>; }`,
			Tsx:     true,
			Output:  []string{`const Hello = (props) => { return <div/>; }`},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 1, Column: 1, EndLine: 1, EndColumn: 47}},
		},

		// ---- Dimension 4: JSX fragment return ----
		// A fragment makes the function a component, but ESTree models it as
		// JSXFragment, which upstream's `hasES6OrJsx` selector does not list —
		// so the synthesized variable is still `var`, not `const`.
		{
			Code:    `function Hello() { return <></>; }`,
			Tsx:     true,
			Output:  []string{`var Hello = function() { return <></>; }`},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 1, Column: 1, EndLine: 1, EndColumn: 35}},
		},

		// Locks in upstream's `varType = node.parent.parent.kind` override:
		// the declarator's own keyword wins over the file-level default.
		{
			Code:    `const Hello = (props) => { return <div/>; };`,
			Tsx:     true,
			Output:  []string{`const Hello = function(props) { return <div/>; }`},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 1, Column: 15, EndLine: 1, EndColumn: 44}},
		},

		// Locks in each `hasES6OrJsx` arm that raises fileVarType to `const`
		// without any JSX or `const`/`let` declaration in the file.
		{
			// `export * from` — ESTree ExportAllDeclaration.
			Code:    "export * from './x';\nfunction Hello(props) { return React.createElement('div'); }",
			Tsx:     true,
			Output:  []string{"export * from './x';\nconst Hello = function(props) { return React.createElement('div'); }"},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 1, EndLine: 2, EndColumn: 61}},
		},
		{
			// `export { x }` — ESTree ExportNamedDeclaration / ExportSpecifier.
			Code:    "var x = 1;\nexport { x };\nfunction Hello(props) { return React.createElement('div'); }",
			Tsx:     true,
			Output:  []string{"var x = 1;\nexport { x };\nconst Hello = function(props) { return React.createElement('div'); }"},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 3, Column: 1, EndLine: 3, EndColumn: 61}},
		},
		{
			// `import x = require(...)` — ESTree TSImportEqualsDeclaration.
			Code:    "import x = require('./x');\nfunction Hello(props) { return React.createElement('div'); }",
			Tsx:     true,
			Output:  []string{"import x = require('./x');\nconst Hello = function(props) { return React.createElement('div'); }"},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 1, EndLine: 2, EndColumn: 61}},
		},

		// ---- Real-user: #3207, "do not break on dollar signs" ----
		// Every `$`-prefixed replacement pattern JavaScript's String#replace
		// understands must survive the template substitution untouched.
		{
			Code:    "function Hello(props) {\n  return <div>{'$&' + \"$`\" + '$1' + \"$$\"}</div>;\n}",
			Tsx:     true,
			Output:  []string{"const Hello = (props) => {\n  return <div>{'$&' + \"$`\" + '$1' + \"$$\"}</div>;\n}"},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 1, Column: 1, EndLine: 3, EndColumn: 2}},
		},
		// ---- Real-user: #3248, "replace var by const in certain situations" ----
		// A `let` anywhere in the file is enough to make the synthesized
		// declaration `const`, even when the component itself returns no JSX.
		{
			Code:    "let counter = 0;\nfunction Hello(props) { return React.createElement('div'); }",
			Tsx:     true,
			Output:  []string{"let counter = 0;\nconst Hello = function(props) { return React.createElement('div'); }"},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 2, Column: 1, EndLine: 2, EndColumn: 61}},
		},

		// Locks in upstream hasOneUnconstrainedTypeParam()'s `length === 1` arm:
		// two type parameters are unambiguous, so the arrow fix is offered.
		{
			Code:    `function Hello<T1, T2>(props: Props<T1, T2>) { return <div/>; }`,
			Tsx:     true,
			Output:  []string{`const Hello = <T1, T2>(props: Props<T1, T2>) => { return <div/>; }`},
			Options: named("arrow-function"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "arrow-function", Line: 1, Column: 1, EndLine: 1, EndColumn: 64}},
		},

		// ---- Dimension 4: multi-declarator variable statement ----
		// Upstream replaces `node.parent.parent.range` — the WHOLE declaration,
		// including sibling declarators. Mirrored verbatim.
		{
			Code:    `var a = 1, Hello = (props) => { return <div/>; };`,
			Tsx:     true,
			Output:  []string{`var Hello = function(props) { return <div/>; }`},
			Options: named("function-expression"),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "function-expression", Line: 1, Column: 20, EndLine: 1, EndColumn: 49}},
		},

		// Locks in the report order for a file with both a named and an unnamed
		// offender: reports follow source order, matching upstream's single
		// `Program:exit` pass over the collected pairs.
		{
			Code:    "var Hello = (props) => { return <div/>; };\nfunction wrap() { return function(props) { return <span/>; }; }",
			Tsx:     true,
			Output:  []string{"var Hello = function(props) { return <div/>; }\nfunction wrap() { return (props) => { return <span/>; }; }"},
			Options: []interface{}{map[string]interface{}{"namedComponents": "function-expression", "unnamedComponents": "arrow-function"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "function-expression", Line: 1, Column: 13, EndLine: 1, EndColumn: 42},
				{MessageId: "arrow-function", Line: 2, Column: 26, EndLine: 2, EndColumn: 61},
			},
		},
	})
}

// TestFunctionComponentDefinitionEditDemand exercises Dimension 3 (autofix
// boundaries): diagnostic count, message, and range must stay identical across
// every edit demand, and the replacement text must materialize only when an
// autofix is actually requested.
func TestFunctionComponentDefinitionEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"var Hello = (props) => { return <div/>; };\nvar World = (props) => { return <span/>; };",
		"edit-demand.tsx",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     FunctionComponentDefinitionRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return FunctionComponentDefinitionRule.Run(ctx, []any{
							map[string]any{"namedComponents": "function-expression"},
						})
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
		if fixes := *allEdits[index].FixesPtr; len(fixes) == 0 {
			t.Fatalf("diagnostic %d: all-edits demand produced no fixes", index)
		}
		if diagnostic := allEdits[index]; diagnostic.Suggestions != nil {
			t.Fatalf("diagnostic %d: autofix-only rule materialized suggestions", index)
		}
	}
}
