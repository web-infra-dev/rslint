// TestOrderBranchCoverage covers reachable upstream arms not isolated by the
// migrated suite, plus two real reports from import-js users. This file also
// locks in the minimatch behavior exposed through pathGroups. Cases use the
// native tsgo parser/resolver and keep upstream diagnostic/fix semantics.
package order_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestOrderBranchCoverage(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			// ---- Real-user #2683: type-group alphabetizing keeps parent before sibling ----
			{
				Code: "import type { CommandOptions } from '../lib/baseCommand';\n" +
					"import type { Options } from './openapi/validate';\n\n" +
					"import Command from '../lib/baseCommand';\n" +
					"import isHidden from '../lib/decorators/isHidden';\n\n" +
					"import OpenAPIValidateCommand from './openapi/validate';",
				Options: map[string]any{
					"alphabetize":             map[string]any{"order": "asc", "caseInsensitive": true},
					"groups":                  []any{"type", "builtin", "external", "internal", "parent", "sibling", "index", "object"},
					"newlines-between":        "always",
					"warnOnUnassignedImports": true,
				},
			},
			// ---- Locks in upstream TSImportEquals arm: exported import-equals is ignored ----
			{Code: "export import sibling = require('./z');\nimport fs = require('fs');"},
			// ---- Locks in resolver arm: a resolved project module named `fs` is internal, not builtin ----
			{
				Code:     "import external from 'async';\nimport local from 'fs';",
				TSConfig: "tsconfig.order-core-local.json",
				Options:  map[string]any{"groups": []any{"external", "internal", "builtin"}},
			},
			// ---- Locks in upstream exportMap arm: chain exports are not sorted when alphabetize is ignore ----
			{
				Code:    "exports.Z = 1;\nmodule.exports.A = 1;",
				Options: map[string]any{"named": map[string]any{"cjsExports": true}},
			},
			// ---- Locks in upstream scope arm: local module/exports bindings suppress CJS handling ----
			{
				Code: "function f() {\n  const module = { exports: {} };\n  const exports = {};\n" +
					"  module.exports = { b, a };\n  exports.B = 1;\n  exports.A = 1;\n}",
				Options: map[string]any{
					"named":       map[string]any{"enabled": true, "cjsExports": true},
					"alphabetize": map[string]any{"order": "asc"},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user #3235: relative paths use stable lexical ordering and converge ----
			{
				Code: "import ExecutionStrategyAbi from './abis/executionStrategy.json';\n" +
					"import { FullConfig } from './config';\n" +
					"import { handleVotingPowerValidationMetadata } from '../common/ipfs';\n" +
					"import L1AvatarExecutionStrategyAbi from './abis/l1/L1AvatarExectionStrategy';\n" +
					"import { Space } from '../../.checkpoint/models';",
				Options: map[string]any{
					"groups":      []any{"builtin", "external", "internal", "type"},
					"alphabetize": map[string]any{"order": "asc", "caseInsensitive": true},
				},
				Output: []string{
					"import { handleVotingPowerValidationMetadata } from '../common/ipfs';\nimport ExecutionStrategyAbi from './abis/executionStrategy.json';\nimport { FullConfig } from './config';\nimport L1AvatarExecutionStrategyAbi from './abis/l1/L1AvatarExectionStrategy';\nimport { Space } from '../../.checkpoint/models';",
					"import { Space } from '../../.checkpoint/models';\nimport { handleVotingPowerValidationMetadata } from '../common/ipfs';\nimport ExecutionStrategyAbi from './abis/executionStrategy.json';\nimport { FullConfig } from './config';\nimport L1AvatarExecutionStrategyAbi from './abis/l1/L1AvatarExectionStrategy';\n",
					"import { Space } from '../../.checkpoint/models';\nimport { handleVotingPowerValidationMetadata } from '../common/ipfs';\nimport ExecutionStrategyAbi from './abis/executionStrategy.json';\nimport L1AvatarExecutionStrategyAbi from './abis/l1/L1AvatarExectionStrategy';\nimport { FullConfig } from './config';\n",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}, {MessageId: "order"}, {MessageId: "order"}},
			},
			// ---- Locks in upstream reverse scan arm: one misplaced first import reports `after` ----
			{
				Code:   "import sibling from './z';\nimport fs from 'fs';\nimport async from 'async';",
				Output: []string{"import fs from 'fs';\nimport async from 'async';\nimport sibling from './z';\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./z` import should occur after import of `async`"}},
			},
			// ---- Locks in upstream named reverse scan arm and `after` named fixer ----
			{
				Code:    `import { c, a, b } from "pkg";`,
				Options: map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}},
				Output:  []string{`import { a, b, c } from "pkg";`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`c` import should occur after import of `b`"}},
			},
			// ---- Locks in upstream equal-name arm: aliases disambiguate the diagnostic ----
			{
				Code:    `import { A as C, A as B } from "pkg";`,
				Options: map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}},
				Output:  []string{`import { A as B, A as C } from "pkg";`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`A as B` import should occur before import of `A as C`"}},
			},
			// ---- Locks in upstream orderImportKind desc secondary comparator ----
			{
				Code: "import type { T } from 'pkg';\nimport { V } from 'pkg';",
				Options: map[string]any{
					"groups":      []any{[]any{"type", "external"}},
					"alphabetize": map[string]any{"order": "asc", "orderImportKind": "desc"},
				},
				Output: []string{"import { V } from 'pkg';\nimport type { T } from 'pkg';\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Locks in upstream pathGroup arm without an explicit position ----
			{
				Code: "import react from 'react';\nimport app from '@app/main';",
				Options: map[string]any{
					"groups":                        []any{"internal", "external"},
					"pathGroups":                    []any{map[string]any{"pattern": "@app/**", "group": "internal"}},
					"pathGroupsExcludedImportTypes": []any{},
				},
				Output: []string{"import app from '@app/main';\nimport react from 'react';\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Locks in upstream maxPosition arm: more than ten `before` groups round to a power of ten ----
			{
				Code:    "import ten from 'p10/x';\nimport zero from 'p00/x';",
				Options: orderManyBeforePathGroups(),
				Output:  []string{"import zero from 'p00/x';\nimport ten from 'p10/x';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Locks in upstream patternOptions arm: matchBase + nocase + extglob ----
			{
				Code: "import react from 'react';\nimport styles from 'styles/MAIN.CSS';",
				Options: map[string]any{
					"groups": []any{"internal", "external"},
					"pathGroups": []any{map[string]any{
						"pattern": "*.+(css|svg)", "group": "internal",
						"patternOptions": map[string]any{"matchBase": true, "nocase": true},
					}},
					"pathGroupsExcludedImportTypes": []any{},
				},
				Output: []string{"import styles from 'styles/MAIN.CSS';\nimport react from 'react';\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Locks in upstream canReorder arm: multi-declarator require reports without a fix ----
			{
				Code:   "const sibling = require('./z'), fs = require('fs');",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Locks in upstream canCross arm: an unassigned import is a fix barrier ----
			{
				Code:   "const sibling = require('./z');\nimport './side-effect';\nconst fs = require('fs');",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
		},
	)
}

func orderManyBeforePathGroups() map[string]any {
	pathGroups := make([]any, 0, 11)
	for i := range 11 {
		pathGroups = append(pathGroups, map[string]any{
			"pattern":  "p" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + "/**",
			"group":    "external",
			"position": "before",
		})
	}
	return map[string]any{
		"pathGroups":                    pathGroups,
		"pathGroupsExcludedImportTypes": []any{},
	}
}

// TestOrderSchema locks in the exact upstream option surface and dependencies.
func TestOrderSchema(t *testing.T) {
	t.Parallel()
	valid := []any{map[string]any{
		"groups": []any{"builtin", []any{"parent", "sibling"}, "type"},
		"named":  map[string]any{"enabled": true, "types": "types-first"},
	}}
	if err := order.OrderRule.Schema.Validate(valid); err != nil {
		t.Fatalf("valid upstream options rejected: %v", err)
	}
	invalid := []struct {
		name    string
		options []any
	}{
		{name: "unknown group", options: []any{map[string]any{"groups": []any{"not-a-group"}}}},
		{name: "duplicate group", options: []any{map[string]any{"groups": []any{"builtin", "builtin"}}}},
		{name: "unknown property", options: []any{map[string]any{"madeUp": true}}},
		{name: "type newline dependency", options: []any{map[string]any{"newlines-between-types": "always"}}},
	}
	for _, test := range invalid {
		if err := order.OrderRule.Schema.Validate(test.options); err == nil {
			t.Errorf("%s: invalid options accepted", test.name)
		}
	}
}

// TestOrderEditDemand proves every fixer is deferred: diagnostic identity is
// demand-independent, and only autofix-capable demands materialize edits.
func TestOrderEditDemand(t *testing.T) {
	t.Parallel()
	const fileName = "order-edit-demand.ts"
	const code = "import b from 'b';\nimport a from 'a';"
	const fixed = "import a from 'a';\nimport b from 'b';\n"

	program, sourceFile := createOrderProgram(t, fileName, code)
	diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		got := lintOrderWithDemand(program, sourceFile, demand, alphabetizeOptionsForDemand())
		if len(got) != 1 {
			t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(got))
		}
		diagnostics[demand] = got[0]
	}

	base := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		want, got := base, diagnostic
		want.FixesPtr, want.Suggestions = nil, nil
		got.FixesPtr, got.Suggestions = nil, nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic identity", demand)
		}
	}
	if diagnostics[rule.EditDemandNone].FixesPtr != nil || diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
		t.Fatal("non-autofix demand materialized fixes")
	}
	autofix := diagnostics[rule.EditDemandAutofix].FixesPtr
	all := diagnostics[rule.EditDemandAll].FixesPtr
	if autofix == nil || all == nil || !reflect.DeepEqual(*autofix, *all) {
		t.Fatal("autofix artifacts differ between autofix and all demand")
	}
	if output, _, _ := linter.ApplyRuleFixes(code, []rule.RuleDiagnostic{diagnostics[rule.EditDemandAll]}); output != fixed {
		t.Fatalf("fixed output = %q, want %q", output, fixed)
	}
}

func alphabetizeOptionsForDemand() []any {
	return []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}
}

func lintOrderWithDemand(program *compiler.Program, sourceFile *ast.SourceFile, demand rule.EditDemand, options []any) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program, File: sourceFile.FileName(), HasTypeInfo: true, ExcludePaths: []string{},
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name: order.OrderRule.Name, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners { return order.OrderRule.Run(ctx, options) },
			}}
		},
		Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})
	return diagnostics
}

