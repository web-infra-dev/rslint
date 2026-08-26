package linter

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func runPatternTraversalTest(t *testing.T, source string, listeners rule.RuleListeners) {
	t.Helper()

	program, paths := createTestProgramWithFiles(t, map[string]string{"input.ts": source})
	programs := wrapTestPrograms(program)
	lintPlan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         programs,
		SingleThreaded:   true,
		TargetsByProgram: [][]string{{paths["input.ts"]}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{{
				Name:     "pattern-traversal",
				Severity: rule.SeverityWarning,
				Run: func(rule.RuleContext) rule.RuleListeners {
					return listeners
				},
			}}
		},
	})
	_, err := RunLinter(RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}
}

func TestPatternTraversalVisitsComputedAssignmentKeyAsExpression(t *testing.T) {
	var events []string
	recordIdentifier := func(suffix string) func(*ast.Node) {
		return func(node *ast.Node) {
			events = append(events, "identifier:"+node.AsIdentifier().Text+suffix)
		}
	}

	runPatternTraversalTest(t, `({ [key()]: target } = source);`, rule.RuleListeners{
		ast.KindComputedPropertyName: func(*ast.Node) {
			events = append(events, "computed:enter")
		},
		rule.ListenerOnAllowPattern(ast.KindComputedPropertyName): func(*ast.Node) {
			events = append(events, "computed:pattern-enter")
		},
		rule.ListenerOnExit(rule.ListenerOnAllowPattern(ast.KindComputedPropertyName)): func(*ast.Node) {
			events = append(events, "computed:pattern-exit")
		},
		rule.ListenerOnExit(ast.KindComputedPropertyName): func(*ast.Node) {
			events = append(events, "computed:exit")
		},
		ast.KindCallExpression: func(*ast.Node) {
			events = append(events, "call:enter")
		},
		rule.ListenerOnAllowPattern(ast.KindCallExpression): func(*ast.Node) {
			events = append(events, "call:pattern-enter")
		},
		rule.ListenerOnExit(ast.KindCallExpression): func(*ast.Node) {
			events = append(events, "call:exit")
		},
		ast.KindIdentifier: recordIdentifier(":enter"),
		rule.ListenerOnAllowPattern(ast.KindIdentifier):                      recordIdentifier(":pattern-enter"),
		rule.ListenerOnExit(rule.ListenerOnAllowPattern(ast.KindIdentifier)): recordIdentifier(":pattern-exit"),
		rule.ListenerOnExit(ast.KindIdentifier):                              recordIdentifier(":exit"),
	})

	want := []string{
		"computed:enter",
		"call:enter",
		"identifier:key:enter",
		"identifier:key:exit",
		"call:exit",
		"computed:exit",
		"identifier:target:enter",
		"identifier:target:pattern-enter",
		"identifier:target:pattern-exit",
		"identifier:target:exit",
		"identifier:source:enter",
		"identifier:source:exit",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("listener events = %#v, want %#v", events, want)
	}
}

func TestPatternTraversalVisitsPlainAssignmentKeyAsExpression(t *testing.T) {
	var events []string
	recordIdentifier := func(suffix string) func(*ast.Node) {
		return func(node *ast.Node) {
			events = append(events, "identifier:"+node.AsIdentifier().Text+suffix)
		}
	}

	runPatternTraversalTest(t, `({ key: target } = source);`, rule.RuleListeners{
		ast.KindIdentifier: recordIdentifier(":enter"),
		rule.ListenerOnAllowPattern(ast.KindIdentifier):                      recordIdentifier(":pattern-enter"),
		rule.ListenerOnExit(rule.ListenerOnAllowPattern(ast.KindIdentifier)): recordIdentifier(":pattern-exit"),
		rule.ListenerOnExit(ast.KindIdentifier):                              recordIdentifier(":exit"),
	})

	want := []string{
		"identifier:key:enter",
		"identifier:key:exit",
		"identifier:target:enter",
		"identifier:target:pattern-enter",
		"identifier:target:pattern-exit",
		"identifier:target:exit",
		"identifier:source:enter",
		"identifier:source:exit",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("listener events = %#v, want %#v", events, want)
	}
}

func TestPatternTraversalVisitsLiteralAssignmentKeys(t *testing.T) {
	strings := 0
	numbers := 0

	runPatternTraversalTest(t, `({ "stringKey": first, 2: second } = source);`, rule.RuleListeners{
		ast.KindStringLiteral:  func(*ast.Node) { strings++ },
		ast.KindNumericLiteral: func(*ast.Node) { numbers++ },
	})

	if strings != 1 || numbers != 1 {
		t.Fatalf("literal key visits = (%d string, %d numeric), want (1, 1)", strings, numbers)
	}
}

