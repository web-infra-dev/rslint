package display_name

// These regressions cover effective pragma resolution against eslint-plugin-react
// v7.37.5. The existing rule suite lives in display_name_test.go.
import (
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestDisplayNameExtrasPragma(t *testing.T) {
	for _, test := range []struct {
		name     string
		code     string
		settings map[string]interface{}
		errors   []rule_tester.InvalidTestCaseError
	}{
		{
			name:     "block",
			code:     "/** @jsx Foo */\nvar Hello = Foo.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0"}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 29, EndLine: 2, EndColumn: 61}},
		},
		{
			name:     "line",
			code:     "// @jsx Foo\nvar Hello = Foo.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0"}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 29, EndLine: 2, EndColumn: 61}},
		},
		{
			name:     "source overrides settings",
			code:     "/** @jsx Foo */\nvar Hello = Foo.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0", "pragma": "Bar"}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 29, EndLine: 2, EndColumn: 61}},
		},
		{
			name:     "settings fallback",
			code:     "// no annotation\nvar Hello = Foo.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0", "pragma": "Foo"}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 29, EndLine: 2, EndColumn: 61}},
		},
		{
			name:     "default fallback",
			code:     "// no annotation\nvar Hello = React.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0"}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 31, EndLine: 2, EndColumn: 63}},
		},
		{
			name:     "dotted source",
			code:     "/** @jsx Foo.h */\nvar Hello = Foo.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0", "pragma": "Bar"}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 29, EndLine: 2, EndColumn: 61}},
		},
		{
			name:     "invalid source",
			code:     "/** @jsx 123 */\nvar Hello = React.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0", "pragma": "Bar"}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 31, EndLine: 2, EndColumn: 63}},
		},
		{
			name:     "first annotation",
			code:     "/** @jsx Foo */ /* @jsx Bar */\nvar Hello = Foo.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0", "pragma": "Bar"}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 29, EndLine: 2, EndColumn: 61}},
		},
		{
			name:     "overridden settings ignored",
			code:     "/** @jsx Foo */\nvar Hello = Bar.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0", "pragma": "Bar"}},
			errors:   []rule_tester.InvalidTestCaseError{},
		},
		{
			name:     "overridden default ignored",
			code:     "/** @jsx Foo */\nvar Hello = React.createClass({ render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0"}},
			errors:   []rule_tester.InvalidTestCaseError{},
		},
		{
			name:     "explicit display name",
			code:     "/** @jsx Foo */\nvar Hello = Foo.createClass({ displayName: 'Hello', render() { return <div />; } });",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0"}},
			errors:   []rule_tester.InvalidTestCaseError{},
		},
		{
			name:     "wrapper substitution",
			code:     "/** @jsx Foo */\nconst Hello = Foo.wrap(() => <div />);",
			settings: map[string]interface{}{"react": map[string]interface{}{"createClass": "createClass", "version": "16.14.0", "pragma": "Bar"}, "componentWrapperFunctions": []any{map[string]interface{}{"object": "<pragma>", "property": "wrap"}}},
			errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noDisplayName", Message: "Component definition is missing display name", Line: 2, Column: 15, EndLine: 2, EndColumn: 38}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			options := map[string]interface{}{"ignoreTranspilerName": true}
			t.Run("checker", func(t *testing.T) {
				var valid []rule_tester.ValidTestCase
				var invalid []rule_tester.InvalidTestCase
				if len(test.errors) == 0 {
					valid = append(valid, rule_tester.ValidTestCase{Code: test.code, Tsx: true, Settings: test.settings, Options: options})
				} else {
					invalid = append(invalid, rule_tester.InvalidTestCase{Code: test.code, Tsx: true, Settings: test.settings, Options: options, Errors: test.errors})
				}
				rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DisplayNameRule, valid, invalid)
			})
			t.Run("source", func(t *testing.T) {
				root := fixtures.GetRootDir()
				fileName := filepath.Join(root.Dir, "pragma.tsx")
				fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: test.code})
				program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
					RootFileNames: []string{fileName}, Host: utils.CreateCompilerHost(root.Dir, fs),
					CompilerOptions: &core.CompilerOptions{Target: core.ScriptTargetESNext}, SingleThreaded: true,
				})
				if err != nil {
					t.Fatal(err)
				}
				var diagnostics []rule.RuleDiagnostic
				plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
					Programs: []*lintprogram.Program{program}, TargetsByProgram: [][]string{{fileName}}, SingleThreaded: true,
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: DisplayNameRule.Name, Severity: rule.SeverityError,
							Environment: &rule.RuleEnvironment{Settings: test.settings},
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								if ctx.TypeChecker != nil {
									t.Fatal("source program unexpectedly has a checker")
								}
								return DisplayNameRule.Run(ctx, []any{options})
							}}}
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				_, err = linter.RunLinter(linter.RunLinterOptions{SingleThreaded: true, LintPlan: plan,
					Consumer: rule.DiagnosticConsumer{Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				if err != nil {
					t.Fatal(err)
				}
				if len(diagnostics) != len(test.errors) {
					t.Fatalf("got %d diagnostics, want %d: %#v", len(diagnostics), len(test.errors), diagnostics)
				}
				for i, diagnostic := range diagnostics {
					want := test.errors[i]
					if diagnostic.Message.Id != want.MessageId || diagnostic.Message.Description != want.Message {
						t.Fatalf("unexpected message: %#v", diagnostic.Message)
					}
					// All oracle cases report on the second line; use raw offsets to also
					// validate the source-only Program's complete diagnostic range.
					lineStart := 0
					for j, ch := range test.code {
						if ch == '\n' {
							lineStart = j + 1
							break
						}
					}
					if diagnostic.Range.Pos() != lineStart+want.Column-1 || diagnostic.Range.End() != lineStart+want.EndColumn-1 {
						t.Fatalf("range = %v, want columns %d-%d on line 2", diagnostic.Range, want.Column, want.EndColumn)
					}
				}
			})
		})
	}
}
