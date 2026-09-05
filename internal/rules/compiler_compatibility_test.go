package rules_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Exercise every catalog rule, including deferred edits, on compiler recovery
// trees and recently introduced syntax. Rule-specific tests assert diagnostic
// and fix semantics; this test guards against panics and silently skipped input.
func TestAllRulesCompilerCompatibility(t *testing.T) {
	const code = `
export {};
interface Base<T> { value: T }
interface Normal extends Base<Array<string>> {}
class Implemented implements Base<string> { value = ''; }
interface Called extends Base() { (): void; }
interface Parenthesized extends (Base) { (): void; }
interface Indexed extends Base['Type'] { (): void; }
interface Props extends Base() { readonly value: string }
function Component(props: Props) { return <div>{props.value}</div>; }
class Private { #value = 1; method(value: typeof this.#value) { return value; } }
function unreachableFor() { for ((() => { throw 1; })();;) { console.log('unreachable'); } }
function unreachableForOf() { for (const value of (() => { throw 1; })()) { console.log(value); } }
`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "compiler-compatibility.tsx")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	compilerProgram, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	file := compilerProgram.GetSourceFile(fileName)
	if file == nil {
		t.Fatal("compiler compatibility fixture was not included")
	}
	if len(file.Diagnostics()) != 0 {
		t.Fatal("fixture has parse diagnostics and would bypass rule execution")
	}
	heritageKinds := map[ast.Kind]int{}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Parent != nil && node.Parent.Kind == ast.KindHeritageClause {
			heritageKinds[node.Kind]++
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if heritageKinds[ast.KindTypeReference] != 2 || heritageKinds[ast.KindExpressionWithTypeArguments] != 4 {
		t.Fatalf("fixture must exercise normal and recovery heritage nodes: %v", heritageKinds)
	}

	catalog := rules.All().AllRules()
	if len(catalog) == 0 {
		t.Fatal("rule catalog must not be empty")
	}
	configured := make([]rule.ConfiguredRule, 0, len(catalog))
	initialized := make(map[string]int, len(catalog))
	for name, impl := range catalog {
		options := rule_tester.ResolveTestCaseOptions(t, &impl, nil)
		configured = append(configured, rule.ConfiguredRule{
			Name:        name,
			Environment: &rule.RuleEnvironment{},
			Severity:    rule.SeverityError,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				initialized[name]++
				return impl.Run(ctx, options)
			},
		})
	}
	plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*program.Program{program.NewFromCompiler(compilerProgram)},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile:  func(*ast.SourceFile) []rule.ConfiguredRule { return configured },
	})
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := 0
	result, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       plan,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: func(rule.RuleDiagnostic) { diagnostics++ },
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.LintedFileCount != 1 || len(result.ExecutedRules) != len(catalog) || diagnostics == 0 {
		t.Fatalf("incomplete rule execution: files=%d, rules=%d/%d, diagnostics=%d", result.LintedFileCount, len(result.ExecutedRules), len(catalog), diagnostics)
	}
	for name := range catalog {
		if initialized[name] != 1 {
			t.Errorf("rule %q initialized %d times, want once", name, initialized[name])
		}
	}
}
