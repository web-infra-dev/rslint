package prefer_set_size_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_set_size"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferSetSizeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_set_size.PreferSetSizeRule,
		[]rule_tester.ValidTestCase{
			// Static bracket access and optional access are distinct from upstream's
			// non-optional dot-member shape.
			{Code: "Array['from'](new Set(values)).length", FileName: "file.js"},
			{Code: "Array.from(new Set(values))?.length", FileName: "file.js"},
			{Code: "[...new Set(values)]?.length", FileName: "file.js"},
			// A TS assertion on the conversion itself is visible to upstream and does
			// not count as one of its two conversion shapes.
			{Code: "([...(set as Set<string>)] as string[]).length", FileName: "file.ts"},
			// Upstream's isMethodCall rejects spread arguments by default: the
			// source cannot replace the conversion as one member-expression object.
			{Code: "declare const args: [Set<number>]; Array.from(...args).length", FileName: "file.ts"},
			{Code: "const set = set; [...set].length", FileName: "file.js"},
			{Code: "class Values extends Array<string> {} const set = new Values(); Array.from(set).length", FileName: "file.ts"},
		},
		[]rule_tester.InvalidTestCase{
			// Const aliases recurse, while `let` deliberately did not match in the
			// JavaScript upstream parser.
			invalid("const first = new Set(values); const second = first; [...second].length", "const first = new Set(values); const second = first; second.size", "file.js"),
			// A Set construction without constructor parentheses needs a wrapper once
			// it becomes the object of a member expression.
			invalid("[...new Set].length", "(new Set).size", "file.js"),
			invalid("[...new (Set)].length", "(new (Set)).size", "file.js"),
			// Rslint-specific runtime helper regressions; these are not part of
			// eslint-plugin-unicorn v74.0.0's prefer-set-size test corpus.
			invalid("[...(flag ? new Set() : new Set())].length", "(flag ? new Set() : new Set()).size", "file.js"),
			invalid("[...(sideEffect(), new Set())].length", "(sideEffect(), new Set()).size", "file.js"),
			invalid("function size(value: unknown) { return [...(new Set() satisfies Set)].length; }", "function size(value: unknown) { return (new Set() satisfies Set).size; }", "file.ts"),
			// Comments nested in Set survive; an outer conversion comment suppresses
			// the whole fix rather than dropping author text.
			invalid("[...new /* retained */ Set(values)].length", "new /* retained */ Set(values).size", "file.js"),
			{Code: "Array.from(/* outer */ new /* retained */ Set(values)).length", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{errorAtLength("Array.from(/* outer */ new /* retained */ Set(values)).length", 0)}},
			// Authored TypeScript wrappers remain visible in the replacement and are
			// parenthesized before they become a member-expression object.
			{Code: "function size(value: unknown) { return Array.from(value satisfies Set<string>).length; }", FileName: "file.ts"},
			// A project-backed non-null assertion narrows the nullable Set before
			// `isSet` unwraps it.
			invalid("declare const set: Set<string> | null; [...set!].length", "declare const set: Set<string> | null; (set!).size", "file.ts"),
			// Checker-backed expressions retain Set inheritance, including calls
			// that cannot be classified through a binding annotation.
			invalid("class StringSet extends Set<string> {} const set = new StringSet(); [...set].length", "class StringSet extends Set<string> {} const set = new StringSet(); set.size", "file.ts"),
			invalid("interface StringSet extends ReadonlySet<string> {} declare function getSet(): StringSet; Array.from(getSet()).length", "interface StringSet extends ReadonlySet<string> {} declare function getSet(): StringSet; getSet().size", "file.ts"),
			// Memoization must not reuse flow-sensitive checker results between
			// references to the same mutable binding.
			invalid("function size(value: Set<string> | string[]) { if (value instanceof Set) return Array.from(value).length; return Array.from(value).length; }", "function size(value: Set<string> | string[]) { if (value instanceof Set) return value.size; return Array.from(value).length; }", "file.ts"),
		},
	)
}

