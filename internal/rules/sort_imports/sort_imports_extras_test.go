// TestSortImportsExtras locks in branches and edge shapes that the upstream
// test suite does not exercise. Each case names the Dimension 4 row,
// real-user issue, or upstream branch it protects. The migrated upstream suite
// lives in sort_imports_upstream_test.go.
//
// Dimension 4 audit:
//   - N/A receiver/expression wrappers and optional chains: the listener reads
//     ImportDeclaration fields, not expression children.
//   - N/A property/access key forms: imports expose bindings rather than keys.
//   - N/A function/class container variants and nesting boundaries: imports are
//     top-level declarations and cannot nest in those containers.
//   - N/A spread/rest, empty function/class bodies, and body-absent members: no
//     such nodes can occur in an import declaration.
//   - Applicable graceful-degradation forms (empty named lists, import type,
//     type-only specifiers, import attributes) are covered below.
package sort_imports

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

func TestSortImportsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SortImportsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: empty named import list degrades to `none` ----
			{Code: "import {} from 'empty';\nimport value from 'value';"},
			// ---- Dimension 4: whole-clause type import participates normally ----
			{Code: "import type A from 'a';\nimport type B from 'b';"},
			// ---- Dimension 4: type-only named specifiers participate normally ----
			{Code: "import { type A, type B } from 'types';"},
			// ---- Dimension 4: import attributes do not change member syntax ----
			{Code: "import data from 'data.json' with { type: 'json' };\nimport value from 'value';"},
			// ---- Real-user: eslint/eslint#12393 — side-effect imports have no local name, so the rule intentionally does not alphabetize their module specifiers ----
			{Code: "import '../b.js';\nimport 'https://example.com/elements/last.js';\nimport '../a.js';"},
			// Locks in usedMemberSyntax() `multiple`: default plus namespace has two specifiers.
			{Code: "import value, * as ns from 'mixed';\nimport z from 'named';", Options: []any{map[string]any{"memberSyntaxSortOrder": []any{"multiple", "none", "all", "single"}}}},
			// Locks in lowercasing's nullable-name arm: a side-effect import followed by a default import does not compare names.
			{Code: "import 'setup';\nimport Value from 'value';", Options: []any{map[string]any{"ignoreCase": true}}},
			// Locks in UTF-16 code-unit ordering: the astral identifier begins with D801, below E000.
			{Code: "import { 𐐀, 豈 } from 'unicode';"},
			// Locks in member sorting's no-error branch and stable handling of equal lowered names.
			{Code: "import { A, a } from 'equal';", Options: []any{map[string]any{"ignoreCase": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: type-only named specifiers retain their complete source text in the fix ----
			{Code: "import { type B, type A } from 'types';", Output: []string{"import { type A, type B } from 'types';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("A", 1, 18, 24)}},
			// ---- Real-user: eslint/eslint#14259 — one named binding is `single`, so it alphabetically compares with a default import ----
			{Code: "import React from 'react';\nimport { Box } from '@material-ui/core';", Options: []any{map[string]any{"memberSyntaxSortOrder": []any{"single", "multiple", "none", "all"}}}, Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 41)}},
			// ---- Real-user: eslint/eslint#18017 — aliases sort by the local binding, not the exported name ----
			{Code: "import { a, b as myAliasName, c } from 'some-module';", Output: []string{"import { a, c, b as myAliasName } from 'some-module';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("c", 1, 31, 32)}},
			// ---- Real-user: eslint/eslint#17062 — default case-sensitive ordering places uppercase locals first ----
			{Code: "import { flexRender, PaginationState, useReactTable } from '@tanstack/react-table';", Output: []string{"import { PaginationState, flexRender, useReactTable } from '@tanstack/react-table';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("PaginationState", 1, 22, 37)}},
			// Locks in upstream's comparator tie arm: a triggered ignoreCase fix reverses case-equivalent local names.
			{Code: "import { b, a, A } from 'values';", Options: []any{map[string]any{"ignoreCase": true}}, Output: []string{"import { A, a, b } from 'values';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 13, 14)}},
			// Locks in getFirstLocalMemberName(): a renamed first member compares by its local alias.
			{Code: "import { x as z } from 'one';\nimport { y as a } from 'two';", Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 30)}},
			// Locks in allowSeparatedGroups' exact threshold: consecutive lines do not reset.
			{Code: "import b from 'b';\nimport a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}, Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 19)}},
			// Locks in fix comment suppression for a comment immediately inside the opening brace.
			{Code: "import { /* keep */ b, a } from 'values';", Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 24, 25)}},
			// Locks in fix comment suppression for a comment immediately inside the closing brace.
			{Code: "import { b, a /* keep */ } from 'values';", Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 13, 14)}},
			// Comments contained within a specifier move with that specifier and do not suppress the fix.
			{Code: "import { b /* keep */ as c, a } from 'values';", Output: []string{"import { a, b /* keep */ as c } from 'values';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 29, 30)}},
			{Code: "import { b as /* keep */ c, a } from 'values';", Output: []string{"import { a, b as /* keep */ c } from 'values';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 29, 30)}},
			{Code: "import { 'b' /* keep */ as c, a } from 'values';", Output: []string{"import { a, 'b' /* keep */ as c } from 'values';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 31, 32)}},
			{Code: "import { type b /* keep */ as c, a } from 'values';", Output: []string{"import { a, type b /* keep */ as c } from 'values';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 34, 35)}},
			// Locks in UTF-16 code-unit ordering in the reverse direction.
			{Code: "import { 豈, 𐐀 } from 'unicode';", Output: []string{"import { 𐐀, 豈 } from 'unicode';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("𐐀", 1, 13, 15)}},
		},
	)
}

func TestSortImportsEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram("import { b, a } from 'values';", "edit-demand.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{Name: SortImportsRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners { return SortImportsRule.Run(ctx, nil) }}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) }},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	none := run(rule.EditDemandNone)
	autofix := run(rule.EditDemandAutofix)
	suggestion := run(rule.EditDemandSuggestion)
	all := run(rule.EditDemandAll)
	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{rule.EditDemandNone: none, rule.EditDemandAutofix: autofix, rule.EditDemandSuggestion: suggestion} {
		if !reflect.DeepEqual(withoutEdits(diagnostic), withoutEdits(all)) {
			t.Errorf("demand %d changed diagnostic identity", demand)
		}
	}
	if none.FixesPtr != nil || suggestion.FixesPtr != nil {
		t.Fatal("a non-autofix demand materialized fixes")
	}
	if autofix.FixesPtr == nil || !reflect.DeepEqual(autofix.FixesPtr, all.FixesPtr) {
		t.Fatal("autofix and all-edits demands did not produce the same fix")
	}
}
