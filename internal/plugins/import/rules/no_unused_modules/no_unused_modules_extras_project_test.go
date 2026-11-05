package no_unused_modules

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// projectRun lints targets with one Program per tsconfig, the way the CLI and
// API bind a multi-project workspace, and returns "file:line:column message".
type projectRun struct {
	files           map[string]string
	tsconfigs       []string
	targets         [][]string
	options         map[string]any
	settings        map[string]any
	projectAnalysis bool
	singleThreaded  bool
}

func (run projectRun) diagnostics(t *testing.T) []string {
	t.Helper()
	root := fixtureRoot(t, run.files)
	options := rule_tester.ResolveTestCaseOptions(t, &NoUnusedModulesRule, run.options)
	programs := make([]*lintprogram.Program, len(run.tsconfigs))
	for index, tsconfig := range run.tsconfigs {
		host := utils.CreateCompilerHost(root.Dir, root.FS)
		compiled, err := utils.CreateProgram(true, root.FS, root.Dir, tsconfig, host)
		if err != nil {
			t.Fatalf("create Program for %s: %v", tsconfig, err)
		}
		programs[index] = lintprogram.NewFromCompiler(compiled)
	}
	plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: run.targets,
		SingleThreaded:   run.singleThreaded,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:        NoUnusedModulesRule.Name,
				Severity:    rule.SeverityError,
				Environment: &rule.RuleEnvironment{Settings: run.settings},
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUnusedModulesRule.Run(ctx, options)
				},
			}}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var results []string
	_, err = linter.RunLinter(linter.RunLinterOptions{
		LintPlan:        plan,
		SingleThreaded:  run.singleThreaded,
		ProjectAnalysis: run.projectAnalysis,
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.Pos())
			mu.Lock()
			defer mu.Unlock()
			results = append(results, fmt.Sprintf("%s:%d:%d %s", diagnostic.FilePath, line+1, column+1, diagnostic.Message.Description))
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(results)
	return results
}