func TestPreferSetSizeProjectFalseTypeSyntax(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		output       string
		noDiagnostic bool
	}{
		{
			name:   "parameter Set annotation",
			code:   "function getSize(set: Set<string>) { return Array.from(set).length; }",
			output: "function getSize(set: Set<string>) { return set.size; }",
		},
		{
			name:   "parameter ReadonlySet annotation",
			code:   "function getSize(set: ReadonlySet<string>) { return Array.from(set).length; }",
			output: "function getSize(set: ReadonlySet<string>) { return set.size; }",
		},
		{
			name:   "spread assertion",
			code:   "function getSize(set: unknown) { return [...(set as Set<string>)].length; }",
			output: "function getSize(set: unknown) { return (set as Set<string>).size; }",
		},
		{
			name:   "Array.from assertion",
			code:   "function getSize(set: unknown) { return Array.from(set as Set<string>).length; }",
			output: "function getSize(set: unknown) { return (set as Set<string>).size; }",
		},
		{
			name:   "type alias annotation",
			code:   "type SetAlias = Set<string>; function getSize(set: SetAlias) { return Array.from(set).length; }",
			output: "type SetAlias = Set<string>; function getSize(set: SetAlias) { return set.size; }",
		},
		{
			name:   "all Set union",
			code:   "function getSize(set: Set<string> | ReadonlySet<string>) { return Array.from(set).length; }",
			output: "function getSize(set: Set<string> | ReadonlySet<string>) { return set.size; }",
		},
		{
			name:   "Set intersection assertion",
			code:   "function getSize(set: unknown) { return [...(set as Set<string> & { tag: string })].length; }",
			output: "function getSize(set: unknown) { return (set as Set<string> & { tag: string }).size; }",
		},
		{
			name:   "interface heritage",
			code:   "interface StringSet extends ReadonlySet<string> {} function getSize(set: StringSet) { return Array.from(set).length; }",
			output: "interface StringSet extends ReadonlySet<string> {} function getSize(set: StringSet) { return set.size; }",
		},
		{
			name:   "class heritage",
			code:   "class StringSet extends Set<string> {} function getSize(set: StringSet) { return Array.from(set).length; }",
			output: "class StringSet extends Set<string> {} function getSize(set: StringSet) { return set.size; }",
		},
		{
			name:         "mixed union is not known Set",
			code:         "function getSize(set: Set<string> | string[]) { return Array.from(set).length; }",
			noDiagnostic: true,
		},
		{
			name:         "nullable union is not known Set",
			code:         "function getSize(set: Set<string> | null) { return Array.from(set).length; }",
			noDiagnostic: true,
		},
		{
			name:         "cyclic type alias",
			code:         "type Recursive = Recursive; function getSize(set: Recursive) { return Array.from(set).length; }",
			noDiagnostic: true,
		},
		{
			name:         "implements is not inheritance",
			code:         "class Box implements ReadonlySet<string> {} function getSize(set: Box) { return Array.from(set).length; }",
			noDiagnostic: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diagnostics := lintPreferSetSizeProjectFalse(t, test.code)
			if test.noDiagnostic {
				if len(diagnostics) != 0 {
					t.Fatalf("diagnostics = %+v, want none", diagnostics)
				}
				return
			}
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostic count = %d, want 1: %+v", len(diagnostics), diagnostics)
			}
			output, unapplied, fixed := linter.ApplyRuleFixes(test.code, diagnostics)
			if !fixed || len(unapplied) != 0 || output != test.output {
				t.Fatalf("source-only output = %q, fixed=%t, unapplied=%+v; want %q", output, fixed, unapplied, test.output)
			}
		})
	}
}

// A cycle guard alone still expands these small shared graphs exponentially.
// Keep the generated cases small in source but deep enough to expose re-walking.
func TestPreferSetSizeSharedAliasGraphs(t *testing.T) {
	const depth = 30
	t.Run("const bindings", func(t *testing.T) {
		var source strings.Builder
		source.WriteString("const flag = true;\nconst set0 = new Set();\n")
		for i := 1; i <= depth; i++ {
			fmt.Fprintf(&source, "const set%d = flag ? set%d : set%d;\n", i, i-1, i-1)
		}
		code := source.String() + fmt.Sprintf("[...set%d].length", depth)
		output := source.String() + fmt.Sprintf("set%d.size", depth)
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_set_size.PreferSetSizeRule,
			nil, []rule_tester.InvalidTestCase{invalid(code, output, "file.js")},
		)
	})
	t.Run("type aliases", func(t *testing.T) {
		var source strings.Builder
		source.WriteString("type Set0 = Set<string>;\n")
		for i := 1; i <= depth; i++ {
			fmt.Fprintf(&source, "type Set%d = Set%d | Set%d;\n", i, i-1, i-1)
		}
		prefix := source.String() + fmt.Sprintf("function getSize(set: Set%d) { return ", depth)
		code := prefix + "Array.from(set).length; }"
		diagnostics := lintPreferSetSizeProjectFalse(t, code)
		if len(diagnostics) != 1 {
			t.Fatalf("diagnostic count = %d, want 1: %+v", len(diagnostics), diagnostics)
		}
		output, unapplied, fixed := linter.ApplyRuleFixes(code, diagnostics)
		if !fixed || len(unapplied) != 0 || output != prefix+"set.size; }" {
			t.Fatalf("output = %q, fixed=%t, unapplied=%+v", output, fixed, unapplied)
		}
	})
}

func lintPreferSetSizeProjectFalse(t *testing.T, code string) []rule.RuleDiagnostic {
	t.Helper()
	dir := tspath.NormalizePath(t.TempDir())
	fileName := tspath.NormalizePath(filepath.Join(dir, "file.ts"))
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            utils.CreateCompilerHost(dir, fs),
		CompilerOptions: &core.CompilerOptions{Target: core.ScriptTargetESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("create project:false program: %v", err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("project:false fixture unexpectedly received a TypeChecker")
	}

	diagnostics := []rule.RuleDiagnostic{}
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    fileName,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     prefer_set_size.PreferSetSizeRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("project:false fixture unexpectedly received a TypeChecker")
					}
					return prefer_set_size.PreferSetSizeRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAutofix,
			Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		},
	})
	return diagnostics
}
