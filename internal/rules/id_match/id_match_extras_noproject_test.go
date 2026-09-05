package id_match_test

// TestIdMatchExtrasNoProject locks in the answers the rule must give for a
// file no tsconfig owns, which rslint lints through a parsed-only Program that
// has no TypeChecker. Every other suite runs against a compiler-backed
// Program, where the checker can supply a declaration the binder scope walk
// never sees; these cases exist so the standard library's type names stay
// exempt either way. Its siblings are id_match_extras_branches_test.go,
// id_match_extras_dim4_test.go, id_match_extras_realuser_test.go and
// id_match_extras_typescript_test.go.

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules/id_match"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestIdMatchExtrasNoProject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		code    string
		options []any
		want    []string
	}{
		{
			// ---- A standard-library type is not the author's to name ----
			name:    "library type",
			code:    "let x: Record<string, number>;",
			options: []any{`^x$`},
		},
		{
			name:    "library type as a type argument",
			code:    "let x: Array<Uppercase<string>>;",
			options: []any{`^x$`},
		},
		{
			name:    "library type in export assignment",
			code:    "export = Record;",
			options: []any{`^x$`},
		},
		{
			name:    "library type in default export",
			code:    "export default Record;",
			options: []any{`^x$`},
		},
		{
			// ---- A qualified member is still an authored name ----
			name:    "library names as qualified type members",
			code:    "type X = Foo.Record<string> | Foo.Array<string> | Foo.Partial<string>;",
			options: []any{`^(X|Foo|string)$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^(X|Foo|string)$'.",
				"Identifier 'Array' does not match the pattern '^(X|Foo|string)$'.",
				"Identifier 'Partial' does not match the pattern '^(X|Foo|string)$'.",
			},
		},
		{
			// ---- A tuple label is authored even when it spells a library type ----
			name:    "library name as tuple label",
			code:    "type X = [Record: string];",
			options: []any{`^(X|string)$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^(X|string)$'.",
			},
		},
		{
			// ---- An import-type member belongs to the imported module ----
			name:    "library name as import type member",
			code:    `type X = import("./m").Record;`,
			options: []any{`^X$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^X$'.",
			},
		},
		{
			// ---- A type query reads a value, not the global type namespace ----
			name:    "library name in type query",
			code:    "type X = typeof Partial;",
			options: []any{`^X$`},
			want: []string{
				"Identifier 'Partial' does not match the pattern '^X$'.",
			},
		},
		{
			// ---- A computed type member also reads a value ----
			name:    "library name as computed type member",
			code:    "type X = { [Partial]: string };",
			options: []any{`^(X|string)$`},
			want: []string{
				"Identifier 'Partial' does not match the pattern '^(X|string)$'.",
			},
		},
		{
			// ---- Heritage type references still use the global type namespace ----
			name:    "library names in heritage type reference",
			code:    "interface X extends Partial<Record<string, string>> {}",
			options: []any{`^(X|string)$`},
		},
		{
			name:    "library name as function type parameter",
			code:    "type X = (Record: string) => void;",
			options: []any{`^(X|string|void)$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^(X|string|void)$'.",
			},
		},
		{
			name:    "library name as property signature",
			code:    "type X = { Record: string };",
			options: []any{`^(X|string)$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^(X|string)$'.",
			},
		},
		{
			name:    "library name as mapped type parameter",
			code:    "type X = { [Record in string]: string };",
			options: []any{`^(X|string)$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^(X|string)$'.",
			},
		},
		{
			name:    "library name as infer binding and reference",
			code:    "type X = string extends infer Record ? Record : never;",
			options: []any{`^(X|string|never)$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^(X|string|never)$'.",
				"Identifier 'Record' does not match the pattern '^(X|string|never)$'.",
			},
		},
		{
			name:    "library names in chained import type qualifier",
			code:    `type X = import("./m").Array.Record;`,
			options: []any{`^X$`},
			want: []string{
				"Identifier 'Array' does not match the pattern '^X$'.",
				"Identifier 'Record' does not match the pattern '^X$'.",
			},
		},
		{
			name:    "library name as import type argument",
			code:    `type X = import("./m").T<Partial<string>>;`,
			options: []any{`^(X|T|string)$`},
		},
		{
			// ---- A name the file declares itself is the author's ----
			name:    "shadowed library type",
			code:    "type Record = 1;\nlet x: Record;",
			options: []any{`^x$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^x$'.",
				"Identifier 'Record' does not match the pattern '^x$'.",
			},
		},
		{
			// ---- A library name spelled in a value position is a value ----
			name:    "library name as a value",
			code:    "let x = Record;",
			options: []any{`^x$`},
			want: []string{
				"Identifier 'Record' does not match the pattern '^x$'.",
			},
		},
		{
			// ---- A type nothing declares is still the author's to fix ----
			name:    "unresolved type",
			code:    "let x: Unknown_1;",
			options: []any{`^[^_]+$`},
			want: []string{
				"Identifier 'Unknown_1' does not match the pattern '^[^_]+$'.",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := lintWithoutProject(t, testCase.code, testCase.options)
			if len(got) != len(testCase.want) {
				t.Fatalf("got %d diagnostics %q, want %d %q", len(got), got, len(testCase.want), testCase.want)
			}
			for index, want := range testCase.want {
				if got[index] != want {
					t.Errorf("diagnostic %d = %q, want %q", index+1, got[index], want)
				}
			}
		})
	}
}

// lintWithoutProject runs the rule over code through a parsed-only Program,
// the source generation rslint builds for a lint target no tsconfig claims.
func lintWithoutProject(t *testing.T, code string, options []any) []string {
	t.Helper()

	const root = "/id-match-no-project"
	fileName := tspath.ResolvePath(root, "file.ts")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{fileName: code})
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(root, fs),
		CompilerOptions: &core.CompilerOptions{},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}

	messages := make([]string, 0, 2)
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program:                sourceProgram,
		Files:                  []string{fileName},
		ExcludedPathSubstrings: []string{},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:        "id-match",
				Environment: &rule.RuleEnvironment{},
				Severity:    rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Error("a parsed-only Program must not supply a TypeChecker")
					}
					return id_match.IdMatchRule.Run(ctx, options)
				},
			}}
		},
		OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
			messages = append(messages, diagnostic.Message.Description)
		},
	})
	return messages
}
