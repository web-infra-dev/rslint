package utils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func TestCollectRstestCallCandidateNames(t *testing.T) {
	sourceFile := parser.ParseSourceFile(
		ast.SourceFileParseOptions{
			FileName: "/candidate.test.ts",
			Path:     "/candidate.test.ts",
		},
		`import { test as importedTest, expect as importedExpect } from "@rstest/core";
const testAlias = importedTest;
const testCase = testAlias;
const expectAlias = importedExpect;
const check = expectAlias;`,
		core.ScriptKindTS,
	)
	candidates := collectRstestCallCandidateNames(sourceFile)
	for _, name := range []string{"importedTest", "testAlias", "testCase"} {
		if !candidates.test[name] {
			t.Errorf("test candidate %q was not propagated", name)
		}
	}
	for _, name := range []string{"importedExpect", "expectAlias", "check"} {
		if !candidates.expect[name] {
			t.Errorf("expect candidate %q was not propagated", name)
		}
	}
}
