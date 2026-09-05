package id_match

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

func TestMayReferenceExternalTypeOrNamespaceCoversExactMeanings(t *testing.T) {
	for _, source := range []string{
		`export = Subject;`,
		`type T = Subject;`,
		`type T = Subject.Member;`,
		`interface I extends Subject {}`,
		`interface I extends Subject.Member {}`,
		`class C implements Subject {}`,
		`class C implements Subject.Member {}`,
		`import Alias = Subject;`,
		`import Alias = Subject.Member;`,
	} {
		t.Run(source, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.ts",
				Path:     "/test.ts",
			}, source, core.ScriptKindTS)
			var subject *ast.Node
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if subject == nil && node.Kind == ast.KindIdentifier && node.Text() == "Subject" &&
					scope.IsReferenceIdentifier(node) {
					subject = node
				}
				node.ForEachChild(visit)
				return false
			}
			sourceFile.AsNode().ForEachChild(visit)
			if subject == nil {
				t.Fatal("Subject reference not found")
			}
			meaning := scope.TypeScriptReferenceMeaning(subject)
			if meaning&(ast.SemanticMeaningType|ast.SemanticMeaningNamespace) == 0 {
				t.Fatalf("meaning = %v, want type or namespace capability", meaning)
			}
			if !mayReferenceExternalTypeOrNamespace(subject) {
				t.Fatal("fast path rejected a type- or namespace-capable reference")
			}
		})
	}
}

func TestIdMatchExternalTypeAndNamespaceReferences(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/external_references.txtar")
	root := tspath.NormalizePath(archive.Materialize(t, "external-references"))
	fs := bundled.WrapFS(osvfs.FS())
	program, err := utils.CreateProgram(true, fs, root, "tsconfig.json", utils.CreateCompilerHost(root, fs))
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	sourceProgram := lintprogram.NewFromCompiler(program)

	tests := []struct {
		name       string
		file       string
		properties bool
		want       []string
	}{
		{
			name: "external type and namespace meanings",
			file: "valid.ts",
		},
		{
			name: "dual export of external namespace",
			file: "namespace-export.ts",
		},
		{
			name:       "external heritage roots with properties enabled",
			file:       "valid.ts",
			properties: true,
		},
		{
			name: "dual export of external class",
			file: "class-export.ts",
		},
		{
			name: "parenthesized external type is a value",
			file: "parenthesized.ts",
			want: []string{"External_Interface"},
		},
		{
			name: "value local and unresolved controls",
			file: "controls.ts",
			want: []string{
				"External_NS",
				"External_Value",
				"External_Class",
				"External_Value",
				"External_Interface",
				"Missing_NS",
				"Local_Type",
				"Local_Type",
			},
		},
		{
			name: "local value wins over external type",
			file: "shadow.ts",
			want: []string{"External_Interface", "External_Interface"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileName := tspath.ResolvePath(root, test.file)
			got := make([]string, 0, len(test.want))
			options := []any{`^[^_]+$`}
			if test.properties {
				options = append(options, map[string]any{"properties": true})
			}
			testutil.LintProgram(t, testutil.LintProgramOptions{
				Program:                sourceProgram,
				Files:                  []string{fileName},
				ExcludedPathSubstrings: []string{},
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name:        IdMatchRule.Name,
						Environment: &rule.RuleEnvironment{},
						Severity:    rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return IdMatchRule.Run(ctx, options)
						},
					}}
				},
				OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
					got = append(got, diagnostic.Message.Description)
				},
			})
			if len(got) != len(test.want) {
				t.Fatalf("diagnostic names = %q, want %q", got, test.want)
			}
			for index, want := range test.want {
				wantMessage := messageNotMatch(want, `^[^_]+$`).Description
				if got[index] != wantMessage {
					t.Errorf("diagnostic %d = %q, want %q", index+1, got[index], wantMessage)
				}
			}
		})
	}
}
