package rules_test

import (
	"context"
	"reflect"
	"slices"
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
		return
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
	// Rules with required options need an explicit fixture configuration.
	fixtureOptions := map[string]any{
		"import/enforce-node-protocol-usage": []any{"always"},
	}
	configured := make([]rule.ConfiguredRule, 0, len(catalog))
	initialized := make(map[string]int, len(catalog))
	for name, impl := range catalog {
		options := rule_tester.ResolveTestCaseOptions(t, &impl, fixtureOptions[name])
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

// Every opt-in is exercised on the same source universe and isolated roots.
// Adding an opt-in automatically extends this compatibility check; other rules
// keep the complete-Program default.
func TestFileIsolatedRulesMatchCompleteSourceUniverse(t *testing.T) {
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "isolated-source-compatibility")
	contents := map[string]string{
		"a.ts":       "import value, * as other from './b'; export const a = value; const unused: any = other; if (true) { debugger; }\n",
		"b.ts":       "import { a } from './a'; export const b = a; export default b; function unused() { return; }\n",
		"global.js":  "/* global custom */\r\nvar duplicate = 1; var duplicate = 2; if (custom == NaN) { debugger; }\n",
		"view.tsx":   "export const View = () => <div>{[1,,2].map(x => x)}</div>;\n",
		"unicode.js": "const face = '😀';\r\nlet unused = face;\u2028debugger;\u2029",
		"bad.ts":     "export const broken: = 1;",
	}
	paths := make([]string, 0, len(contents))
	files := make(map[string]string, len(contents))
	for name, text := range contents {
		path := tspath.ResolvePath(root.Dir, name)
		paths = append(paths, path)
		files[path] = text
	}
	slices.Sort(paths)
	fs := utils.NewOverlayVFS(root.FS, files)
	build := func(names []string) (*program.Program, error) {
		return program.NewFromRoots(program.RootOptions{
			RootFileNames: names, Host: utils.CreateCompilerHost(root.Dir, fs),
			CompilerOptions: program.SourceOnlyCompilerOptions(), SingleThreaded: true,
		})
	}
	complete, err := build(paths)
	if err != nil {
		t.Fatal(err)
	}
	selected := 0
	for name, impl := range rules.All().AllRules() {
		if !impl.SupportsFileIsolation {
			continue
		}
		selected++
		t.Run(name, func(t *testing.T) {
			if impl.RequiresTypeInfo || impl.IsEslintPluginRule {
				t.Fatal("source-only scope was granted to a non-native or checker-only rule")
			}
			options := rule_tester.ResolveTestCaseOptions(t, &impl, nil)
			configured := []rule.ConfiguredRule{{Name: name, SupportsFileIsolation: true, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners { return impl.Run(ctx, options) }}}
			run := func(isolated, singleThreaded bool) linter.NativeObservation {
				t.Helper()
				generation := linter.Generation{Native: linter.NativeGeneration{
					SingleThreaded: singleThreaded, Cwd: root.Dir,
					RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule { return configured },
				}}
				if isolated {
					generation.Native.DeferredRoots = &program.DeferredRoots{FileNames: paths, Build: func(_ context.Context, path string) (*program.Program, error) { return build([]string{path}) }}
				} else {
					generation.Native.Programs = []*program.Program{complete}
					generation.Native.TargetsByProgram = [][]string{paths}
				}
				result, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(linter.GenerationProviderFunc(func(context.Context, linter.SourceSnapshot) (linter.Generation, linter.ReleaseFunc, error) {
					return generation, nil, nil
				}), linter.ObservationPolicy{}, nil))
				if err != nil {
					t.Fatal(err)
				}
				observation := result.Observation.Native
				linter.StableSortDiagnosticsByFileAndStart(observation.Diagnostics)
				// Materialize line maps before comparing their lazy projections.
				for _, diagnostic := range observation.Diagnostics {
					diagnostic.SourceFile.ECMALineMap()
				}
				return observation
			}
			want := run(false, true)
			for _, singleThreaded := range []bool{true, false} {
				got := run(true, singleThreaded)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("isolated diagnostics or accounting changed (singleThreaded=%v):\ncomplete=%+v\nisolated=%+v", singleThreaded, want, got)
				}
			}
		})
	}
	if selected == 0 {
		t.Fatal("no file-isolated rules exercised")
	}
}
