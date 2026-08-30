// no_unsafe_enum_comparison_extras_test.go covers rslint-specific edge cases.
// Migrated upstream cases live in no_unsafe_enum_comparison_upstream_test.go.
package no_unsafe_enum_comparison

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnsafeEnumComparisonDoesNotOfferUnsafeSuggestions(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string]string
		target string
	}{
		{
			name: "plain literal",
			files: map[string]string{
				"use.ts": `enum E { A = 1 }
const value = 1 as E;
value === 1;`,
			},
			target: "use.ts",
		},
		{
			name: "computed member declared in another file",
			files: map[string]string{
				"enum.ts": strings.Repeat("// padding\n", 100) + `export enum E { ['x'] = 1 }`,
				"use.ts": `import { E } from './enum.js';
const value = 1 as E;
value === 1;`,
			},
			target: "use.ts",
		},
		{
			name: "aliased import",
			files: map[string]string{
				"enum.ts": `export enum Original { A = 1 }`,
				"use.ts": `import { Original as Alias } from './enum.js';
const value = 1 as Alias;
value === 1;`,
			},
			target: "use.ts",
		},
		{
			name: "namespace import",
			files: map[string]string{
				"enum.ts": `export enum E { A = 1 }`,
				"use.ts": `import * as N from './enum.js';
const value = 1 as N.E;
value === 1;`,
			},
			target: "use.ts",
		},
		{
			name: "type-only import",
			files: map[string]string{
				"enum.ts": `export enum E { A = 1 }`,
				"use.ts": `import type { E } from './enum.js';
declare const value: E;
value === 1;`,
			},
			target: "use.ts",
		},
		{
			name: "escaped string member",
			files: map[string]string{
				"use.ts": `enum E { 'a\\b' = 1 }
const value = 1 as E;
value === 1;`,
			},
			target: "use.ts",
		},
		{
			name: "assignment result",
			files: map[string]string{
				"use.ts": `enum E { A = 1 }
declare const value: E;
let written = 0;
value === (written = 1);`,
			},
			target: "use.ts",
		},
		{
			name: "sequence with call",
			files: map[string]string{
				"use.ts": `enum E { A = 1 }
declare const value: E;
declare function log(): void;
value === (log(), 1);`,
			},
			target: "use.ts",
		},
		{
			name: "mutated enum object",
			files: map[string]string{
				"use.ts": `enum E { A = 1 }
(E as any).A = 2;
const value = 1 as E;
value === 1;`,
			},
			target: "use.ts",
		},
		{
			name: "cyclic initialization",
			files: map[string]string{
				"enum.ts": `import './use.js';
export enum E { A = 1 }`,
				"use.ts": `import { E } from './enum.js';
const value = 1 as E;
value === 1;`,
			},
			target: "use.ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := lintNoUnsafeEnumComparisonFiles(t, tt.files, tt.target)
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
			}
			if diagnostics[0].Suggestions != nil {
				t.Fatalf("suggestions = %#v, want nil", *diagnostics[0].Suggestions)
			}
		})
	}
}

func lintNoUnsafeEnumComparisonFiles(t testing.TB, files map[string]string, target string) []rule.RuleDiagnostic {
	t.Helper()
	root := fixtures.GetRootDir()
	overlay := make(map[string]string, len(files))
	for name, code := range files {
		overlay[tspath.ResolvePath(root.Dir, name)] = code
	}
	fs := utils.NewOverlayVFS(root.FS, overlay)
	compilerProgram, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
	if err != nil {
		t.Fatalf("create program: %v", err)
	}
	sourceFile := compilerProgram.GetSourceFile(target)
	if sourceFile == nil {
		t.Fatalf("missing target %q", target)
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     lintprogram.NewFromCompiler(compilerProgram),
		File:        sourceFile.FileName(),
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:             NoUnsafeEnumComparisonRule.Name,
				Severity:         rule.SeverityError,
				RequiresTypeInfo: true,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUnsafeEnumComparisonRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandSuggestion,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}