func TestPatternTraversalVisitsNestedComputedAssignmentKeysOnce(t *testing.T) {
	ordinary := make(map[string]int)
	patterns := make(map[string]int)
	computed := 0
	computedPatterns := 0

	runPatternTraversalTest(t, `({
  [outerKey]: { [innerKey]: nestedTarget },
  [consume((({ [recursiveKey]: recursiveTarget } = recursiveSource)))]: target,
} = source);`, rule.RuleListeners{
		ast.KindComputedPropertyName: func(*ast.Node) {
			computed++
		},
		rule.ListenerOnAllowPattern(ast.KindComputedPropertyName): func(*ast.Node) {
			computedPatterns++
		},
		ast.KindIdentifier: func(node *ast.Node) {
			ordinary[node.AsIdentifier().Text]++
		},
		rule.ListenerOnAllowPattern(ast.KindIdentifier): func(node *ast.Node) {
			patterns[node.AsIdentifier().Text]++
		},
	})

	wantOrdinary := map[string]int{
		"outerKey":        1,
		"innerKey":        1,
		"nestedTarget":    1,
		"consume":         1,
		"recursiveKey":    1,
		"recursiveTarget": 1,
		"recursiveSource": 1,
		"target":          1,
		"source":          1,
	}
	wantPatterns := map[string]int{
		"nestedTarget":    1,
		"recursiveTarget": 1,
		"target":          1,
	}
	if computed != 4 || computedPatterns != 0 {
		t.Errorf("computed-key visits = (%d ordinary, %d pattern), want (4, 0)", computed, computedPatterns)
	}
	if !reflect.DeepEqual(ordinary, wantOrdinary) {
		t.Errorf("ordinary identifier visits = %#v, want %#v", ordinary, wantOrdinary)
	}
	if !reflect.DeepEqual(patterns, wantPatterns) {
		t.Errorf("pattern identifier visits = %#v, want %#v", patterns, wantPatterns)
	}
}

func TestPatternTraversalKeepsObjectExpressionsInsideComputedKeysOutOfPatternContext(t *testing.T) {
	allowPattern := 0
	notAllowPattern := 0

	runPatternTraversalTest(t, `({ [({ [valueKey]: value })]: target } = source);`, rule.RuleListeners{
		rule.ListenerOnAllowPattern(ast.KindObjectLiteralExpression): func(*ast.Node) {
			allowPattern++
		},
		rule.ListenerOnNotAllowPattern(ast.KindObjectLiteralExpression): func(*ast.Node) {
			notAllowPattern++
		},
	})

	if allowPattern != 1 || notAllowPattern != 1 {
		t.Fatalf("object contexts = (%d pattern, %d expression), want (1, 1)", allowPattern, notAllowPattern)
	}
}

func TestPatternTraversalComputedKeyRouteMatrix(t *testing.T) {
	keys := map[string]struct{}{
		"directKey": {},
		"arrayKey":  {},
		"forOfKey":  {},
		"forInKey":  {},
		"arrowKey":  {},
		"classKey":  {},
	}
	ordinary := make(map[string]int, len(keys))
	patterns := make(map[string]int, len(keys))
	record := func(counts map[string]int) func(*ast.Node) {
		return func(node *ast.Node) {
			name := node.AsIdentifier().Text
			if _, tracked := keys[name]; tracked {
				counts[name]++
			}
		}
	}

	runPatternTraversalTest(t, `
({ [directKey]: directTarget } = source);
([{ [arrayKey]: arrayTarget }] = source);
for ({ [forOfKey]: forOfTarget } of source) {}
for ({ [forInKey]: forInTarget } in source) {}
({ [(() => arrowKey)()]: arrowTarget } = source);
({ [class { [classKey]() {} }]: classTarget } = source);
`, rule.RuleListeners{
		ast.KindIdentifier: record(ordinary),
		rule.ListenerOnAllowPattern(ast.KindIdentifier): record(patterns),
	})

	for name := range keys {
		if ordinary[name] != 1 || patterns[name] != 0 {
			t.Errorf("%s visits = (%d ordinary, %d pattern), want (1, 0)", name, ordinary[name], patterns[name])
		}
	}
}
