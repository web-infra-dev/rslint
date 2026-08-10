package linter

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type linterFileCacheTestKey struct{}

func TestRunLinterCachesOncePerFileAcrossRules(t *testing.T) {
	program, paths := createTestProgramWithFiles(t, map[string]string{
		"first.test.ts":  `test("first", () => {});`,
		"second.test.ts": `test("second", () => {});`,
	})
	values := make(map[string][]*int)
	builds := make(map[string]int)

	makeRule := func(name string) ConfiguredRule {
		return ConfiguredRule{
			Name:     name,
			Severity: rule.SeverityWarning,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				fileName := ctx.SourceFile.FileName()
				value := rule.CachedByFile(ctx, linterFileCacheTestKey{}, func() *int {
					builds[fileName]++
					return new(int)
				})
				values[fileName] = append(values[fileName], value)
				return rule.RuleListeners{}
			},
		}
	}

	_, err := RunLinter(RunLinterOptions{
		Programs:       []*compiler.Program{program},
		SingleThreaded: true,
		TargetFiles: [][]string{{
			paths["first.test.ts"],
			paths["second.test.ts"],
		}},
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule {
			return []ConfiguredRule{makeRule("first-rule"), makeRule("second-rule")}
		},
		Consumer: rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) {}},
	})
	if err != nil {
		t.Fatalf("RunLinter error: %v", err)
	}

	first := values[paths["first.test.ts"]]
	second := values[paths["second.test.ts"]]
	if builds[paths["first.test.ts"]] != 1 || len(first) != 2 || first[0] != first[1] {
		t.Fatalf("first file built %d times with values %#v", builds[paths["first.test.ts"]], first)
	}
	if builds[paths["second.test.ts"]] != 1 || len(second) != 2 || second[0] != second[1] {
		t.Fatalf("second file built %d times with values %#v", builds[paths["second.test.ts"]], second)
	}
	if first[0] == second[0] {
		t.Fatal("different files shared a cached value")
	}
}
