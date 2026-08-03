package prefer_node_protocol_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_node_protocol"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestPreferNodeProtocolExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
func TestPreferNodeProtocolExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_node_protocol.PreferNodeProtocolRule,
		[]rule_tester.ValidTestCase{
			// ---- Regression: tsgo-reparsed JSDoc imports are invisible upstream ----
			{
				Code:     "/** @type {import(\"fs\").Stats} */\nconst stats = {};",
				FileName: "jsdoc-import-type.js",
			},
			{
				Code:     "/** @param {import(\"fs\").Stats} stats */\nfunction read(stats) {}",
				FileName: "jsdoc-param-import-type.js",
			},
			{
				Code:     "/** @import {Stats} from \"fs\" */\nconst stats = {};",
				FileName: "jsdoc-import-tag.js",
			},
			{
				Code:     "/** @typedef {import(\"fs\").Stats} Stats */\nconst stats = {};",
				FileName: "jsdoc-typedef-import-type.js",
			},
			{
				Code:     "/** @type {typeof import(\"fs\")} */\nconst fsModule = {};",
				FileName: "jsdoc-typeof-import-type.js",
			},
			{
				Code:     "/** @import * as fs from \"fs\" */\nconst stats = {};",
				FileName: "jsdoc-namespace-import-tag.js",
			},

			// ---- Branch lock-in: upstream listens to ExportNamedDeclaration only ----
			{Code: `export * from "fs";`},
			{Code: `export * as ns from "fs";`},
			{Code: `export type * from "fs";`, FileName: "export-all.ts"},
			{Code: `import fs = require("fs");`, FileName: "import-equals.ts"},

			// ---- Dimension 4: optional chains stay excluded through parentheses ----
			{Code: `const fs = ((require))?.("fs");`},
			{Code: `const fs = (process?.getBuiltinModule)("fs");`},

			// ---- Dimension 4: template-literal specifiers are not string literals ----
			// tsgo splits ESTree's `Literal` into StringLiteral vs
			// NoSubstitutionTemplateLiteral; the rule matches only the former.
			{Code: "const fs = require(`fs`);"},
			{Code: "async function foo() {\n\treturn import(`fs`);\n}"},

			// ---- Dimension 4: optional-chain require / getBuiltinModule are excluded ----
			{Code: `const fs = require?.("fs");`},
			{Code: `const fs = process.getBuiltinModule?.("fs");`},
			{Code: `const fs = process?.getBuiltinModule("fs");`},

			// ---- Dimension 4: computed / non-`process` receiver for getBuiltinModule ----
			{Code: `const fs = process["getBuiltinModule"]("fs");`},
			{Code: `const fs = notProcess.getBuiltinModule("fs");`},

			// ---- Dimension 4: already-prefixed and non-builtin specifiers ----
			{Code: `import fs from "node:fs/promises";`},
			{Code: `const fs = require("node:fs");`},
			{Code: `type fs = import("node:fs");`},
			{Code: `type fs = import("./local");`},

			// ---- Asymmetric builtin-modules list entries (node:-only, no bare form) ----
			// These names exist only under the `node:` prefix in builtin-modules,
			// so the precomputed dual-membership set keeps the bare form unflagged.
			{Code: `import "sea";`},
			{Code: `import "sqlite";`},
			{Code: `import "quic";`},
			{Code: `import "test";`},

			// ---- Dimension 4: dynamic import with non-literal / no argument ----
			{Code: "async function foo() {\n\treturn import(name);\n}"},
			{Code: "async function foo() {\n\treturn import();\n}"},

			// ---- Real-user: type-position import in a generic argument stays untouched when non-builtin ----
			{Code: `type T = import("./types").Foo<import("./other").Bar>;`},

			// ---- Real-user: export * / export named from a non-builtin path ----
			{Code: `export * from "./barrel";`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Regression: a real import beside JSDoc still reports exactly once ----
			{
				Code:     "/** @type {import(\"fs\").Stats} */\nimport fs from \"fs\";",
				FileName: "jsdoc-and-runtime-import.js",
				Output:   []string{"/** @type {import(\"fs\").Stats} */\nimport fs from \"node:fs\";"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 2, Column: 16}},
			},

			// ---- Branch lock-in: a named type export is still ExportNamedDeclaration upstream ----
			{
				Code:     `export type {Stats} from "fs";`,
				FileName: "named-type-export.ts",
				Output:   []string{`export type {Stats} from "node:fs";`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 26}},
			},

			// ---- Branch lock-in: parenthesized `process` receiver for getBuiltinModule ----
			// ESTree flattens parentheses, so `(process).getBuiltinModule("fs")`
			// presents a bare `process` identifier object to upstream. tsgo keeps
			// the ParenthesizedExpression, so the rule applies SkipParentheses to
			// the object to stay 1:1.
			{
				Code:   `const fs = (process).getBuiltinModule("buffer")`,
				Output: []string{`const fs = (process).getBuiltinModule("node:buffer")`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 39}},
			},

			// ---- Branch lock-in: parenthesized string argument to require ----
			// ESTree flattens the paren, so `require(("fs"))` is still a static
			// string-literal argument.
			{
				Code:   `const fs = require(("fs"))`,
				Output: []string{`const fs = require(("node:fs"))`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 21}},
			},

			// ---- Branch lock-in: parenthesized require callee ----
			// ESTree flattens parentheses, so `(require)("fs")` presents a bare
			// `require` identifier callee to upstream. tsgo keeps the
			// ParenthesizedExpression, so isStaticRequire applies SkipParentheses
			// to the callee to stay 1:1.
			{
				Code:   `const fs = (require)("fs")`,
				Output: []string{`const fs = (require)("node:fs")`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 22}},
			},
			{
				Code:   `const fs = ((require))("fs")`,
				Output: []string{`const fs = ((require))("node:fs")`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 24}},
			},
			{
				Code:   `const fs = ((require))((("fs")));`,
				Output: []string{`const fs = ((require))((("node:fs")));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 26}},
			},

			// ---- Branch lock-in: direct (unparenthesized) process.getBuiltinModule ----
			{
				Code:   `const fs = process.getBuiltinModule("buffer")`,
				Output: []string{`const fs = process.getBuiltinModule("node:buffer")`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 37}},
			},
			{
				Code:   `const fs = (process.getBuiltinModule)("fs");`,
				Output: []string{`const fs = (process.getBuiltinModule)("node:fs");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 39}},
			},

			// ---- Dimension 3: dynamic import options and static import attributes preserve fixes ----
			{
				Code:   `const fs = import("fs", {with: {type: "json"}});`,
				Output: []string{`const fs = import("node:fs", {with: {type: "json"}});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 19}},
			},
			{
				Code:   `import fs from "fs" with {type: "json"};`,
				Output: []string{`import fs from "node:fs" with {type: "json"};`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 16}},
			},

			// ---- Dimension 4: single-quote fix preserves the original quote ----
			{
				Code:   `import fs from 'fs';`,
				Output: []string{`import fs from 'node:fs';`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 16}},
			},

			// ---- Locks in: normal JavaScript imports remain ImportDeclaration ----
			{
				Code:     `import fs from "fs";`,
				FileName: "js-import.js",
				Output:   []string{`import fs from "node:fs";`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 16}},
			},

			// ---- Real-user: subpath builtin (`fs/promises`) via require ----
			{
				Code:   `const {readFile} = require("fs/promises");`,
				Output: []string{`const {readFile} = require("node:fs/promises");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 1, Column: 28}},
			},

			// ---- Real-user: multi-line import statement position ----
			{
				Code:   "import {\n\treadFile,\n} from \"fs/promises\";",
				Output: []string{"import {\n\treadFile,\n} from \"node:fs/promises\";"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-node-protocol", Line: 3, Column: 8}},
			},

			// ---- Locks in the message text (data interpolation of moduleName) ----
			{
				Code:   `import "child_process";`,
				Output: []string{`import "node:child_process";`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "prefer-node-protocol",
					Message:   "Prefer `node:child_process` over `child_process`.",
					Line:      1, Column: 8,
				}},
			},
		},
	)
}

