// TestOrderBranchCoverage covers configuration, diagnostic, and fixer behavior
// not isolated by the migrated suite, plus two real import-js user reports.
// It also exercises the minimatch options exposed through pathGroups using the
// native tsgo parser and resolver.
package order_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
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
			// Node resolves an empty external-module folder to the package root,
			// making a local path alias external just like `async`.
			{
				Code:     "import external from 'async';\nimport local from 'local-api';",
				TSConfig: "tsconfig.order-core-local.json",
				Options: map[string]any{
					"groups":      []any{"internal", "external"},
					"alphabetize": map[string]any{"order": "asc"},
				},
				Settings: map[string]any{"import/external-module-folders": []any{""}},
			},
			// ---- Real-user: #2683 type-group alphabetizing keeps parent before sibling ----
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
			// Exported import-equals declarations do not participate in ordering.
			{Code: "export import sibling = require('./z');\nimport fs = require('fs');"},
			// Requires declared by an exported variable do not participate either;
			// tsgo represents `export` as a declaration modifier.
			{
				Code:    "export const sibling = require('./z');\nconst fs = require('fs');",
				Options: map[string]any{"alphabetize": map[string]any{"order": "asc"}},
			},
			// Empty named bindings have no import bindings and are ignored unless
			// warnOnUnassignedImports is enabled.
			{
				Code:    "import {} from './z';\nimport a from './a';",
				Options: map[string]any{"alphabetize": map[string]any{"order": "asc"}},
			},
			// A Node builtin remains builtin even when TypeScript paths maps the
			// exact specifier to a project file. The configured TypeScript resolver
			// checks Node builtins before attempting filesystem resolution.
			{
				Code:     "import builtin from 'buffer';\nimport external from 'async';",
				TSConfig: "tsconfig.order-core-local.json",
				Options:  map[string]any{"groups": []any{"builtin", "external", "internal"}},
			},
			// import/internal-regex retains the upstream precedence over builtins.
			{
				Code:     "import overridden from 'buffer';\nimport external from 'async';",
				TSConfig: "tsconfig.order-core-local.json",
				Options:  map[string]any{"groups": []any{"internal", "external", "builtin"}},
				Settings: map[string]any{"import/internal-regex": "^buffer$"},
			},
			// Non-exact builtin subpath specifiers remain resolution-sensitive.
			{
				Code:     "import local from 'fs/not-a-builtin';\nimport external from 'async';",
				TSConfig: "tsconfig.order-core-local.json",
				Options:  map[string]any{"groups": []any{"internal", "external", "builtin"}},
			},
			// Configured core modules also remain resolution-sensitive.
			{
				Code:     "import local from 'virtual';\nimport external from 'async';",
				TSConfig: "tsconfig.order-core-local.json",
				Options:  map[string]any{"groups": []any{"internal", "external", "builtin"}},
				Settings: map[string]any{"import/core-modules": []any{"virtual"}},
			},
			// CommonJS export assignments are not sorted when alphabetize is ignored.
			{
				Code:    "exports.Z = 1;\nmodule.exports.A = 1;",
				Options: map[string]any{"named": map[string]any{"cjsExports": true}},
			},
			// Local `module` and `exports` bindings suppress CommonJS handling.
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
			// Duplicate display names gain aliases as diagnostics are produced;
			// later diagnostics sharing the first entry retain that disambiguation.
			{
				Code:    `import { A as E, A as D, A as C, A as B, A } from "./x";`,
				Options: map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}},
				Output: []string{
					`import { A as D, A as E, A as C, A as B, A } from "./x";`,
					`import { A as C, A as D, A as E, A as B, A } from "./x";`,
					`import { A as B, A as C, A as D, A as E, A } from "./x";`,
					`import { A, A as B, A as C, A as D, A as E } from "./x";`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "order", Message: "`A as D` import should occur before import of `A as E`", Line: 1, Column: 18},
					{MessageId: "order", Message: "`A` import should occur before import of `A as E`", Line: 1, Column: 26},
					{MessageId: "order", Message: "`A` import should occur before import of `A as E`", Line: 1, Column: 34},
					{MessageId: "order", Message: "`A` import should occur before import of `A as E`", Line: 1, Column: 42},
				},
			},
			// uniqueItems is local to each JSON array. When a type appears once as
			// a scalar and again in a nested group, the later group rank wins.
			{
				Code: "import external from 'package-name';\n" +
					"import builtin from 'fs';",
				Options: map[string]any{
					"groups":      []any{"external", []any{"external", "builtin"}},
					"alphabetize": map[string]any{"order": "asc"},
				},
				Output: []string{"import builtin from 'fs';\n" +
					"import external from 'package-name';\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// A user-configured core-module name is builtin only when resolution
			// fails, even when the written name otherwise looks relative.
			{
				Code: "import external from 'package-name';\n" +
					"import customBuiltin from '../definitely-missing-order-core';",
				Settings: map[string]any{"import/core-modules": []any{".."}},
				Output: []string{"import customBuiltin from '../definitely-missing-order-core';\n" +
					"import external from 'package-name';\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// Named require destructuring accepts every JavaScript literal argument,
			// not just module-name strings.
			{
				Code:    "const { b, a } = require(1);",
				Options: map[string]any{"named": map[string]any{"require": true}, "alphabetize": map[string]any{"order": "asc"}},
				Output:  []string{"const { a, b } = require(1);"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 1, Column: 12}},
			},
			// An explicitly empty groups array is truthy in JavaScript and must not
			// fall back to the default groups. Omitted types still share a rank and
			// therefore participate in alphabetizing.
			{
				Code:    "import b from 'b';\nimport a from 'a';",
				Options: map[string]any{"groups": []any{}, "alphabetize": map[string]any{"order": "asc"}},
				Output:  []string{"import a from 'a';\nimport b from 'b';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// Optional require chains are not movable ordering candidates and remain
			// an autofix barrier in tsgo's direct optional-chain representation.
			{
				Code: "const sibling = require('./z');\n" +
					"const optional = require('middle').value?.();\n" +
					"const fs = require('fs');",
				Options: map[string]any{"alphabetize": map[string]any{"order": "asc"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 3, Column: 12}},
			},
			// With warnings enabled an empty import participates in ordering, but
			// its zero specifiers still make it unsafe for the fixer to cross.
			{
				Code: "import {} from './z';\n" +
					"import a from './a';",
				Options: map[string]any{
					"warnOnUnassignedImports": true,
					"alphabetize":             map[string]any{"order": "asc"},
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// The same zero-specifier import is a fixer barrier between two imports
			// that do participate in ordering.
			{
				Code: "import z from './z';\n" +
					"import {} from './middle';\n" +
					"import a from './a';",
				Options: map[string]any{"alphabetize": map[string]any{"order": "asc"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 3, Column: 1}},
			},
			// The reverse scan emits fewer diagnostics here and preserves the
			// no-final-newline replacement behavior.
			{
				Code:   "import sibling from './z';\nimport fs from 'fs';\nimport async from 'async';",
				Output: []string{"import fs from 'fs';\nimport async from 'async';import sibling from './z';\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`./z` import should occur after import of `async`"}},
			},
			// Named sorting likewise chooses the shorter reverse scan and `after` fix.
			{
				Code:    `import { c, a, b } from "pkg";`,
				Options: map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}},
				Output:  []string{`import { a, b, c } from "pkg";`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`c` import should occur after import of `b`"}},
			},
			// Aliases disambiguate otherwise equal display names.
			{
				Code:    `import { A as C, A as B } from "pkg";`,
				Options: map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}},
				Output:  []string{`import { A as B, A as C } from "pkg";`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Message: "`A as B` import should occur before import of `A as C`"}},
			},
			// orderImportKind supports a descending secondary comparison.
			{
				Code: "import type { T } from 'pkg';\nimport { V } from 'pkg';",
				Options: map[string]any{
					"groups":      []any{[]any{"type", "external"}},
					"alphabetize": map[string]any{"order": "asc", "orderImportKind": "desc"},
				},
				Output: []string{"import { V } from 'pkg';\nimport type { T } from 'pkg';\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// A pathGroup without a position uses its target group's base rank.
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
			// More than ten `before` path groups still receive distinct fractional ranks.
			{
				Code:    "import ten from 'p10/x';\nimport zero from 'p00/x';",
				Options: orderManyBeforePathGroups(),
				Output:  []string{"import zero from 'p00/x';\nimport ten from 'p10/x';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// patternOptions compose matchBase, nocase, and extended globs.
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
			// A multi-declarator require reports but is not a safe whole-statement fix.
			{
				Code:   "const sibling = require('./z'), fs = require('fs');",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// ---- Real-user: #1548 a side-effect import is an autofix barrier ----
			{
				Code:   "const sibling = require('./z');\nimport './side-effect';\nconst fs = require('fs');",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "order"}},
			},
			// Upstream 2.32.0 sorts the numeric body indexes as strings. For indexes
			// 2 and 10 that produces an empty safety-check range and offers a fix
			// across side-effect imports; numeric source order must suppress it.
			{
				Code: "import a from './a';\n" +
					"import './side-0';\n" +
					"import z from './z';\n" +
					"import './side-1';\n" +
					"import './side-2';\n" +
					"import './side-3';\n" +
					"import './side-4';\n" +
					"import './side-5';\n" +
					"import './side-6';\n" +
					"import './side-7';\n" +
					"import b from './b';",
				Options: map[string]any{"alphabetize": map[string]any{"order": "asc"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 11, Column: 1}},
			},
			// Although an exported require is ignored for ordering, it remains an
			// unrelated statement and blocks a fix between surrounding imports.
			{
				Code:    "import z from './z';\nexport const middle = require('./middle');\nimport a from './a';",
				Options: map[string]any{"alphabetize": map[string]any{"order": "asc"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 3, Column: 1}},
			},
			// A non-string literal require is safe to cross during a fix even though
			// it is not itself registered as a module import.
			{
				Code:    "import z from \"./z\";\nconst middle = require(1);\nimport a from \"./a\";",
				Options: map[string]any{"alphabetize": map[string]any{"order": "asc"}},
				Output:  []string{"import a from \"./a\";\nimport z from \"./z\";\nconst middle = require(1);\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 3, Column: 1}},
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
	crossNestedDuplicate := []any{map[string]any{
		"groups": []any{"external", []any{"external", "builtin"}},
	}}
	if err := order.OrderRule.Schema.Validate(crossNestedDuplicate); err != nil {
		t.Fatalf("cross-nested duplicate rejected: %v", err)
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
	tests := []struct {
		name    string
		code    string
		fixed   string
		options []any
	}{
		{
			name:    "module-swap",
			code:    "import b from 'b';\nimport a from 'a';",
			fixed:   "import a from 'a';\nimport b from 'b';\n",
			options: alphabetizeOptionsForDemand(),
		},
		{
			name:    "named-swap",
			code:    "import { b, a } from 'pkg';",
			fixed:   "import { a, b } from 'pkg';",
			options: []any{map[string]any{"named": true, "alphabetize": map[string]any{"order": "asc"}}},
		},
		{
			name:    "insert-group-newline",
			code:    "import fs from 'fs';\nimport sibling from './sibling';",
			fixed:   "import fs from 'fs';\n\nimport sibling from './sibling';",
			options: []any{map[string]any{"newlines-between": "always"}},
		},
		{
			name:    "remove-group-newline",
			code:    "import fs from 'fs';\n\nimport sibling from './sibling';",
			fixed:   "import fs from 'fs';\nimport sibling from './sibling';",
			options: []any{map[string]any{"newlines-between": "never"}},
		},
		{
			name: "consolidate-multiline-island",
			code: "import a from 'a';\n" +
				"import {\n" +
				"  b,\n" +
				"} from 'b';",
			fixed: "import a from 'a';\n\n" +
				"import {\n" +
				"  b,\n" +
				"} from 'b';",
			options: []any{map[string]any{
				"newlines-between":   "always-and-inside-groups",
				"consolidateIslands": "inside-groups",
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			program, sourceFile := createOrderProgram(t, "order-edit-demand-"+test.name+".ts", test.code)
			diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{}
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				got := lintOrderWithDemand(program, sourceFile, demand, test.options)
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
			if output, _, _ := linter.ApplyRuleFixes(test.code, []rule.RuleDiagnostic{diagnostics[rule.EditDemandAll]}); output != test.fixed {
				t.Fatalf("fixed output = %q, want %q", output, test.fixed)
			}
		})
	}
}

// TestOrderCommonJSControlFlowFixRoots covers the safe movable unit for nested
// assignments. An unbraced if/else pair and assignments in one switch clause
// share an export sequence, but both entries resolve to the same statement
// root, so the fix only normalizes its trailing newline. A braced block keeps
// each ExpressionStatement independently movable and can be sorted.
func TestOrderCommonJSControlFlowFixRoots(t *testing.T) {
	t.Parallel()

	options := []any{map[string]any{
		"named":       map[string]any{"cjsExports": true},
		"alphabetize": map[string]any{"order": "asc"},
	}}
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "if-else",
			code: "if (ready) exports.Z = 1;\n" +
				"else exports.A = 1;",
			want: "if (ready) exports.Z = 1;\n" +
				"else exports.A = 1;\n",
		},
		{
			name: "switch-clause",
			code: "switch (kind) {\n" +
				"case 1:\n" +
				"  exports.Z = 1;\n" +
				"  exports.A = 1;\n" +
				"}",
			want: "switch (kind) {\n" +
				"case 1:\n" +
				"  exports.Z = 1;\n" +
				"  exports.A = 1;\n" +
				"}\n",
		},
		{
			name: "braced-block",
			code: "if (ready) {\n" +
				"  exports.Z = 1;\n" +
				"  exports.A = 1;\n" +
				"}",
			want: "if (ready) {\n" +
				"  exports.A = 1;\n" +
				"  exports.Z = 1;\n" +
				"}",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			program, sourceFile := createOrderProgram(t, "order-cjs-root-"+test.name+".ts", test.code)
			diagnostics := lintOrderWithDemand(program, sourceFile, rule.EditDemandAutofix, options)
			if len(diagnostics) != 1 {
				t.Fatalf("got %d diagnostics, want 1", len(diagnostics))
			}
			output, _, fixed := linter.ApplyRuleFixes(test.code, diagnostics)
			if !fixed {
				t.Fatal("expected a fix")
			}
			if output != test.want {
				t.Fatalf("fixed output = %q, want %q", output, test.want)
			}
		})
	}
}

func alphabetizeOptionsForDemand() []any {
	return []any{map[string]any{"alphabetize": map[string]any{"order": "asc"}}}
}

func lintOrderWithDemand(program *compiler.Program, sourceFile *ast.SourceFile, demand rule.EditDemand, options []any) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(), HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
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
			// An explicitly supplied empty patternOptions object keeps minimatch's
			// leading-comment behavior, so a leading-# pattern matches nothing.
			{
				Code:    "import react from 'react';\nimport alias from '#alias';",
				Options: minimatchPathGroupOptions("#alias", map[string]any{}),
			},
		},
		[]rule_tester.InvalidTestCase{
			// With patternOptions omitted, import/order disables leading comments
			// and treats a leading # as an ordinary package-import pattern.
			{
				Code:    "import react from 'react';\nimport alias from '#alias';",
				Options: minimatchPathGroupOptions("#alias", nil),
				Output:  []string{"import alias from '#alias';\nimport react from 'react';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
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
			// A partial globstar may consume the complete written suffix while the
			// remainder of the configured pattern is still pending.
			{
				Code:    "import react from 'react';\nimport feature from 'scope/feature';",
				Options: minimatchPathGroupOptions("scope/**/package", map[string]any{"partial": true}),
				Output:  []string{"import feature from 'scope/feature';\nimport react from 'react';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// minimatch 3.1.5 places sections between multiple `**` tokens using
			// partial-mode cutoffs. Once the written path is too short for the
			// first section, a dot-name beyond that cutoff remains a valid prefix.
			{
				Code:    "import react from 'react';\nimport hidden from 'scope/.hidden';",
				Options: minimatchPathGroupOptions("scope/**/package/**/tail", map[string]any{"partial": true}),
				Output:  []string{"import hidden from 'scope/.hidden';\nimport react from 'react';\n"},
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