func assertDiagnostics(t *testing.T, got []string, want ...string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("diagnostics:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// Each importer resolves with the tsconfig of the Program that owns it; no
// merged Program or type checker is needed. The importer's alias exists only
// in two/tsconfig.json.
func TestNoUnusedModulesAcrossTSConfigs(t *testing.T) {
	files := map[string]string{
		"/multi/one/tsconfig.json": `{"compilerOptions": {"module": "esnext", "moduleResolution": "bundler"}, "include": ["**/*.ts"]}`,
		"/multi/two/tsconfig.json": `{"compilerOptions": {"module": "esnext", "moduleResolution": "bundler", "baseUrl": ".", "paths": {"@one/*": ["../one/*"]}}, "include": ["**/*.ts"]}`,
		"/multi/one/a.ts":          `export const a = 1; export const b = 2;`,
		"/multi/two/use.ts":        `import { a } from '@one/a';`,
	}
	options := map[string]any{"unusedExports": true, "src": []any{"/multi"}}
	assertDiagnostics(t, projectRun{
		files: files, tsconfigs: []string{"multi/one/tsconfig.json", "multi/two/tsconfig.json"},
		targets: [][]string{{"/multi/one/a.ts"}, nil}, options: options, projectAnalysis: true, singleThreaded: true,
	}.diagnostics(t), "/multi/one/a.ts:1:21 exported declaration 'b' not used within other modules")

	// Without the importer's own Program, its alias cannot be resolved; the
	// linted file's tsconfig is not borrowed for it.
	assertDiagnostics(t, projectRun{
		files: files, tsconfigs: []string{"multi/one/tsconfig.json"},
		targets: [][]string{{"/multi/one/a.ts"}}, options: options, projectAnalysis: true, singleThreaded: true,
	}.diagnostics(t),
		"/multi/one/a.ts:1:1 exported declaration 'a' not used within other modules",
		"/multi/one/a.ts:1:21 exported declaration 'b' not used within other modules")
}

// Editor passes have no frozen project view: unusedExports reports nothing
// rather than guessing from one Program, while missingExports still works.
func TestNoUnusedModulesWithoutProjectAnalysis(t *testing.T) {
	files := map[string]string{
		"/editor/tsconfig.json": `{"include": ["**/*.ts"]}`,
		"/editor/a.ts":          `export const a = 1;`,
		"/editor/b.ts":          `const b = 1;`,
	}
	run := projectRun{
		files: files, tsconfigs: []string{"editor/tsconfig.json"}, targets: [][]string{{"/editor/a.ts", "/editor/b.ts"}},
		options: map[string]any{"unusedExports": true, "missingExports": true, "src": []any{"/editor"}}, singleThreaded: true,
	}
	assertDiagnostics(t, run.diagnostics(t), "/editor/b.ts:1:1 No exports found")
	run.projectAnalysis = true
	assertDiagnostics(t, run.diagnostics(t),
		"/editor/a.ts:1:1 exported declaration 'a' not used within other modules",
		"/editor/b.ts:1:1 No exports found")
}

// Every target and every worker shares one index, so the result does not
// depend on target order or parallelism.
func TestNoUnusedModulesTargetOrderAndParallelism(t *testing.T) {
	files := map[string]string{"/order/tsconfig.json": `{"compilerOptions": {"module": "esnext", "moduleResolution": "bundler"}, "include": ["**/*.ts"]}`}
	var targets []string
	var want []string
	for index := range 40 {
		name := fmt.Sprintf("/order/m%02d.ts", index)
		next := fmt.Sprintf("./m%02d", (index+1)%40)
		files[name] = fmt.Sprintf("import { used } from '%s';\nexport const used = 1;\nexport const unused%d = 2;", next, index)
		targets = append(targets, name)
		want = append(want, fmt.Sprintf("%s:3:1 exported declaration 'unused%d' not used within other modules", name, index))
	}
	slices.Sort(want)
	options := map[string]any{"unusedExports": true, "src": []any{"/order"}}
	reversed := slices.Clone(targets)
	slices.Reverse(reversed)
	for _, run := range []projectRun{
		{files: files, tsconfigs: []string{"order/tsconfig.json"}, targets: [][]string{targets}, options: options, projectAnalysis: true, singleThreaded: true},
		{files: files, tsconfigs: []string{"order/tsconfig.json"}, targets: [][]string{reversed}, options: options, projectAnalysis: true},
	} {
		assertDiagnostics(t, run.diagnostics(t), want...)
	}
}

func TestNoUnusedModulesSchema(t *testing.T) {
	for _, valid := range []any{
		nil,
		map[string]any{"missingExports": true},
		map[string]any{"unusedExports": true},
		map[string]any{"unusedExports": true, "src": []any{"src"}, "ignoreExports": []any{"src/index.ts"}, "ignoreUnusedTypeExports": true},
		map[string]any{"missingExports": true, "unusedExports": false},
		// src minItems applies only to the unusedExports branch.
		map[string]any{"missingExports": true, "src": []any{}},
	} {
		if err := NoUnusedModulesRule.Schema.Validate(rule.NormalizeOptions(valid)); err != nil {
			t.Errorf("options %#v rejected: %v", valid, err)
		}
	}
	for _, invalid := range []any{
		map[string]any{},
		map[string]any{"missingExports": false, "unusedExports": false},
		map[string]any{"unusedExports": true, "src": []any{}},
		map[string]any{"unusedExports": true, "src": []any{""}},
		map[string]any{"unusedExports": true, "src": []any{"a", "a"}},
		map[string]any{"missingExports": "yes"},
		[]any{map[string]any{"missingExports": true}, map[string]any{"missingExports": true}},
	} {
		if err := NoUnusedModulesRule.Schema.Validate(rule.NormalizeOptions(invalid)); err == nil {
			t.Errorf("options %#v accepted", invalid)
		}
	}
}