// TestPreferNodeProtocolEditDemand verifies the deferred fix builder runs only
// under a matching edit demand and never changes the diagnostic identity. The
// rule emits an autofix and no suggestions.
func TestPreferNodeProtocolEditDemand(t *testing.T) {
	t.Parallel()

	const fileName = "edit-demand.ts"
	const code = `import fs from "fs";`
	const fixOutput = `import fs from "node:fs";`

	program, sourceFile := createPreferNodeProtocolProgram(t, fileName, code)
	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		got := lintPreferNodeProtocolWithDemand(program, sourceFile, demand)
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		diagnostics[demand] = got[0]
	}

	// Diagnostic identity (message + range) must be stable across demands.
	base := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		requireSamePreferNodeProtocolDiagnostic(t, base, diagnostic, demand)
	}

	// None and suggestion-only demands materialize no autofixes.
	if diagnostics[rule.EditDemandNone].FixesPtr != nil {
		t.Errorf("none demand unexpectedly materialized fixes")
	}
	if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
		t.Errorf("suggestion-only demand unexpectedly materialized fixes")
	}
	// The rule never emits suggestions.
	for demand, diagnostic := range diagnostics {
		if diagnostic.Suggestions != nil {
			t.Errorf("demand %d unexpectedly materialized suggestions", demand)
		}
	}

	autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
	allFixes := diagnostics[rule.EditDemandAll].FixesPtr
	if autofixOnly == nil || allFixes == nil || !reflect.DeepEqual(*autofixOnly, *allFixes) {
		t.Fatalf("autofix artifacts differ between autofix-only and all demand")
	}
	if applyPreferNodeProtocolFixes(code, *autofixOnly) != fixOutput {
		t.Fatalf("autofix output = %q, want %q", applyPreferNodeProtocolFixes(code, *autofixOnly), fixOutput)
	}
}

func lintPreferNodeProtocolWithDemand(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     program,
		File:        sourceFile.FileName(),
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     prefer_node_protocol.PreferNodeProtocolRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return prefer_node_protocol.PreferNodeProtocolRule.Run(ctx, nil)
				},
			}}
		},
		ExcludePaths: []string{},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}

func requireSamePreferNodeProtocolDiagnostic(t *testing.T, want, got rule.RuleDiagnostic, demand rule.EditDemand) {
	t.Helper()
	want.FixesPtr = nil
	want.Suggestions = nil
	got.FixesPtr = nil
	got.Suggestions = nil
	if !reflect.DeepEqual(got, want) {
		t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
	}
}

func createPreferNodeProtocolProgram(t testing.TB, fileName, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

// applyPreferNodeProtocolFixes applies non-overlapping RuleFix edits to code, in
// descending range order so earlier offsets stay valid.
func applyPreferNodeProtocolFixes(code string, fixes []rule.RuleFix) string {
	sorted := make([]rule.RuleFix, len(fixes))
	copy(sorted, fixes)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j-1].Range.Pos() < sorted[j].Range.Pos(); j-- {
			sorted[j-1], sorted[j] = sorted[j], sorted[j-1]
		}
	}
	result := code
	for _, fix := range sorted {
		result = result[:fix.Range.Pos()] + fix.Text + result[fix.Range.End():]
	}
	return result
}
