package linter

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestRefStoreCollectionCompletesBeforeFileFinalize(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const value = 1; consume(value);",
	})
	var events []string
	sourceFileExitCalls := 0
	makeRule := func(name string, request bool) ConfiguredRule {
		configured := ConfiguredRule{
			Name:     name,
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				events = append(events, name+":run")
				if request && !ctx.RequestRefs(rule.RefNeedReferences) {
					t.Fatal("linter context omitted RefStore")
				}
				return rule.RuleListeners{
					ast.KindIdentifier: func(node *ast.Node) {
						if node.Text() == "value" {
							events = append(events, name+":identifier")
						}
					},
					rule.ListenerOnExit(ast.KindIdentifier): func(node *ast.Node) {
						if node.Text() == "value" {
							events = append(events, name+":identifier-exit")
						}
					},
					rule.ListenerOnExit(ast.KindSourceFile): func(*ast.Node) {
						sourceFileExitCalls++
					},
					rule.ListenerOnFileFinalize(): func(file *ast.Node) {
						events = append(events, name+":finalize")
						if !request {
							return
						}
						var valueNodes []*ast.Node
						var visit func(*ast.Node) bool
						visit = func(node *ast.Node) bool {
							if node.Kind == ast.KindIdentifier && node.Text() == "value" {
								valueNodes = append(valueNodes, node)
							}
							node.ForEachChild(visit)
							return false
						}
						file.ForEachChild(visit)
						if len(valueNodes) != 2 {
							t.Fatalf("value nodes = %d, want declaration and reference", len(valueNodes))
						}
						got := ctx.Refs.References(valueNodes[0].Parent.Symbol())
						if len(got) != 1 || got[0] != valueNodes[1] {
							t.Fatalf("finalize References = %v, want complete reference %v", got, valueNodes[1])
						}
					},
				}
			},
		}
		if request {
			configured.Needs = rule.RuleNeeds{Refs: rule.RefNeedReferences}
		}
		return configured
	}

	_, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["a.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{makeRule("a", true), makeRule("b", false)}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if sourceFileExitCalls != 0 {
		t.Fatalf("ListenerOnExit(SourceFile) calls = %d, want 0", sourceFileExitCalls)
	}
	want := []string{
		"a:run", "b:run",
		"a:identifier", "b:identifier", "a:identifier-exit", "b:identifier-exit",
		"a:identifier", "b:identifier", "a:identifier-exit", "b:identifier-exit",
		"a:finalize", "b:finalize",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestFileFinalizeRunsOnceForEveryFileIncludingEmpty(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"empty.ts": "",
		"value.ts": "const value = 1;",
	})
	finalized := make(map[string]int)

	_, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["empty.ts"], paths["value.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name: "finalize",
				Run: func(rule.RuleContext) rule.RuleListeners {
					return rule.RuleListeners{
						rule.ListenerOnFileFinalize(): func(file *ast.Node) {
							finalized[file.AsSourceFile().FileName()]++
						},
					}
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	for _, name := range []string{"empty.ts", "value.ts"} {
		if got := finalized[paths[name]]; got != 1 {
			t.Errorf("%s finalize calls = %d, want 1", name, got)
		}
	}
}

func TestRefStoreCollectionCoversPatternTraversalExactlyOnce(t *testing.T) {
	const source = `
let outerKey, innerKey, nestedTarget, consume, recursiveKey;
let recursiveTarget, recursiveSource, target, sourceValue, shared;
({
  [outerKey]: { [innerKey]: nestedTarget },
  [consume((({ [recursiveKey]: recursiveTarget } = recursiveSource)))]: target,
} = sourceValue);
({ [shared]: shared } = shared);
`
	program, paths := createTestProgramWithFiles(t, map[string]string{"pattern.ts": source})
	wantNames := []string{
		"outerKey", "innerKey", "nestedTarget", "consume", "recursiveKey",
		"recursiveTarget", "recursiveSource", "target", "sourceValue",
		"shared",
	}

	_, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles:    [][]string{{paths["pattern.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:  "pattern-ref-facts",
				Needs: rule.RuleNeeds{Refs: rule.RefNeedReferences | rule.RefNeedBindingDeclarations},
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					ctx.RequestRefs(rule.RefNeedReferences | rule.RefNeedBindingDeclarations)
					return rule.RuleListeners{
						rule.ListenerOnFileFinalize(): func(*ast.Node) {
							declarations := make(map[string]*ast.Symbol)
							for _, declaration := range ctx.Refs.BindingDeclarations() {
								if declaration.Symbol != nil {
									declarations[declaration.Name.Text()] = declaration.Symbol
								}
							}
							for _, name := range wantNames {
								symbol := declarations[name]
								if symbol == nil {
									t.Errorf("missing streamed declaration for %q", name)
									continue
								}
								references := ctx.Refs.References(symbol)
								wantCount := 1
								if name == "shared" {
									wantCount = 3
								}
								if len(references) != wantCount {
									t.Errorf("%s references = %v, want exactly %d", name, references, wantCount)
									continue
								}
								for index, reference := range references {
									if reference.Text() != name || index > 0 && references[index-1].Pos() >= reference.Pos() {
										t.Errorf("%s references are not unique source order: %v", name, references)
										break
									}
								}
							}
						},
					}
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
}