func createOrderProgram(t testing.TB, fileName, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()
	root := fixtures.GetRootDir()
	fs := utils.NewOverlayVFS(root.FS, map[string]string{tspath.ResolvePath(root.Dir, fileName): code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

func TestOrderMinimatchCompatibility(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    "import react from 'react';\nimport scope from 'scope';",
				Options: minimatchPathGroupOptions("scope/package", map[string]any{}),
			},
		},
		[]rule_tester.InvalidTestCase{
			// Empty module specifiers and patterns are both legal strings;
			// minimatch 3 treats them as an exact match.
			{
				Code:    "import react from 'react';\nimport empty from '';",
				Options: minimatchPathGroupOptions("", nil),
				Output:  []string{"import empty from '';\nimport react from 'react';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// import-js 2.32 uses minimatch 3.1.5, whose brace expansion includes
			// numeric/letter ranges in addition to comma alternatives.
			{
				Code:    "import react from 'react';\nimport ranged from 'pkg/2';",
				Options: minimatchPathGroupOptions("pkg/{1..3}", nil),
				Output:  []string{"import ranged from 'pkg/2';\nimport react from 'react';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// partial accepts a module specifier that is a matching prefix of the
			// configured pattern.
			{
				Code:    "import react from 'react';\nimport scope from 'scope';",
				Options: minimatchPathGroupOptions("scope/package", map[string]any{"partial": true}),
				Output:  []string{"import scope from 'scope';\nimport react from 'react';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// flipNegate makes the non-negated portion of a leading-! pattern the
			// positive match, matching minimatch's option semantics.
			{
				Code:    "import other from 'other';\nimport react from 'react';",
				Options: minimatchPathGroupOptions("!react", map[string]any{"flipNegate": true}),
				Output:  []string{"import react from 'react';\nimport other from 'other';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
		},
	)
}

func minimatchPathGroupOptions(pattern string, patternOptions map[string]any) map[string]any {
	pathGroup := map[string]any{
		"pattern": pattern,
		"group":   "internal",
	}
	if patternOptions != nil {
		pathGroup["patternOptions"] = patternOptions
	}
	return map[string]any{
		"groups":                        []any{"internal", "external"},
		"pathGroups":                    []any{pathGroup},
		"pathGroupsExcludedImportTypes": []any{},
	}
}
