package no_undef

import (
	"fmt"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestNoUndefRuleNilTypeChecker pins the rule's behavior on files linted
// without a TypeChecker (gap files outside every tsconfig): the binder scope
// walk, ctx.Globals, and the ECMAScript built-in table are then the only
// sources of defined names, matching ESLint's default of recognizing
// ECMAScript built-ins but not host globals unless configured.
func TestNoUndefRuleNilTypeChecker(t *testing.T) {
	t.Parallel()
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "no-undef-nil-tc.ts")
	code := `var declaredLocal = 1; declaredLocal;
console.log(1);
Promise.resolve();
var buf = new Float16Array(8);
myConfiguredGlobal;
myOffGlobal;
var myOffLocal = 1; myOffLocal;
undeclaredName123;
`
	fs := utils.NewOverlayVFSForFile(filePath, code)
	program, err := utils.CreateProgram(
		true, fs, rootDir, "tsconfig.json", utils.CreateCompilerHost(rootDir, fs),
	)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatalf("source file not found for %s", filePath)
		return
	}

	var reported []string
	ctx := (rule.RuleContext{
		SourceFile:  sourceFile,
		Program:     program,
		TypeChecker: nil, // explicitly nil — this is the path under test
		Refs:        rule.NewRefStore(sourceFile, program.Options(), nil),
		Globals: map[string]bool{
			"myConfiguredGlobal": true,
			"myOffGlobal":        false,
			"myOffLocal":         false,
		},
	}).WithReporter("test/no-undef", rule.SeverityError, func(d rule.RuleDiagnostic) {
		reported = append(reported, d.Message.Description)
	})

	listeners := NoUndefRule.Run(ctx, nil)

	var walk func(n *ast.Node) bool
	walk = func(n *ast.Node) bool {
		if cb, ok := listeners[n.Kind]; ok {
			cb(n)
		}
		n.ForEachChild(walk)
		return false
	}
	walk(sourceFile.AsNode())

	want := []string{
		"'console' is not defined.",
		"'myOffGlobal' is not defined.",
		"'undeclaredName123' is not defined.",
	}
	if fmt.Sprint(reported) != fmt.Sprint(want) {
		t.Fatalf("reported = %v, want %v", reported, want)
	}
}
